#!/bin/bash

# SuperRay-TUI Build Script
# Quick build for current platform

set -e

# Get script directory (project root)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Go path
export PATH=/Volumes/mindata/Library/go/bin:$PATH
export GOPATH=$HOME/go

# Enable CGO
export CGO_ENABLED=1

# SuperRay library paths (using project's third_party)
SUPERRAY_DIR="$SCRIPT_DIR/third_party/superray"
SUPERRAY_INCLUDE="$SUPERRAY_DIR/include"
SUPERRAY_LIB="$SUPERRAY_DIR/lib/darwin/universal"
SUPERRAY_GEOIP="$SUPERRAY_DIR/geoip"

# CGO flags
export CGO_CFLAGS="-I$SUPERRAY_INCLUDE"
export CGO_LDFLAGS="-L$SUPERRAY_LIB -lsuperray -rpath @executable_path/../lib"

# Detect architecture
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
    TARGET="darwin-arm64"
else
    TARGET="darwin-amd64"
fi

# Build directory
BUILD_DIR="./build/$TARGET"
mkdir -p "$BUILD_DIR/lib" "$BUILD_DIR/geoip"

# Application name
APP_NAME="superray-tui"

echo "========================================"
echo "SuperRay TUI Build Script"
echo "========================================"
echo ""
echo "Go version: $(go version)"
echo "Target: $TARGET"
echo "SuperRay Include: $SUPERRAY_INCLUDE"
echo "SuperRay Library: $SUPERRAY_LIB"
echo ""

# Download dependencies
echo "Downloading dependencies..."
go mod tidy

# Build
echo ""
echo "Building $APP_NAME..."
go build -o "$BUILD_DIR/$APP_NAME" .

# Copy files
cp "$SUPERRAY_LIB/libsuperray.dylib" "$BUILD_DIR/lib/"
cp "$SUPERRAY_GEOIP/"* "$BUILD_DIR/geoip/" 2>/dev/null || true
cp ".env.example" "$BUILD_DIR/.env.example" 2>/dev/null || true

# Check if build was successful
if [ -f "$BUILD_DIR/$APP_NAME" ]; then
    echo ""
    echo "========================================"
    echo "Build successful!"
    echo ""
    echo "Output:"
    echo "  $BUILD_DIR/$APP_NAME"
    echo "  $BUILD_DIR/lib/libsuperray.dylib"
    echo "  $BUILD_DIR/geoip/"
    echo ""
    echo "To run:"
    echo "  cd $BUILD_DIR && ./$APP_NAME"
    echo ""
    echo "Or use Makefile:"
    echo "  make run"
    echo "========================================"
else
    echo "Build failed!"
    exit 1
fi
