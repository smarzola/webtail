# Webtail Agent Guidelines

## Build Commands
- **Build**: `go build -o webtail .` or `make build`
- **Cross-platform**: `make build-all` (Linux AMD64/ARM64, macOS ARM64)
- **Release**: `make VERSION=v1.0.0 release-archives`
- **Dependencies**: `go mod download && go mod tidy` or `make deps`

## Test Commands
- **Run all tests**: `go test ./...` or `make test`
- **Run single test**: `go test -run TestName ./...`
- **With coverage**: `go test -cover ./...`

## Code Style Guidelines
- **Formatting**: Use `gofmt` (standard Go formatting)
- **Imports**: Group standard library first, then third-party packages
- **Naming**: PascalCase for exported identifiers, camelCase for unexported
- **Error handling**: Use `fmt.Errorf` with `%w` verb for error wrapping
- **Struct tags**: Use proper JSON tags for configuration structs
- **Concurrency**: Use `context.Context` for cancellation, `sync.WaitGroup` for coordination
- **Logging**: Use `log.Printf` for consistent logging format
- **Cleanup**: Use `defer` statements for resource cleanup
- **URL handling**: Parse target URLs properly to support http/https schemes

## Configuration Changes
- **Target field**: Use `target` instead of `upstream_host` to support full URLs with schemes
- **URL parsing**: Always parse target URLs and default to http if no scheme provided
- **Validation**: Ensure target URLs are properly formatted

## Development Workflow
- **Lint**: No specific linter configured, use `go vet ./...` for basic checks
- **Format**: Run `gofmt -w .` before committing
- **Version**: Set via `-ldflags="-X main.version=v1.0.0"` at build time