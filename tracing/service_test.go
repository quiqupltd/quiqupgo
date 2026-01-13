package tracing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/quiqupltd/quiqupgo/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// testHelper creates a BaseService with an in-memory span recorder for testing.
func testHelper(t *testing.T) (*tracing.BaseService, *tracetest.InMemoryExporter) {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())
	})

	tracer := tp.Tracer("test")
	meter := metricnoop.NewMeterProvider().Meter("test")
	base := tracing.NewBaseService(tracer, meter, "test.component")

	return &base, exporter
}

func TestNewBaseService(t *testing.T) {
	tracer := sdktrace.NewTracerProvider().Tracer("test")
	meter := metricnoop.NewMeterProvider().Meter("test")

	base := tracing.NewBaseService(tracer, meter, "my.service")

	assert.Equal(t, "my.service", base.ComponentName())
	assert.NotNil(t, base.Tracer())
	assert.NotNil(t, base.Meter())
}

func TestBaseService_Trace_Success(t *testing.T) {
	base, exporter := testHelper(t)

	// Simulate a function with named return that succeeds
	doWork := func(ctx context.Context) (result string, err error) {
		_, end := base.Trace(ctx, "DoWork")
		defer end(&err)

		return "success", nil
	}

	ctx := context.Background()
	result, err := doWork(ctx)

	require.NoError(t, err)
	assert.Equal(t, "success", result)

	// Verify span was created with correct name
	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "test.component.DoWork", spans[0].Name)
	assert.Equal(t, codes.Unset, spans[0].Status.Code)
	assert.Empty(t, spans[0].Events) // No error events
}

func TestBaseService_Trace_Error(t *testing.T) {
	base, exporter := testHelper(t)

	expectedErr := errors.New("something went wrong")

	// Simulate a function with named return that fails
	doWork := func(ctx context.Context) (result string, err error) {
		_, end := base.Trace(ctx, "DoWork")
		defer end(&err)

		return "", expectedErr
	}

	ctx := context.Background()
	result, err := doWork(ctx)

	require.ErrorIs(t, err, expectedErr)
	assert.Empty(t, result)

	// Verify span was created with error recorded
	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "test.component.DoWork", spans[0].Name)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
	assert.Equal(t, "something went wrong", spans[0].Status.Description)

	// Verify error was recorded as event
	require.Len(t, spans[0].Events, 1)
	assert.Equal(t, "exception", spans[0].Events[0].Name)
}

func TestBaseService_Trace_ErrorAssignedLater(t *testing.T) {
	base, exporter := testHelper(t)

	// This test verifies that errors assigned AFTER the defer are still captured
	// This is the key difference from the broken pattern where err is captured at defer time
	doWork := func(ctx context.Context) (err error) {
		ctx, end := base.Trace(ctx, "DoWork")
		defer end(&err)

		// Do some work...
		_ = ctx

		// Error is assigned after defer was registered
		err = errors.New("late error")
		return err
	}

	ctx := context.Background()
	err := doWork(ctx)

	require.Error(t, err)

	// Verify the late error was captured
	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
	assert.Equal(t, "late error", spans[0].Status.Description)
}

func TestBaseService_Trace_WithSpanOptions(t *testing.T) {
	base, exporter := testHelper(t)

	doWork := func(ctx context.Context) (err error) {
		_, end := base.Trace(ctx, "DoWork",
			trace.WithAttributes(attribute.String("user.id", "123")),
		)
		defer end(&err)
		return nil
	}

	ctx := context.Background()
	err := doWork(ctx)
	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	// Verify attributes were applied
	attrs := spans[0].Attributes
	var found bool
	for _, attr := range attrs {
		if attr.Key == "user.id" && attr.Value.AsString() == "123" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected attribute user.id=123 not found")
}

func TestBaseService_Trace_NestedSpans(t *testing.T) {
	base, exporter := testHelper(t)

	innerWork := func(ctx context.Context) (err error) {
		ctx, end := base.Trace(ctx, "InnerWork")
		defer end(&err)
		_ = ctx
		return nil
	}

	outerWork := func(ctx context.Context) (err error) {
		ctx, end := base.Trace(ctx, "OuterWork")
		defer end(&err)
		return innerWork(ctx)
	}

	ctx := context.Background()
	err := outerWork(ctx)
	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 2)

	// Spans are recorded in order of completion (inner first, then outer)
	assert.Equal(t, "test.component.InnerWork", spans[0].Name)
	assert.Equal(t, "test.component.OuterWork", spans[1].Name)

	// Verify parent-child relationship
	assert.Equal(t, spans[1].SpanContext.SpanID(), spans[0].Parent.SpanID())
}

func TestBaseService_WithSpan_Success(t *testing.T) {
	base, exporter := testHelper(t)

	ctx := context.Background()
	err := base.WithSpan(ctx, "ProcessItem", func(ctx context.Context) error {
		// Verify context has span
		span := trace.SpanFromContext(ctx)
		assert.True(t, span.SpanContext().IsValid())
		return nil
	})

	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "test.component.ProcessItem", spans[0].Name)
	assert.Equal(t, codes.Unset, spans[0].Status.Code)
}

func TestBaseService_WithSpan_Error(t *testing.T) {
	base, exporter := testHelper(t)

	expectedErr := errors.New("process failed")

	ctx := context.Background()
	err := base.WithSpan(ctx, "ProcessItem", func(ctx context.Context) error {
		return expectedErr
	})

	require.ErrorIs(t, err, expectedErr)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
	assert.Equal(t, "process failed", spans[0].Status.Description)
}

func TestBaseService_WithSpan_WithOptions(t *testing.T) {
	base, exporter := testHelper(t)

	ctx := context.Background()
	err := base.WithSpan(ctx, "ProcessItem", func(ctx context.Context) error {
		return nil
	}, trace.WithAttributes(attribute.Int("item.count", 42)))

	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	var found bool
	for _, attr := range spans[0].Attributes {
		if attr.Key == "item.count" && attr.Value.AsInt64() == 42 {
			found = true
			break
		}
	}
	assert.True(t, found, "expected attribute item.count=42 not found")
}

func TestWithSpanResult_Success(t *testing.T) {
	base, exporter := testHelper(t)

	ctx := context.Background()
	result, err := tracing.WithSpanResult(ctx, base, "FetchUser", func(ctx context.Context) (string, error) {
		return "user-123", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "user-123", result)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "test.component.FetchUser", spans[0].Name)
	assert.Equal(t, codes.Unset, spans[0].Status.Code)
}

func TestWithSpanResult_Error(t *testing.T) {
	base, exporter := testHelper(t)

	expectedErr := errors.New("user not found")

	ctx := context.Background()
	result, err := tracing.WithSpanResult(ctx, base, "FetchUser", func(ctx context.Context) (string, error) {
		return "", expectedErr
	})

	require.ErrorIs(t, err, expectedErr)
	assert.Empty(t, result)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
	assert.Equal(t, "user not found", spans[0].Status.Description)
}

func TestWithSpanResult_WithOptions(t *testing.T) {
	base, exporter := testHelper(t)

	ctx := context.Background()
	result, err := tracing.WithSpanResult(ctx, base, "FetchUser", func(ctx context.Context) (int, error) {
		return 42, nil
	}, trace.WithAttributes(attribute.String("query.type", "by-id")))

	require.NoError(t, err)
	assert.Equal(t, 42, result)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)

	var found bool
	for _, attr := range spans[0].Attributes {
		if attr.Key == "query.type" && attr.Value.AsString() == "by-id" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected attribute query.type=by-id not found")
}

func TestWithSpanResult_ComplexType(t *testing.T) {
	base, exporter := testHelper(t)

	type User struct {
		ID   string
		Name string
	}

	ctx := context.Background()
	result, err := tracing.WithSpanResult(ctx, base, "FetchUser", func(ctx context.Context) (*User, error) {
		return &User{ID: "123", Name: "Alice"}, nil
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "123", result.ID)
	assert.Equal(t, "Alice", result.Name)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
}

func TestBaseService_Tracer(t *testing.T) {
	tracer := sdktrace.NewTracerProvider().Tracer("test")
	meter := metricnoop.NewMeterProvider().Meter("test")
	base := tracing.NewBaseService(tracer, meter, "my.service")

	assert.Equal(t, tracer, base.Tracer())
}

func TestBaseService_Meter(t *testing.T) {
	tracer := sdktrace.NewTracerProvider().Tracer("test")
	meter := metricnoop.NewMeterProvider().Meter("test")
	base := tracing.NewBaseService(tracer, meter, "my.service")

	assert.Equal(t, meter, base.Meter())
}

func TestBaseService_ComponentName(t *testing.T) {
	tracer := sdktrace.NewTracerProvider().Tracer("test")
	meter := metricnoop.NewMeterProvider().Meter("test")
	base := tracing.NewBaseService(tracer, meter, "geocoder.domain")

	assert.Equal(t, "geocoder.domain", base.ComponentName())
}

func TestBaseService_Trace_NilErrorPointer(t *testing.T) {
	base, exporter := testHelper(t)

	// Test that passing nil error pointer doesn't panic
	doWork := func(ctx context.Context) {
		ctx, end := base.Trace(ctx, "DoWork")
		defer end(nil) // Pass nil explicitly
		_ = ctx
	}

	ctx := context.Background()
	doWork(ctx)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Unset, spans[0].Status.Code)
}

// ExampleUserService demonstrates embedding BaseService in a real service.
type ExampleUserService struct {
	tracing.BaseService
}

func NewExampleUserService(tracer trace.Tracer, meter metric.Meter) *ExampleUserService {
	return &ExampleUserService{
		BaseService: tracing.NewBaseService(tracer, meter, "user.service"),
	}
}

func (s *ExampleUserService) GetUser(ctx context.Context, id string) (user string, err error) {
	// In real usage, ctx would be passed to downstream calls (DB, HTTP, etc.)
	_, end := s.Trace(ctx, "GetUser")
	defer end(&err)

	if id == "" {
		return "", errors.New("id is required")
	}
	return "user-" + id, nil
}

func (s *ExampleUserService) DeleteUser(ctx context.Context, id string) error {
	return s.WithSpan(ctx, "DeleteUser", func(ctx context.Context) error {
		if id == "" {
			return errors.New("id is required")
		}
		return nil
	})
}

func TestExampleUserService_GetUser_Success(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	svc := NewExampleUserService(
		tp.Tracer("test"),
		metricnoop.NewMeterProvider().Meter("test"),
	)

	user, err := svc.GetUser(context.Background(), "123")
	require.NoError(t, err)
	assert.Equal(t, "user-123", user)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "user.service.GetUser", spans[0].Name)
}

func TestExampleUserService_GetUser_Error(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	svc := NewExampleUserService(
		tp.Tracer("test"),
		metricnoop.NewMeterProvider().Meter("test"),
	)

	_, err := svc.GetUser(context.Background(), "")
	require.Error(t, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status.Code)
}

func TestExampleUserService_DeleteUser(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	svc := NewExampleUserService(
		tp.Tracer("test"),
		metricnoop.NewMeterProvider().Meter("test"),
	)

	err := svc.DeleteUser(context.Background(), "123")
	require.NoError(t, err)

	spans := exporter.GetSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, "user.service.DeleteUser", spans[0].Name)
}
