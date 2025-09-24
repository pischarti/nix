#!/bin/bash

# Test runner script for the Go Kubernetes CLI project

set -e

echo "🧪 Running Go tests for Kubernetes CLI project..."
echo

# Run tests with coverage
echo "📊 Running tests with coverage..."
go test ./... -v -cover

echo
echo "📈 Detailed coverage report..."
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out

echo
echo "🎯 HTML coverage report (optional - uncomment to generate):"
echo "# go tool cover -html=coverage.out -o coverage.html"
echo "# open coverage.html"

echo
echo "✅ All tests completed successfully!"
