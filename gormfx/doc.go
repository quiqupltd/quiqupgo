// Package gormfx provides an uber/fx module for GORM database connections.
//
// The package is named gormfx to avoid import conflicts with gorm.io/gorm.
// It exports *gorm.DB through dependency injection with OpenTelemetry tracing
// enabled via the otelgorm plugin.
//
// This module depends on:
//   - trace.TracerProvider (from tracing module)
//
// Example usage:
//
//	fx.New(
//	    tracing.Module(),
//	    fx.Provide(func(db *sql.DB) gormfx.Config {
//	        return &gormfx.StandardConfig{DB: db}
//	    }),
//	    gormfx.Module(),
//	)
package gormfx
