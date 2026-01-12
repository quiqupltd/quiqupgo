package testutil

import (
	"fmt"
	"sync"

	"github.com/quiqupltd/quiqupgo/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// LogEntry represents a captured log entry for assertions.
type LogEntry struct {
	Level   string
	Message string
	Fields  map[string]interface{}
}

// BufferLogger captures log output for test assertions.
type BufferLogger struct {
	mu       sync.Mutex
	entries  []LogEntry
	observed *observer.ObservedLogs
	zapCore  zapcore.Core
}

// NewBufferLogger creates a new BufferLogger.
func NewBufferLogger() *BufferLogger {
	core, observed := observer.New(zapcore.DebugLevel)
	return &BufferLogger{
		observed: observed,
		zapCore:  core,
	}
}

// Debug logs a message at debug level.
func (l *BufferLogger) Debug(msg string, keyvals ...interface{}) {
	l.log("debug", msg, keyvals...)
}

// Info logs a message at info level.
func (l *BufferLogger) Info(msg string, keyvals ...interface{}) {
	l.log("info", msg, keyvals...)
}

// Warn logs a message at warn level.
func (l *BufferLogger) Warn(msg string, keyvals ...interface{}) {
	l.log("warn", msg, keyvals...)
}

// Error logs a message at error level.
func (l *BufferLogger) Error(msg string, keyvals ...interface{}) {
	l.log("error", msg, keyvals...)
}

// With returns a new Logger with the given key-value pairs added to the context.
func (l *BufferLogger) With(keyvals ...interface{}) logger.Logger {
	// For simplicity, return the same logger (fields are captured in log calls)
	return l
}

// Debugf logs a formatted message at debug level.
func (l *BufferLogger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

// Infof logs a formatted message at info level.
func (l *BufferLogger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted message at warn level.
func (l *BufferLogger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted message at error level.
func (l *BufferLogger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *BufferLogger) log(level, msg string, keyvals ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	fields := make(map[string]interface{})
	for i := 0; i < len(keyvals)-1; i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keyvals[i])
		}
		fields[key] = keyvals[i+1]
	}

	l.entries = append(l.entries, LogEntry{
		Level:   level,
		Message: msg,
		Fields:  fields,
	})
}

// GetEntries returns all captured log entries.
func (l *BufferLogger) GetEntries() []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]LogEntry(nil), l.entries...)
}

// Clear removes all captured log entries.
func (l *BufferLogger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = nil
}

// Len returns the number of captured log entries.
func (l *BufferLogger) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}

// ZapLogger returns a *zap.Logger that writes to this buffer.
func (l *BufferLogger) ZapLogger() *zap.Logger {
	return zap.New(l.zapCore)
}

// ObservedLogs returns the observed logs from the underlying zap observer.
// Use this for more detailed zap-level assertions.
func (l *BufferLogger) ObservedLogs() *observer.ObservedLogs {
	return l.observed
}

// BufferModule provides a buffer logger for testing with assertions.
// It returns both the fx.Option and the BufferLogger for accessing captured logs.
//
// Example:
//
//	logMod, buffer := loggertest.BufferModule()
//	app := fx.New(
//	    logMod,
//	    // ... other modules
//	)
//	// After running code...
//	entries := buffer.GetEntries()
//	assert.Len(t, entries, 1)
func BufferModule() (fx.Option, *BufferLogger) {
	buffer := NewBufferLogger()

	module := fx.Module("logger-test-buffer",
		fx.Provide(
			func() *zap.Logger {
				return buffer.ZapLogger()
			},
			func() logger.Logger {
				return buffer
			},
		),
	)

	return module, buffer
}

// Ensure BufferLogger implements Logger.
var _ logger.Logger = (*BufferLogger)(nil)
