// Package kafka provides an uber/fx module for Kafka messaging.
//
// It exports Producer and Consumer through dependency injection with
// OpenTelemetry tracing for message propagation.
//
// This module depends on:
//   - trace.Tracer (from tracing module)
//   - *zap.Logger (from logger module)
//
// Example usage:
//
//	fx.New(
//	    tracing.Module(),
//	    logger.Module(),
//	    fx.Provide(func() kafka.Config {
//	        return &kafka.StandardConfig{
//	            Brokers:       []string{"kafka:9092"},
//	            ConsumerGroup: "my-service",
//	        }
//	    }),
//	    kafka.Module(),
//	)
package kafka
