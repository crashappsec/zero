#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
# 
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Commit Message Validation Script
# Validates commit messages follow conventional commit format
# Usage: ./check-commit-message.sh <commit-message-file>
#        ./check-commit-message.sh --check-last
#############################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Conventional commit types
VALID_TYPES="feat|fix|docs|style|refactor|test|chore|revert"

# Get commit message
if [[ "$1" == "--check-last" ]]; then
    COMMIT_MSG=$(git log -1 --pretty=%B)
elif [[ -f "$1" ]]; then
    COMMIT_MSG=$(cat "$1")
else
    echo -e "${RED}Error: No commit message provided${NC}"
    echo "Usage: $0 <commit-message-file>"
    echo "       $0 --check-last"
    exit 1
fi

# Get first line of commit message
FIRST_LINE=$(echo "$COMMIT_MSG" | head -n 1)

echo "Validating commit message..."
echo "Message: $FIRST_LINE"
echo ""

# Check if message matches conventional commit format
if ! echo "$FIRST_LINE" | grep -qE "^($VALID_TYPES)(\(.+\))?: .+"; then
    echo -e "${RED}❌ Invalid commit message format${NC}"
    echo ""
    echo "Commit messages must follow the conventional commit format:"
    echo "  <type>[optional scope]: <description>"
    echo ""
    echo "Valid types: $VALID_TYPES"
    echo ""
    echo "Examples:"
    echo "  feat: add new certificate validation"
    echo "  fix(skills): correct regex pattern"
    echo "  docs: update README with usage examples"
    echo "  chore(deps): update dependencies"
    echo ""
    exit 1
fi

# Check message length (should be <= 100 characters)
if [[ ${#FIRST_LINE} -gt 100 ]]; then
    echo -e "${YELLOW}⚠️  Warning: Commit message is longer than 100 characters${NC}"
    echo "   Consider making it more concise"
fi

# Check for period at end (should not have one)
if echo "$FIRST_LINE" | grep -qE "\.$"; then
    echo -e "${YELLOW}⚠️  Warning: Commit message should not end with a period${NC}"
fi

# Check that description starts with lowercase (after type)
TYPE_AND_DESC=$(echo "$FIRST_LINE" | sed -E 's/^[a-z]+(\([^)]+\))?:[ ]*//')
if ! echo "$TYPE_AND_DESC" | grep -qE "^[a-z]"; then
    echo -e "${YELLOW}⚠️  Warning: Description should start with lowercase${NC}"
fi

echo -e "${GREEN}✓ Commit message format is valid${NC}"
exit 0
