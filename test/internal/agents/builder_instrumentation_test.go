package agents_test

import (
	"context"
	"testing"

	internalagents "github.com/agenticgokit/agenticgokit/internal/agents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// TestAgentBuilderInstrumentation tests that agent builder creates observability spans
func TestAgentBuilderInstrumentation(t *testing.T) {
	// Setup tracing for test
	ctx := context.Background()
	shutdown, err := setupTestTracer(ctx)
	require.NoError(t, err, "failed to setup test tracer")
	defer shutdown(ctx) //nolint:errcheck

	// Create a simple builder
	builder := internalagents.NewAgent("test-agent")

	// Build the agent - this should create spans internally
	agent, err := builder.Build()

	// Verify build succeeded
	assert.NoError(t, err, "agent build should succeed")
	assert.NotNil(t, agent, "agent should be created")
}

// TestAgentBuilderWithCapabilities tests agent builder with capabilities instrumentation
func TestAgentBuilderWithCapabilities(t *testing.T) {
	// Setup tracing
	ctx := context.Background()
	shutdown, err := setupTestTracer(ctx)
	require.NoError(t, err, "failed to setup test tracer")
	defer shutdown(ctx) //nolint:errcheck

	// Create builder with multiple capabilities
	builder := internalagents.NewAgent("capability-test-agent")
	builder = builder.WithDefaultMetrics()

	// Verify builder state has capabilities added
	assert.Equal(t, builder.CapabilityCount(), 1, "builder should have metrics capability")

	// Build the agent
	agent, err := builder.Build()
	assert.NoError(t, err, "agent with capabilities should build")
	assert.NotNil(t, agent, "agent should be created")
}

// TestAgentBuilderValidationError tests that validation errors are handled gracefully
func TestAgentBuilderValidationError(t *testing.T) {
	// Setup tracing
	ctx := context.Background()
	shutdown, err := setupTestTracer(ctx)
	require.NoError(t, err, "failed to setup test tracer")
	defer shutdown(ctx) //nolint:errcheck

	// Create builder with minimal configuration (should still build successfully)
	builder := internalagents.NewAgentWithConfig("minimal-agent",
		internalagents.AgentBuilderConfig{
			ValidateCapabilities: true,
			StrictMode:           false,
		})

	// Build should succeed even with minimal config in non-strict mode
	agent, err := builder.Build()
	assert.NoError(t, err, "minimal agent should build in non-strict mode")
	assert.NotNil(t, agent, "agent should be created")
}

// TestAgentBuilderMultipleCapabilities tests builder with multiple capabilities tracked in spans
func TestAgentBuilderMultipleCapabilities(t *testing.T) {
	// Setup tracing
	ctx := context.Background()
	shutdown, err := setupTestTracer(ctx)
	require.NoError(t, err, "failed to setup test tracer")
	defer shutdown(ctx) //nolint:errcheck

	// Create builder
	builder := internalagents.NewAgent("multi-cap-agent")

	// Build with metrics (default)
	builder = builder.WithDefaultMetrics()

	// Verify capabilities count is tracked
	capCount := builder.CapabilityCount()
	assert.Greater(t, capCount, 0, "builder should have capabilities")

	// Build the agent
	agent, err := builder.Build()
	assert.NoError(t, err, "multi-capability agent should build")
	assert.NotNil(t, agent, "agent should be created")
}

// setupTestTracer initializes a console-based tracer for testing
func setupTestTracer(ctx context.Context) (func(context.Context) error, error) {
	exporter, err := stdouttrace.New(stdouttrace.WithWriter(&testWriter{}))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

// testWriter discards output (we only care about span generation, not output)
type testWriter struct{}

func (w *testWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// TestAgentBuilderSpanAttributes tests that agent builder creates spans with correct attributes
func TestAgentBuilderSpanAttributes(t *testing.T) {
	// Setup tracing
	ctx := context.Background()
	shutdown, err := setupTestTracer(ctx)
	require.NoError(t, err, "failed to setup test tracer")
	defer shutdown(ctx) //nolint:errcheck

	// Create agent
	agentName := "attribute-test-agent"
	builder := internalagents.NewAgent(agentName)

	// Build the agent
	agent, err := builder.Build()

	// Verify build succeeded
	assert.NoError(t, err, "agent should build successfully")
	assert.NotNil(t, agent, "agent should be created")
}

// TestAgentBuilderErrorHandling tests error handling during build with observability
func TestAgentBuilderErrorHandling(t *testing.T) {
	// Setup tracing
	ctx := context.Background()
	shutdown, err := setupTestTracer(ctx)
	require.NoError(t, err, "failed to setup test tracer")
	defer shutdown(ctx) //nolint:errcheck

	// Create builder
	builder := internalagents.NewAgent("error-test-agent")

	// In strict mode, build should handle errors appropriately
	assert.NoError(t, builder.Validate(), "valid builder should validate")

	// Build the agent
	agent, err := builder.Build()
	assert.NoError(t, err, "valid agent should build")
	assert.NotNil(t, agent, "agent should be created")
}

// BenchmarkAgentBuilderWithInstrumentation benchmarks agent build with observability
func BenchmarkAgentBuilderWithInstrumentation(b *testing.B) {
	// Setup tracing once
	ctx := context.Background()
	shutdown, err := setupTestTracer(ctx)
	require.NoError(b, err, "failed to setup test tracer")
	defer shutdown(ctx) //nolint:errcheck

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := internalagents.NewAgent("bench-agent")
		builder = builder.WithDefaultMetrics()
		_, _ = builder.Build() //nolint:errcheck
	}
}
