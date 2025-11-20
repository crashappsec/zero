#!/bin/bash
# Copyright (c) 2024 Crash Override Inc
# 101 Fulton St, 416, New York 10038
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Bootstrap Script
# Makes all skill scripts executable in one command
# Usage: chmod +x bootstrap.sh && ./bootstrap.sh
#############################################################################

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo ""
echo "========================================="
echo "  Skills and Prompts Bootstrap"
echo "========================================="
echo ""

# Get script directory (repo root)
REPO_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Find all .sh files in skills directory
echo -e "${BLUE}Finding all shell scripts in skills/ directory...${NC}"
echo ""

SCRIPT_COUNT=0
while IFS= read -r -d '' script; do
    # Make executable
    chmod +x "$script"

    # Show relative path
    RELATIVE_PATH="${script#$REPO_ROOT/}"
    echo -e "${GREEN}✓${NC} Made executable: $RELATIVE_PATH"

    ((SCRIPT_COUNT++))
done < <(find "$REPO_ROOT/skills" -type f -name "*.sh" -print0)

echo ""
echo "========================================="
echo -e "${GREEN}  Bootstrap Complete${NC}"
echo "========================================="
echo ""
echo "Made $SCRIPT_COUNT script(s) executable"
echo ""
echo "Next steps:"
echo "  1. Copy .env.example to .env:"
echo "     cp .env.example .env"
echo ""
echo "  2. Add your Anthropic API key to .env:"
echo "     ANTHROPIC_API_KEY=sk-ant-xxx"
echo ""
echo "  3. Run any skill script:"
echo "     ./skills/code-ownership/ownership-analyzer-claude.sh --help"
echo ""
