# Makefile for webtail
.PHONY: build build-all clean test deps docker-build docker-build-all docker-push help

# Build for current platform
build:
	go build -o webtail .

# Build for all platforms
build-all:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -o dist/webtail-linux-amd64/webtail .
	GOOS=linux GOARCH=arm64 go build -o dist/webtail-linux-arm64/webtail .
	GOOS=darwin GOARCH=arm64 go build -o dist/webtail-darwin-arm64/webtail .

# Build with version info
build-version:
	go build -ldflags="-X main.version=$(VERSION)" -o webtail .

# Create release archives (requires VERSION to be set)
release-archives:
	@if [ -z "$(VERSION)" ]; then echo "Please set VERSION variable (e.g., make VERSION=v1.0.0 release-archives)"; exit 1; fi
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o dist/webtail-linux-amd64/webtail .
	GOOS=linux GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION)" -o dist/webtail-linux-arm64/webtail .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION)" -o dist/webtail-darwin-arm64/webtail .
	cd dist && for dir in */; do if [ -d "$$dir" ]; then tar -czf "$${dir%/}-$(VERSION).tar.gz" -C "$$dir" .; fi; done

# Build Docker image for current platform
docker-build:
	docker build --build-arg VERSION=$(or $(VERSION),dev) -t ghcr.io/smarzola/webtail:$(or $(VERSION),dev) .

# Build multi-arch Docker image (linux/amd64 + linux/arm64)
docker-build-all:
	@if [ -z "$(VERSION)" ]; then echo "Please set VERSION variable (e.g., make VERSION=v1.0.0 docker-build-all)"; exit 1; fi
	docker buildx build --platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		-t ghcr.io/smarzola/webtail:$(VERSION) \
		-t ghcr.io/smarzola/webtail:latest \
		.

# Push Docker image to ghcr.io (requires docker login ghcr.io first)
docker-push:
	@if [ -z "$(VERSION)" ]; then echo "Please set VERSION variable (e.g., make VERSION=v1.0.0 docker-push)"; exit 1; fi
	docker buildx build --platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		-t ghcr.io/smarzola/webtail:$(VERSION) \
		-t ghcr.io/smarzola/webtail:latest \
		--push \
		.

# Clean build artifacts
clean:
	rm -f webtail webtail-*

# Run tests
test:
	go test ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Show version
version:
	./webtail --version

# Show help
help:
	@echo "Available targets:"
	@echo "  build             - Build for current platform"
	@echo "  build-all         - Build for all supported platforms"
	@echo "  build-version     - Build with version info (set VERSION variable)"
	@echo "  release-archives  - Build and create release archives (requires VERSION)"
	@echo "  clean             - Remove build artifacts"
	@echo "  test              - Run tests"
	@echo "  deps              - Download and tidy dependencies"
	@echo "  version           - Show current version"
	@echo "  docker-build      - Build Docker image for current platform"
	@echo "  docker-build-all  - Build multi-arch Docker image (amd64 + arm64)"
	@echo "  docker-push       - Build and push multi-arch image to ghcr.io (requires VERSION)"
	@echo "  help              - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make VERSION=v1.0.0 build-version"
	@echo "  make VERSION=v1.0.0 release-archives"
	@echo "  make build-all"
	@echo "  make VERSION=v1.0.0 docker-push"