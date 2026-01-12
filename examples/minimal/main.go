// Package main demonstrates a minimal application using only the logger module.
//
// This is the simplest possible setup - just structured logging without
// any additional infrastructure.
//
// Usage:
//
//	go run ./examples/minimal
package main

import (
	"context"

	"github.com/quiqupltd/quiqupgo/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		// Provide logger configuration
		fx.Provide(newLoggerConfig),

		// Include the logger module
		logger.Module(),

		// Run the application
		fx.Invoke(run),
	).Run()
}

// newLoggerConfig creates the logger configuration.
func newLoggerConfig() logger.Config {
	return &logger.StandardConfig{
		ServiceName: "minimal-example",
		Environment: "development",
	}
}

// run is the main application entry point.
func run(lc fx.Lifecycle, log *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("application started",
				zap.String("example", "minimal"),
			)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("application stopped")
			return nil
		},
	})
}
