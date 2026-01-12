package temporal

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.uber.org/zap"
)

// ZapLoggerAdapter adapts a *zap.Logger to Temporal's log.Logger interface.
type ZapLoggerAdapter struct {
	logger *zap.Logger
}

// NewZapLoggerAdapter creates a new ZapLoggerAdapter.
func NewZapLoggerAdapter(logger *zap.Logger) *ZapLoggerAdapter {
	return &ZapLoggerAdapter{logger: logger}
}

// Debug logs at debug level.
func (a *ZapLoggerAdapter) Debug(msg string, keyvals ...interface{}) {
	a.logger.Debug(msg, toZapFields(keyvals...)...)
}

// Info logs at info level.
func (a *ZapLoggerAdapter) Info(msg string, keyvals ...interface{}) {
	a.logger.Info(msg, toZapFields(keyvals...)...)
}

// Warn logs at warn level.
func (a *ZapLoggerAdapter) Warn(msg string, keyvals ...interface{}) {
	a.logger.Warn(msg, toZapFields(keyvals...)...)
}

// Error logs at error level.
func (a *ZapLoggerAdapter) Error(msg string, keyvals ...interface{}) {
	a.logger.Error(msg, toZapFields(keyvals...)...)
}

// With returns a new logger with the given key-value pairs added.
func (a *ZapLoggerAdapter) With(keyvals ...interface{}) log.Logger {
	return &ZapLoggerAdapter{
		logger: a.logger.With(toZapFields(keyvals...)...),
	}
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

// Ensure ZapLoggerAdapter implements log.Logger.
var _ log.Logger = (*ZapLoggerAdapter)(nil)
