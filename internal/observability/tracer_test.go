package observability

import (
	"context"
	"path/filepath"
	"testing"
)

func TestClampSampleRate(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{-1, 0},
		{0, 0},
		{0.5, 0.5},
		{1, 1},
		{5, 1},
	}

	for _, tc := range cases {
		got := clampSampleRate(tc.in)
		if got != tc.want {
			t.Fatalf("clampSampleRate(%v)=%v want %v", tc.in, got, tc.want)
		}
	}
}

func TestSelectExporterErrors(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	_, err := selectExporter(ctx, "file", TracerConfig{})
	if err == nil {
		t.Fatal("expected error for missing file path")
	}

	_, err = selectExporter(ctx, "unknown", TracerConfig{})
	if err == nil {
		t.Fatal("expected error for unsupported exporter")
	}
}

func TestSelectExporterFile(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "trace.json")

	exp, err := selectExporter(ctx, "file", TracerConfig{FilePath: path, Debug: true})
	if err != nil {
		t.Fatalf("selectExporter(file) error = %v", err)
	}

	if err := exp.Shutdown(ctx); err != nil {
		t.Fatalf("exporter shutdown error = %v", err)
	}
}

func TestSetupTracerConsole(t *testing.T) {
	ctx := context.Background()
	shutdown, err := SetupTracer(ctx, TracerConfig{
		ServiceName:    "test-service",
		ServiceVersion: "v0.0.1",
		Environment:    "test",
		Exporter:       "console",
		SampleRate:     1.0,
		Debug:          true,
	})
	if err != nil {
		t.Fatalf("SetupTracer error = %v", err)
	}

	if shutdown == nil {
		t.Fatal("shutdown should not be nil")
	}

	if err := shutdown(ctx); err != nil {
		t.Fatalf("shutdown error = %v", err)
	}
}
