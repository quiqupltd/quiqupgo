# Migration Guide

This guide helps you migrate from existing fx-based setups to quiqupgo modules.

## Overview

The key change from typical fx setups is moving from **monolithic global configuration** to **per-module configuration interfaces**.

### Before (Monolithic Config)

```go
// Old pattern: One global config struct
type GlobalConfig struct {
    ServiceName     string
    Environment     string
    OTLPEndpoint    string
    LogLevel        string
    TemporalHost    string
    TemporalNS      string
    DBHost          string
    DBName          string
    KafkaBrokers    []string
    // ... dozens more fields
}

// Single config provider
fx.Provide(func() *GlobalConfig {
    return loadConfig()
})

// Modules reach into global config
fx.Provide(func(cfg *GlobalConfig) *zap.Logger {
    return createLogger(cfg.ServiceName, cfg.LogLevel)
})
```

### After (Per-Module Config)

```go
// New pattern: Each module gets its own config
fx.Provide(
    newTracingConfig,
    newLoggerConfig,
    newTemporalConfig,
    newGormConfig,
)

// Each config function adapts from your app config
func newTracingConfig(app *AppConfig) tracing.Config {
    return &tracing.StandardConfig{
        ServiceName:     app.ServiceName,
        EnvironmentName: app.Environment,
        OTLPEndpoint:    app.Observability.OTLPEndpoint,
    }
}
```

## Step-by-Step Migration

### Step 1: Add quiqupgo Dependency

```bash
go get github.com/quiqupltd/quiqupgo
```

### Step 2: Identify Current Modules

List the fx modules you currently use:
- [ ] Tracing/Telemetry
- [ ] Logging
- [ ] Database (GORM)
- [ ] Temporal
- [ ] Kafka/PubSub
- [ ] HTTP Middleware

### Step 3: Create Config Adapters

For each module, create a function that adapts your existing config:

```go
// Keep your existing AppConfig
type AppConfig struct {
    Name        string
    Environment string
    Telemetry   TelemetryConfig
    Database    DatabaseConfig
    // ...
}

// Create adapters for each module
func newTracingConfig(app *AppConfig) tracing.Config {
    return &tracing.StandardConfig{
        ServiceName:     app.Name,
        EnvironmentName: app.Environment,
        OTLPEndpoint:    app.Telemetry.OTLPEndpoint,
    }
}

func newLoggerConfig(app *AppConfig) logger.Config {
    return &logger.StandardConfig{
        ServiceName: app.Name,
        Environment: app.Environment,
    }
}

func newGormConfig(app *AppConfig, sqlDB *sql.DB) gormfx.Config {
    return &gormfx.StandardConfig{
        DB:            sqlDB,
        MaxOpenConns:  app.Database.MaxOpenConns,
        MaxIdleConns:  app.Database.MaxIdleConns,
    }
}
```

### Step 4: Replace Old Modules

Replace your custom fx modules with quiqupgo modules:

```go
// Before
func Module() fx.Option {
    return fx.Module("global",
        fx.Provide(provideLogger),
        fx.Provide(provideTracer),
        fx.Provide(provideDB),
    )
}

// After
func Module() fx.Option {
    return fx.Module("global",
        // Load your app config
        fx.Provide(loadAppConfig),

        // Adapt to module configs
        fx.Provide(
            newTracingConfig,
            newLoggerConfig,
            newGormConfig,
        ),

        // Use quiqupgo modules
        tracing.Module(),
        logger.Module(),
        gormfx.Module(),
    )
}
```

### Step 5: Update Tests

Replace test mocks with quiqupgo test utilities:

```go
// Before
func TestMyService(t *testing.T) {
    app := fx.New(
        fx.Provide(mockLogger),
        fx.Provide(mockTracer),
        // ...
    )
}

// After
func TestMyService(t *testing.T) {
    app := fxutil.TestApp(t,
        tracingtest.NoopModule(),
        loggertest.NoopModule(),
        // ...
    )
}
```

### Step 6: Update HTTP Middleware

Replace custom tracing middleware:

```go
// Before
e.Use(customTracingMiddleware(tracer))

// After
e.Use(middleware.EchoTracing(tracerProvider, "my-service"))
```

## Common Migration Patterns

### Pattern 1: Keeping Existing Config Loader

```go
// Your existing config loading
func loadConfig() *AppConfig {
    cfg := &AppConfig{}
    viper.Unmarshal(cfg)
    return cfg
}

// New: provide both app config and module configs
fx.Provide(loadConfig),
fx.Provide(func(cfg *AppConfig) tracing.Config {
    return adaptToTracingConfig(cfg)
}),
```

### Pattern 2: Environment Variables

```go
func newTracingConfig() tracing.Config {
    return &tracing.StandardConfig{
        ServiceName:  os.Getenv("SERVICE_NAME"),
        OTLPEndpoint: os.Getenv("OTLP_ENDPOINT"),
    }
}
```

### Pattern 3: Multiple Environments

```go
func newLoggerConfig() logger.Config {
    env := os.Getenv("ENVIRONMENT")
    return &logger.StandardConfig{
        ServiceName: os.Getenv("SERVICE_NAME"),
        Environment: env, // "development" = console, else = JSON
    }
}
```

## Troubleshooting

### Error: Missing Config Dependency

```
fx.New failed: missing dependency: tracing.Config
```

**Solution**: Add a provider for the module config:
```go
fx.Provide(func() tracing.Config {
    return &tracing.StandardConfig{...}
})
```

### Error: Type Mismatch

```
cannot use &MyConfig{} as tracing.Config
```

**Solution**: Ensure your config implements all interface methods.

### Error: Circular Dependency

**Solution**: Use `fx.Annotate` to break cycles:
```go
fx.Provide(
    fx.Annotate(
        newTracingConfig,
        fx.As(new(tracing.Config)),
    ),
)
```

## Gradual Migration

You can migrate module by module:

1. **Week 1**: Migrate tracing + logger
2. **Week 2**: Migrate GORM
3. **Week 3**: Migrate Temporal + PubSub
4. **Week 4**: Migrate middleware + update tests

Each step can be merged independently.

## Getting Help

- Check the `examples/` directory for working code
- Review [Configuration Guide](configuration.md) for all options
- Open an issue for migration problems
