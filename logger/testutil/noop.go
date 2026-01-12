// Package testutil provides testing utilities for the logger module.
package testutil

import (
	"github.com/quiqupltd/quiqupgo/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// NoopModule provides a no-op logger for testing.
// Use this in tests where you don't need actual logging output.
//
// Example:
//
//	app := fx.New(
//	    loggertest.NoopModule(),
//	    // ... other modules
//	)
func NoopModule() fx.Option {
	return fx.Module("logger-test-noop",
		fx.Provide(
			provideNoopZapLogger,
			provideNoopLogger,
		),
	)
}

func provideNoopZapLogger() *zap.Logger {
	return zap.NewNop()
}

func provideNoopLogger(zapLogger *zap.Logger) logger.Logger {
	return logger.NewZapLogger(zapLogger)
}

// NoopConfig is a test configuration for the logger module.
type NoopConfig struct {
	ServiceName string
	Environment string
}

// NewNoopConfig creates a NoopConfig with test defaults.
func NewNoopConfig() *NoopConfig {
	return &NoopConfig{
		ServiceName: "test-service",
		Environment: "test",
	}
}

func (c *NoopConfig) GetServiceName() string { return c.ServiceName }
func (c *NoopConfig) GetEnvironment() string { return c.Environment }
