// Package middleware provides HTTP middleware for OpenTelemetry tracing.
//
// Unlike other packages, middleware does not provide an fx.Module.
// Instead, it provides standalone middleware functions for Echo and net/http.
//
// Example usage with Echo:
//
//	e := echo.New()
//	e.Use(middleware.EchoTracing(tracerProvider, "my-service"))
//
// Example usage with net/http:
//
//	mux := http.NewServeMux()
//	handler := middleware.HTTPTracing(tracerProvider, "my-service")(mux)
package middleware
