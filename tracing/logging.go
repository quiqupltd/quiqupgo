package tracing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

var (
	// loggerProviders caches LoggerProviders by service name to avoid creating duplicates.
	loggerProviders   = make(map[string]*sdklog.LoggerProvider)
	loggerProvidersMu sync.Mutex
)

// GetLoggerProvider returns a LoggerProvider for the given configuration.
// It caches providers by service name to avoid creating duplicates.
// If OTLP endpoint is not configured, returns nil (no-op logging).
//
// Options can be passed to customize the provider.
func GetLoggerProvider(ctx context.Context, cfg Config, opts ...ModuleOption) (*sdklog.LoggerProvider, error) {
	serviceName := cfg.GetServiceName()

	// Check cache first
	loggerProvidersMu.Lock()
	if lp, ok := loggerProviders[serviceName]; ok {
		loggerProvidersMu.Unlock()
		return lp, nil
	}
	loggerProvidersMu.Unlock()

	// Build options
	options := defaultModuleOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}

	// Create new provider
	lp, err := createLoggerProvider(ctx, cfg, options)
	if err != nil {
		return nil, err
	}

	// Double-check locking to avoid race
	loggerProvidersMu.Lock()
	defer loggerProvidersMu.Unlock()
	if existingLP, ok := loggerProviders[serviceName]; ok {
		// Another goroutine created it, shut down ours
		if lp != nil {
			_ = lp.Shutdown(ctx)
		}
		return existingLP, nil
	}

	loggerProviders[serviceName] = lp
	return lp, nil
}

// createLoggerProvider creates a new LoggerProvider with OTLP exporter.
func createLoggerProvider(ctx context.Context, cfg Config, opts *moduleOptions) (*sdklog.LoggerProvider, error) {
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
	exporterOpts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(endpoint),
	}

	if cfg.GetOTLPInsecure() {
		exporterOpts = append(exporterOpts, otlploghttp.WithInsecure())
	}

	// Add TLS config if provided
	tlsCfg, err := GetTLSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %w", err)
	}
	if tlsCfg != nil {
		exporterOpts = append(exporterOpts, otlploghttp.WithTLSClientConfig(tlsCfg))
	}

	// Create exporter
	exporter, err := otlploghttp.New(ctx, exporterOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	// Create LoggerProvider with batch processor
	lp := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)

	// Register as global provider
	global.SetLoggerProvider(lp)

	return lp, nil
}

// ShutdownLoggerProvider gracefully shuts down the LoggerProvider.
func ShutdownLoggerProvider(ctx context.Context, lp *sdklog.LoggerProvider) error {
	if lp == nil {
		return nil
	}

	// Use a timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return lp.Shutdown(shutdownCtx)
}

// ClearLoggerProviderCache clears the cached LoggerProviders.
// This is mainly useful for testing.
func ClearLoggerProviderCache() {
	loggerProvidersMu.Lock()
	defer loggerProvidersMu.Unlock()
	loggerProviders = make(map[string]*sdklog.LoggerProvider)
}
