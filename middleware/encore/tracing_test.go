package encore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestConvertTraceID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValid bool
	}{
		{
			name:      "valid trace ID",
			input:     "0123456789abcdef0123456789abcdef",
			wantValid: true,
		},
		{
			name:      "empty string",
			input:     "",
			wantValid: false,
		},
		{
			name:      "invalid base32",
			input:     "!!!invalid!!!",
			wantValid: false,
		},
		{
			name:      "short input produces partial ID",
			input:     "0123456789abcdef",
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertTraceID(tt.input)
			assert.Equal(t, tt.wantValid, result.IsValid(), "validity mismatch for input: %s", tt.input)
		})
	}
}

func TestConvertTraceID_Deterministic(t *testing.T) {
	// Same input should always produce same output
	input := "0123456789abcdef0123456789abcdef"
	result1 := ConvertTraceID(input)
	result2 := ConvertTraceID(input)
	assert.Equal(t, result1, result2)
	assert.True(t, result1.IsValid())
}

func TestConvertSpanID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValid bool
	}{
		{
			name:      "valid span ID",
			input:     "0123456789abcdef",
			wantValid: true,
		},
		{
			name:      "empty string",
			input:     "",
			wantValid: false,
		},
		{
			name:      "invalid base32",
			input:     "!!!invalid!!!",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertSpanID(tt.input)
			assert.Equal(t, tt.wantValid, result.IsValid(), "validity mismatch for input: %s", tt.input)
		})
	}
}

func TestConvertSpanID_Deterministic(t *testing.T) {
	// Same input should always produce same output
	input := "0123456789abcdef"
	result1 := ConvertSpanID(input)
	result2 := ConvertSpanID(input)
	assert.Equal(t, result1, result2)
	assert.True(t, result1.IsValid())
}

func TestStartSpan(t *testing.T) {
	// Create a test tracer provider with span recorder
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	ctx := context.Background()
	info := &TraceInfo{
		TraceID:       "0123456789abcdef0123456789abcdef",
		SpanID:        "0123456789abcdef",
		ParentTraceID: "fedcba9876543210fedcba9876543210",
		ParentSpanID:  "fedcba9876543210",
	}

	ctx, span := StartSpan(ctx, tp, info, "test-span",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(attribute.String("custom", "value")),
	)
	require.NotNil(t, span)
	span.End()

	// Verify span was created
	spans := recorder.Ended()
	require.Len(t, spans, 1)

	recordedSpan := spans[0]
	assert.Equal(t, "test-span", recordedSpan.Name())
	assert.Equal(t, trace.SpanKindServer, recordedSpan.SpanKind())

	// Verify trace ID was converted correctly
	expectedTraceID := ConvertTraceID(info.TraceID)
	assert.Equal(t, expectedTraceID, recordedSpan.SpanContext().TraceID())

	// Verify encore attributes are present
	attrs := recordedSpan.Attributes()
	attrMap := make(map[string]string)
	for _, a := range attrs {
		if a.Value.Type() == attribute.STRING {
			attrMap[string(a.Key)] = a.Value.AsString()
		}
	}

	assert.Equal(t, info.TraceID, attrMap["encore.trace_id"])
	assert.Equal(t, info.SpanID, attrMap["encore.span_id"])
	assert.Equal(t, info.ParentTraceID, attrMap["encore.parent_trace_id"])
	assert.Equal(t, info.ParentSpanID, attrMap["encore.parent_span_id"])
	assert.Equal(t, "value", attrMap["custom"])

	// Verify span context is propagated
	spanCtx := trace.SpanContextFromContext(ctx)
	assert.True(t, spanCtx.IsValid())
	assert.Equal(t, expectedTraceID, spanCtx.TraceID())
}

func TestStartSpan_WithoutParentInfo(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	ctx := context.Background()
	info := &TraceInfo{
		TraceID: "0123456789abcdef0123456789abcdef",
		SpanID:  "0123456789abcdef",
		// No parent info
	}

	_, span := StartSpan(ctx, tp, info, "test-span")
	span.End()

	spans := recorder.Ended()
	require.Len(t, spans, 1)

	// Verify only trace_id and span_id attributes are present (no parent attributes)
	attrs := spans[0].Attributes()
	attrMap := make(map[string]string)
	for _, a := range attrs {
		if a.Value.Type() == attribute.STRING {
			attrMap[string(a.Key)] = a.Value.AsString()
		}
	}

	assert.Equal(t, info.TraceID, attrMap["encore.trace_id"])
	assert.Equal(t, info.SpanID, attrMap["encore.span_id"])
	_, hasParentTrace := attrMap["encore.parent_trace_id"]
	_, hasParentSpan := attrMap["encore.parent_span_id"]
	assert.False(t, hasParentTrace)
	assert.False(t, hasParentSpan)
}

func TestStartSpanWithParent(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	ctx := context.Background()
	info := &TraceInfo{
		TraceID: "0123456789abcdef0123456789abcdef",
		SpanID:  "0123456789abcdef",
	}

	_, span := StartSpanWithParent(ctx, tp, info, "child-span")
	require.NotNil(t, span)
	span.End()

	spans := recorder.Ended()
	require.Len(t, spans, 1)

	recordedSpan := spans[0]
	assert.Equal(t, "child-span", recordedSpan.Name())

	// Verify both trace ID and span ID were set correctly
	expectedTraceID := ConvertTraceID(info.TraceID)
	expectedSpanID := ConvertSpanID(info.SpanID)
	assert.Equal(t, expectedTraceID, recordedSpan.SpanContext().TraceID())

	// The parent span ID should be the Encore span ID
	assert.Equal(t, expectedSpanID, recordedSpan.Parent().SpanID())
}

func TestStartSpan_EmptyTraceInfo(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	ctx := context.Background()
	info := &TraceInfo{} // Empty trace info

	_, span := StartSpan(ctx, tp, info, "test-span")
	span.End()

	// Should still create a span, just with empty/zero trace ID
	spans := recorder.Ended()
	require.Len(t, spans, 1)
}

func TestTracerName(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	ctx := context.Background()
	info := &TraceInfo{
		TraceID: "0123456789abcdef0123456789abcdef",
		SpanID:  "0123456789abcdef",
	}

	_, span := StartSpan(ctx, tp, info, "test-span")
	span.End()

	spans := recorder.Ended()
	require.Len(t, spans, 1)

	// Verify the tracer name
	assert.Equal(t, "github.com/quiqupltd/quiqupgo/middleware/encore", spans[0].InstrumentationScope().Name)
}
