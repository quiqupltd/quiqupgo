package tracing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	// meterProviders caches MeterProviders by service name to avoid creating duplicates.
	meterProviders   = make(map[string]*sdkmetric.MeterProvider)
	meterProvidersMu sync.Mutex
)

// GetMeterProvider returns a MeterProvider for the given configuration.
// It caches providers by service name to avoid creating duplicates.
// If OTLP endpoint is not configured, returns nil (no-op metrics).
//
// Options can be passed to customize the provider (e.g., WithMetricInterval).
func GetMeterProvider(ctx context.Context, cfg Config, opts ...ModuleOption) (*sdkmetric.MeterProvider, error) {
	serviceName := cfg.GetServiceName()

	// Check cache first
	meterProvidersMu.Lock()
	if mp, ok := meterProviders[serviceName]; ok {
		meterProvidersMu.Unlock()
		return mp, nil
	}
	meterProvidersMu.Unlock()

	// Build options
	options := defaultModuleOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}

	// Create new provider
	mp, err := createMeterProvider(ctx, cfg, options)
	if err != nil {
		return nil, err
	}

	// Double-check locking to avoid race
	meterProvidersMu.Lock()
	defer meterProvidersMu.Unlock()
	if existingMP, ok := meterProviders[serviceName]; ok {
		// Another goroutine created it, shut down ours
		if mp != nil {
			_ = mp.Shutdown(ctx)
		}
		return existingMP, nil
	}

	meterProviders[serviceName] = mp
	return mp, nil
}

// GetMeter returns a Meter from the MeterProvider.
// If mp is nil, returns a no-op meter from the global provider.
func GetMeter(mp *sdkmetric.MeterProvider) metric.Meter {
	if mp == nil {
		// Return no-op meter from global (which might be no-op if not set)
		return otel.Meter(TracerName())
	}
	return mp.Meter(TracerName())
}

// createMeterProvider creates a new MeterProvider with OTLP exporter.
func createMeterProvider(ctx context.Context, cfg Config, opts *moduleOptions) (*sdkmetric.MeterProvider, error) {
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
	exporterOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(endpoint),
	}

	if cfg.GetOTLPInsecure() {
		exporterOpts = append(exporterOpts, otlpmetrichttp.WithInsecure())
	}

	// Add TLS config if provided
	tlsCfg, err := GetTLSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS config: %w", err)
	}
	if tlsCfg != nil {
		exporterOpts = append(exporterOpts, otlpmetrichttp.WithTLSClientConfig(tlsCfg))
	}

	// Create exporter
	exporter, err := otlpmetrichttp.New(ctx, exporterOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}

	// Create periodic reader with configurable interval
	readerInterval := opts.metricInterval
	if readerInterval == 0 {
		readerInterval = 10 * time.Second
	}

	reader := sdkmetric.NewPeriodicReader(exporter,
		sdkmetric.WithInterval(readerInterval),
	)

	// Create MeterProvider
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)

	// Register as global provider
	otel.SetMeterProvider(mp)

	return mp, nil
}

// ShutdownMeterProvider gracefully shuts down the MeterProvider.
func ShutdownMeterProvider(ctx context.Context, mp *sdkmetric.MeterProvider) error {
	if mp == nil {
		return nil
	}

	// Use a timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return mp.Shutdown(shutdownCtx)
}

// ClearMeterProviderCache clears the cached MeterProviders.
// This is mainly useful for testing.
func ClearMeterProviderCache() {
	meterProvidersMu.Lock()
	defer meterProvidersMu.Unlock()
	meterProviders = make(map[string]*sdkmetric.MeterProvider)
}
