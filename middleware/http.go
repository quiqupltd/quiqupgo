package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// HTTPTracing returns an http.Handler middleware that adds OpenTelemetry tracing.
//
// The middleware:
//   - Extracts trace context from incoming request headers
//   - Creates a span for each request with HTTP attributes
//   - Injects trace context into response headers
//   - Records errors and status codes
//
// Usage:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/api", handler)
//	tracedHandler := middleware.HTTPTracing(tracerProvider, "my-service")(mux)
//	http.ListenAndServe(":8080", tracedHandler)
func HTTPTracing(tp trace.TracerProvider, serviceName string, opts ...TracingOption) func(http.Handler) http.Handler {
	cfg := newTracingConfig(tp, serviceName, opts...)
	tracer := tp.Tracer("github.com/quiqupltd/quiqupgo/middleware")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Skip tracing for configured paths
			if cfg.skipPaths[path] {
				next.ServeHTTP(w, r)
				return
			}

			// Extract trace context from incoming request
			ctx := cfg.propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// Start span
			ctx, span := tracer.Start(ctx, spanName(r.Method, path),
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(httpAttributes(r)...),
			)
			defer span.End()

			// Inject trace context into response headers
			cfg.propagator.Inject(ctx, propagation.HeaderCarrier(w.Header()))

			// Wrap response writer to capture status code
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call next handler with updated context
			next.ServeHTTP(wrappedWriter, r.WithContext(ctx))

			// Record status code
			span.SetAttributes(httpStatusAttributes(wrappedWriter.statusCode)...)

			// Record error status
			if wrappedWriter.statusCode >= 400 {
				span.SetStatus(codes.Error, "HTTP error")
			}
		})
	}
}

// HTTPTracingHandler returns an http.Handler that wraps the provided handler with tracing.
// This is a convenience function for single handlers.
func HTTPTracingHandler(tp trace.TracerProvider, serviceName string, handler http.Handler, opts ...TracingOption) http.Handler {
	return HTTPTracing(tp, serviceName, opts...)(handler)
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures 200 status if WriteHeader wasn't called.
func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// Unwrap returns the original ResponseWriter.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}
