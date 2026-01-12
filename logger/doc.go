// Package logger provides an uber/fx module for structured logging with zap.
//
// It exports *zap.Logger and a Logger interface through dependency injection.
// Configure it by providing an implementation of the Config interface.
//
// In development environment, logs are formatted with colors and human-readable output.
// In production, logs are JSON formatted for machine processing.
//
// Example usage:
//
//	fx.New(
//	    fx.Provide(func() logger.Config {
//	        return &logger.StandardConfig{
//	            ServiceName: "my-service",
//	            Environment: "development",
//	        }
//	    }),
//	    logger.Module(),
//	)
package logger
