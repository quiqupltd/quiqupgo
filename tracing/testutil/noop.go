// Package testutil provides testing utilities for the tracing module.
package testutil

import (
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"
)

var (
	// Singleton no-op providers to avoid creating multiple instances
	noopTracerProvider = tracenoop.NewTracerProvider()
	noopMeterProvider  = metricnoop.NewMeterProvider()
)

// NoopModule provides no-op OpenTelemetry providers for testing.
// Use this in tests where you don't need actual tracing.
//
// Example:
//
//	app := fx.New(
//	    tracingtest.NoopModule(),
//	    // ... other modules
//	)
func NoopModule() fx.Option {
	return fx.Module("tracing-test",
		fx.Provide(
			provideNoopTracerProvider,
			provideNoopTracer,
			provideNoopMeterProvider,
			provideNoopMeter,
		),
	)
}

func provideNoopTracerProvider() trace.TracerProvider {
	return noopTracerProvider
}

func provideNoopTracer(tp trace.TracerProvider) trace.Tracer {
	return tp.Tracer("test")
}

func provideNoopMeterProvider() metric.MeterProvider {
	return noopMeterProvider
}

func provideNoopMeter(mp metric.MeterProvider) metric.Meter {
	return mp.Meter("test")
}

// NoopConfig is a test configuration that disables OTLP export.
type NoopConfig struct {
	ServiceName     string
	EnvironmentName string
}

// NewNoopConfig creates a NoopConfig with test defaults.
func NewNoopConfig() *NoopConfig {
	return &NoopConfig{
		ServiceName:     "test-service",
		EnvironmentName: "test",
	}
}

func (c *NoopConfig) GetServiceName() string     { return c.ServiceName }
func (c *NoopConfig) GetEnvironmentName() string { return c.EnvironmentName }
func (c *NoopConfig) GetOTLPEndpoint() string    { return "" } // Disabled
func (c *NoopConfig) GetOTLPInsecure() bool      { return false }
func (c *NoopConfig) GetOTLPTLSCert() string     { return "" }
func (c *NoopConfig) GetOTLPTLSKey() string      { return "" }
func (c *NoopConfig) GetOTLPTLSCA() string       { return "" }
