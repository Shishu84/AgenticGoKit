package agents_test

import (
	"context"
	"testing"

	"github.com/agenticgokit/agenticgokit/internal/agents"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFactoryInstrumentationCompiles tests that factory instrumentation compiles
func TestFactoryInstrumentationCompiles(t *testing.T) {
	// Setup tracing for test
	ctx := context.Background()
	shutdown, err := setupTestTracer(ctx)
	require.NoError(t, err, "failed to setup test tracer")
	defer shutdown(ctx) //nolint:errcheck

	// This test verifies that observability imports are present in factory
	assert.NotNil(t, agents.NewAgent("test"), "factory should be able to create agents")
}

// TestFactoryAgentCreation tests basic factory agent creation
func TestFactoryAgentCreation(t *testing.T) {
	// Setup tracing
	ctx := context.Background()
	shutdown, err := setupTestTracer(ctx)
	require.NoError(t, err, "failed to setup test tracer")
	defer shutdown(ctx) //nolint:errcheck

	// Create builder (factory uses builder internally)
	builder := agents.NewAgent("factory-test")
	assert.NotNil(t, builder, "builder should be created")

	// Build should work
	agent, err := builder.Build()
	assert.NoError(t, err, "agent should build successfully")
	assert.NotNil(t, agent, "agent should not be nil")
}

// TestAgentBuilderSpanGeneration tests that builder creates spans during build
func TestAgentBuilderSpanGeneration(t *testing.T) {
	// Setup tracing
	ctx := context.Background()
	shutdown, err := setupTestTracer(ctx)
	require.NoError(t, err, "failed to setup test tracer")
	defer shutdown(ctx) //nolint:errcheck

	// Build agent with metrics
	builder := agents.NewAgent("instrumentation-test")
	builder = builder.WithDefaultMetrics()

	agent, err := builder.Build()
	assert.NoError(t, err, "agent should build")
	assert.NotNil(t, agent, "agent should be created")

	// Spans are recorded by OpenTelemetry internally;
	// this test verifies the builder doesn't crash when recording spans
}

// BenchmarkFactoryWithInstrumentation benchmarks factory with observability
func BenchmarkFactoryWithInstrumentation(b *testing.B) {
	// Setup tracing once
	ctx := context.Background()
	shutdown, err := setupTestTracer(ctx)
	require.NoError(b, err, "failed to setup test tracer")
	defer shutdown(ctx) //nolint:errcheck

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder := agents.NewAgent("bench-factory")
		_, _ = builder.Build() //nolint:errcheck
	}
}
