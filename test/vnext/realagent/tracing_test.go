package realagent_test

import (
	"testing"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTracingEnabledInConfig tests tracing can be configured via Config
func TestTracingEnabledInConfig(t *testing.T) {
	config := &vnext.Config{
		Name:         "trace-config-agent",
		SystemPrompt: "Test",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "ollama",
			Model:    "gemma3:1b",
		},
		Tracing: &vnext.TracingConfig{
			Enabled: true,
			Level:   "debug",
		},
	}

	agent, err := vnext.NewBuilder(config.Name).
		WithConfig(config).
		Build()

	require.NoError(t, err, "Should create agent without error")
	require.NotNil(t, agent, "Agent should not be nil")

	// Verify config has tracing enabled
	assert.NotNil(t, agent.Config().Tracing, "Tracing config should be present")
	assert.True(t, agent.Config().Tracing.Enabled, "Tracing should be enabled")
	assert.Equal(t, "debug", agent.Config().Tracing.Level, "Tracing level should be debug")
}

// TestTracingDisabled tests tracing can be disabled
func TestTracingDisabled(t *testing.T) {
	config := &vnext.Config{
		Name:         "no-trace-agent",
		SystemPrompt: "Test",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider: "ollama",
			Model:    "gemma3:1b",
		},
		Tracing: &vnext.TracingConfig{
			Enabled: false,
			Level:   "none",
		},
	}

	agent, err := vnext.NewBuilder(config.Name).
		WithConfig(config).
		Build()

	require.NoError(t, err, "Should create agent without error")
	require.NotNil(t, agent, "Agent should not be nil")

	// Verify tracing is disabled
	assert.NotNil(t, agent.Config().Tracing, "Tracing config should be present")
	assert.False(t, agent.Config().Tracing.Enabled, "Tracing should be disabled")
}

// TestTraceIDInResult tests that results include trace information fields
func TestTraceIDInResult(t *testing.T) {
	result := &vnext.Result{
		Success:   true,
		Content:   "Test response",
		Duration:  100 * time.Millisecond,
		TraceID:   "test-trace-123",
		SessionID: "session-456",
		Metadata:  map[string]interface{}{},
	}

	assert.Equal(t, "test-trace-123", result.TraceID, "TraceID should be set in result")
	assert.Equal(t, "session-456", result.SessionID, "SessionID should be set in result")
	assert.True(t, result.IsSuccess(), "Result should be successful")
}

// TestRunOptionsTracingFields tests tracing fields in RunOptions
func TestRunOptionsTracingFields(t *testing.T) {
	opts := vnext.NewRunOptions()
	assert.NotNil(t, opts, "RunOptions should be created")
	assert.False(t, opts.TraceEnabled, "TraceEnabled should default to false")
	assert.Equal(t, "", opts.TraceLevel, "TraceLevel should default to empty")
}

// TestStreamMetadataTraceID tests that streaming preserves trace ID in metadata
func TestStreamMetadataTraceID(t *testing.T) {
	metadata := &vnext.StreamMetadata{
		AgentName: "test-agent",
		StartTime: time.Now(),
		Model:     "test-model",
		TraceID:   "stream-trace-789",
		SessionID: "stream-session-123",
		Extra:     map[string]interface{}{},
	}

	assert.Equal(t, "stream-trace-789", metadata.TraceID, "TraceID should be in stream metadata")
	assert.Equal(t, "stream-session-123", metadata.SessionID, "SessionID should be in stream metadata")
}

// TestRunOptionsIncludeTrace tests IncludeTrace option
func TestRunOptionsIncludeTrace(t *testing.T) {
	opts := vnext.RunWithDetailedResult()
	assert.NotNil(t, opts, "RunWithDetailedResult should create options")
	assert.True(t, opts.DetailedResult, "DetailedResult should be true")
	assert.True(t, opts.IncludeTrace, "IncludeTrace should be true with detailed result")
}

// TestAgentConfigWithTracingAndMemory tests agent config with both tracing and memory enabled
func TestAgentConfigWithTracingAndMemory(t *testing.T) {
	config := &vnext.Config{
		Name:         "traced-memory-agent",
		SystemPrompt: "Test",
		Timeout:      30 * time.Second,
		LLM: vnext.LLMConfig{
			Provider:    "ollama",
			Model:       "gemma3:1b",
			Temperature: 0.7,
		},
		Tracing: &vnext.TracingConfig{
			Enabled: true,
			Level:   "enhanced",
		},
		Memory: &vnext.MemoryConfig{
			Enabled:  true,
			Provider: "in_memory",
		},
	}

	agent, err := vnext.NewBuilder(config.Name).
		WithConfig(config).
		Build()

	require.NoError(t, err, "Should create agent without error")
	require.NotNil(t, agent, "Agent should not be nil")

	// Verify both tracing and memory are configured
	assert.NotNil(t, agent.Config().Tracing, "Tracing should be configured")
	assert.True(t, agent.Config().Tracing.Enabled, "Tracing should be enabled")
	assert.NotNil(t, agent.Config().Memory, "Memory should be configured")
	assert.True(t, agent.Config().Memory.Enabled, "Memory should be enabled")
}
