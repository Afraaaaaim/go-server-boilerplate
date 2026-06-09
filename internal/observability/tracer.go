package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/Afraaaaaim/go-server-boilerplate/internal/config"
)

// TracerProvider wraps the SDK provider so the caller can shut it down cleanly.
type TracerProvider struct {
	provider *sdktrace.TracerProvider
}

// Shutdown flushes and stops the tracer provider.
// Always call this on application exit: defer tp.Shutdown(ctx)
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.provider == nil {
		return nil
	}
	return tp.provider.Shutdown(ctx)
}

// InitTracer sets up the global OTel tracer provider.
// If OTLP is not configured, it installs a no-op provider so all
// otel.Tracer() calls still work without panicking.
func InitTracer(ctx context.Context, cfg *config.Config) (*TracerProvider, error) {
	res, err := buildResource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build OTel resource: %w", err)
	}

	// No OTLP endpoint configured — use a no-op provider so trace calls
	// are safe but nothing is exported.
	if !cfg.OTLPEnabled() {
		noop := sdktrace.NewTracerProvider(sdktrace.WithResource(res))
		otel.SetTracerProvider(noop)
		otel.SetTextMapPropagator(buildPropagator())
		return &TracerProvider{provider: noop}, nil
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		// In production set WithTLSClientConfig; insecure is fine for local/docker
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter), // async batch export, better than WithSyncer in prod
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // swap for ParentBased in high-traffic
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(buildPropagator())

	return &TracerProvider{provider: provider}, nil
}

// buildResource creates the OTel resource that identifies this service
// in your observability backend (Jaeger, Grafana, etc).
func buildResource(ctx context.Context, cfg *config.Config) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironment(cfg.Env),
			),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithHost(),
	)
}

// buildPropagator returns a composite propagator that handles both
// W3C TraceContext and Baggage headers — the current OTel standard.
func buildPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}