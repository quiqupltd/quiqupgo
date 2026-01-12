package gormfx

import (
	"database/sql"
)

// Config is the interface that applications must implement to configure the GORM module.
// Applications can either implement this interface on their own config struct or use
// StandardConfig.
type Config interface {
	// GetDB returns the underlying *sql.DB connection.
	// This allows apps to manage the raw connection externally and pass it in.
	GetDB() *sql.DB

	// GetMaxOpenConns returns the maximum number of open connections.
	// Return 0 to use GORM's default.
	GetMaxOpenConns() int

	// GetMaxIdleConns returns the maximum number of idle connections.
	// Return 0 to use GORM's default.
	GetMaxIdleConns() int

	// GetEnableTracing returns whether OpenTelemetry tracing should be enabled.
	GetEnableTracing() bool
}

// StandardConfig is a standard implementation of Config that applications can use.
type StandardConfig struct {
	// DB is the underlying *sql.DB connection.
	// This must be provided by the application.
	DB *sql.DB

	// MaxOpenConns is the maximum number of open connections.
	// 0 means use GORM's default.
	MaxOpenConns int

	// MaxIdleConns is the maximum number of idle connections.
	// 0 means use GORM's default.
	MaxIdleConns int

	// EnableTracing enables OpenTelemetry tracing.
	// Defaults to true if not set.
	EnableTracing *bool
}

// GetDB returns the underlying *sql.DB connection.
func (c *StandardConfig) GetDB() *sql.DB {
	return c.DB
}

// GetMaxOpenConns returns the maximum number of open connections.
func (c *StandardConfig) GetMaxOpenConns() int {
	return c.MaxOpenConns
}

// GetMaxIdleConns returns the maximum number of idle connections.
func (c *StandardConfig) GetMaxIdleConns() int {
	return c.MaxIdleConns
}

// GetEnableTracing returns whether OpenTelemetry tracing should be enabled.
// Defaults to true if not explicitly set.
func (c *StandardConfig) GetEnableTracing() bool {
	if c.EnableTracing == nil {
		return true
	}
	return *c.EnableTracing
}

// Ensure StandardConfig implements Config.
var _ Config = (*StandardConfig)(nil)
