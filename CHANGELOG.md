# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.1] - 2026-01-12

### Added

- **Developer Tooling**: Added Serena and Claude configuration files
  - `CLAUDE.md` with project instructions for Claude Code
  - `.serena/` configuration for Serena IDE integration

## [0.2.0] - 2026-01-12

### Added

- **Temporal Docker Image**: Custom image for CI with dev server baked in
  - `Dockerfile.temporal` with `server start-dev` command
  - Published to public GCR for unauthenticated service container pulls
  - Built with Kaniko for efficient caching

### Changed

- **CI/CD Improvements**:
  - Switched to Blacksmith runners for faster builds
  - Added workflow-level concurrency to cancel in-progress runs
  - Simplified Temporal setup using `temporalio/admin-tools:1.29` dev server
  - Main workflow now runs on both `main` and `dev` branches
  - Added race detector to coverage job
  - Coverage HTML uploaded as artifact

- **Documentation**:
  - Added Versioning & Releases section to README
  - Added CHANGELOG.md following Keep a Changelog format
  - Documented Git Flow branching strategy (`feat/*`, `rel/*`, `hotfix/*`)

## [0.1.0] - 2025-01-12

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

- **PubSub Module** (`pubsub/`): Kafka messaging with tracing
  - Producer and Consumer with OTEL context propagation
  - Configurable consumer groups and topics
  - Test utilities: `testutil.MockModule()` with in-memory pubsub

- **Middleware** (`middleware/`): HTTP tracing middleware
  - Echo framework integration
  - Standard net/http middleware

- **Examples**: Minimal, API service, and worker service examples

- **CI/CD**: GitHub Actions workflows for PR and main branch
  - Lint, test, and coverage checks
  - Integration tests with Postgres, Redpanda, Temporal, OTEL Collector
  - Coverage badge uploaded to GCS

[Unreleased]: https://github.com/quiqupltd/quiqupgo/compare/v0.2.1...HEAD
[0.2.1]: https://github.com/quiqupltd/quiqupgo/releases/tag/v0.2.1
[0.2.0]: https://github.com/quiqupltd/quiqupgo/releases/tag/v0.2.0
[0.1.0]: https://github.com/quiqupltd/quiqupgo/releases/tag/v0.1.0
