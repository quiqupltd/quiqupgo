package logger

// Logger defines a standard logging interface.
// This allows for easy mocking in tests and abstraction from the underlying implementation.
type Logger interface {
	// Debug logs a message at debug level.
	Debug(msg string, keyvals ...interface{})

	// Info logs a message at info level.
	Info(msg string, keyvals ...interface{})

	// Warn logs a message at warn level.
	Warn(msg string, keyvals ...interface{})

	// Error logs a message at error level.
	Error(msg string, keyvals ...interface{})

	// With returns a new Logger with the given key-value pairs added to the context.
	With(keyvals ...interface{}) Logger

	// Debugf logs a formatted message at debug level.
	Debugf(format string, args ...interface{})

	// Infof logs a formatted message at info level.
	Infof(format string, args ...interface{})

	// Warnf logs a formatted message at warn level.
	Warnf(format string, args ...interface{})

	// Errorf logs a formatted message at error level.
	Errorf(format string, args ...interface{})
}
