package v1beta

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// EVAL SERVER - HTTP Server for Testing Agents and Workflows
// =============================================================================

// EvalServer is an HTTP server that exposes agents and workflows for evaluation
type EvalServer struct {
	port     int
	traceDir string
	logger   *log.Logger

	// Registered agents and workflows
	agents    map[string]Agent
	workflows map[string]Workflow

	// Trace storage
	traces   map[string]*EvalTrace
	tracesMu sync.RWMutex

	// Session storage
	sessions   map[string]*EvalSession
	sessionsMu sync.RWMutex

	// HTTP server
	server *http.Server
	mux    *http.ServeMux
}

// EvalServerOption is a functional option for configuring EvalServer
type EvalServerOption func(*EvalServer)

// WithEvalAgent registers an agent with the eval server
func WithEvalAgent(name string, agent Agent) EvalServerOption {
	return func(s *EvalServer) {
		s.agents[name] = agent
	}
}

// WithEvalWorkflow registers a workflow with the eval server
func WithEvalWorkflow(name string, workflow Workflow) EvalServerOption {
	return func(s *EvalServer) {
		s.workflows[name] = workflow
	}
}

// WithEvalPort sets the port for the eval server
func WithEvalPort(port int) EvalServerOption {
	return func(s *EvalServer) {
		s.port = port
	}
}

// WithTraceDir sets the directory for persisting traces
func WithTraceDir(dir string) EvalServerOption {
	return func(s *EvalServer) {
		s.traceDir = dir
	}
}

// WithEvalLogger sets a custom logger for the eval server
func WithEvalLogger(logger *log.Logger) EvalServerOption {
	return func(s *EvalServer) {
		s.logger = logger
	}
}

// NewEvalServer creates a new EvalServer with the given options
func NewEvalServer(opts ...EvalServerOption) *EvalServer {
	s := &EvalServer{
		port:      8787,
		agents:    make(map[string]Agent),
		workflows: make(map[string]Workflow),
		traces:    make(map[string]*EvalTrace),
		sessions:  make(map[string]*EvalSession),
		logger:    log.New(os.Stdout, "[EvalServer] ", log.LstdFlags),
		mux:       http.NewServeMux(),
	}

	for _, opt := range opts {
		opt(s)
	}

	// Create trace directory if specified
	if s.traceDir != "" {
		if err := os.MkdirAll(s.traceDir, 0755); err != nil {
			s.logger.Printf("Warning: failed to create trace dir: %v", err)
		}
	}

	// Register routes
	s.registerRoutes()

	return s
}

// registerRoutes sets up the HTTP routes
func (s *EvalServer) registerRoutes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/invoke", s.handleInvoke)
	s.mux.HandleFunc("/invoke/", s.handleInvokeNamed)
	s.mux.HandleFunc("/traces/", s.handleTraces)
	s.mux.HandleFunc("/session", s.handleSessionCreate)
	s.mux.HandleFunc("/session/", s.handleSession)
	s.mux.HandleFunc("/list", s.handleList)
}

// ListenAndServe starts the HTTP server
func (s *EvalServer) ListenAndServe() error {
	addr := fmt.Sprintf(":%d", s.port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}

	s.logger.Printf("Starting eval server on %s", addr)
	s.logger.Printf("Registered agents: %v", s.agentNames())
	s.logger.Printf("Registered workflows: %v", s.workflowNames())

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *EvalServer) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// agentNames returns the names of registered agents
func (s *EvalServer) agentNames() []string {
	names := make([]string, 0, len(s.agents))
	for name := range s.agents {
		names = append(names, name)
	}
	return names
}

// workflowNames returns the names of registered workflows
func (s *EvalServer) workflowNames() []string {
	names := make([]string, 0, len(s.workflows))
	for name := range s.workflows {
		names = append(names, name)
	}
	return names
}

// getOrCreateSession gets an existing session or creates a new one
func (s *EvalServer) getOrCreateSession(sessionID string) *EvalSession {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()

	if sessionID != "" {
		if session, ok := s.sessions[sessionID]; ok {
			session.UpdateLastUsed()
			return session
		}
	}

	// Create new session
	newID := uuid.NewString()
	session := &EvalSession{
		ID:        newID,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		History:   make([]SessionMessage, 0),
	}
	s.sessions[newID] = session
	return session
}

// storeTrace stores a trace in memory and optionally to disk
func (s *EvalServer) storeTrace(trace *EvalTrace) {
	s.tracesMu.Lock()
	defer s.tracesMu.Unlock()
	s.traces[trace.ID] = trace
}

// getTrace retrieves a trace by ID
func (s *EvalServer) getTrace(id string) (*EvalTrace, bool) {
	s.tracesMu.RLock()
	defer s.tracesMu.RUnlock()
	trace, ok := s.traces[id]
	return trace, ok
}
