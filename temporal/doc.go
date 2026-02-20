// Package temporal provides an uber/fx module for Temporal workflow client.
//
// It exports client.Client through dependency injection with OpenTelemetry tracing
// and metrics integration. The client automatically handles TLS for remote connections.
//
// This module depends on:
//   - *zap.Logger (from logger module)
//   - trace.Tracer (from tracing module)
//   - metric.Meter (from tracing module)
//
// # Basic Usage
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
//
// # Metrics
//
// OpenTelemetry metrics are enabled automatically when a metric.Meter is available
// in the fx container (provided by the tracing module). No additional configuration
// is needed — the Temporal SDK emits workflow latency, task queue depth, and other
// metrics through the OTel MetricsHandler.
//
// # Lazy Client
//
// By default, Module() uses client.Dial which eagerly connects to the Temporal server.
// Use WithLazyClient() to defer connection until the first RPC call:
//
//	temporal.Module(temporal.WithLazyClient())
//
// # Worker Tracing
//
// The module provides OpenTelemetry tracing for the client automatically. For workers,
// use the worker tracing helpers to enable tracing of workflow and activity execution.
//
// Using the helper function directly:
//
//	interceptors, err := temporal.WorkerInterceptors()
//	if err != nil {
//	    return err
//	}
//	w := worker.New(client, "task-queue", worker.Options{
//	    Interceptors: interceptors,
//	})
//
// Or apply to existing options:
//
//	opts := worker.Options{
//	    MaxConcurrentActivityExecutionSize: 100,
//	}
//	temporal.ApplyWorkerInterceptors(&opts)
//	w := worker.New(client, taskQueue, opts)
//
// Or via fx dependency injection:
//
//	fx.New(
//	    temporal.Module(temporal.WithWorkerInterceptors()),
//	    fx.Invoke(func(client client.Client, interceptors temporal.WorkerInterceptorSlice) {
//	        w := worker.New(client, "task-queue", worker.Options{
//	            Interceptors: interceptors,
//	        })
//	    }),
//	)
//
// For workers without a client (e.g., separate worker services), use the standalone module:
//
//	fx.New(
//	    temporal.WorkerInterceptorsModule(),
//	    fx.Invoke(func(interceptors temporal.WorkerInterceptorSlice) {
//	        // Use with externally provided client
//	    }),
//	)
//
// # Local Module
//
// LocalModule() provides a second client.Client tagged with name:"temporallocal" for
// connecting to an in-cluster (e.g., GKE) Temporal server. It uses NewLazyClient by
// default, requires no TLS, and includes OTel tracing + metrics.
//
//	fx.New(
//	    tracing.Module(),
//	    logger.Module(),
//	    fx.Provide(func() temporal.LocalConfig {
//	        return &temporal.StandardLocalConfig{
//	            HostPort:  "temporal.local:7233",
//	            Namespace: "default",
//	        }
//	    }),
//	    temporal.LocalModule(),
//	)
//
// # Dual-Client Pattern (Cloud + Local)
//
// Run both cloud and local Temporal clients side by side for gradual migration:
//
//	fx.New(
//	    tracing.Module(),
//	    logger.Module(),
//	    fx.Provide(func() temporal.Config {
//	        return &temporal.StandardConfig{
//	            HostPort: "cloud.temporal.io:7233",
//	            TLSCert:  certPEM,
//	            TLSKey:   keyPEM,
//	        }
//	    }),
//	    fx.Provide(func() temporal.LocalConfig {
//	        return &temporal.StandardLocalConfig{
//	            HostPort: "temporal.local:7233",
//	        }
//	    }),
//	    temporal.Module(),
//	    temporal.LocalModule(),
//	)
//
// Inject both clients using fx tags:
//
//	type WorkerService struct {
//	    CloudClient client.Client
//	    LocalClient client.Client `name:"temporallocal"`
//	}
package temporal
