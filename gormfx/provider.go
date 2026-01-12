package gormfx

import (
	"fmt"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDB creates a new GORM database connection with optional OpenTelemetry tracing.
// It wraps an existing *sql.DB connection and configures GORM with the otelgorm plugin.
func NewDB(cfg Config, tp trace.TracerProvider) (*gorm.DB, error) {
	sqlDB := cfg.GetDB()
	if sqlDB == nil {
		return nil, fmt.Errorf("sql.DB is required")
	}

	// Configure connection pool if specified
	if maxOpen := cfg.GetMaxOpenConns(); maxOpen > 0 {
		sqlDB.SetMaxOpenConns(maxOpen)
	}
	if maxIdle := cfg.GetMaxIdleConns(); maxIdle > 0 {
		sqlDB.SetMaxIdleConns(maxIdle)
	}

	// Create GORM DB using the existing connection
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open GORM connection: %w", err)
	}

	// Add OpenTelemetry tracing plugin if enabled
	if cfg.GetEnableTracing() && tp != nil {
		if err := db.Use(otelgorm.NewPlugin(
			otelgorm.WithTracerProvider(tp),
			otelgorm.WithDBName("postgres"),
		)); err != nil {
			return nil, fmt.Errorf("failed to add otelgorm plugin: %w", err)
		}
	}

	return db, nil
}
