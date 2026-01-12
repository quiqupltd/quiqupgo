package middleware

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// EchoTracing returns an Echo middleware that adds OpenTelemetry tracing.
//
// The middleware:
//   - Extracts trace context from incoming request headers
//   - Creates a span for each request with HTTP attributes
//   - Injects trace context into response headers
//   - Records errors and status codes
//
// Usage:
//
//	e := echo.New()
//	e.Use(middleware.EchoTracing(tracerProvider, "my-service"))
func EchoTracing(tp trace.TracerProvider, serviceName string, opts ...TracingOption) echo.MiddlewareFunc {
	cfg := newTracingConfig(tp, serviceName, opts...)
	tracer := tp.Tracer("github.com/quiqupltd/quiqupgo/middleware")

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			path := c.Path()

			// Skip tracing for configured paths
			if cfg.skipPaths[path] {
				return next(c)
			}

			// Extract trace context from incoming request
			ctx := cfg.propagator.Extract(req.Context(), propagation.HeaderCarrier(req.Header))

			// Start span
			ctx, span := tracer.Start(ctx, spanName(req.Method, path),
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(httpAttributes(req)...),
			)
			defer span.End()

			// Store span in context
			c.SetRequest(req.WithContext(ctx))

			// Inject trace context into response headers
			cfg.propagator.Inject(ctx, propagation.HeaderCarrier(c.Response().Header()))

			// Call next handler
			err := next(c)

			// Record status code
			statusCode := c.Response().Status
			span.SetAttributes(httpStatusAttributes(statusCode)...)

			// Record error if present
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			} else if statusCode >= 400 {
				span.SetStatus(codes.Error, "HTTP error")
			}

			return err
		}
	}
}

// EchoTracingWithConfig returns an Echo middleware with the provided configuration.
// This is an alias for EchoTracing with options pre-applied.
func EchoTracingWithConfig(tp trace.TracerProvider, serviceName string, skipPaths []string) echo.MiddlewareFunc {
	return EchoTracing(tp, serviceName, WithSkipPaths(skipPaths...))
}
