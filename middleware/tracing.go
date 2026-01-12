package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// tracingConfig holds common configuration for tracing middleware.
type tracingConfig struct {
	tracerProvider trace.TracerProvider
	serviceName    string
	propagator     propagation.TextMapPropagator
	skipPaths      map[string]bool
}

// TracingOption is a functional option for configuring tracing middleware.
type TracingOption func(*tracingConfig)

// WithPropagator sets a custom propagator for trace context propagation.
func WithPropagator(propagator propagation.TextMapPropagator) TracingOption {
	return func(cfg *tracingConfig) {
		cfg.propagator = propagator
	}
}

// WithSkipPaths sets paths that should not be traced (e.g., health checks).
func WithSkipPaths(paths ...string) TracingOption {
	return func(cfg *tracingConfig) {
		for _, p := range paths {
			cfg.skipPaths[p] = true
		}
	}
}

// defaultPropagator returns the default propagator for trace context.
func defaultPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// newTracingConfig creates a new tracing config with defaults.
func newTracingConfig(tp trace.TracerProvider, serviceName string, opts ...TracingOption) *tracingConfig {
	cfg := &tracingConfig{
		tracerProvider: tp,
		serviceName:    serviceName,
		propagator:     defaultPropagator(),
		skipPaths:      make(map[string]bool),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// httpAttributes returns common HTTP attributes for a span.
func httpAttributes(r *http.Request) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		semconv.HTTPMethod(r.Method),
		semconv.HTTPURL(r.URL.String()),
		semconv.HTTPScheme(r.URL.Scheme),
		semconv.NetHostName(r.Host),
	}

	if r.URL.Path != "" {
		attrs = append(attrs, semconv.HTTPTarget(r.URL.Path))
	}

	if ua := r.UserAgent(); ua != "" {
		attrs = append(attrs, attribute.String("http.user_agent", ua))
	}

	return attrs
}

// httpStatusAttributes returns HTTP status code attributes.
func httpStatusAttributes(statusCode int) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.HTTPStatusCode(statusCode),
	}
}

// spanName generates a span name from the HTTP method and path.
func spanName(method, path string) string {
	return method + " " + path
}
