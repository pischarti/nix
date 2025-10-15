#!/bin/bash
# E2E Test Runner for kaws operator
# This script runs the complete end-to-end tests including operator deployment

set -e

echo "════════════════════════════════════════════════════════════════"
echo "🚀 KAWS Operator E2E Test Suite"
echo "════════════════════════════════════════════════════════════════"
echo ""

# Change to the kaws directory
cd "$(dirname "$0")"

# Build the operator Docker image (from repo root for monorepo structure)
echo "📦 Building operator Docker image..."
cd ../..
docker build -f go/kaws/Dockerfile -t kaws-operator:test .
cd go/kaws
echo "✅ Image built"
echo ""

# Run E2E tests
echo "🧪 Running E2E tests..."
echo "   This will:"
echo "   1. Create a kind cluster"
echo "   2. Deploy CRDs and RBAC"
echo "   3. Deploy operator with 3 replicas"
echo "   4. Test leader election"
echo "   5. Test informer functionality"
echo "   6. Test EventRecycler reconciliation"
echo ""

go test -v -tags=integration -run TestOperatorE2E ./cmd/operator/ -timeout 10m

echo ""
echo "════════════════════════════════════════════════════════════════"
echo "✅ E2E Tests Complete!"
echo "════════════════════════════════════════════════════════════════"

