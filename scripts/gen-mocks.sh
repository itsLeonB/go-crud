#!/bin/bash

# Script to generate gomock mocks for each file with interfaces
# Only searches root level (current directory only)
# Creates <filename>_mock.go next to each source file

set -e

echo "Scanning for interfaces in root directory...
"

# Counter for files processed
count=0
success=0
failed=0

# Find all Go files in root directory only (maxdepth 1)
find . -maxdepth 1 -name "*.go" \
  -not -name "*_test.go" \
  -not -name "*_mock.go" \
  -type f | while read -r file; do
  
  # Check if file contains interface definitions
  if grep -q "type.*interface\s*{" "$file"; then
    count=$((count + 1))
    
    # Get base filename
    base=$(basename "$file" .go)
    output="./${base}_mock.go"
    
    echo "Processing: $file"
    
    # Get package name from the file
    package=$(grep -m 1 "^package " "$file" | awk '{print $2}')
    
    # Generate mock file
    if GO111MODULE=on mockgen -source="$file" -destination="$output" -package="$package" 2>/dev/null; then
      echo "  ✓ Generated: $output"
      success=$((success + 1))
    else
      echo "  ✗ Failed to generate mock for: $file"
      failed=$((failed + 1))
    fi
  fi
done

echo ""
echo "======================================"
echo "Mock generation complete!"
echo "Files with interfaces found and processed"
echo "Check *_mock.go files in current directory"
echo "======================================"
