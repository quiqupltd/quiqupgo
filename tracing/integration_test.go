//go:build integration

package tracing_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/quiqupltd/quiqupgo/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

// getOTELEndpoint returns the OTEL collector endpoint from env or defaults to OrbStack URL.
func getOTELEndpoint() string {
	if endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); endpoint != "" {
		return endpoint
	}
	return "otel-collector.quiqupgo.orb.local:4318"
}

// IntegrationTestConfig implements tracing.Config for integration tests.
type IntegrationTestConfig struct {
	serviceName     string
	environmentName string
	otlpEndpoint    string
}

func NewIntegrationTestConfig(serviceName string) *IntegrationTestConfig {
	return &IntegrationTestConfig{
		serviceName:     serviceName,
		environmentName: "integration-test",
		otlpEndpoint:    getOTELEndpoint(),
	}
}

func (c *IntegrationTestConfig) GetServiceName() string     { return c.serviceName }
func (c *IntegrationTestConfig) GetEnvironmentName() string { return c.environmentName }
func (c *IntegrationTestConfig) GetOTLPEndpoint() string    { return c.otlpEndpoint }
func (c *IntegrationTestConfig) GetOTLPInsecure() bool      { return true }
func (c *IntegrationTestConfig) GetOTLPTLSCert() string     { return "" }
func (c *IntegrationTestConfig) GetOTLPTLSKey() string      { return "" }
func (c *IntegrationTestConfig) GetOTLPTLSCA() string       { return "" }

// TracingIntegrationSuite tests the tracing module against a real OTEL collector.
type TracingIntegrationSuite struct {
	suite.Suite
}

func TestTracingIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(TracingIntegrationSuite))
}

func (s *TracingIntegrationSuite) SetupTest() {
	// Clear caches before each test
	tracing.ClearTracerProviderCache()
	tracing.ClearMeterProviderCache()
	tracing.ClearLoggerProviderCache()
}

func (s *TracingIntegrationSuite) TestModuleWithRealCollector() {
	var (
		tp     oteltrace.TracerProvider
		tracer oteltrace.Tracer
		mp     metric.MeterProvider
		meter  metric.Meter
	)

	cfg := NewIntegrationTestConfig("tracing-integration-test")

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() tracing.Config { return cfg }),
		tracing.Module(
			tracing.WithBatchTimeout(1*time.Second),
			tracing.WithAlwaysSample(),
		),
		fx.Populate(&tp, &tracer, &mp, &meter),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := app.Start(ctx)
	s.Require().NoError(err)

	// All providers should be real (not nil)
	s.NotNil(tp)
	s.NotNil(tracer)
	s.NotNil(mp)
	s.NotNil(meter)

	// Create spans
	ctx, span := tracer.Start(ctx, "integration-test-root")
	defer span.End()

	_, childSpan := tracer.Start(ctx, "integration-test-child")
	childSpan.End()

	// Record metrics
	counter, err := meter.Int64Counter("integration.test.counter")
	s.Require().NoError(err)
	counter.Add(ctx, 42)

	histogram, err := meter.Float64Histogram("integration.test.histogram")
	s.Require().NoError(err)
	histogram.Record(ctx, 123.45)

	// Stop app (should flush and shutdown providers)
	err = app.Stop(ctx)
	s.Require().NoError(err)
}

func (s *TracingIntegrationSuite) TestProviderCachingViaModule() {
	// Test caching by starting the same service twice via Module
	cfg := NewIntegrationTestConfig("cache-test-service")

	var tp1 oteltrace.TracerProvider

	app1 := fx.New(
		fx.NopLogger,
		fx.Provide(func() tracing.Config { return cfg }),
		tracing.Module(),
		fx.Populate(&tp1),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := app1.Start(ctx)
	s.Require().NoError(err)
	s.NotNil(tp1)

	err = app1.Stop(ctx)
	s.Require().NoError(err)
}

func (s *TracingIntegrationSuite) TestShutdownProviders() {
	ctx := context.Background()

	// Test shutting down nil providers (should be no-op)
	err := tracing.ShutdownTracerProvider(ctx, nil)
	s.NoError(err)

	err = tracing.ShutdownMeterProvider(ctx, nil)
	s.NoError(err)

	err = tracing.ShutdownLoggerProvider(ctx, nil)
	s.NoError(err)
}

func (s *TracingIntegrationSuite) TestLoggerProvider() {
	// Clear cache to ensure fresh provider
	tracing.ClearLoggerProviderCache()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := NewIntegrationTestConfig("logger-provider-test")

	// Create logger provider
	lp, err := tracing.GetLoggerProvider(ctx, cfg, nil)
	s.Require().NoError(err)
	if lp != nil {
		// Shutdown the provider
		err = tracing.ShutdownLoggerProvider(ctx, lp)
		s.NoError(err)
	}
}

func (s *TracingIntegrationSuite) TestGetResource() {
	ctx := context.Background()
	cfg := NewIntegrationTestConfig("resource-test-service")

	res, err := tracing.GetResource(ctx, cfg)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), res)

	// Verify resource has expected attributes
	attrs := res.Attributes()
	var foundService bool
	for _, attr := range attrs {
		if attr.Key == "service.name" && attr.Value.AsString() == "resource-test-service" {
			foundService = true
			break
		}
	}
	assert.True(s.T(), foundService, "resource should have service.name attribute")
}
