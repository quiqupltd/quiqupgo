package tracing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var (
	// tracerProviders caches TracerProviders by service name to avoid creating duplicates.
	tracerProviders   = make(map[string]*trace.TracerProvider)
	tracerProvidersMu sync.Mutex
)

// GetTracerProvider returns a TracerProvider for the given configuration.
// It caches providers by service name to avoid creating duplicates.
// If OTLP endpoint is not configured, returns nil (no-op tracing).
//
// Options can be passed to customize the provider (e.g., WithBatchTimeout, WithSampler).
func GetTracerProvider(ctx context.Context, cfg Config, opts ...ModuleOption) (*trace.TracerProvider, error) {
	serviceName := cfg.GetServiceName()

	// Check cache first
	tracerProvidersMu.Lock()
	if tp, ok := tracerProviders[serviceName]; ok {
		tracerProvidersMu.Unlock()
		return tp, nil
	}
	tracerProvidersMu.Unlock()

	// Build options
	options := defaultModuleOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}

	// Create new provider
	tp, err := createTracerProvider(ctx, cfg, options)
	if err != nil {
		return nil, err
	}

	// Double-check locking to avoid race
	tracerProvidersMu.Lock()
	defer tracerProvidersMu.Unlock()
	if existingTP, ok := tracerProviders[serviceName]; ok {
		// Another goroutine created it, shut down ours
		if tp != nil {
			_ = tp.Shutdown(ctx)
		}
		return existingTP, nil
	}

	tracerProviders[serviceName] = tp
	return tp, nil
}

// GetTracer returns a Tracer from the TracerProvider.
// If tp is nil, returns a no-op tracer from the global provider.
func GetTracer(tp *trace.TracerProvider) oteltrace.Tracer {
	if tp == nil {
		// Return no-op tracer from global (which might be no-op if not set)
		return otel.Tracer(TracerName())
	}
	return tp.Tracer(TracerName())
}

// createTracerProvider creates a new TracerProvider with OTLP exporter.
func createTracerProvider(ctx context.Context, cfg Config, opts *moduleOptions) (*trace.TracerProvider, error) {
	endpoint := cfg.GetOTLPEndpoint()
	if endpoint == "" {
		// No endpoint configured, return nil (graceful degradation)
		return nil, nil
	}

	// Create resource
	res, err := GetResource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Build exporter options
	exporterOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
	}

	if cfg.GetOTLPInsecure() {
		exporterOpts = append(exporterOpts, otlptracehttp.WithInsecure())
	}

	// Add TLS config if provided
	tlsCfg, err := GetTLSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %w", err)
	}
	if tlsCfg != nil {
		exporterOpts = append(exporterOpts, otlptracehttp.WithTLSClientConfig(tlsCfg))
	}

	// Create exporter
	exporter, err := otlptracehttp.New(ctx, exporterOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Build TracerProvider options
	tpOpts := []trace.TracerProviderOption{
		trace.WithResource(res),
		trace.WithBatcher(exporter,
			trace.WithBatchTimeout(opts.batchTimeout),
		),
	}

	if opts.sampler != nil {
		tpOpts = append(tpOpts, trace.WithSampler(opts.sampler))
	}

	tp := trace.NewTracerProvider(tpOpts...)

	// Register as global provider
	otel.SetTracerProvider(tp)

	return tp, nil
}

// ShutdownTracerProvider gracefully shuts down the TracerProvider.
func ShutdownTracerProvider(ctx context.Context, tp *trace.TracerProvider) error {
	if tp == nil {
		return nil
	}

	// Use a timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return tp.Shutdown(shutdownCtx)
}

// ClearTracerProviderCache clears the cached TracerProviders.
// This is mainly useful for testing.
func ClearTracerProviderCache() {
	tracerProvidersMu.Lock()
	defer tracerProvidersMu.Unlock()
	tracerProviders = make(map[string]*trace.TracerProvider)
}
