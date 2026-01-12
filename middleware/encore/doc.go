// Package encore provides OpenTelemetry tracing middleware helpers for Encore.dev applications.
//
// This package helps bridge Encore's built-in tracing with OpenTelemetry by converting
// Encore's base32-encoded trace and span IDs to OTEL format and creating properly
// correlated spans.
//
// # Usage
//
// In your Encore application, create a thin middleware wrapper that calls the helpers:
//
//	//encore:middleware global target=all
//	func TracingMiddleware(req middleware.Request, next middleware.Next) middleware.Response {
//	    reqData := req.Data()
//
//	    tp := getTracerProvider() // your tracer provider
//
//	    ctx, span := encore.StartSpan(req.Context(), tp, &encore.TraceInfo{
//	        TraceID:       reqData.Trace.TraceID,
//	        SpanID:        reqData.Trace.SpanID,
//	        ParentTraceID: reqData.Trace.ParentTraceID,
//	        ParentSpanID:  reqData.Trace.ParentSpanID,
//	    }, reqData.Endpoint,
//	        trace.WithSpanKind(trace.SpanKindServer),
//	    )
//	    defer span.End()
//
//	    resp := next(req.WithContext(ctx))
//	    if resp.Err != nil {
//	        span.RecordError(resp.Err)
//	    }
//	    return resp
//	}
//
// # Trace Correlation
//
// The middleware creates spans that share Encore's trace ID for correlation in your
// tracing backend (Jaeger, etc.), but does NOT attempt to parent under Encore's spans
// since Encore exports traces separately. This avoids "root span not yet received"
// issues in your tracing UI.
package encore
