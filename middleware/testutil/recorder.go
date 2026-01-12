// Package testutil provides testing utilities for the middleware package.
package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// SpanRecorder records spans for testing.
type SpanRecorder struct {
	exporter *tracetest.InMemoryExporter
	tp       *sdktrace.TracerProvider
}

// NewSpanRecorder creates a new SpanRecorder for testing.
func NewSpanRecorder() *SpanRecorder {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)

	return &SpanRecorder{
		exporter: exporter,
		tp:       tp,
	}
}

// TracerProvider returns the TracerProvider for use with middleware.
func (r *SpanRecorder) TracerProvider() trace.TracerProvider {
	return r.tp
}

// Spans returns all recorded spans.
func (r *SpanRecorder) Spans() tracetest.SpanStubs {
	return r.exporter.GetSpans()
}

// Reset clears all recorded spans.
func (r *SpanRecorder) Reset() {
	r.exporter.Reset()
}

// Shutdown shuts down the tracer provider.
func (r *SpanRecorder) Shutdown() error {
	return r.tp.Shutdown(context.Background())
}

// FindSpanByName finds a span by its name.
func (r *SpanRecorder) FindSpanByName(name string) *tracetest.SpanStub {
	for i := range r.Spans() {
		if r.Spans()[i].Name == name {
			return &r.Spans()[i]
		}
	}
	return nil
}

// SpanHasAttribute checks if a span has an attribute with the given key.
func SpanHasAttribute(span tracetest.SpanStub, key string) bool {
	for _, attr := range span.Attributes {
		if string(attr.Key) == key {
			return true
		}
	}
	return false
}

// GetSpanAttribute gets an attribute value from a span.
func GetSpanAttribute(span tracetest.SpanStub, key string) (attribute.Value, bool) {
	for _, attr := range span.Attributes {
		if string(attr.Key) == key {
			return attr.Value, true
		}
	}
	return attribute.Value{}, false
}

// RequestRecorder records HTTP requests for testing.
type RequestRecorder struct {
	mu       sync.RWMutex
	requests []*RecordedRequest
}

// RecordedRequest represents a recorded HTTP request.
type RecordedRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
}

// NewRequestRecorder creates a new RequestRecorder.
func NewRequestRecorder() *RequestRecorder {
	return &RequestRecorder{
		requests: make([]*RecordedRequest, 0),
	}
}

// Handler returns an http.Handler that records requests.
func (r *RequestRecorder) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.mu.Lock()
		r.requests = append(r.requests, &RecordedRequest{
			Method:  req.Method,
			Path:    req.URL.Path,
			Headers: req.Header.Clone(),
		})
		r.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})
}

// Requests returns all recorded requests.
func (r *RequestRecorder) Requests() []*RecordedRequest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]*RecordedRequest(nil), r.requests...)
}

// Reset clears all recorded requests.
func (r *RequestRecorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = make([]*RecordedRequest, 0)
}

// TestServer is a wrapper around httptest.Server with span recording.
type TestServer struct {
	Server       *httptest.Server
	SpanRecorder *SpanRecorder
}

// NewTestServer creates a new test server with span recording.
func NewTestServer(handler http.Handler) *TestServer {
	recorder := NewSpanRecorder()
	server := httptest.NewServer(handler)

	return &TestServer{
		Server:       server,
		SpanRecorder: recorder,
	}
}

// Close closes the test server and shuts down the span recorder.
func (ts *TestServer) Close() {
	ts.Server.Close()
	ts.SpanRecorder.Shutdown()
}
