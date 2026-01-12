package logger_test

import (
	"context"
	"testing"
	"time"

	"github.com/quiqupltd/quiqupgo/logger"
	"github.com/quiqupltd/quiqupgo/logger/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func TestModule_Development(t *testing.T) {
	var (
		zapLogger *zap.Logger
		log       logger.Logger
	)

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() logger.Config {
			return &logger.StandardConfig{
				ServiceName: "test-service",
				Environment: "development",
			}
		}),
		logger.Module(),
		fx.Populate(&zapLogger, &log),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	require.NoError(t, err)

	assert.NotNil(t, zapLogger)
	assert.NotNil(t, log)

	// Test logging (shouldn't panic)
	log.Info("test message", "key", "value")
	log.Debug("debug message")
	log.Warn("warn message")
	log.Error("error message")

	err = app.Stop(ctx)
	require.NoError(t, err)
}

func TestModule_Production(t *testing.T) {
	var (
		zapLogger *zap.Logger
		log       logger.Logger
	)

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() logger.Config {
			return &logger.StandardConfig{
				ServiceName: "test-service",
				Environment: "production",
			}
		}),
		logger.Module(),
		fx.Populate(&zapLogger, &log),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	require.NoError(t, err)

	assert.NotNil(t, zapLogger)
	assert.NotNil(t, log)

	err = app.Stop(ctx)
	require.NoError(t, err)
}

func TestNoopModule(t *testing.T) {
	var (
		zapLogger *zap.Logger
		log       logger.Logger
	)

	app := fx.New(
		fx.NopLogger,
		testutil.NoopModule(),
		fx.Populate(&zapLogger, &log),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	require.NoError(t, err)

	assert.NotNil(t, zapLogger)
	assert.NotNil(t, log)

	// Should not panic
	log.Info("test message", "key", "value")

	err = app.Stop(ctx)
	require.NoError(t, err)
}

func TestBufferModule(t *testing.T) {
	logMod, buffer := testutil.BufferModule()

	var log logger.Logger

	app := fx.New(
		fx.NopLogger,
		logMod,
		fx.Populate(&log),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Start(ctx)
	require.NoError(t, err)

	// Log some messages
	log.Info("test info", "key", "value")
	log.Debug("test debug")
	log.Warn("test warn")
	log.Error("test error")

	// Check captured entries
	entries := buffer.GetEntries()
	require.Len(t, entries, 4)

	assert.Equal(t, "info", entries[0].Level)
	assert.Equal(t, "test info", entries[0].Message)
	assert.Equal(t, "value", entries[0].Fields["key"])

	assert.Equal(t, "debug", entries[1].Level)
	assert.Equal(t, "warn", entries[2].Level)
	assert.Equal(t, "error", entries[3].Level)

	// Test Clear
	buffer.Clear()
	assert.Equal(t, 0, buffer.Len())

	err = app.Stop(ctx)
	require.NoError(t, err)
}

func TestStandardConfig(t *testing.T) {
	cfg := &logger.StandardConfig{
		ServiceName: "test-service",
		Environment: "production",
	}

	assert.Equal(t, "test-service", cfg.GetServiceName())
	assert.Equal(t, "production", cfg.GetEnvironment())
	assert.False(t, cfg.IsDevelopment())

	cfg.Environment = "development"
	assert.True(t, cfg.IsDevelopment())

	cfg.Environment = "local"
	assert.True(t, cfg.IsDevelopment())

	cfg.Environment = "dev"
	assert.True(t, cfg.IsDevelopment())
}

func TestZapLogger_Interface(t *testing.T) {
	zapLog := zap.NewNop()
	log := logger.NewZapLogger(zapLog)

	// Test all interface methods don't panic
	log.Debug("debug", "key", "value")
	log.Info("info", "key", "value")
	log.Warn("warn", "key", "value")
	log.Error("error", "key", "value")

	log.Debugf("debug %s", "formatted")
	log.Infof("info %s", "formatted")
	log.Warnf("warn %s", "formatted")
	log.Errorf("error %s", "formatted")

	// Test With
	withLog := log.With("context", "value")
	assert.NotNil(t, withLog)
	withLog.Info("with context")

	// Test Unwrap
	assert.Equal(t, zapLog, log.Unwrap())
}
