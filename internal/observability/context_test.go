package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

func TestLoggerContextRoundTrip(t *testing.T) {
	logger := zerolog.New(&bytes.Buffer{})
	ctx := WithLogger(context.Background(), &logger)

	got := LoggerFromContext(ctx)
	if got != &logger {
		t.Fatalf("expected logger pointer from context")
	}
}

func TestRunIDPropagation(t *testing.T) {
	ctx := WithRunID(context.Background(), "run-123")
	if got := RunIDFromContext(ctx); got != "run-123" {
		t.Fatalf("RunIDFromContext got %s want %s", got, "run-123")
	}
}

func TestEnrichLoggerAddsTraceAndRunID(t *testing.T) {
	var buf bytes.Buffer
	base := zerolog.New(&buf)

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 2},
		SpanID:     trace.SpanID{0, 0, 0, 0, 0, 0, 0, 3},
		TraceFlags: trace.FlagsSampled,
	})

	ctx := trace.ContextWithSpanContext(context.Background(), sc)
	ctx = WithRunID(ctx, "run-abc")

	enriched := EnrichLogger(ctx, &base)
	enriched.Info().Msg("hello")

	var line map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatalf("failed to unmarshal log: %v", err)
	}

	if line["trace_id"] != sc.TraceID().String() {
		t.Fatalf("trace_id mismatch: got %v", line["trace_id"])
	}

	if line["span_id"] != sc.SpanID().String() {
		t.Fatalf("span_id mismatch: got %v", line["span_id"])
	}

	if line["run_id"] != "run-abc" {
		t.Fatalf("run_id mismatch: got %v", line["run_id"])
	}
}

func TestCreateChildLogger(t *testing.T) {
	var buf bytes.Buffer
	base := zerolog.New(&buf)
	ctx := WithLogger(context.Background(), &base)

	child := CreateChildLogger(ctx, "component-x")
	child.Info().Msg("hi")

	var line map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatalf("failed to unmarshal log: %v", err)
	}

	if line["component"] != "component-x" {
		t.Fatalf("component field missing, got %v", line["component"])
	}
}
