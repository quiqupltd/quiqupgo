// Package main demonstrates an API service using tracing, logger, and GORM modules.
//
// This example shows a typical HTTP API service with:
//   - OpenTelemetry tracing
//   - Structured logging with zap
//   - Database access with GORM
//   - HTTP middleware for request tracing
//
// Usage:
//
//	go run ./examples/api-service
package main

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/quiqupltd/quiqupgo/gormfx"
	"github.com/quiqupltd/quiqupgo/logger"
	"github.com/quiqupltd/quiqupgo/middleware"
	"github.com/quiqupltd/quiqupgo/tracing"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // SQLite driver for demo
)

func main() {
	fx.New(
		// Provide configurations
		fx.Provide(
			newTracingConfig,
			newLoggerConfig,
			newGormConfig,
		),

		// Include modules
		tracing.Module(),
		logger.Module(),
		gormfx.Module(),

		// Provide HTTP server
		fx.Provide(newEchoServer),

		// Start the server
		fx.Invoke(registerRoutes),
	).Run()
}

// newTracingConfig creates the tracing configuration.
func newTracingConfig() tracing.Config {
	return &tracing.StandardConfig{
		ServiceName:     "api-service",
		EnvironmentName: "development",
		OTLPEndpoint:    "", // Empty = disabled for demo
	}
}

// newLoggerConfig creates the logger configuration.
func newLoggerConfig() logger.Config {
	return &logger.StandardConfig{
		ServiceName: "api-service",
		Environment: "development",
	}
}

// newGormConfig creates the GORM configuration with an in-memory SQLite database.
func newGormConfig() (gormfx.Config, error) {
	// Open an in-memory SQLite database for demo
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}

	enableTracing := false // Disable OTEL tracing for SQLite
	return &gormfx.StandardConfig{
		DB:            db,
		EnableTracing: &enableTracing,
	}, nil
}

// newEchoServer creates a new Echo server with tracing middleware.
func newEchoServer(tp trace.TracerProvider) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Add tracing middleware
	e.Use(middleware.EchoTracing(tp, "api-service",
		middleware.WithSkipPaths("/health"),
	))

	return e
}

// User model for demo.
type User struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:100"`
}

// registerRoutes sets up the HTTP routes and starts the server.
func registerRoutes(lc fx.Lifecycle, e *echo.Echo, db *gorm.DB, log *zap.Logger) {
	// Auto-migrate the schema
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("failed to migrate", zap.Error(err))
	}

	// Seed some data
	db.Create(&User{Name: "Alice"})
	db.Create(&User{Name: "Bob"})

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// List users
	e.GET("/users", func(c echo.Context) error {
		var users []User
		if err := db.Find(&users).Error; err != nil {
			log.Error("failed to list users", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "database error"})
		}
		return c.JSON(http.StatusOK, users)
	})

	// Create user
	e.POST("/users", func(c echo.Context) error {
		var user User
		if err := c.Bind(&user); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := db.Create(&user).Error; err != nil {
			log.Error("failed to create user", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "database error"})
		}
		log.Info("user created", zap.Uint("id", user.ID), zap.String("name", user.Name))
		return c.JSON(http.StatusCreated, user)
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("starting HTTP server", zap.String("addr", ":8080"))
			go func() {
				if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
					log.Error("server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("stopping HTTP server")
			return e.Shutdown(ctx)
		},
	})
}
