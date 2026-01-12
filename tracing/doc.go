// Package tracing provides an uber/fx module for OpenTelemetry tracing and metrics.
//
// It exports TracerProvider, Tracer, MeterProvider, and Meter through dependency injection.
// Configure it by providing an implementation of the Config interface.
//
// Example usage:
//
//	fx.New(
//	    fx.Provide(func() tracing.Config {
//	        return &tracing.StandardConfig{
//	            ServiceName:     "my-service",
//	            EnvironmentName: "production",
//	            OTLPEndpoint:    "otel-collector:4318",
//	        }
//	    }),
//	    tracing.Module(),
//	)
package tracing
