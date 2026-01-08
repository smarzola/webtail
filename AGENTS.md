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
- **Simplified config**: Removed redundant `tailnet_domain` field (determined by auth key)
- **Node names**: Use simple hostnames (e.g., "plex") instead of full domain names
- **Forwarder options**: Added configurable `pass_host_header` and `trust_forward_header` per service
- **Security defaults**: Both forwarder options default to `false` for security
- **Tailscale integration**: Uses `server.Up()` for proper domain access and certificate handling

## Docker Integration
- **Enable Docker mode**: Use `-docker` flag to enable Docker container discovery
- **Docker network**: Configure `docker.network` in config.json (required for Docker mode)
- **Docker client config**: Optional `docker.host`, `docker.api_version`, `docker.cert_path`, `docker.tls_verify` in config.json
- **Docker labels**: Containers must have `webtail.enabled=true` to be proxied
- **Required labels**: None (only `webtail.enabled=true`)
- **Optional labels**: `webtail.node_name` (defaults to container name), `webtail.port` (auto-detected from lowest exposed port), `webtail.protocol` (default: http), `webtail.pass_host_header`, `webtail.trust_forward_header` (both default to false)
- **Port auto-detection**: If `webtail.port` is not set, uses lowest exposed port (e.g., 80 preferred over 8080)
- **Node name auto-detection**: If `webtail.node_name` is not set, uses the container name
- **Dynamic target**: Target URL built as `{protocol}://{container_name}.{docker_network}:{port}`
- **Lifecycle management**: Proxies are automatically created/removed when containers start/stop
- **Existing containers**: On startup, webtail scans running containers for webtail labels
- **Combined mode**: Can use both config file and Docker discovery simultaneously
- **Docker env vars**: `DOCKER_HOST` (server URL), `DOCKER_API_VERSION` (API version), `DOCKER_CERT_PATH` (TLS certs dir), `DOCKER_TLS_VERIFY` (enable TLS verification)

## Development Workflow
- **Lint**: No specific linter configured, use `go vet ./...` for basic checks
- **Format**: Run `gofmt -w .` before committing
- **Version**: Set via `-ldflags="-X main.version=v1.0.0"` at build time