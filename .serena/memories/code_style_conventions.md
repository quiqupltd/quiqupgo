# Code Style & Conventions

## Formatting
- Use `go fmt` for base formatting
- Use `goimports` for import organization (standard → external → internal)
- Run `task tools:fmt` before committing

## Linting (golangci-lint)
Enabled linters include:
- errcheck, gosimple, govet, ineffassign, staticcheck, unused
- gofmt, goimports, misspell
- gocritic, revive (with specific rules)
- gosec (security), nilerr, bodyclose, errorlint
- exhaustive, noctx, prealloc, predeclared, whitespace

## Module Pattern
Each fx module follows this structure:
```
module/
├── config.go       # Config interface + StandardConfig implementation
├── module.go       # Module() returns fx.Option with providers and lifecycle hooks
├── doc.go          # Package documentation
├── *_test.go       # Unit tests
├── integration_test.go  # Integration tests (build tag: integration)
└── testutil/       # Test helpers (NoopModule, MockModule, etc.)
```

## Config Interface Pattern
Each module defines a Config interface with getter methods:
```go
type Config interface {
    GetServiceName() string
    GetEnvironmentName() string
    // ...
}

type StandardConfig struct {
    ServiceName     string
    EnvironmentName string
    // ...
}

func (c *StandardConfig) GetServiceName() string { return c.ServiceName }
```

## fx Module Pattern
```go
func Module(opts ...ModuleOption) fx.Option {
    options := defaultModuleOptions()
    for _, opt := range opts {
        opt(options)
    }
    return fx.Module("modulename",
        fx.Supply(options),
        fx.Provide(...),
        fx.Invoke(registerLifecycleHooks),
    )
}
```

## Testing
- Use `testify/assert` and `testify/require` for assertions
- Each module provides testutil helpers:
  - `NoopModule()` - No-op implementation
  - `MockModule()` - Mock implementation
  - `BufferModule()` - Captures output for assertions
- Integration tests use build tag `//go:build integration`

## Naming Conventions
- Config interfaces: `Config`
- Standard implementations: `StandardConfig`
- Module functions: `Module()`
- Provider functions: `provide*` or `new*`
- Test utilities in `testutil/` subdirectory
