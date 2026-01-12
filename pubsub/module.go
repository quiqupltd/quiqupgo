package pubsub

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module returns an fx.Option that provides Kafka producer and consumer.
//
// It provides:
//   - pubsub.Producer (Kafka producer with optional OTEL tracing)
//   - pubsub.Consumer (Kafka consumer with optional OTEL tracing)
//
// It requires:
//   - pubsub.Config (must be provided by the application)
//   - trace.Tracer (from tracing module)
//   - *zap.Logger (from logger module)
func Module(opts ...ModuleOption) fx.Option {
	options := defaultModuleOptions()
	for _, opt := range opts {
		opt(options)
	}

	return fx.Module("pubsub",
		fx.Supply(options),
		fx.Provide(
			provideProducer,
			provideConsumer,
		),
		fx.Invoke(registerLifecycleHooks),
	)
}

// provideProducer creates a Kafka producer.
func provideProducer(cfg Config, tracer trace.Tracer, logger *zap.Logger) (Producer, error) {
	return NewProducer(cfg, tracer, logger.Named("pubsub.producer"))
}

// provideConsumer creates a Kafka consumer.
func provideConsumer(cfg Config, tracer trace.Tracer, logger *zap.Logger) (Consumer, error) {
	return NewConsumer(cfg, tracer, logger.Named("pubsub.consumer"))
}

// registerLifecycleHooks registers shutdown hooks for graceful cleanup.
func registerLifecycleHooks(lc fx.Lifecycle, producer Producer, consumer Consumer) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := producer.Close(); err != nil {
				return err
			}
			return consumer.Close()
		},
	})
}

// moduleOptions holds the configurable options for the pubsub module.
type moduleOptions struct {
	// Currently no options, but kept for future extensibility
}

// defaultModuleOptions returns the default module options.
func defaultModuleOptions() *moduleOptions {
	return &moduleOptions{}
}

// ModuleOption is a functional option for configuring the pubsub module.
type ModuleOption func(*moduleOptions)
