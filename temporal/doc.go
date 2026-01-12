// Package temporal provides an uber/fx module for Temporal workflow client.
//
// It exports client.Client through dependency injection with OpenTelemetry tracing
// integration. The client automatically handles TLS for remote connections.
//
// This module depends on:
//   - *zap.Logger (from logger module)
//   - trace.Tracer (from tracing module)
//
// Example usage:
//
//	fx.New(
//	    tracing.Module(),
//	    logger.Module(),
//	    fx.Provide(func() temporal.Config {
//	        return &temporal.StandardConfig{
//	            HostPort:  "localhost:7233",
//	            Namespace: "default",
//	        }
//	    }),
//	    temporal.Module(),
//	)
package temporal
