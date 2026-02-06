package v1beta

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// HTTP HANDLERS - Endpoints for EvalServer
// =============================================================================

// handleHealth returns server health status
func (s *EvalServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"agents":    s.agentNames(),
		"workflows": s.workflowNames(),
	})
}

// handleList returns the list of available agents and workflows
func (s *EvalServer) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"agents":    s.agentNames(),
		"workflows": s.workflowNames(),
	})
}

// handleInvoke invokes the default agent or first registered agent
func (s *EvalServer) handleInvoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Find default agent (first registered, or one named "main" or "default")
	var agent Agent
	var agentName string

	if a, ok := s.agents["main"]; ok {
		agent = a
		agentName = "main"
	} else if a, ok := s.agents["default"]; ok {
		agent = a
		agentName = "default"
	} else if len(s.agents) == 1 {
		for name, a := range s.agents {
			agent = a
			agentName = name
			break
		}
	}

	if agent == nil {
		// Try workflows
		var workflow Workflow
		if wf, ok := s.workflows["main"]; ok {
			workflow = wf
			agentName = "main"
		} else if wf, ok := s.workflows["default"]; ok {
			workflow = wf
			agentName = "default"
		} else if len(s.workflows) == 1 {
			for name, wf := range s.workflows {
				workflow = wf
				agentName = name
				break
			}
		}

		if workflow != nil {
			s.invokeWorkflow(w, r, agentName, workflow)
			return
		}

		http.Error(w, "No default agent or workflow found", http.StatusNotFound)
		return
	}

	s.invokeAgent(w, r, agentName, agent)
}

// handleInvokeNamed invokes a specific named agent or workflow
func (s *EvalServer) handleInvokeNamed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract name from path: /invoke/{name}
	name := strings.TrimPrefix(r.URL.Path, "/invoke/")
	if name == "" {
		http.Error(w, "Name required", http.StatusBadRequest)
		return
	}

	// Try agent first
	if agent, ok := s.agents[name]; ok {
		s.invokeAgent(w, r, name, agent)
		return
	}

	// Try workflow
	if workflow, ok := s.workflows[name]; ok {
		s.invokeWorkflow(w, r, name, workflow)
		return
	}

	http.Error(w, fmt.Sprintf("Agent or workflow '%s' not found", name), http.StatusNotFound)
}

// invokeAgent executes an agent and returns the result
func (s *EvalServer) invokeAgent(w http.ResponseWriter, r *http.Request, name string, agent Agent) {
	// Parse request
	var req InvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Get or create session
	session := s.getOrCreateSession(req.SessionID)
	session.AddMessage("user", req.Input)

	// Generate trace ID
	traceID := fmt.Sprintf("run-%s-%s", time.Now().Format("20060102-150405"), uuid.NewString()[:8])

	// Create context with timeout
	ctx := r.Context()
	if timeout, ok := req.Options["timeout"].(float64); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	// Execute agent with streaming to avoid timeout issues
	startTime := time.Now()
	stream, err := agent.RunStream(ctx, req.Input)
	if err != nil {
		resp := InvokeResponse{
			TraceID:    traceID,
			SessionID:  session.ID,
			DurationMs: 0,
			Success:    false,
			Error:      err.Error(),
		}
		s.logger.Printf("Agent %s stream creation error: %v", name, err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Collect all streamed chunks
	var contentBuilder strings.Builder
	for chunk := range stream.Chunks() {
		if chunk.Type == "text" && chunk.Content != "" {
			contentBuilder.WriteString(chunk.Content)
		}
	}

	// Wait for final result
	finalResult, streamErr := stream.Wait()
	duration := time.Since(startTime)

	// Build response
	resp := InvokeResponse{
		TraceID:    traceID,
		SessionID:  session.ID,
		DurationMs: duration.Milliseconds(),
		Success:    streamErr == nil,
	}

	if streamErr != nil {
		resp.Error = streamErr.Error()
		s.logger.Printf("Agent %s error: %v", name, streamErr)
	} else {
		// Use accumulated content from all chunks
		finalContent := contentBuilder.String()
		if finalContent == "" && finalResult != nil {
			finalContent = finalResult.Content
		}
		if finalResult != nil {
			resp.Output = finalContent
			resp.Success = finalResult.Success
			if finalResult.ToolsCalled != nil {
				resp.ToolsCalled = finalResult.ToolsCalled
			}
		} else {
			resp.Output = finalContent
		}
		session.AddMessage("assistant", finalContent)
	}

	// Store trace
	trace := &EvalTrace{
		ID:        traceID,
		StartTime: startTime,
		EndTime:   time.Now(),
		Duration:  duration.Milliseconds(),
	}
	s.storeTrace(trace)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Printf("Failed to encode response: %v", err)
	}
}

// invokeWorkflow executes a workflow and returns the result
func (s *EvalServer) invokeWorkflow(w http.ResponseWriter, r *http.Request, name string, workflow Workflow) {
	// Parse request
	var req InvokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Get or create session
	session := s.getOrCreateSession(req.SessionID)
	session.AddMessage("user", req.Input)

	// Generate trace ID
	traceID := fmt.Sprintf("run-%s-%s", time.Now().Format("20060102-150405"), uuid.NewString()[:8])

	// Create context with timeout
	ctx := r.Context()
	if timeout, ok := req.Options["timeout"].(float64); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	// Execute workflow with streaming to avoid timeout issues
	startTime := time.Now()
	stream, err := workflow.RunStream(ctx, req.Input)
	if err != nil {
		resp := InvokeResponse{
			TraceID:    traceID,
			SessionID:  session.ID,
			DurationMs: 0,
			Success:    false,
			Error:      err.Error(),
		}
		s.logger.Printf("Workflow %s stream creation error: %v", name, err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Collect all streamed chunks
	var contentBuilder strings.Builder
	for chunk := range stream.Chunks() {
		if chunk.Type == "text" && chunk.Content != "" {
			contentBuilder.WriteString(chunk.Content)
		}
	}

	// Wait for final result (note: Stream.Wait() returns *Result, but we need WorkflowResult)
	// For workflows, we'll use the accumulated content from chunks
	_, streamErr := stream.Wait()
	duration := time.Since(startTime)

	// Build response
	resp := InvokeResponse{
		TraceID:    traceID,
		SessionID:  session.ID,
		DurationMs: duration.Milliseconds(),
		Success:    streamErr == nil,
	}

	if streamErr != nil {
		resp.Error = streamErr.Error()
		s.logger.Printf("Workflow %s error: %v", name, streamErr)
	} else {
		finalContent := contentBuilder.String()
		resp.Output = finalContent
		resp.Success = true
		session.AddMessage("assistant", finalContent)
	}

	// Store trace
	trace := &EvalTrace{
		ID:        traceID,
		StartTime: startTime,
		EndTime:   time.Now(),
		Duration:  duration.Milliseconds(),
	}
	s.storeTrace(trace)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Printf("Failed to encode response: %v", err)
	}
}

// handleTraces retrieves trace data by ID
func (s *EvalServer) handleTraces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract trace ID from path: /traces/{id}
	traceID := strings.TrimPrefix(r.URL.Path, "/traces/")
	if traceID == "" {
		http.Error(w, "Trace ID required", http.StatusBadRequest)
		return
	}

	trace, ok := s.getTrace(traceID)
	if !ok {
		http.Error(w, "Trace not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trace)
}

// handleSessionCreate creates a new session
func (s *EvalServer) handleSessionCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := s.getOrCreateSession("")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": session.ID,
		"created_at": session.CreatedAt,
	})
}

// handleSession manages session operations
func (s *EvalServer) handleSession(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from path: /session/{id}
	sessionID := strings.TrimPrefix(r.URL.Path, "/session/")
	if sessionID == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Get session history
		s.sessionsMu.RLock()
		session, ok := s.sessions[sessionID]
		s.sessionsMu.RUnlock()

		if !ok {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)

	case http.MethodDelete:
		// Delete session
		s.sessionsMu.Lock()
		delete(s.sessions, sessionID)
		s.sessionsMu.Unlock()

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
