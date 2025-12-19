# Makefile for bergo project
# Uses Zig for CGO cross-compilation

# Project settings
BINARY_NAME := bergo
MAIN_PACKAGE := ./main.go
VERSION ?= $(shell git describe --tags 2>/dev/null || echo "v0.0.0")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go build flags
LDFLAGS := -ldflags "-X bergo/version.Version=$(VERSION) -X bergo/version.BuildTime=$(BUILD_TIME) -X bergo/version.CommitHash=$(COMMIT_HASH)"
GO_BUILD_FLAGS := -trimpath

# Zig settings for CGO cross-compilation
# Zig acts as a drop-in replacement for gcc/clang
export CC = zig cc
export CXX = zig c++

# Target platforms for cross-compilation
# Format: GOOS/GOARCH/Zig-target
PLATFORMS := \
	darwin/amd64/x86_64-macos \
	darwin/arm64/aarch64-macos \
	linux/amd64/x86_64-linux-musl \
	linux/arm64/aarch64-linux-musl \
	windows/amd64/x86_64-windows \
	windows/arm64/aarch64-windows

# Default target
.PHONY: all
all: build

# Build for current platform
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
cross: $(PLATFORMS)

# Pattern rule for cross-compilation
define cross-compile
.PHONY: $(1)
$(1):
	@echo "Building for $(1)..."
	@mkdir -p dist
	$(eval GOOS := $(word 1,$(subst /, ,$(1))))
	$(eval GOARCH := $(word 2,$(subst /, ,$(1))))
	$(eval ZIG_TARGET := $(word 3,$(subst /, ,$(1))))
	@echo "  GOOS=$(GOOS) GOARCH=$(GOARCH) ZIG_TARGET=$(ZIG_TARGET)"
	CC="zig cc -target $(ZIG_TARGET)" \
	CXX="zig c++ -target $(ZIG_TARGET)" \
	CGO_ENABLED=1 \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	go build $(GO_BUILD_FLAGS) $(LDFLAGS) -o dist/$(BINARY_NAME)-$(GOOS)-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe,) $(MAIN_PACKAGE)
endef

# Generate cross-compilation targets
$(foreach platform,$(PLATFORMS),$(eval $(call cross-compile,$(platform))))

# Quick cross-compile for common platforms
.PHONY: darwin
darwin: darwin/amd64 darwin/arm64

.PHONY: linux
linux: linux/amd64 linux/arm64

.PHONY: windows
windows: windows/amd64 windows/arm64

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
	@echo "  all/build    - Build for current platform"
	@echo "  clean        - Remove build artifacts"
	@echo "  cross        - Cross-compile for all platforms"
	@echo "  darwin       - Cross-compile for macOS (amd64 + arm64)"
	@echo "  linux        - Cross-compile for Linux (amd64 + arm64)"
	@echo "  windows      - Cross-compile for Windows (amd64 + arm64)"
	@echo "  run          - Run the application"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  help         - Show this help"
	@echo ""
	@echo "Individual platform targets:"
	@for platform in $(PLATFORMS); do \
		platform_name=$$(echo $$platform | cut -d'/' -f1,2); \
		echo "  $$platform_name"; \
	done
