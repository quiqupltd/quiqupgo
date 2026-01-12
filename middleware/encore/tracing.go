package encore

import (
	"context"
	"encoding/base32"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// encoreBase32Encoding is Encore's custom base32 encoding for trace/span IDs.
var encoreBase32Encoding = base32.NewEncoding("0123456789abcdefghijklmnopqrstuv").WithPadding(base32.NoPadding)

// TraceInfo contains trace information from an Encore request.
// Extract these values from middleware.Request.Data().Trace in your Encore middleware.
type TraceInfo struct {
	// TraceID is the current trace ID (base32 encoded by Encore)
	TraceID string
	// SpanID is the current span ID (base32 encoded by Encore)
	SpanID string
	// ParentTraceID is the parent trace ID if this is a child span
	ParentTraceID string
	// ParentSpanID is the parent span ID if this is a child span
	ParentSpanID string
}

// ConvertTraceID converts Encore's base32-encoded trace ID to OpenTelemetry format.
// Returns a zero TraceID if the input is empty or invalid.
func ConvertTraceID(encoreTraceID string) trace.TraceID {
	var traceIDBytes [16]byte
	if encoreTraceID != "" {
		if decoded, err := encoreBase32Encoding.DecodeString(encoreTraceID); err == nil {
			copy(traceIDBytes[:], decoded)
		}
	}
	return trace.TraceID(traceIDBytes)
}

// ConvertSpanID converts Encore's base32-encoded span ID to OpenTelemetry format.
// Returns a zero SpanID if the input is empty or invalid.
func ConvertSpanID(encoreSpanID string) trace.SpanID {
	var spanIDBytes [8]byte
	if encoreSpanID != "" {
		if decoded, err := encoreBase32Encoding.DecodeString(encoreSpanID); err == nil {
			copy(spanIDBytes[:], decoded)
		}
	}
	return trace.SpanID(spanIDBytes)
}

// StartSpan creates a new OpenTelemetry span correlated with Encore's trace context.
//
// The span shares Encore's trace ID for correlation but is created as a root span
// in OpenTelemetry (not as a child of Encore's span). This is intentional because
// Encore exports its spans separately, and attempting to parent under Encore's spans
// would result in "root span not yet received" errors in tracing UIs.
//
// The span automatically includes attributes for the original Encore trace/span IDs
// to aid in debugging and correlation.
func StartSpan(
	ctx context.Context,
	tp trace.TracerProvider,
	info *TraceInfo,
	spanName string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	traceID := ConvertTraceID(info.TraceID)

	// Create a span context with just the trace ID for correlation.
	// We intentionally don't set SpanID here - this makes our span a root span
	// that shares the trace ID with Encore for correlation, avoiding orphaned
	// parent references.
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
	ctx = trace.ContextWithSpanContext(ctx, spanCtx)

	// Add Encore trace correlation attributes
	encoreAttrs := []attribute.KeyValue{
		attribute.String("encore.trace_id", info.TraceID),
		attribute.String("encore.span_id", info.SpanID),
	}
	if info.ParentTraceID != "" {
		encoreAttrs = append(encoreAttrs, attribute.String("encore.parent_trace_id", info.ParentTraceID))
	}
	if info.ParentSpanID != "" {
		encoreAttrs = append(encoreAttrs, attribute.String("encore.parent_span_id", info.ParentSpanID))
	}

	// Prepend our attributes to any user-provided options
	opts = append([]trace.SpanStartOption{trace.WithAttributes(encoreAttrs...)}, opts...)

	tracer := tp.Tracer("github.com/quiqupltd/quiqupgo/middleware/encore")
	return tracer.Start(ctx, spanName, opts...)
}

// StartSpanWithParent creates an OpenTelemetry span as a child of Encore's current span.
//
// WARNING: Only use this if Encore exports traces to the SAME backend as your
// OpenTelemetry traces. Otherwise, you'll see "root span not yet received" errors
// because the parent span (from Encore) won't exist in your tracing backend.
//
// For most use cases, prefer StartSpan which creates correlated but independent spans.
func StartSpanWithParent(
	ctx context.Context,
	tp trace.TracerProvider,
	info *TraceInfo,
	spanName string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	traceID := ConvertTraceID(info.TraceID)
	spanID := ConvertSpanID(info.SpanID)

	// Create span context with both trace and span ID - this makes our span
	// a child of Encore's span
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
	ctx = trace.ContextWithSpanContext(ctx, spanCtx)

	// Add Encore trace correlation attributes
	encoreAttrs := []attribute.KeyValue{
		attribute.String("encore.trace_id", info.TraceID),
		attribute.String("encore.span_id", info.SpanID),
	}
	if info.ParentTraceID != "" {
		encoreAttrs = append(encoreAttrs, attribute.String("encore.parent_trace_id", info.ParentTraceID))
	}
	if info.ParentSpanID != "" {
		encoreAttrs = append(encoreAttrs, attribute.String("encore.parent_span_id", info.ParentSpanID))
	}

	opts = append([]trace.SpanStartOption{trace.WithAttributes(encoreAttrs...)}, opts...)

	tracer := tp.Tracer("github.com/quiqupltd/quiqupgo/middleware/encore")
	return tracer.Start(ctx, spanName, opts...)
}
