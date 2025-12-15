# SuperRay-TUI Makefile
# Cross-platform build with static linking support

# Application info
APP_NAME := superray-tui
VERSION := 1.0.0

# Go settings
GO := /Volumes/mindata/Library/go/bin/go
GOPATH := $(HOME)/go
CGO_ENABLED := 1

# Directories
PROJECT_DIR := $(shell pwd)
THIRD_PARTY := $(PROJECT_DIR)/third_party/superray
BUILD_DIR := $(PROJECT_DIR)/build
DIST_DIR := $(PROJECT_DIR)/dist

# SuperRay paths
SUPERRAY_INCLUDE := $(THIRD_PARTY)/include
SUPERRAY_GEOIP := $(THIRD_PARTY)/geoip

# Platform detection
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# Set default target based on current platform
ifeq ($(UNAME_S),Darwin)
    ifeq ($(UNAME_M),arm64)
        DEFAULT_TARGET := darwin-arm64
    else
        DEFAULT_TARGET := darwin-amd64
    endif
else ifeq ($(UNAME_S),Linux)
    ifeq ($(UNAME_M),aarch64)
        DEFAULT_TARGET := linux-arm64
    else
        DEFAULT_TARGET := linux-amd64
    endif
else
    DEFAULT_TARGET := windows-amd64
endif

# Targets
.PHONY: all build clean deps help run install native version
.PHONY: darwin-arm64 darwin-amd64 darwin-all darwin-universal
.PHONY: linux-amd64 linux-arm64 windows-amd64
.PHONY: package package-all package-darwin

# Default target
all: build

# Build for current platform
build: deps $(DEFAULT_TARGET)

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GO) mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)

# ============================================================
# macOS Builds (Dynamic Linking - portable with @executable_path)
# ============================================================

# Build for macOS ARM64
darwin-arm64: deps
	@echo "Building for macOS ARM64..."
	@mkdir -p $(BUILD_DIR)/darwin-arm64/lib $(BUILD_DIR)/darwin-arm64/geoip
	CGO_ENABLED=$(CGO_ENABLED) \
		GOOS=darwin \
		GOARCH=arm64 \
		CGO_CFLAGS="-I$(SUPERRAY_INCLUDE)" \
		CGO_LDFLAGS="-L$(THIRD_PARTY)/lib/macos/arm64 -lsuperray -Wl,-rpath,@executable_path/lib" \
		$(GO) build -o $(BUILD_DIR)/darwin-arm64/$(APP_NAME) .
	@cp $(THIRD_PARTY)/lib/macos/arm64/libsuperray.dylib $(BUILD_DIR)/darwin-arm64/lib/
	@install_name_tool -change libsuperray.dylib @executable_path/lib/libsuperray.dylib $(BUILD_DIR)/darwin-arm64/$(APP_NAME) 2>/dev/null || true
	@cp $(SUPERRAY_GEOIP)/* $(BUILD_DIR)/darwin-arm64/geoip/ 2>/dev/null || true
	@cp .env.example $(BUILD_DIR)/darwin-arm64/.env.example 2>/dev/null || true
	@echo "Built: $(BUILD_DIR)/darwin-arm64/$(APP_NAME)"

# Build for macOS AMD64
darwin-amd64: deps
	@echo "Building for macOS AMD64..."
	@mkdir -p $(BUILD_DIR)/darwin-amd64/lib $(BUILD_DIR)/darwin-amd64/geoip
	CGO_ENABLED=$(CGO_ENABLED) \
		GOOS=darwin \
		GOARCH=amd64 \
		CGO_CFLAGS="-I$(SUPERRAY_INCLUDE)" \
		CGO_LDFLAGS="-L$(THIRD_PARTY)/lib/macos/x86_64 -lsuperray -Wl,-rpath,@executable_path/lib" \
		$(GO) build -o $(BUILD_DIR)/darwin-amd64/$(APP_NAME) .
	@cp $(THIRD_PARTY)/lib/macos/x86_64/libsuperray.dylib $(BUILD_DIR)/darwin-amd64/lib/
	@install_name_tool -change libsuperray.dylib @executable_path/lib/libsuperray.dylib $(BUILD_DIR)/darwin-amd64/$(APP_NAME) 2>/dev/null || true
	@cp $(SUPERRAY_GEOIP)/* $(BUILD_DIR)/darwin-amd64/geoip/ 2>/dev/null || true
	@cp .env.example $(BUILD_DIR)/darwin-amd64/.env.example 2>/dev/null || true
	@echo "Built: $(BUILD_DIR)/darwin-amd64/$(APP_NAME)"

# Build macOS Universal Binary (ARM64 + AMD64)
darwin-universal: darwin-arm64 darwin-amd64
	@echo "Creating macOS Universal Binary..."
	@mkdir -p $(BUILD_DIR)/darwin-universal/lib $(BUILD_DIR)/darwin-universal/geoip
	@lipo -create \
		$(BUILD_DIR)/darwin-arm64/$(APP_NAME) \
		$(BUILD_DIR)/darwin-amd64/$(APP_NAME) \
		-output $(BUILD_DIR)/darwin-universal/$(APP_NAME)
	@lipo -create \
		$(BUILD_DIR)/darwin-arm64/lib/libsuperray.dylib \
		$(BUILD_DIR)/darwin-amd64/lib/libsuperray.dylib \
		-output $(BUILD_DIR)/darwin-universal/lib/libsuperray.dylib
	@cp $(SUPERRAY_GEOIP)/* $(BUILD_DIR)/darwin-universal/geoip/ 2>/dev/null || true
	@cp .env.example $(BUILD_DIR)/darwin-universal/.env.example 2>/dev/null || true
	@echo "Built: $(BUILD_DIR)/darwin-universal/$(APP_NAME) (universal)"

# Build both macOS architectures
darwin-all: darwin-arm64 darwin-amd64 darwin-universal
	@echo "All macOS builds complete"

# ============================================================
# Linux Builds (Dynamic Linking - requires libsuperray.so)
# ============================================================

# Build for Linux AMD64 (Dynamic) - using zig as cross-compiler
linux-amd64: deps
	@echo "Building for Linux AMD64..."
	@mkdir -p $(BUILD_DIR)/linux-amd64/lib $(BUILD_DIR)/linux-amd64/geoip
	CGO_ENABLED=$(CGO_ENABLED) \
		GOOS=linux \
		GOARCH=amd64 \
		CGO_CFLAGS="-I$(SUPERRAY_INCLUDE)" \
		CGO_LDFLAGS="-L$(THIRD_PARTY)/lib/linux/amd64 -lsuperray -Wl,-rpath,\$$ORIGIN/lib" \
		CC="zig cc -target x86_64-linux-gnu" \
		CXX="zig c++ -target x86_64-linux-gnu" \
		$(GO) build -o $(BUILD_DIR)/linux-amd64/$(APP_NAME) .
	@cp $(THIRD_PARTY)/lib/linux/amd64/libsuperray.so $(BUILD_DIR)/linux-amd64/lib/
	@cp $(SUPERRAY_GEOIP)/* $(BUILD_DIR)/linux-amd64/geoip/ 2>/dev/null || true
	@cp .env.example $(BUILD_DIR)/linux-amd64/.env.example 2>/dev/null || true
	@echo "Built: $(BUILD_DIR)/linux-amd64/$(APP_NAME)"

# Build for Linux ARM64 (Dynamic) - using zig as cross-compiler
linux-arm64: deps
	@echo "Building for Linux ARM64..."
	@mkdir -p $(BUILD_DIR)/linux-arm64/lib $(BUILD_DIR)/linux-arm64/geoip
	CGO_ENABLED=$(CGO_ENABLED) \
		GOOS=linux \
		GOARCH=arm64 \
		CGO_CFLAGS="-I$(SUPERRAY_INCLUDE)" \
		CGO_LDFLAGS="-L$(THIRD_PARTY)/lib/linux/arm64 -lsuperray -Wl,-rpath,\$$ORIGIN/lib" \
		CC="zig cc -target aarch64-linux-gnu" \
		CXX="zig c++ -target aarch64-linux-gnu" \
		$(GO) build -o $(BUILD_DIR)/linux-arm64/$(APP_NAME) .
	@cp $(THIRD_PARTY)/lib/linux/arm64/libsuperray.so $(BUILD_DIR)/linux-arm64/lib/
	@cp $(SUPERRAY_GEOIP)/* $(BUILD_DIR)/linux-arm64/geoip/ 2>/dev/null || true
	@cp .env.example $(BUILD_DIR)/linux-arm64/.env.example 2>/dev/null || true
	@echo "Built: $(BUILD_DIR)/linux-arm64/$(APP_NAME)"

# ============================================================
# Windows Build (Dynamic Linking - requires superray.dll)
# ============================================================

# Build for Windows AMD64 (Dynamic) - using zig as cross-compiler
windows-amd64: deps
	@echo "Building for Windows AMD64..."
	@mkdir -p $(BUILD_DIR)/windows-amd64/geoip
	CGO_ENABLED=$(CGO_ENABLED) \
		GOOS=windows \
		GOARCH=amd64 \
		CGO_CFLAGS="-I$(SUPERRAY_INCLUDE)" \
		CGO_LDFLAGS="-L$(THIRD_PARTY)/lib/windows/amd64 -lsuperray" \
		CC="zig cc -target x86_64-windows-gnu" \
		CXX="zig c++ -target x86_64-windows-gnu" \
		$(GO) build -o $(BUILD_DIR)/windows-amd64/$(APP_NAME).exe .
	@cp $(THIRD_PARTY)/lib/windows/amd64/superray.dll $(BUILD_DIR)/windows-amd64/
	@cp $(SUPERRAY_GEOIP)/* $(BUILD_DIR)/windows-amd64/geoip/ 2>/dev/null || true
	@cp .env.example $(BUILD_DIR)/windows-amd64/.env.example 2>/dev/null || true
	@echo "Built: $(BUILD_DIR)/windows-amd64/$(APP_NAME).exe"

# ============================================================
# Native build (current platform)
# ============================================================

native: deps
	@echo "Building for current platform ($(DEFAULT_TARGET))..."
	@$(MAKE) $(DEFAULT_TARGET)

# ============================================================
# Package targets
# ============================================================

package-darwin-arm64: darwin-arm64
	@echo "Packaging darwin-arm64..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR)/darwin-arm64 && tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-arm64.tar.gz *
	@echo "Created: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-arm64.tar.gz"

package-darwin-amd64: darwin-amd64
	@echo "Packaging darwin-amd64..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR)/darwin-amd64 && tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64.tar.gz *
	@echo "Created: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-amd64.tar.gz"

package-darwin-universal: darwin-universal
	@echo "Packaging darwin-universal..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR)/darwin-universal && tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-universal.tar.gz *
	@echo "Created: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-darwin-universal.tar.gz"

package-linux-amd64: linux-amd64
	@echo "Packaging linux-amd64..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR)/linux-amd64 && tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz *
	@echo "Created: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz"

package-linux-arm64: linux-arm64
	@echo "Packaging linux-arm64..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR)/linux-arm64 && tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-arm64.tar.gz *
	@echo "Created: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-arm64.tar.gz"

package-windows-amd64: windows-amd64
	@echo "Packaging windows-amd64..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR)/windows-amd64 && zip -qr $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64.zip *
	@echo "Created: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-windows-amd64.zip"

# Package all platforms
package-all: package-darwin-arm64 package-darwin-amd64 package-darwin-universal package-linux-amd64 package-linux-arm64 package-windows-amd64
	@echo ""
	@echo "All packages created in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

# Package for current platform
package: native
	@echo "Packaging for current platform ($(DEFAULT_TARGET))..."
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR)/$(DEFAULT_TARGET) && tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-$(DEFAULT_TARGET).tar.gz *
	@echo "Created: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-$(DEFAULT_TARGET).tar.gz"

# Package macOS (all architectures)
package-darwin: package-darwin-arm64 package-darwin-amd64 package-darwin-universal
	@echo "macOS packages created"

# ============================================================
# Run / Install / Uninstall
# ============================================================

run: native
	@echo "Running $(APP_NAME)..."
	@cd $(BUILD_DIR)/$(DEFAULT_TARGET) && SUPERRAY_GEO_PATH=./geoip ./$(APP_NAME)

install: native
	@echo "Installing $(APP_NAME)..."
	@sudo cp $(BUILD_DIR)/$(DEFAULT_TARGET)/$(APP_NAME) /usr/local/bin/
	@sudo mkdir -p /usr/local/share/superray/geoip
	@sudo cp $(BUILD_DIR)/$(DEFAULT_TARGET)/geoip/* /usr/local/share/superray/geoip/ 2>/dev/null || true
	@echo "Installed to /usr/local/bin/$(APP_NAME)"

uninstall:
	@echo "Uninstalling $(APP_NAME)..."
	@sudo rm -f /usr/local/bin/$(APP_NAME)
	@sudo rm -rf /usr/local/share/superray
	@echo "Uninstalled"

# ============================================================
# Info
# ============================================================

version:
	@echo "$(APP_NAME) version $(VERSION)"
	@echo "Go: $$($(GO) version)"

# Check binary dependencies
check: native
	@echo "Checking binary dependencies..."
ifeq ($(UNAME_S),Darwin)
	@otool -L $(BUILD_DIR)/$(DEFAULT_TARGET)/$(APP_NAME)
else
	@ldd $(BUILD_DIR)/$(DEFAULT_TARGET)/$(APP_NAME) 2>/dev/null || echo "ldd not available"
endif

# ============================================================
# Help
# ============================================================

help:
	@echo "SuperRay-TUI Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build targets (macOS - portable with bundled dylib):"
	@echo "  darwin-arm64     - macOS ARM64 (Apple Silicon)"
	@echo "  darwin-amd64     - macOS AMD64 (Intel)"
	@echo "  darwin-universal - macOS Universal (ARM64 + AMD64)"
	@echo "  darwin-all       - All macOS builds"
	@echo ""
	@echo "Build targets (Linux/Windows - DYNAMIC, needs .so/.dll):"
	@echo "  linux-amd64      - Linux AMD64 (requires cross-compiler)"
	@echo "  linux-arm64      - Linux ARM64 (requires cross-compiler)"
	@echo "  windows-amd64    - Windows AMD64 (requires mingw)"
	@echo ""
	@echo "Common targets:"
	@echo "  all, build       - Build for current platform"
	@echo "  native           - Same as 'all'"
	@echo "  clean            - Remove build artifacts"
	@echo "  deps             - Download Go dependencies"
	@echo ""
	@echo "Package targets:"
	@echo "  package          - Package current platform"
	@echo "  package-darwin   - Package all macOS (arm64, amd64, universal)"
	@echo "  package-all      - Package all platforms"
	@echo ""
	@echo "Other targets:"
	@echo "  run              - Build and run"
	@echo "  install          - Install to /usr/local/bin"
	@echo "  uninstall        - Remove from /usr/local"
	@echo "  check            - Show binary dependencies"
	@echo "  version          - Show version info"
	@echo "  help             - Show this help"
	@echo ""
	@echo "Current platform: $(DEFAULT_TARGET)"
	@echo "Version: $(VERSION)"
	@echo ""
	@echo "Note: All platforms use dynamic linking."
	@echo "      Libraries are bundled with the executable."
