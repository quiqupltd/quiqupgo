// Package temporal provides an uber/fx module for Temporal workflow client.
//
// It exports client.Client through dependency injection with OpenTelemetry tracing
// integration. The client automatically handles TLS for remote connections.
//
// This module depends on:
//   - *zap.Logger (from logger module)
//   - trace.Tracer (from tracing module)
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
package temporal
