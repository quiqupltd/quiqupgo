package gormfx

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

// Module returns an fx.Option that provides a GORM database connection.
//
// It provides:
//   - *gorm.DB (GORM database connection with optional OTEL tracing)
//
// It requires:
//   - gormfx.Config (must be provided by the application)
//   - trace.TracerProvider (optional, from tracing module - for OTEL tracing)
func Module(opts ...ModuleOption) fx.Option {
	options := defaultModuleOptions()
	for _, opt := range opts {
		opt(options)
	}

	return fx.Module("gormfx",
		fx.Supply(options),
		fx.Provide(provideGormDB),
		fx.Invoke(registerLifecycleHooks),
	)
}

// provideGormDB creates the GORM database connection.
func provideGormDB(cfg Config, tp trace.TracerProvider, opts *moduleOptions) (*gorm.DB, error) {
	return NewDB(cfg, tp)
}

// registerLifecycleHooks registers shutdown hooks for graceful cleanup.
func registerLifecycleHooks(lc fx.Lifecycle, db *gorm.DB) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})
}

// moduleOptions holds the configurable options for the gormfx module.
type moduleOptions struct {
	// Currently no options, but kept for future extensibility
}

// defaultModuleOptions returns the default module options.
func defaultModuleOptions() *moduleOptions {
	return &moduleOptions{}
}

// ModuleOption is a functional option for configuring the gormfx module.
type ModuleOption func(*moduleOptions)
