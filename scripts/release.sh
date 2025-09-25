#!/bin/bash

# Release script for Go AWS CLI (gaws)
# Usage: ./scripts/release.sh [patch|minor|major]

set -e

VERSION_TYPE=${1:-patch}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "🚀 Starting release process for version type: $VERSION_TYPE"

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Error: Not in a git repository"
    exit 1
fi

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    echo "❌ Error: Working directory is not clean. Please commit or stash changes."
    exit 1
fi

# Get current version
CURRENT_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
echo "📋 Current version: $CURRENT_TAG"

# Calculate new version
CURRENT=${CURRENT_TAG#v}
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"

case "$VERSION_TYPE" in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
    *)
        echo "❌ Error: Invalid version type. Use: patch, minor, or major"
        exit 1
        ;;
esac

NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"
echo "📦 New version: $NEW_VERSION"

# Confirm release
read -p "🤔 Do you want to create release $NEW_VERSION? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Release cancelled"
    exit 1
fi

# Create and push tag
echo "🏷️  Creating tag $NEW_VERSION..."
git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"
git push origin "$NEW_VERSION"

echo "✅ Tag $NEW_VERSION created and pushed!"
echo "🔄 GitHub Actions will now build and release the artifacts."
echo "📋 You can monitor the progress at: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/actions"
