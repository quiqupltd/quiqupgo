package temporal

import (
	"fmt"

	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

// WorkerInterceptors returns OpenTelemetry tracing interceptors for Temporal workers.
//
// Use this function when creating workers to enable tracing for workflow and activity
// execution. The returned interceptors should be added to worker.Options.Interceptors.
//
// Example:
//
//	interceptors, err := temporal.WorkerInterceptors()
//	if err != nil {
//	    return err
//	}
//	w := worker.New(client, taskQueue, worker.Options{
//	    Interceptors: interceptors,
//	})
func WorkerInterceptors() ([]interceptor.WorkerInterceptor, error) {
	tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create tracing interceptor: %w", err)
	}
	return []interceptor.WorkerInterceptor{tracingInterceptor}, nil
}

// WorkerInterceptorSlice is a slice of WorkerInterceptors for fx dependency injection.
type WorkerInterceptorSlice []interceptor.WorkerInterceptor

// provideWorkerInterceptors creates worker interceptors for fx injection.
func provideWorkerInterceptors() (WorkerInterceptorSlice, error) {
	interceptors, err := WorkerInterceptors()
	if err != nil {
		return nil, err
	}
	return WorkerInterceptorSlice(interceptors), nil
}

// WithWorkerInterceptors is a module option that also provides worker interceptors
// via fx dependency injection.
//
// When this option is enabled, consumers can inject WorkerInterceptorSlice:
//
//	func NewWorker(
//	    client client.Client,
//	    interceptors temporal.WorkerInterceptorSlice,
//	) worker.Worker {
//	    return worker.New(client, "task-queue", worker.Options{
//	        Interceptors: interceptors,
//	    })
//	}
func WithWorkerInterceptors() ModuleOption {
	return func(o *moduleOptions) {
		o.provideWorkerInterceptors = true
	}
}

// WorkerInterceptorsModule returns an fx.Option that provides worker interceptors
// for tracing workflow and activity execution.
//
// This is a standalone module that can be used independently of the main temporal.Module()
// when you only need worker interceptors without a client.
//
// It provides:
//   - WorkerInterceptorSlice ([]interceptor.WorkerInterceptor)
//
// Example:
//
//	fx.New(
//	    temporal.WorkerInterceptorsModule(),
//	    fx.Invoke(func(interceptors temporal.WorkerInterceptorSlice) {
//	        w := worker.New(client, "task-queue", worker.Options{
//	            Interceptors: interceptors,
//	        })
//	    }),
//	)
func WorkerInterceptorsModule() fx.Option {
	return fx.Module("temporal-worker-interceptors",
		fx.Provide(provideWorkerInterceptors),
	)
}

// ApplyWorkerInterceptors is a convenience function that applies OpenTelemetry
// tracing interceptors to existing worker.Options.
//
// Example:
//
//	opts := worker.Options{
//	    MaxConcurrentActivityExecutionSize: 100,
//	}
//	if err := temporal.ApplyWorkerInterceptors(&opts); err != nil {
//	    return err
//	}
//	w := worker.New(client, taskQueue, opts)
func ApplyWorkerInterceptors(opts *worker.Options) error {
	interceptors, err := WorkerInterceptors()
	if err != nil {
		return err
	}
	opts.Interceptors = append(opts.Interceptors, interceptors...)
	return nil
}
