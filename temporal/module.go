package temporal

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module returns an fx.Option that provides a Temporal workflow client.
//
// It provides:
//   - client.Client (Temporal client)
//
// It requires:
//   - temporal.Config (must be provided by the application)
//   - *zap.Logger (from logger module)
//   - trace.Tracer (from tracing module)
func Module(opts ...ModuleOption) fx.Option {
	options := defaultModuleOptions()
	for _, opt := range opts {
		opt(options)
	}

	fxOpts := []fx.Option{
		fx.Supply(options),
		fx.Provide(provideTemporalClient),
		fx.Invoke(registerLifecycleHooks),
	}

	// Optionally provide worker interceptors for tracing
	if options.provideWorkerInterceptors {
		fxOpts = append(fxOpts, fx.Provide(provideWorkerInterceptors))
	}

	return fx.Module("temporal", fxOpts...)
}

// provideTemporalClient creates the Temporal client.
func provideTemporalClient(
	lc fx.Lifecycle,
	cfg Config,
	logger *zap.Logger,
	tracer trace.Tracer,
	opts *moduleOptions,
) (client.Client, error) {
	ctx := context.Background()
	return NewClient(ctx, cfg, logger, tracer)
}

// registerLifecycleHooks registers shutdown hooks for graceful cleanup.
func registerLifecycleHooks(lc fx.Lifecycle, c client.Client) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			c.Close()
			return nil
		},
	})
}

// moduleOptions holds the configurable options for the temporal module.
type moduleOptions struct {
	// provideWorkerInterceptors enables fx provision of worker interceptors.
	provideWorkerInterceptors bool
}

// defaultModuleOptions returns the default module options.
func defaultModuleOptions() *moduleOptions {
	return &moduleOptions{}
}

// ModuleOption is a functional option for configuring the temporal module.
type ModuleOption func(*moduleOptions)
