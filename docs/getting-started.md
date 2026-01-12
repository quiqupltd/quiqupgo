# Getting Started

This guide will help you set up quiqupgo modules in your application.

## Installation

```bash
go get github.com/quiqupltd/quiqupgo
```

## Quick Start

### Minimal Setup (Logger Only)

The simplest setup uses only the logger module:

```go
package main

import (
    "github.com/quiqupltd/quiqupgo/logger"
    "go.uber.org/fx"
    "go.uber.org/zap"
)

func main() {
    fx.New(
        // Provide logger configuration
        fx.Provide(func() logger.Config {
            return &logger.StandardConfig{
                ServiceName: "my-service",
                Environment: "development",
            }
        }),

        // Include the logger module
        logger.Module(),

        // Your application code
        fx.Invoke(func(log *zap.Logger) {
            log.Info("Hello, world!")
        }),
    ).Run()
}
```

### API Service Setup

For an API service with tracing, logging, and database:

```go
package main

import (
    "database/sql"

    "github.com/quiqupltd/quiqupgo/gormfx"
    "github.com/quiqupltd/quiqupgo/logger"
    "github.com/quiqupltd/quiqupgo/tracing"
    "go.uber.org/fx"
)

func main() {
    fx.New(
        // Provide configurations
        fx.Provide(
            newTracingConfig,
            newLoggerConfig,
            newGormConfig,
        ),

        // Include modules
        tracing.Module(),
        logger.Module(),
        gormfx.Module(),

        // Your application code
        fx.Invoke(runServer),
    ).Run()
}

func newTracingConfig() tracing.Config {
    return &tracing.StandardConfig{
        ServiceName:     "my-api",
        EnvironmentName: "production",
        OTLPEndpoint:    "otel-collector:4318",
    }
}

func newLoggerConfig() logger.Config {
    return &logger.StandardConfig{
        ServiceName: "my-api",
        Environment: "production",
    }
}

func newGormConfig(db *sql.DB) gormfx.Config {
    return &gormfx.StandardConfig{
        DB:           db,
        MaxOpenConns: 25,
        MaxIdleConns: 5,
    }
}
```

### Worker Service Setup

For a worker service with Temporal and Kafka:

```go
package main

import (
    "github.com/quiqupltd/quiqupgo/logger"
    "github.com/quiqupltd/quiqupgo/pubsub"
    "github.com/quiqupltd/quiqupgo/temporal"
    "github.com/quiqupltd/quiqupgo/tracing"
    "go.uber.org/fx"
)

func main() {
    fx.New(
        // Provide configurations
        fx.Provide(
            newTracingConfig,
            newLoggerConfig,
            newTemporalConfig,
            newPubSubConfig,
        ),

        // Include modules
        tracing.Module(),
        logger.Module(),
        temporal.Module(),
        pubsub.Module(),

        // Your worker code
        fx.Invoke(runWorker),
    ).Run()
}

func newTemporalConfig() temporal.Config {
    return &temporal.StandardConfig{
        HostPort:  "temporal:7233",
        Namespace: "my-namespace",
    }
}

func newPubSubConfig() pubsub.Config {
    return &pubsub.StandardConfig{
        Brokers:       []string{"kafka:9092"},
        ConsumerGroup: "my-worker",
    }
}
```

## Module Dependencies

Some modules depend on others:

| Module | Dependencies |
|--------|-------------|
| `tracing` | None |
| `logger` | None (optional: `tracing` for OTEL integration) |
| `temporal` | `logger`, `tracing` |
| `gormfx` | `tracing` |
| `pubsub` | `logger`, `tracing` |
| `middleware` | `tracing` |

## Testing

Each module provides test utilities in its `testutil` subpackage:

```go
package myservice_test

import (
    "testing"

    "github.com/quiqupltd/quiqupgo/fxutil"
    "github.com/quiqupltd/quiqupgo/logger/testutil"
    tracingtest "github.com/quiqupltd/quiqupgo/tracing/testutil"
    "go.uber.org/fx"
)

func TestMyService(t *testing.T) {
    var svc *MyService

    app := fxutil.TestApp(t,
        tracingtest.NoopModule(),
        testutil.NoopModule(),
        fx.Provide(NewMyService),
        fx.Populate(&svc),
    )

    app.RequireStart()
    defer app.RequireStop()

    // Test your service
}
```

## Next Steps

- Read the [Configuration Guide](configuration.md) for detailed configuration options
- Check the [Migration Guide](migration-guide.md) if migrating from an existing setup
- See the `examples/` directory for complete working examples
