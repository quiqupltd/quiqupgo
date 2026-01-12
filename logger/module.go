package logger

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module returns an fx.Option that provides structured logging with zap.
//
// It provides:
//   - *zap.Logger (the underlying zap logger)
//   - Logger (the abstracted logger interface)
//
// It requires:
//   - logger.Config (must be provided by the application)
func Module(opts ...ModuleOption) fx.Option {
	options := defaultModuleOptions()
	for _, opt := range opts {
		opt(options)
	}

	return fx.Module("logger",
		fx.Supply(options),
		fx.Provide(
			provideZapLogger,
			provideLogger,
		),
		fx.Invoke(registerLifecycleHooks),
	)
}

// provideZapLogger creates the *zap.Logger.
func provideZapLogger(cfg Config, opts *moduleOptions) (*zap.Logger, error) {
	logger, err := NewLogger(cfg)
	if err != nil {
		return nil, err
	}

	// Replace the global logger
	zap.ReplaceGlobals(logger)

	return logger, nil
}

// provideLogger creates the Logger interface wrapper.
func provideLogger(zapLogger *zap.Logger) Logger {
	return NewZapLogger(zapLogger)
}

// registerLifecycleHooks registers shutdown hooks for graceful cleanup.
func registerLifecycleHooks(lc fx.Lifecycle, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			// Sync the logger to flush any buffered entries
			_ = logger.Sync()
			return nil
		},
	})
}

// moduleOptions holds the configurable options for the logger module.
type moduleOptions struct {
	// Currently no options, but kept for future extensibility
}

// defaultModuleOptions returns the default module options.
func defaultModuleOptions() *moduleOptions {
	return &moduleOptions{}
}

// ModuleOption is a functional option for configuring the logger module.
type ModuleOption func(*moduleOptions)
