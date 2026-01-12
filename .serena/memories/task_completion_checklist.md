# Task Completion Checklist

When completing a task in this codebase, run the following checks:

## 1. Format Code
```bash
task tools:fmt
```
Runs `go fmt` and `goimports` to ensure consistent formatting.

## 2. Tidy Dependencies
```bash
task tools:tidy
```
Runs `go mod tidy` to clean up go.mod and go.sum.

## 3. Run Linter
```bash
task tools:lint
```
Runs golangci-lint with the project's configuration. Fix any issues before proceeding.

For auto-fixable issues:
```bash
task tools:lint-fix
```

## 4. Run Unit Tests
```bash
task test:unit
```
Ensure all tests pass with race detection enabled.

For a specific module:
```bash
task test:module MODULE=<module_name>
```

## 5. Build Verification
```bash
task build:default
```
Ensure the code compiles successfully.

## Full Verification (All Steps)
```bash
task verify
```
Runs tidy → fmt → lint → test → build in sequence.

## Before Creating a PR
1. Run `task verify` to ensure all checks pass
2. If you modified integration test behavior, run `task docker:up && task test:integration`
3. Update documentation in `doc.go` files if adding new public APIs
4. Update README.md if adding new modules or changing usage patterns
