# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.0] - 2026-01-13

### Added

- **Temporal Module**: Worker tracing support for OpenTelemetry instrumentation
  - `WorkerInterceptors()` - Returns interceptors for `worker.Options.Interceptors`
  - `ApplyWorkerInterceptors(&opts)` - Convenience function to apply to existing options
  - `WorkerInterceptorsModule()` - Standalone fx module for worker interceptors
  - `WithWorkerInterceptors()` - Module option to provide interceptors via fx DI
  - `WorkerInterceptorSlice` - Injectable type for fx consumers
  - Enables tracing of workflow and activity execution on workers (complements existing client-side tracing)

## [0.4.0] - 2026-01-13

### Added

- **Tracing Module**: `BaseService` embeddable struct for consistent service tracing
  - `Trace(ctx, name)` - Creates spans with automatic component name prefixing and proper error capture
  - `WithSpan(ctx, name, fn)` - Callback pattern for simpler tracing without named returns
  - `WithSpanResult[T](ctx, svc, name, fn)` - Generic helper for functions returning values
  - Automatic error recording and span status setting
  - See `tracing/doc.go` for comprehensive usage examples

## [0.3.4] - 2026-01-13

### Fixed

- **Tracing Module**: Fixed `GetTracerProvider`, `GetMeterProvider`, and `GetLoggerProvider` functions
  having private `*moduleOptions` parameter type, preventing external usage. Changed signatures to
  accept variadic `...ModuleOption` (which is public), enabling direct use outside the fx module:
  ```go
  tp, err := tracing.GetTracerProvider(ctx, cfg)
  tp, err := tracing.GetTracerProvider(ctx, cfg, tracing.WithSampler(...))
  ```

## [0.3.3] - 2026-01-12

### Added

- **Encore Middleware** (`middleware/encore`): New package for Encore.dev tracing integration
  - `StartSpan()` - Creates OTEL spans correlated with Encore's trace context
  - `StartSpanWithParent()` - Creates child spans (use only if Encore exports to same backend)
  - `ConvertTraceID()` / `ConvertSpanID()` - Convert Encore's base32 IDs to OTEL format
  - Avoids "root span not yet received" errors by creating correlated root spans

## [0.3.2] - 2026-01-12

### Fixed

- **Tracing Module**: Fixed resource detection failing in minimal containers with error
  `user: Current requires cgo or $USER set in environment`. Replaced `resource.WithProcess()`
  with specific process detectors that don't require `os/user.Current()`.

## [0.3.0] - 2026-01-12

Initial public release of quiqupgo - a collection of reusable uber/fx modules for Go microservices.

### Added

- **Tracing Module** (`tracing/`): OpenTelemetry tracing and metrics with OTLP export
  - TracerProvider and MeterProvider with lifecycle management
  - TLS support for secure OTLP endpoints
  - Resource attributes for service identification
  - Test utilities: `testutil.NoopModule()` for unit tests

- **Logger Module** (`logger/`): Structured logging with zap
  - Environment-aware configuration (development vs production)
  - Optional OpenTelemetry integration for trace correlation
  - Test utilities: `testutil.NoopModule()`, `testutil.BufferModule()` for assertions

- **Temporal Module** (`temporal/`): Temporal workflow client
  - OTEL tracing interceptor for distributed tracing
  - TLS support for secure connections
  - Zap logger adapter
  - Workflow utilities: `ListWorkflows()`, `GetWorkflowStatus()`
  - Test utilities: `testutil.MockModule()` with mock client

- **GORM Module** (`gormfx/`): Database with OTEL tracing
  - GORM integration with OpenTelemetry plugin
  - Accepts existing `*sql.DB` for connection pooling control
  - Test utilities: `testutil.MockModule()`

- **Kafka Module** (`kafka/`): Kafka messaging with tracing
  - Producer and Consumer with OTEL context propagation
  - TLS and SASL authentication support
  - Configurable consumer groups and topics
  - Test utilities: `testutil.TestModule()` with in-memory kafka

- **Middleware** (`middleware/`): HTTP tracing middleware
  - Echo framework integration
  - Standard net/http middleware

- **Examples**: Minimal, API service, and worker service examples

- **CI/CD**: GitHub Actions workflows with Blacksmith runners
  - Lint, test, and coverage checks
  - Integration tests with Postgres, Redpanda, Temporal, OTEL Collector
  - Coverage badge uploaded to GCS
  - Custom Temporal Docker image for CI

- **Developer Tooling**:
  - `CLAUDE.md` with project instructions for Claude Code
  - `.serena/` configuration for Serena IDE integration

[Unreleased]: https://github.com/quiqupltd/quiqupgo/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/quiqupltd/quiqupgo/compare/v0.3.4...v0.4.0
[0.3.4]: https://github.com/quiqupltd/quiqupgo/compare/v0.3.3...v0.3.4
[0.3.3]: https://github.com/quiqupltd/quiqupgo/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/quiqupltd/quiqupgo/compare/v0.3.0...v0.3.2
[0.3.0]: https://github.com/quiqupltd/quiqupgo/releases/tag/v0.3.0
