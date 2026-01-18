package observability

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type exporterWithCloser struct {
	sdktrace.SpanExporter
	closer io.Closer
}

func (e *exporterWithCloser) Shutdown(ctx context.Context) error {
	shutdownErr := e.SpanExporter.Shutdown(ctx)
	closeErr := e.closer.Close()

	if shutdownErr != nil {
		return shutdownErr
	}

	return closeErr
}

func newConsoleExporter(pretty bool) (sdktrace.SpanExporter, error) {
	opts := []stdouttrace.Option{
		stdouttrace.WithWriter(os.Stdout),
	}

	if pretty {
		opts = append(opts, stdouttrace.WithPrettyPrint())
	}

	return stdouttrace.New(opts...)
}

func newFileExporter(path string, pretty bool) (sdktrace.SpanExporter, error) {
	if path == "" {
		return nil, fmt.Errorf("file exporter requires a path")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create exporter directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open exporter file: %w", err)
	}

	opts := []stdouttrace.Option{stdouttrace.WithWriter(file)}
	if pretty {
		opts = append(opts, stdouttrace.WithPrettyPrint())
	}

	exp, err := stdouttrace.New(opts...)
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	return &exporterWithCloser{SpanExporter: exp, closer: file}, nil
}
