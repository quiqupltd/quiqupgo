// Package testutil provides testing utilities for the gormfx module.
package testutil

import (
	"database/sql"

	"github.com/quiqupltd/quiqupgo/gormfx"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NoopConfig is a test configuration for the gormfx module.
// It does not provide a real database connection.
type NoopConfig struct {
	db            *sql.DB
	maxOpenConns  int
	maxIdleConns  int
	enableTracing bool
}

// NewNoopConfig creates a NoopConfig with test defaults.
func NewNoopConfig() *NoopConfig {
	return &NoopConfig{
		db:            nil,
		enableTracing: false,
	}
}

func (c *NoopConfig) GetDB() *sql.DB         { return c.db }
func (c *NoopConfig) GetMaxOpenConns() int   { return c.maxOpenConns }
func (c *NoopConfig) GetMaxIdleConns() int   { return c.maxIdleConns }
func (c *NoopConfig) GetEnableTracing() bool { return c.enableTracing }

// Ensure NoopConfig implements Config.
var _ gormfx.Config = (*NoopConfig)(nil)

// NewTestDB creates an in-memory SQLite database for testing.
// This is useful for integration tests that need a real GORM database
// without connecting to a real PostgreSQL instance.
func NewTestDB() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

// TestModule returns an fx.Option that provides a test GORM database.
// It uses an in-memory SQLite database that is suitable for testing.
//
// Usage:
//
//	app := fx.New(
//	    gormfx_testutil.TestModule(),
//	    // ... other modules that depend on *gorm.DB
//	)
func TestModule() fx.Option {
	return fx.Module("gormfx-test",
		fx.Provide(func() (*gorm.DB, error) {
			return NewTestDB()
		}),
	)
}

// NoopTracerProviderModule returns an fx.Option that provides a no-op TracerProvider.
// Use this when testing with gormfx.Module() and tracing is not needed.
func NoopTracerProviderModule() fx.Option {
	return fx.Module("noop-tracer-provider",
		fx.Provide(func() trace.TracerProvider {
			return tracenoop.NewTracerProvider()
		}),
	)
}
