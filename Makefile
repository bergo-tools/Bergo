# Makefile for bergo project
# Uses Zig for CGO cross-compilation

# Project settings
BINARY_NAME := bergo
MAIN_PACKAGE := ./main.go
VERSION ?= $(shell git describe --tags 2>/dev/null || echo "v0.0.0")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
SDK_PATH = $(shell xcrun --show-sdk-path)

# Go build flags
LDFLAGS := -ldflags "-X bergo/version.Version=$(VERSION) -X bergo/version.BuildTime=$(BUILD_TIME) -X bergo/version.CommitHash=$(COMMIT_HASH)"
GO_BUILD_FLAGS := -trimpath

# Default target
.PHONY: all
all: build

# Build for current platform (no CGO cross-compile needed)
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -rf dist/

# Cross-compile for all platforms
.PHONY: cross
cross: linux-amd64 linux-arm64 windows-amd64 windows-arm64 darwin-amd64 darwin-arm64


# macOS amd64
.PHONY: darwin-amd64
darwin-amd64:
	@echo "Building for darwin/amd64..."
	@mkdir -p dist
	CC="clang -arch x86_64 -isysroot ${SDK_PATH}" CXX="clang++ -arch x86_64 -isysroot ${SDK_PATH}" \
	CGO_ENABLED=1 \
	GOOS=darwin \
	GOARCH=amd64 \
	go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)

# macOS arm64
.PHONY: darwin-arm64
darwin-arm64:
	@echo "Building for darwin/arm64..."
	@mkdir -p dist
	CC="clang -arch arm64 -isysroot ${SDK_PATH}" CXX="clang++ -arch arm64 -isysroot ${SDK_PATH}" \
	CGO_ENABLED=1 \
	GOOS=darwin \
	GOARCH=arm64 \
	go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)

# Linux amd64
.PHONY: linux-amd64
linux-amd64:
	@echo "Building for linux/amd64..."
	@mkdir -p dist
	CC="zig cc -target x86_64-linux-musl" \
	CXX="zig c++ -target x86_64-linux-musl" \
	CGO_ENABLED=1 \
	GOOS=linux \
	GOARCH=amd64 \
	go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)

# Linux arm64
.PHONY: linux-arm64
linux-arm64:
	@echo "Building for linux/arm64..."
	@mkdir -p dist
	CC="zig cc -target aarch64-linux-musl" \
	CXX="zig c++ -target aarch64-linux-musl" \
	CGO_ENABLED=1 \
	GOOS=linux \
	GOARCH=arm64 \
	go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)

# Windows amd64
.PHONY: windows-amd64
windows-amd64:
	@echo "Building for windows/amd64..."
	@mkdir -p dist
	CC="zig cc -target x86_64-windows-gnu" \
	CXX="zig c++ -target x86_64-windows-gnu" \
	CGO_ENABLED=1 \
	GOOS=windows \
	GOARCH=amd64 \
	go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

# Windows arm64
.PHONY: windows-arm64
windows-arm64:
	@echo "Building for windows/arm64..."
	@mkdir -p dist
	CC="zig cc -target aarch64-windows-gnu" \
	CXX="zig c++ -target aarch64-windows-gnu" \
	CGO_ENABLED=1 \
	GOOS=windows \
	GOARCH=arm64 \
	go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-arm64.exe $(MAIN_PACKAGE)

# Quick cross-compile for common platforms
.PHONY: darwin
darwin: darwin-amd64 darwin-arm64

.PHONY: linux
linux: linux-amd64 linux-arm64

.PHONY: windows
windows: windows-amd64 windows-arm64

# Run the application
.PHONY: run
run:
	go run $(MAIN_PACKAGE)

# Install dependencies
.PHONY: deps
deps:
	go mod download
	go mod tidy

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all/build      - Build for current platform"
	@echo "  clean          - Remove build artifacts"
	@echo "  cross          - Cross-compile for all platforms"
	@echo "  darwin         - Cross-compile for macOS (amd64 + arm64)"
	@echo "  linux          - Cross-compile for Linux (amd64 + arm64)"
	@echo "  windows        - Cross-compile for Windows (amd64 + arm64)"
	@echo "  run            - Run the application"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Individual platform targets:"
	@echo "  darwin-amd64   - macOS Intel"
	@echo "  darwin-arm64   - macOS Apple Silicon"
	@echo "  linux-amd64    - Linux x86_64"
	@echo "  linux-arm64    - Linux ARM64"
	@echo "  windows-amd64  - Windows x86_64"
	@echo "  windows-arm64  - Windows ARM64"
