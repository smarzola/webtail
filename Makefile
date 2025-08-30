# Makefile for webtail
.PHONY: build build-all clean test deps help

# Build for current platform
build:
	go build -o webtail .

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o webtail-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o webtail-linux-arm64 .
	GOOS=darwin GOARCH=arm64 go build -o webtail-darwin-arm64 .

# Build with version info
build-version:
	go build -ldflags="-X main.version=$(VERSION)" -o webtail .

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
	@echo "  build         - Build for current platform"
	@echo "  build-all     - Build for all supported platforms"
	@echo "  build-version - Build with version info (set VERSION variable)"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run tests"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  version       - Show current version"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make VERSION=v1.0.0 build-version"
	@echo "  make build-all"