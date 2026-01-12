# Configuration Guide

This guide covers all configuration options for each module.

## Configuration Pattern

Each module uses a **Config interface** pattern:

1. Module defines a `Config` interface
2. Module provides a `StandardConfig` struct implementing the interface
3. Your app can use `StandardConfig` or implement custom config logic

```go
// Using StandardConfig
fx.Provide(func() tracing.Config {
    return &tracing.StandardConfig{
        ServiceName: "my-service",
    }
})

// Using custom config (e.g., from environment/Viper)
fx.Provide(func(appCfg *AppConfig) tracing.Config {
    return &myTracingConfig{appCfg: appCfg}
})
```

## Tracing Module

### Interface

```go
type Config interface {
    GetServiceName() string
    GetEnvironmentName() string
    GetOTLPEndpoint() string
    GetOTLPInsecure() bool
    GetOTLPTLSCert() string   // base64/PEM encoded
    GetOTLPTLSKey() string    // base64/PEM encoded
    GetOTLPTLSCA() string     // base64/PEM encoded
}
```

### StandardConfig Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ServiceName` | string | `""` | Service name for spans |
| `EnvironmentName` | string | `""` | Deployment environment |
| `OTLPEndpoint` | string | `""` | OTLP collector endpoint (empty = disabled) |
| `OTLPInsecure` | bool | `false` | Use insecure connection |
| `OTLPTLSCert` | string | `""` | TLS certificate (PEM) |
| `OTLPTLSKey` | string | `""` | TLS private key (PEM) |
| `OTLPTLSCA` | string | `""` | TLS CA certificate (PEM) |

### Example

```go
&tracing.StandardConfig{
    ServiceName:     "order-service",
    EnvironmentName: "production",
    OTLPEndpoint:    "otel-collector.observability:4318",
    OTLPInsecure:    false,
}
```

## Logger Module

### Interface

```go
type Config interface {
    GetServiceName() string
    GetEnvironment() string
}
```

### StandardConfig Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ServiceName` | string | `""` | Service name for log entries |
| `Environment` | string | `""` | Environment (`development`/`local` = console, else JSON) |

### Example

```go
&logger.StandardConfig{
    ServiceName: "order-service",
    Environment: "production",  // JSON output
}

&logger.StandardConfig{
    ServiceName: "order-service",
    Environment: "development",  // Human-readable console output
}
```

## Temporal Module

### Interface

```go
type Config interface {
    GetHostPort() string
    GetNamespace() string
    GetTLSCert() string  // PEM encoded
    GetTLSKey() string   // PEM encoded
}
```

### StandardConfig Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `HostPort` | string | `localhost:7233` | Temporal server address |
| `Namespace` | string | `default` | Temporal namespace |
| `TLSCert` | string | `""` | TLS certificate (PEM) |
| `TLSKey` | string | `""` | TLS private key (PEM) |

### Example

```go
&temporal.StandardConfig{
    HostPort:  "temporal.internal:7233",
    Namespace: "my-namespace",
    TLSCert:   os.Getenv("TEMPORAL_TLS_CERT"),
    TLSKey:    os.Getenv("TEMPORAL_TLS_KEY"),
}
```

## GORM Module

### Interface

```go
type Config interface {
    GetDB() *sql.DB
    GetMaxOpenConns() int
    GetMaxIdleConns() int
    GetEnableTracing() bool
}
```

### StandardConfig Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `DB` | `*sql.DB` | required | Underlying database connection |
| `MaxOpenConns` | int | `0` (GORM default) | Max open connections |
| `MaxIdleConns` | int | `0` (GORM default) | Max idle connections |
| `EnableTracing` | `*bool` | `true` | Enable OTEL tracing |

### Example

```go
// The sql.DB is typically created elsewhere
sqlDB, _ := sql.Open("postgres", dsn)

&gormfx.StandardConfig{
    DB:            sqlDB,
    MaxOpenConns:  25,
    MaxIdleConns:  5,
    EnableTracing: ptr(true),
}
```

## PubSub Module

### Interface

```go
type Config interface {
    GetBrokers() []string
    GetConsumerGroup() string
    GetProducerTimeout() time.Duration
    GetConsumerTimeout() time.Duration
    GetEnableTracing() bool
    GetTLSEnabled() bool
    GetTLSCert() string
    GetTLSKey() string
    GetTLSCA() string
    GetSASLEnabled() bool
    GetSASLMechanism() string
    GetSASLUsername() string
    GetSASLPassword() string
}
```

### StandardConfig Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Brokers` | `[]string` | `["localhost:9092"]` | Kafka broker addresses |
| `ConsumerGroup` | string | `"default"` | Consumer group ID |
| `ProducerTimeout` | `time.Duration` | `10s` | Producer timeout |
| `ConsumerTimeout` | `time.Duration` | `10s` | Consumer timeout |
| `EnableTracing` | `*bool` | `true` | Enable OTEL tracing |
| `TLSEnabled` | bool | `false` | Enable TLS |
| `TLSCert` | string | `""` | TLS certificate |
| `TLSKey` | string | `""` | TLS private key |
| `TLSCA` | string | `""` | TLS CA certificate |
| `SASLEnabled` | bool | `false` | Enable SASL auth |
| `SASLMechanism` | string | `"PLAIN"` | SASL mechanism |
| `SASLUsername` | string | `""` | SASL username |
| `SASLPassword` | string | `""` | SASL password |

### Example

```go
&pubsub.StandardConfig{
    Brokers:       []string{"kafka1:9092", "kafka2:9092"},
    ConsumerGroup: "order-processor",
    TLSEnabled:    true,
    SASLEnabled:   true,
    SASLMechanism: "SCRAM-SHA-256",
    SASLUsername:  os.Getenv("KAFKA_USER"),
    SASLPassword:  os.Getenv("KAFKA_PASS"),
}
```

## HTTP Middleware

The middleware package doesn't require fx configuration. Use it directly:

```go
// Echo
e := echo.New()
e.Use(middleware.EchoTracing(tracerProvider, "my-service",
    middleware.WithSkipPaths("/health", "/ready"),
))

// net/http
handler := middleware.HTTPTracing(tracerProvider, "my-service")(mux)
```

### Options

| Option | Description |
|--------|-------------|
| `WithSkipPaths(paths...)` | Skip tracing for specific paths |
| `WithPropagator(p)` | Custom trace context propagator |

## Environment-Based Configuration

Example of loading config from environment:

```go
type AppConfig struct {
    ServiceName string `env:"SERVICE_NAME"`
    Environment string `env:"ENVIRONMENT"`
    OTLPEndpoint string `env:"OTLP_ENDPOINT"`
    // ... other fields
}

// Adapt to module configs
fx.Provide(
    fx.Annotate(
        func(cfg *AppConfig) tracing.Config {
            return &tracing.StandardConfig{
                ServiceName:     cfg.ServiceName,
                EnvironmentName: cfg.Environment,
                OTLPEndpoint:    cfg.OTLPEndpoint,
            }
        },
        fx.As(new(tracing.Config)),
    ),
)
```
