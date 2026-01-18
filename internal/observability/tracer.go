package observability

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type TracerConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string
	Exporter       string
	SampleRate     float64
	Debug          bool
	FilePath       string
}

func SetupTracer(ctx context.Context, cfg TracerConfig) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("service.version", cfg.ServiceVersion),
			attribute.String("deployment.environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	exporterName := strings.ToLower(strings.TrimSpace(cfg.Exporter))
	if exporterName == "" {
		exporterName = "console"
	}

	sampleRate := clampSampleRate(cfg.SampleRate)
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(sampleRate))
	if cfg.Debug {
		sampler = sdktrace.AlwaysSample()
	}

	exporter, err := selectExporter(ctx, exporterName, cfg)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

func selectExporter(ctx context.Context, exporter string, cfg TracerConfig) (sdktrace.SpanExporter, error) {
	switch exporter {
	case "otlp", "otlphttp":
		return otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(cfg.Endpoint), otlptracehttp.WithInsecure())
	case "console":
		return newConsoleExporter(cfg.Debug)
	case "file":
		return newFileExporter(cfg.FilePath, cfg.Debug)
	default:
		return nil, fmt.Errorf("unsupported exporter: %s", exporter)
	}
}

func clampSampleRate(rate float64) float64 {
	switch {
	case rate < 0:
		return 0
	case rate > 1:
		return 1
	default:
		return rate
	}
}
