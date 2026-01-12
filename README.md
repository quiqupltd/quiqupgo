# Quiqup Shared GO

![Coverage](https://storage.googleapis.com/quiqup-public-assets/coverbadges/quiqupltd/quiqupgo/coverage.svg)

Reusable [uber/fx](https://github.com/uber-go/fx) modules for Go microservices, providing standardized infrastructure for tracing, logging, database, temporal workflows, pub/sub messaging, and HTTP middleware.

## Features

- **Modular Design**: Import only the modules you need
- **Per-Module Configuration**: Each module defines its own config interface - no monolithic global config
- **OpenTelemetry Native**: Full OTEL integration for tracing, metrics, and logging
- **Testing Utilities**: Every module includes test helpers (NoopModule, MockModule, BufferLogger)
- **Uber/fx Integration**: Seamless dependency injection with lifecycle management

## Installation

```bash
go get github.com/quiqupltd/quiqupgo
```

## Available Modules

| Module | Package | Description |
|--------|---------|-------------|
| **Tracing** | `github.com/quiqupltd/quiqupgo/tracing` | OpenTelemetry tracing + metrics |
| **Logger** | `github.com/quiqupltd/quiqupgo/logger` | Structured logging with zap + OTEL |
| **Temporal** | `github.com/quiqupltd/quiqupgo/temporal` | Temporal workflow client |
| **GORM** | `github.com/quiqupltd/quiqupgo/gormfx` | GORM database with OTEL plugin |
| **Kafka** | `github.com/quiqupltd/quiqupgo/kafka` | Kafka messaging with tracing |
| **Middleware** | `github.com/quiqupltd/quiqupgo/middleware` | HTTP tracing middleware (Echo/net/http) |
| **Encore Middleware** | `github.com/quiqupltd/quiqupgo/middleware/encore` | Encore.dev tracing integration |

## Quick Start

### Minimal Example (Logger Only)

```go
package main

import (
    "github.com/quiqupltd/quiqupgo/logger"
    "go.uber.org/fx"
    "go.uber.org/zap"
)

func main() {
    fx.New(
        // Provide logger config
        fx.Provide(func() logger.Config {
            return &logger.StandardConfig{
                ServiceName: "my-service",
                Environment: "development",
            }
        }),

        // Include logger module
        logger.Module(),

        // Use the logger
        fx.Invoke(func(log *zap.Logger) {
            log.Info("application started")
        }),
    ).Run()
}
```

### Full Stack Example (All Modules)

```go
package main

import (
    "database/sql"

    "github.com/quiqupltd/quiqupgo/gormfx"
    "github.com/quiqupltd/quiqupgo/logger"
    "github.com/quiqupltd/quiqupgo/temporal"
    "github.com/quiqupltd/quiqupgo/tracing"
    "go.uber.org/fx"
)

// AppConfig is your application's configuration
type AppConfig struct {
    ServiceName string
    Environment string
    OTLP        struct {
        Endpoint string
        Insecure bool
    }
    Temporal struct {
        HostPort  string
        Namespace string
    }
    DB *sql.DB
}

func main() {
    fx.New(
        // Provide app config
        fx.Provide(newAppConfig),

        // Adapt app config to module configs
        fx.Provide(
            fx.Annotate(newTracingConfig, fx.As(new(tracing.Config))),
            fx.Annotate(newLoggerConfig, fx.As(new(logger.Config))),
            fx.Annotate(newTemporalConfig, fx.As(new(temporal.Config))),
            fx.Annotate(newGormConfig, fx.As(new(gormfx.Config))),
        ),

        // Include modules
        tracing.Module(),
        logger.Module(),
        temporal.Module(),
        gormfx.Module(),

        // Your application code
        fx.Invoke(run),
    ).Run()
}

func newTracingConfig(app *AppConfig) tracing.Config {
    return &tracing.StandardConfig{
        ServiceName:     app.ServiceName,
        EnvironmentName: app.Environment,
        OTLPEndpoint:    app.OTLP.Endpoint,
        OTLPInsecure:    app.OTLP.Insecure,
    }
}

func newLoggerConfig(app *AppConfig) logger.Config {
    return &logger.StandardConfig{
        ServiceName: app.ServiceName,
        Environment: app.Environment,
    }
}

func newTemporalConfig(app *AppConfig) temporal.Config {
    return &temporal.StandardConfig{
        HostPort:  app.Temporal.HostPort,
        Namespace: app.Temporal.Namespace,
    }
}

func newGormConfig(app *AppConfig) gormfx.Config {
    return &gormfx.StandardConfig{
        DB: app.DB,
    }
}
```

## Module Details

### Tracing Module

Provides OpenTelemetry tracing and metrics infrastructure.

```go
import "github.com/quiqupltd/quiqupgo/tracing"

// What it provides via fx:
// - trace.TracerProvider
// - trace.Tracer
// - metric.MeterProvider
// - metric.Meter

// Configuration interface
type Config interface {
    GetServiceName() string
    GetEnvironmentName() string
    GetOTLPEndpoint() string
    GetOTLPInsecure() bool
    GetOTLPTLSCert() string   // base64 encoded
    GetOTLPTLSKey() string    // base64 encoded
    GetOTLPTLSCA() string     // base64 encoded
}

// Usage
fx.New(
    fx.Provide(func() tracing.Config {
        return &tracing.StandardConfig{
            ServiceName:     "my-service",
            EnvironmentName: "production",
            OTLPEndpoint:    "otel-collector:4318",
        }
    }),
    tracing.Module(
        tracing.WithBatchTimeout(5 * time.Second),
        tracing.WithSampler(sdktrace.AlwaysSample()),
    ),
)
```

### Logger Module

Provides structured logging with zap and optional OpenTelemetry integration.

```go
import "github.com/quiqupltd/quiqupgo/logger"

// What it provides via fx:
// - *zap.Logger
// - logger.Logger (interface)

// Configuration interface
type Config interface {
    GetServiceName() string
    GetEnvironment() string  // "development" for pretty logs, else JSON
}

// Usage
fx.New(
    fx.Provide(func() logger.Config {
        return &logger.StandardConfig{
            ServiceName: "my-service",
            Environment: "development",
        }
    }),
    logger.Module(),
)
```

### Temporal Module

Provides Temporal workflow client with OTEL tracing.

```go
import "github.com/quiqupltd/quiqupgo/temporal"

// What it provides via fx:
// - client.Client (Temporal client)

// Configuration interface
type Config interface {
    GetHostPort() string
    GetNamespace() string
    GetTLSCert() string  // PEM encoded
    GetTLSKey() string   // PEM encoded
}

// Dependencies (must be provided):
// - *zap.Logger
// - trace.Tracer

// Usage
fx.New(
    // ... logger and tracing modules first
    fx.Provide(func() temporal.Config {
        return &temporal.StandardConfig{
            HostPort:  "localhost:7233",
            Namespace: "default",
        }
    }),
    temporal.Module(),
)
```

### GORM Module

Provides GORM database connection with OTEL tracing plugin.

```go
import "github.com/quiqupltd/quiqupgo/gormfx"

// What it provides via fx:
// - *gorm.DB

// Configuration interface
type Config interface {
    GetDB() *sql.DB
}

// Dependencies (must be provided):
// - trace.TracerProvider

// Usage
fx.New(
    // ... tracing module first
    fx.Provide(func(db *sql.DB) gormfx.Config {
        return &gormfx.StandardConfig{DB: db}
    }),
    gormfx.Module(),
)
```

### Kafka Module

Provides Kafka producer and consumer with OTEL tracing.

```go
import "github.com/quiqupltd/quiqupgo/kafka"

// What it provides via fx:
// - kafka.Producer
// - kafka.Consumer

// Configuration interface
type Config interface {
    GetBrokers() []string
    GetConsumerGroup() string
    GetTopics() []string
}

// Dependencies (must be provided):
// - trace.Tracer
// - *zap.Logger

// Usage
fx.New(
    // ... logger and tracing modules first
    fx.Provide(func() kafka.Config {
        return &kafka.StandardConfig{
            Brokers:       []string{"kafka:9092"},
            ConsumerGroup: "my-service",
            Topics:        []string{"events"},
        }
    }),
    kafka.Module(),
)
```

### Middleware Module

Provides HTTP middleware for tracing (not an fx.Module - standalone functions).

```go
import "github.com/quiqupltd/quiqupgo/middleware"

// Echo middleware
e := echo.New()
e.Use(middleware.EchoTracing(tracerProvider, "my-service"))

// Standard http middleware
mux := http.NewServeMux()
handler := middleware.HTTPTracing(tracerProvider, "my-service")(mux)
```

### Encore Middleware

Provides helpers for integrating OpenTelemetry tracing with [Encore.dev](https://encore.dev) applications.

```go
import "github.com/quiqupltd/quiqupgo/middleware/encore"

// In your Encore application, create middleware that wraps these helpers:

//encore:middleware global target=all
func TracingMiddleware(req middleware.Request, next middleware.Next) middleware.Response {
    reqData := req.Data()
    tp := getTracerProvider() // your tracer provider

    ctx, span := encore.StartSpan(req.Context(), tp, &encore.TraceInfo{
        TraceID:       reqData.Trace.TraceID,
        SpanID:        reqData.Trace.SpanID,
        ParentTraceID: reqData.Trace.ParentTraceID,
        ParentSpanID:  reqData.Trace.ParentSpanID,
    }, reqData.Endpoint.Name,
        trace.WithSpanKind(trace.SpanKindServer),
    )
    defer span.End()

    resp := next(req.WithContext(ctx))
    if resp.Err != nil {
        span.RecordError(resp.Err)
    }
    return resp
}
```

**Key Functions:**
- `StartSpan()` - Creates OTEL spans correlated with Encore's trace context (recommended)
- `StartSpanWithParent()` - Creates child spans under Encore's span (only if Encore exports to same backend)
- `ConvertTraceID()` / `ConvertSpanID()` - Convert Encore's base32 IDs to OTEL format

**Why `StartSpan` instead of `StartSpanWithParent`?**

Encore exports its own traces separately. Using `StartSpanWithParent` would create child spans referencing a parent that may not exist in your tracing backend, causing "root span not yet received" errors. `StartSpan` creates correlated but independent root spans that share the same trace ID for correlation.

## Testing

Each module provides test utilities:

```go
import (
    loggertest "github.com/quiqupltd/quiqupgo/logger/testutil"
    tracingtest "github.com/quiqupltd/quiqupgo/tracing/testutil"
    temporaltest "github.com/quiqupltd/quiqupgo/temporal/testutil"
)

func TestMyService(t *testing.T) {
    // Get buffer logger for assertions
    logMod, buffer := loggertest.BufferModule()

    var svc *MyService
    app := fx.New(
        fx.NopLogger,
        tracingtest.NoopModule(),  // No-op tracing
        logMod,                     // Buffer logger
        temporaltest.MockModule(), // Mock temporal client
        fx.Provide(NewMyService),
        fx.Populate(&svc),
    )

    ctx := context.Background()
    require.NoError(t, app.Start(ctx))
    defer app.Stop(ctx)

    // Test your service
    svc.DoSomething()

    // Assert on logs
    entries := buffer.GetEntries()
    assert.Contains(t, entries[0].Message, "did something")
}
```

## Creating Your Own App Module

Best practice is to create your own composition module that adapts your app config:

```go
// internal/fxglobal/module.go
package fxglobal

import (
    "github.com/quiqupltd/quiqupgo/gormfx"
    "github.com/quiqupltd/quiqupgo/logger"
    "github.com/quiqupltd/quiqupgo/tracing"
    "go.uber.org/fx"
)

func Module(serviceName string, db *sql.DB) fx.Option {
    return fx.Module("global",
        fx.Provide(func() *AppConfig {
            return loadAppConfig()
        }),

        // Adapt to module configs
        fx.Provide(
            fx.Annotate(newTracingConfig, fx.As(new(tracing.Config))),
            fx.Annotate(newLoggerConfig, fx.As(new(logger.Config))),
            fx.Annotate(func() gormfx.Config {
                return &gormfx.StandardConfig{DB: db}
            }, fx.As(new(gormfx.Config))),
        ),

        // Include modules (only what you need)
        tracing.Module(),
        logger.Module(),
        gormfx.Module(),
        // temporal.Module() - not needed for this service
    )
}
```

## Development

### Prerequisites

- Go 1.24+ (via asdf: `asdf install`)
- [Task](https://taskfile.dev/) (via asdf: `asdf install`)
- Docker or [OrbStack](https://orbstack.dev/)
- [direnv](https://direnv.net/) (optional, for automatic environment loading)

### Setup

```bash
# Install Go and Task via asdf
asdf install

# Install development tools
task tools:install

# Setup git hooks
task tools:hooks

# Run tests
task test:unit

# Run linter
task tools:lint

# Format code
task tools:fmt
```

### Running Integration Tests

Integration tests require local infrastructure (Redpanda, Postgres, Temporal, etc.):

```bash
# Start all services via OrbStack/Docker
task docker:up

# Wait for services to be healthy
task docker:status

# Run integration tests
task test:integration

# Stop services when done
task docker:down
```

Services are accessible at:
- Postgres: `localhost:5432`
- Redpanda/Kafka: `localhost:19092`
- Temporal: `localhost:7233`
- Jaeger UI: `localhost:16686`
- Redpanda Console: `localhost:8080`

### Available Tasks

```bash
task --list
```

## Versioning & Releases

This project uses [Semantic Versioning](https://semver.org/) and [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/) for release management.

### Consuming Specific Versions

```bash
# Latest version
go get github.com/quiqupltd/quiqupgo

# Specific version
go get github.com/quiqupltd/quiqupgo@v0.1.0

# Latest patch for a minor version
go get github.com/quiqupltd/quiqupgo@v0.1
```

### Branching Strategy

| Branch | Purpose |
|--------|---------|
| `main` | Production-ready releases, tagged with versions |
| `develop` | Integration branch for features |
| `feat/*` | New features and enhancements |
| `rel/*` | Release preparation |
| `hotfix/*` | Critical production fixes |

### Version History

See [CHANGELOG.md](CHANGELOG.md) for detailed release notes and [GitHub Releases](https://github.com/quiqupltd/quiqupgo/releases) for downloadable artifacts.

## Project Structure

```
quiqupgo/
├── tracing/          # OpenTelemetry tracing + metrics module
├── logger/           # Structured logging module
├── temporal/         # Temporal workflow client module
├── gormfx/           # GORM database module
├── kafka/           # Kafka messaging module
├── middleware/       # HTTP middleware
│   └── encore/       # Encore.dev tracing helpers
├── fxutil/           # Shared utilities
├── examples/         # Example applications
│   ├── minimal/      # Logger only
│   ├── api-service/  # Tracing + Logger + GORM
│   └── worker-service/ # All modules
└── docs/             # Documentation
```

## License

MIT License - see LICENSE file for details.
