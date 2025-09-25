#!/bin/bash

# Build script for Glappy Bird game

set -e

echo "🐦 Building Glappy Bird..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go version $GO_VERSION is too old. Please install Go $REQUIRED_VERSION or later."
    exit 1
fi

echo "✅ Go version $GO_VERSION detected"

# Install dependencies (using root go.mod)
echo "📦 Installing dependencies..."
cd .. && go mod tidy && cd glappy

# Build the game (now using internal/game package)
echo "🔨 Building glappy..."
go build -o glappy .

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    echo "🎮 To run the game:"
    echo "   ./glappy"
    echo ""
    echo "🎯 Game controls:"
    echo "   SPACE - Jump"
    echo "   R - Restart (when game over)"
    echo "   ESC - Quit"
else
    echo "❌ Build failed!"
    exit 1
fi
