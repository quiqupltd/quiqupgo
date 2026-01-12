package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger wraps a *zap.Logger and implements the Logger interface.
type ZapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// NewZapLogger creates a new ZapLogger from a *zap.Logger.
func NewZapLogger(logger *zap.Logger) *ZapLogger {
	return &ZapLogger{
		logger: logger,
		sugar:  logger.Sugar(),
	}
}

// Debug logs a message at debug level.
func (l *ZapLogger) Debug(msg string, keyvals ...interface{}) {
	l.sugar.Debugw(msg, keyvals...)
}

// Info logs a message at info level.
func (l *ZapLogger) Info(msg string, keyvals ...interface{}) {
	l.sugar.Infow(msg, keyvals...)
}

// Warn logs a message at warn level.
func (l *ZapLogger) Warn(msg string, keyvals ...interface{}) {
	l.sugar.Warnw(msg, keyvals...)
}

// Error logs a message at error level.
func (l *ZapLogger) Error(msg string, keyvals ...interface{}) {
	l.sugar.Errorw(msg, keyvals...)
}

// With returns a new Logger with the given key-value pairs added to the context.
func (l *ZapLogger) With(keyvals ...interface{}) Logger {
	return &ZapLogger{
		logger: l.logger.With(toZapFields(keyvals...)...),
		sugar:  l.sugar.With(keyvals...),
	}
}

// Debugf logs a formatted message at debug level.
func (l *ZapLogger) Debugf(format string, args ...interface{}) {
	l.sugar.Debugf(format, args...)
}

// Infof logs a formatted message at info level.
func (l *ZapLogger) Infof(format string, args ...interface{}) {
	l.sugar.Infof(format, args...)
}

// Warnf logs a formatted message at warn level.
func (l *ZapLogger) Warnf(format string, args ...interface{}) {
	l.sugar.Warnf(format, args...)
}

// Errorf logs a formatted message at error level.
func (l *ZapLogger) Errorf(format string, args ...interface{}) {
	l.sugar.Errorf(format, args...)
}

// Unwrap returns the underlying *zap.Logger.
func (l *ZapLogger) Unwrap() *zap.Logger {
	return l.logger
}

// toZapFields converts key-value pairs to zap.Fields.
func toZapFields(keyvals ...interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(keyvals)/2)
	for i := 0; i < len(keyvals)-1; i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keyvals[i])
		}
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}
	return fields
}

// NewLogger creates a new *zap.Logger based on the configuration.
// In development mode, it uses a human-readable console format.
// In production mode, it uses JSON structured logging.
func NewLogger(cfg Config) (*zap.Logger, error) {
	var zapCfg zap.Config

	env := cfg.GetEnvironment()
	isDev := env == "development" || env == "local" || env == "dev"

	if isDev {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapCfg = zap.NewProductionConfig()
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := zapCfg.Build(
		zap.AddCallerSkip(1), // Skip wrapper functions
		zap.Fields(
			zap.String("service", cfg.GetServiceName()),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return logger, nil
}

// Ensure ZapLogger implements Logger.
var _ Logger = (*ZapLogger)(nil)
