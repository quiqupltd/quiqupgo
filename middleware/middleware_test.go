package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/quiqupltd/quiqupgo/middleware"
	"github.com/quiqupltd/quiqupgo/middleware/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestEchoTracing(t *testing.T) {
	recorder := testutil.NewSpanRecorder()
	defer recorder.Shutdown()

	e := echo.New()
	e.Use(middleware.EchoTracing(recorder.TracerProvider(), "test-service"))

	e.GET("/api/users", func(c echo.Context) error {
		return c.String(http.StatusOK, "users")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Check span was created
	spans := recorder.Spans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "GET /api/users", span.Name)
	assert.Equal(t, trace.SpanKindServer, span.SpanKind)

	// Check HTTP attributes
	assert.True(t, testutil.SpanHasAttribute(span, "http.method"))
	assert.True(t, testutil.SpanHasAttribute(span, "http.status_code"))
}

func TestEchoTracing_WithSkipPaths(t *testing.T) {
	recorder := testutil.NewSpanRecorder()
	defer recorder.Shutdown()

	e := echo.New()
	e.Use(middleware.EchoTracing(recorder.TracerProvider(), "test-service",
		middleware.WithSkipPaths("/health", "/ready"),
	))

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.GET("/api/users", func(c echo.Context) error {
		return c.String(http.StatusOK, "users")
	})

	// Request to skipped path
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Request to traced path
	req = httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Only the /api/users request should be traced
	spans := recorder.Spans()
	require.Len(t, spans, 1)
	assert.Equal(t, "GET /api/users", spans[0].Name)
}

func TestEchoTracing_Error(t *testing.T) {
	recorder := testutil.NewSpanRecorder()
	defer recorder.Shutdown()

	e := echo.New()
	e.Use(middleware.EchoTracing(recorder.TracerProvider(), "test-service"))

	e.GET("/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	})

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	spans := recorder.Spans()
	require.Len(t, spans, 1)

	span := spans[0]
	// Check that error was recorded
	assert.True(t, span.Status.Code != 0 || len(span.Events) > 0)
}

func TestHTTPTracing(t *testing.T) {
	recorder := testutil.NewSpanRecorder()
	defer recorder.Shutdown()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	})

	traced := middleware.HTTPTracing(recorder.TracerProvider(), "test-service")(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	rec := httptest.NewRecorder()

	traced.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "hello", rec.Body.String())

	// Check span was created
	spans := recorder.Spans()
	require.Len(t, spans, 1)

	span := spans[0]
	assert.Equal(t, "GET /api/hello", span.Name)
	assert.Equal(t, trace.SpanKindServer, span.SpanKind)
}

func TestHTTPTracing_WithSkipPaths(t *testing.T) {
	recorder := testutil.NewSpanRecorder()
	defer recorder.Shutdown()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	traced := middleware.HTTPTracing(recorder.TracerProvider(), "test-service",
		middleware.WithSkipPaths("/health"),
	)(handler)

	// Request to skipped path
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	traced.ServeHTTP(rec, req)

	// Request to traced path
	req = httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rec = httptest.NewRecorder()
	traced.ServeHTTP(rec, req)

	// Only the /api/data request should be traced
	spans := recorder.Spans()
	require.Len(t, spans, 1)
	assert.Equal(t, "GET /api/data", spans[0].Name)
}

func TestHTTPTracing_ErrorStatus(t *testing.T) {
	recorder := testutil.NewSpanRecorder()
	defer recorder.Shutdown()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	traced := middleware.HTTPTracing(recorder.TracerProvider(), "test-service")(handler)

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	traced.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	spans := recorder.Spans()
	require.Len(t, spans, 1)

	// Check status code was recorded
	statusCode, ok := testutil.GetSpanAttribute(spans[0], "http.status_code")
	require.True(t, ok)
	assert.Equal(t, int64(http.StatusInternalServerError), statusCode.AsInt64())
}

func TestHTTPTracingHandler(t *testing.T) {
	recorder := testutil.NewSpanRecorder()
	defer recorder.Shutdown()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	traced := middleware.HTTPTracingHandler(recorder.TracerProvider(), "test-service", handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	traced.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	spans := recorder.Spans()
	require.Len(t, spans, 1)
}

func TestSpanRecorder(t *testing.T) {
	recorder := testutil.NewSpanRecorder()
	defer recorder.Shutdown()

	tp := recorder.TracerProvider()
	tracer := tp.Tracer("test")

	_, span1 := tracer.Start(context.Background(), "span-1")
	span1.End()

	_, span2 := tracer.Start(context.Background(), "span-2")
	span2.End()

	spans := recorder.Spans()
	require.Len(t, spans, 2)

	// Find by name
	found := recorder.FindSpanByName("span-1")
	require.NotNil(t, found)
	assert.Equal(t, "span-1", found.Name)

	// Reset
	recorder.Reset()
	assert.Len(t, recorder.Spans(), 0)
}

func TestRequestRecorder(t *testing.T) {
	recorder := testutil.NewRequestRecorder()

	handler := recorder.Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/create", nil)
	req.Header.Set("X-Custom", "value")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	requests := recorder.Requests()
	require.Len(t, requests, 1)
	assert.Equal(t, http.MethodPost, requests[0].Method)
	assert.Equal(t, "/api/create", requests[0].Path)
	assert.Equal(t, "value", requests[0].Headers.Get("X-Custom"))

	// Reset
	recorder.Reset()
	assert.Len(t, recorder.Requests(), 0)
}
