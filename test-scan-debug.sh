#!/bin/bash
# Debug script to test single repo scan

set -x  # Enable debug output

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Test with one repo
cd "$SCRIPT_DIR"

echo "Testing bootstrap.sh scan for phantom-tests/react..."
./utils/zero/scripts/bootstrap.sh --scan-only --security phantom-tests/react

echo "Exit code: $?"
