# Quiqupgo - Project Overview

## Purpose
Quiqupgo is a collection of reusable [uber/fx](https://github.com/uber-go/fx) modules for Go microservices. It provides standardized infrastructure components for building production-ready services at Quiqup.

## Tech Stack
- **Language**: Go 1.24+
- **DI Framework**: uber/fx for dependency injection and lifecycle management
- **Observability**: OpenTelemetry (tracing, metrics, logging)
- **Logging**: uber/zap with OTEL integration
- **Database**: GORM with PostgreSQL support and OTEL tracing plugin
- **Messaging**: Kafka (via segmentio/kafka-go) with tracing
- **Workflows**: Temporal with OTEL tracing
- **HTTP**: Echo framework middleware, standard net/http middleware
- **Task Runner**: [Task](https://taskfile.dev/) (Taskfile.yml)
- **Version Management**: asdf (.tool-versions)

## Available Modules

| Module | Package | Provides |
|--------|---------|----------|
| **Tracing** | `tracing` | TracerProvider, Tracer, MeterProvider, Meter |
| **Logger** | `logger` | *zap.Logger, Logger interface |
| **Temporal** | `temporal` | Temporal client.Client |
| **GORM** | `gormfx` | *gorm.DB with OTEL plugin |
| **Kafka** | `kafka` | Kafka Producer, Consumer |
| **Middleware** | `middleware` | HTTP tracing middleware (Echo/net/http) |

## Module Dependencies
```
tracing.Module() → provides: TracerProvider, Tracer, MeterProvider, Meter
    ↓
logger.Module() → provides: *zap.Logger, Logger interface
    ↓
temporal.Module() → requires: *zap.Logger, Tracer → provides: client.Client
gormfx.Module()   → requires: TracerProvider → provides: *gorm.DB
kafka.Module()   → requires: *zap.Logger, Tracer → provides: Producer, Consumer
```

## Repository Structure
```
quiqupgo/
├── tracing/          # OpenTelemetry tracing + metrics module
├── logger/           # Structured logging module
├── temporal/         # Temporal workflow client module
├── gormfx/           # GORM database module
├── kafka/           # Kafka/Kafka messaging module
├── middleware/       # HTTP middleware (not an fx module)
├── fxutil/           # Shared fx utilities
├── examples/         # Example applications
├── docs/             # Documentation
└── taskfiles/        # Task definitions
```

## Private Repository
This is a private Quiqup repository. Consumers need:
- `GOPRIVATE=github.com/quiqupltd/*`
- Git configured for SSH or PAT authentication
