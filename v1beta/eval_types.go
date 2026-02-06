package v1beta

import (
	"sync"
	"time"
)

// =============================================================================
// EVAL TYPES - Request/Response and Trace Storage
// =============================================================================

// InvokeRequest is the request body for /invoke endpoints
type InvokeRequest struct {
	Input     string                 `json:"input"`
	SessionID string                 `json:"session_id,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// InvokeResponse is the response body for /invoke endpoints
type InvokeResponse struct {
	Output      string   `json:"output"`
	TraceID     string   `json:"trace_id"`
	SessionID   string   `json:"session_id"`
	DurationMs  int64    `json:"duration_ms"`
	Success     bool     `json:"success"`
	ToolsCalled []string `json:"tools_called,omitempty"`
	Error       string   `json:"error,omitempty"`
}

// EvalTrace represents collected trace data for a single invocation
type EvalTrace struct {
	ID        string       `json:"id"`
	StartTime time.Time    `json:"start_time"`
	EndTime   time.Time    `json:"end_time"`
	Duration  int64        `json:"duration_ms"`
	Spans     []*EvalSpan  `json:"spans,omitempty"`
	Events    []*EvalEvent `json:"events,omitempty"`
}

// EvalSpan represents a span within a trace (e.g., agent execution, tool call)
type EvalSpan struct {
	Name       string                 `json:"name"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
	Duration   int64                  `json:"duration_ms"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Events     []*EvalEvent           `json:"events,omitempty"`
}

// EvalEvent represents an event within a span or trace
type EvalEvent struct {
	Name       string                 `json:"name"`
	Time       time.Time              `json:"time"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// EvalSession represents a conversation session with memory
type EvalSession struct {
	ID        string           `json:"id"`
	CreatedAt time.Time        `json:"created_at"`
	LastUsed  time.Time        `json:"last_used"`
	Memory    Memory           `json:"-"`
	History   []SessionMessage `json:"history"`
	mu        sync.Mutex       `json:"-"`
}

// SessionMessage represents a message in the session history
type SessionMessage struct {
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// UpdateLastUsed updates the last used timestamp
func (s *EvalSession) UpdateLastUsed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastUsed = time.Now()
}

// AddMessage adds a message to the session history
func (s *EvalSession) AddMessage(role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.History = append(s.History, SessionMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
}
