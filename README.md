# Quiqup Shared GO

![Coverage](https://storage.googleapis.com/quiqup-public-assets/coverbadges/quiqupltd/quiqupgo/coverage.svg)

Reusable [uber/fx](https://github.com/uber-go/fx) modules for Go microservices, providing standardized infrastructure for tracing, logging, database, temporal workflows, pub/sub messaging, and HTTP middleware.

## Features

- **Modular Design**: Import only the modules you need
- **Per-Module Configuration**: Each module defines its own config interface - no monolithic global config
- **OpenTelemetry Native**: Full OTEL integration for tracing, metrics, and logging
- **Testing Utilities**: Every module includes test helpers (NoopModule, MockModule, BufferLogger)
- **Uber/fx Integration**: Seamless dependency injection with lifecycle management

## Prerequisites

This is a private repository. Before installing, you need to configure Go and Git to access private Quiqup modules.

### Local Development Setup

1. **Set GOPRIVATE** to skip the public proxy for Quiqup modules:

```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
export GOPRIVATE=github.com/quiqupltd/*
```

2. **Configure Git** to use SSH for GitHub:

```bash
git config --global url."git@github.com:".insteadOf "https://github.com/"
```

Or if using a GitHub Personal Access Token:

```bash
git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
```

### CI/CD Setup

For GitHub Actions, add these environment variables to your workflow:

```yaml
env:
  GOPRIVATE: github.com/quiqupltd/*

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Configure Git for private modules
        run: git config --global url."https://${{ secrets.GH_PAT }}@github.com/".insteadOf "https://github.com/"

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - run: go mod download
      - run: go build ./...
```

> **Note**: `GH_PAT` should be a GitHub Personal Access Token with `repo` scope, stored as a repository or organization secret.

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
| **PubSub** | `github.com/quiqupltd/quiqupgo/pubsub` | Kafka messaging with tracing |
| **Middleware** | `github.com/quiqupltd/quiqupgo/middleware` | HTTP tracing middleware (Echo/net/http) |

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

### PubSub Module

Provides Kafka producer and consumer with OTEL tracing.

```go
import "github.com/quiqupltd/quiqupgo/pubsub"

// What it provides via fx:
// - pubsub.Producer
// - pubsub.Consumer

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
    fx.Provide(func() pubsub.Config {
        return &pubsub.StandardConfig{
            Brokers:       []string{"kafka:9092"},
            ConsumerGroup: "my-service",
            Topics:        []string{"events"},
        }
    }),
    pubsub.Module(),
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
- [OrbStack](https://orbstack.dev/) (recommended for local Docker)
- [direnv](https://direnv.net/) (for automatic environment loading)

### Setup

```bash
# Install Go and Task via asdf
asdf install

# Install development tools
task tools:install

# Setup git hooks
task tools:hooks

# Allow direnv to load .env file
direnv allow

# Copy example env (already done, but for reference)
cp .env.example .env

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

With OrbStack, services are accessible via `.orb.local` URLs (no port mapping needed):
- `postgres.quiqupgo.orb.local:5432`
- `redpanda.quiqupgo.orb.local:9092`
- `temporal.quiqupgo.orb.local:7233`
- `jaeger.quiqupgo.orb.local:16686` (trace UI)
- `redpanda-console.quiqupgo.orb.local:8080` (Kafka UI)

### CI/CD Integration Tests

For GitHub Actions, override the environment variables:

```yaml
env:
  KAFKA_BROKER: localhost:9092
  POSTGRES_HOST: localhost
  TEMPORAL_HOST: localhost:7233
```

### Available Tasks

```bash
task --list
```

## Project Structure

```
quiqupgo/
├── tracing/          # OpenTelemetry tracing + metrics module
├── logger/           # Structured logging module
├── temporal/         # Temporal workflow client module
├── gormfx/           # GORM database module
├── pubsub/           # Kafka/PubSub messaging module
├── middleware/       # HTTP middleware
├── fxutil/           # Shared utilities
├── examples/         # Example applications
│   ├── minimal/      # Logger only
│   ├── api-service/  # Tracing + Logger + GORM
│   └── worker-service/ # All modules
└── docs/             # Documentation
```

## License

MIT License - see LICENSE file for details.
