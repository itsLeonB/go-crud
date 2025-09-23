#!/bin/sh
#
# Pre-push hook that runs linting and tests before allowing push

echo "Running pre-push checks..."

# Run linting
echo "\n=== Running linting ==="
if ! make lint; then
    echo "❌ Linting failed! Please fix the issues before pushing."
    exit 1
fi

# Run tests
echo "\n=== Running tests ==="
if ! make test-all; then
    echo "❌ Tests failed! Please fix the test issues before pushing."
    exit 1
fi

echo "\n✅ All checks passed! Pushing can continue...\n"
