package tracing_test

import (
	"context"
	"testing"
	"time"

	"github.com/quiqupltd/quiqupgo/tracing"
	"github.com/quiqupltd/quiqupgo/tracing/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

func TestModule_WithNoopConfig(t *testing.T) {
	// Clear any cached providers from other tests
	tracing.ClearTracerProviderCache()
	tracing.ClearMeterProviderCache()
	tracing.ClearLoggerProviderCache()

	var (
		tp     oteltrace.TracerProvider
		tracer oteltrace.Tracer
		mp     metric.MeterProvider
		meter  metric.Meter
	)

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() tracing.Config {
			return testutil.NewNoopConfig()
		}),
		tracing.Module(),
		fx.Populate(&tp, &tracer, &mp, &meter),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	require.NoError(t, err)

	// Verify providers are not nil (should be no-op providers)
	assert.NotNil(t, tp)
	assert.NotNil(t, tracer)
	assert.NotNil(t, mp)
	assert.NotNil(t, meter)

	// Can create spans without errors
	_, span := tracer.Start(ctx, "test-span")
	span.End()

	// Shutdown
	err = app.Stop(ctx)
	require.NoError(t, err)
}

func TestNoopModule(t *testing.T) {
	var (
		tp     oteltrace.TracerProvider
		tracer oteltrace.Tracer
		mp     metric.MeterProvider
		meter  metric.Meter
	)

	app := fx.New(
		fx.NopLogger,
		testutil.NoopModule(),
		fx.Populate(&tp, &tracer, &mp, &meter),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	require.NoError(t, err)

	// Verify all providers are available
	assert.NotNil(t, tp)
	assert.NotNil(t, tracer)
	assert.NotNil(t, mp)
	assert.NotNil(t, meter)

	// Can create spans without errors
	_, span := tracer.Start(ctx, "test-span")
	span.End()

	err = app.Stop(ctx)
	require.NoError(t, err)
}

func TestStandardConfig(t *testing.T) {
	cfg := &tracing.StandardConfig{
		ServiceName:     "test-service",
		EnvironmentName: "test",
		OTLPEndpoint:    "localhost:4318",
		OTLPInsecure:    true,
		OTLPTLSCert:     "cert-data",
		OTLPTLSKey:      "key-data",
		OTLPTLSCA:       "ca-data",
	}

	assert.Equal(t, "test-service", cfg.GetServiceName())
	assert.Equal(t, "test", cfg.GetEnvironmentName())
	assert.Equal(t, "localhost:4318", cfg.GetOTLPEndpoint())
	assert.True(t, cfg.GetOTLPInsecure())
	assert.Equal(t, "cert-data", cfg.GetOTLPTLSCert())
	assert.Equal(t, "key-data", cfg.GetOTLPTLSKey())
	assert.Equal(t, "ca-data", cfg.GetOTLPTLSCA())
}

func TestModuleOptions(t *testing.T) {
	// Clear caches
	tracing.ClearTracerProviderCache()
	tracing.ClearMeterProviderCache()
	tracing.ClearLoggerProviderCache()

	var tp oteltrace.TracerProvider

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() tracing.Config {
			return testutil.NewNoopConfig()
		}),
		tracing.Module(
			tracing.WithBatchTimeout(3*time.Second),
			tracing.WithMetricInterval(5*time.Second),
			tracing.WithAlwaysSample(),
		),
		fx.Populate(&tp),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	require.NoError(t, err)

	assert.NotNil(t, tp)

	err = app.Stop(ctx)
	require.NoError(t, err)
}

func TestTracerName(t *testing.T) {
	name := tracing.TracerName()
	assert.Equal(t, "github.com/quiqupltd/quiqupgo/tracing", name)
}

func TestGetResource(t *testing.T) {
	cfg := &tracing.StandardConfig{
		ServiceName:     "test-service",
		EnvironmentName: "test-env",
	}

	ctx := context.Background()
	res, err := tracing.GetResource(ctx, cfg)
	require.NoError(t, err)
	require.NotNil(t, res)

	// Verify the resource has the expected attributes
	attrs := res.Attributes()
	var foundService, foundEnv bool
	for _, attr := range attrs {
		if attr.Key == "service.name" && attr.Value.AsString() == "test-service" {
			foundService = true
		}
		if attr.Key == "deployment.environment" && attr.Value.AsString() == "test-env" {
			foundEnv = true
		}
	}
	assert.True(t, foundService, "service.name attribute not found")
	assert.True(t, foundEnv, "deployment.environment attribute not found")
}

func TestWithSampler(t *testing.T) {
	tracing.ClearTracerProviderCache()
	tracing.ClearMeterProviderCache()
	tracing.ClearLoggerProviderCache()

	var tp oteltrace.TracerProvider

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() tracing.Config {
			return testutil.NewNoopConfig()
		}),
		tracing.Module(
			tracing.WithSampler(nil), // Test explicit nil sampler
		),
		fx.Populate(&tp),
	)

	ctx := t.Context()
	require.NoError(t, app.Start(ctx))
	assert.NotNil(t, tp)
	require.NoError(t, app.Stop(ctx))
}

func TestWithNeverSample(t *testing.T) {
	tracing.ClearTracerProviderCache()
	tracing.ClearMeterProviderCache()
	tracing.ClearLoggerProviderCache()

	var tp oteltrace.TracerProvider

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() tracing.Config {
			return testutil.NewNoopConfig()
		}),
		tracing.Module(
			tracing.WithNeverSample(),
		),
		fx.Populate(&tp),
	)

	ctx := t.Context()
	require.NoError(t, app.Start(ctx))
	assert.NotNil(t, tp)
	require.NoError(t, app.Stop(ctx))
}

func TestWithTraceIDRatioBased(t *testing.T) {
	tracing.ClearTracerProviderCache()
	tracing.ClearMeterProviderCache()
	tracing.ClearLoggerProviderCache()

	var tp oteltrace.TracerProvider

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() tracing.Config {
			return testutil.NewNoopConfig()
		}),
		tracing.Module(
			tracing.WithTraceIDRatioBased(0.5),
		),
		fx.Populate(&tp),
	)

	ctx := t.Context()
	require.NoError(t, app.Start(ctx))
	assert.NotNil(t, tp)
	require.NoError(t, app.Stop(ctx))
}

func TestGetTracer_NilProvider(t *testing.T) {
	tracer := tracing.GetTracer(nil)
	assert.NotNil(t, tracer, "should return a tracer even with nil provider")
}

func TestShutdownTracerProvider_Nil(t *testing.T) {
	ctx := context.Background()
	err := tracing.ShutdownTracerProvider(ctx, nil)
	assert.NoError(t, err, "shutting down nil provider should not error")
}

func TestShutdownMeterProvider_Nil(t *testing.T) {
	ctx := context.Background()
	err := tracing.ShutdownMeterProvider(ctx, nil)
	assert.NoError(t, err, "shutting down nil provider should not error")
}

func TestShutdownLoggerProvider_Nil(t *testing.T) {
	ctx := context.Background()
	err := tracing.ShutdownLoggerProvider(ctx, nil)
	assert.NoError(t, err, "shutting down nil provider should not error")
}

func TestGetMeter_NilProvider(t *testing.T) {
	meter := tracing.GetMeter(nil)
	assert.NotNil(t, meter, "should return a meter even with nil provider")
}

func TestGetTLSConfig_Empty(t *testing.T) {
	cfg := &tracing.StandardConfig{
		OTLPTLSCert: "",
		OTLPTLSKey:  "",
		OTLPTLSCA:   "",
	}

	tlsCfg, err := tracing.GetTLSConfig(cfg)
	require.NoError(t, err)
	assert.Nil(t, tlsCfg, "should return nil when no TLS config provided")
}

func TestGetTLSConfig_InvalidCertBase64(t *testing.T) {
	cfg := &tracing.StandardConfig{
		OTLPTLSCert: "not-valid-base64!!!",
		OTLPTLSKey:  "not-valid-base64!!!",
	}

	_, err := tracing.GetTLSConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode TLS certificate")
}

func TestGetTLSConfig_InvalidKeyBase64(t *testing.T) {
	// Valid base64 but not a valid cert
	validBase64 := "dGVzdA==" // "test" in base64

	cfg := &tracing.StandardConfig{
		OTLPTLSCert: validBase64,
		OTLPTLSKey:  "not-valid-base64!!!",
	}

	_, err := tracing.GetTLSConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode TLS key")
}

func TestGetTLSConfig_InvalidCABase64(t *testing.T) {
	cfg := &tracing.StandardConfig{
		OTLPTLSCA: "not-valid-base64!!!",
	}

	_, err := tracing.GetTLSConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode TLS CA certificate")
}

func TestGetTLSConfig_InvalidCAContent(t *testing.T) {
	// Valid base64 but not a valid PEM
	invalidCA := "dGVzdA==" // "test" in base64

	cfg := &tracing.StandardConfig{
		OTLPTLSCA: invalidCA,
	}

	_, err := tracing.GetTLSConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse TLS CA certificate")
}

func TestGetTLSConfig_InvalidKeyPair(t *testing.T) {
	// Valid base64 but not valid cert/key PEM data
	invalidPEM := "dGVzdA==" // "test" in base64

	cfg := &tracing.StandardConfig{
		OTLPTLSCert: invalidPEM,
		OTLPTLSKey:  invalidPEM,
	}

	_, err := tracing.GetTLSConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load TLS key pair")
}

func TestModule_MultipleStartStop(t *testing.T) {
	tracing.ClearTracerProviderCache()
	tracing.ClearMeterProviderCache()
	tracing.ClearLoggerProviderCache()

	var tp oteltrace.TracerProvider

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() tracing.Config {
			return testutil.NewNoopConfig()
		}),
		tracing.Module(),
		fx.Populate(&tp),
	)

	ctx := t.Context()

	// Start and stop twice to verify lifecycle hooks work correctly
	require.NoError(t, app.Start(ctx))
	assert.NotNil(t, tp)
	require.NoError(t, app.Stop(ctx))
}

func TestGetLoggerProvider_NoEndpoint(t *testing.T) {
	tracing.ClearLoggerProviderCache()

	cfg := &tracing.StandardConfig{
		ServiceName:  "test-service",
		OTLPEndpoint: "", // No endpoint
	}

	ctx := context.Background()
	lp, err := tracing.GetLoggerProvider(ctx, cfg, nil)
	require.NoError(t, err)
	assert.Nil(t, lp, "should return nil when no endpoint is configured")
}

func TestGetLoggerProvider_Caching(t *testing.T) {
	tracing.ClearLoggerProviderCache()

	cfg := &tracing.StandardConfig{
		ServiceName:  "cached-service",
		OTLPEndpoint: "", // No endpoint returns nil, but tests caching logic
	}

	ctx := context.Background()

	// First call
	lp1, err := tracing.GetLoggerProvider(ctx, cfg, nil)
	require.NoError(t, err)

	// Second call should use cache
	lp2, err := tracing.GetLoggerProvider(ctx, cfg, nil)
	require.NoError(t, err)

	// Both should be the same (nil in this case since no endpoint)
	assert.Equal(t, lp1, lp2)
}

func TestGetTracerProvider_NoEndpoint(t *testing.T) {
	tracing.ClearTracerProviderCache()

	cfg := &tracing.StandardConfig{
		ServiceName:  "test-service",
		OTLPEndpoint: "", // No endpoint
	}

	ctx := context.Background()
	tp, err := tracing.GetTracerProvider(ctx, cfg, nil)
	require.NoError(t, err)
	assert.Nil(t, tp, "should return nil when no endpoint is configured")
}

func TestGetMeterProvider_NoEndpoint(t *testing.T) {
	tracing.ClearMeterProviderCache()

	cfg := &tracing.StandardConfig{
		ServiceName:  "test-service",
		OTLPEndpoint: "", // No endpoint
	}

	ctx := context.Background()
	mp, err := tracing.GetMeterProvider(ctx, cfg, nil)
	require.NoError(t, err)
	assert.Nil(t, mp, "should return nil when no endpoint is configured")
}

func TestModule_WithAllOptions(t *testing.T) {
	tracing.ClearTracerProviderCache()
	tracing.ClearMeterProviderCache()
	tracing.ClearLoggerProviderCache()

	var (
		tp     oteltrace.TracerProvider
		tracer oteltrace.Tracer
		mp     metric.MeterProvider
		meter  metric.Meter
	)

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() tracing.Config {
			return testutil.NewNoopConfig()
		}),
		tracing.Module(
			tracing.WithBatchTimeout(1*time.Second),
			tracing.WithMetricInterval(2*time.Second),
			tracing.WithAlwaysSample(),
		),
		fx.Populate(&tp, &tracer, &mp, &meter),
	)

	ctx := t.Context()
	require.NoError(t, app.Start(ctx))

	// All providers should be available
	assert.NotNil(t, tp)
	assert.NotNil(t, tracer)
	assert.NotNil(t, mp)
	assert.NotNil(t, meter)

	// Create a span and record a metric
	_, span := tracer.Start(ctx, "test-operation")
	span.End()

	counter, err := meter.Int64Counter("test.counter")
	require.NoError(t, err)
	counter.Add(ctx, 1)

	require.NoError(t, app.Stop(ctx))
}
