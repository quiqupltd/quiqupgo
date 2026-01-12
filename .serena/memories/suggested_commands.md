# Suggested Commands

## Setup
```bash
# Install Go and Task via asdf
asdf install

# Install development tools (golangci-lint, goimports, godoc)
task tools:install

# Install git pre-commit hook for linting
task tools:hooks

# Allow direnv to load .env file
direnv allow
```

## Development
```bash
# Run all unit tests
task test:unit

# Run tests without verbose output
task test:short

# Test a specific module
task test:module MODULE=tracing
task test:module MODULE=logger
task test:module MODULE=kafka

# Generate coverage report (creates coverage.html)
task coverage
```

## Linting & Formatting
```bash
# Run linters
task tools:lint

# Run linters with auto-fix
task tools:lint-fix

# Format code (go fmt + goimports)
task tools:fmt

# Tidy dependencies
task tools:tidy
```

## Building
```bash
# Build all packages
task build:default

# Build all examples
task build:examples
```

## Integration Tests
```bash
# Start infrastructure (Postgres, Kafka/Redpanda, Temporal, Jaeger)
task docker:up

# Wait for services to be healthy
task docker:status

# Run integration tests
task test:integration

# Run all tests (unit + integration)
task test:all

# Stop infrastructure
task docker:down

# View service logs
task docker:logs
```

## Full Verification
```bash
# Run all checks (tidy, fmt, lint, test, build)
task verify

# CI pipeline checks
task ci
```

## Documentation
```bash
# Serve godoc locally at http://localhost:6060
task tools:docs
```

## System Commands
Standard Linux/Unix commands are available:
- `git` - Version control
- `ls`, `cd`, `grep`, `find` - File system navigation
- `docker`, `docker compose` - Container management
