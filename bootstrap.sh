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

# Check for syft (SBOM generator)
echo -e "${BLUE}Checking for syft (SBOM generator)...${NC}"
if command -v syft &> /dev/null; then
    SYFT_VERSION=$(syft version 2>&1 | head -1 || echo "unknown")
    echo -e "${GREEN}✓${NC} syft is installed: $SYFT_VERSION"
    echo ""
else
    echo -e "${YELLOW}⚠${NC} syft is not installed"
    echo ""
    echo "syft is required to generate SBOMs for repositories without existing SBOMs."
    echo ""
    echo "Install syft:"
    echo "  - macOS:   brew install syft"
    echo "  - Linux:   curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s"
    echo "  - Manual:  https://github.com/anchore/syft#installation"
    echo ""
fi

# Check for osv-scanner
echo -e "${BLUE}Checking for osv-scanner...${NC}"
if command -v osv-scanner &> /dev/null; then
    OSV_VERSION=$(osv-scanner --version 2>&1 | head -1 || echo "unknown")
    echo -e "${GREEN}✓${NC} osv-scanner is installed: $OSV_VERSION"
    echo ""
else
    echo -e "${YELLOW}⚠${NC} osv-scanner is not installed"
    echo ""
    echo "osv-scanner is required for SBOM vulnerability analysis and taint analysis."
    echo ""

    # Check if Go is installed
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version)
        echo -e "${GREEN}✓${NC} Go is installed: $GO_VERSION"
        echo ""

        # Prompt to install osv-scanner
        read -p "Would you like to install osv-scanner now? (y/n) " -n 1 -r
        echo ""

        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo ""
            echo -e "${BLUE}Installing osv-scanner...${NC}"
            if go install github.com/google/osv-scanner/cmd/osv-scanner@latest; then
                echo -e "${GREEN}✓${NC} osv-scanner installed successfully"

                # Check if GOPATH/bin is in PATH
                if ! command -v osv-scanner &> /dev/null; then
                    echo ""
                    echo -e "${YELLOW}⚠${NC} osv-scanner was installed but is not in your PATH"
                    echo ""
                    GOPATH_BIN=$(go env GOPATH)/bin
                    echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
                    echo "  export PATH=\"\$PATH:$GOPATH_BIN\""
                    echo ""
                    echo "Then reload your shell or run:"
                    echo "  source ~/.bashrc  # or ~/.zshrc"
                    echo ""
                fi
            else
                echo -e "${YELLOW}✗${NC} Failed to install osv-scanner"
                echo ""
                echo "Try installing manually:"
                echo "  go install github.com/google/osv-scanner/cmd/osv-scanner@latest"
                echo ""
            fi
        else
            echo ""
            echo "To install osv-scanner later, run:"
            echo "  go install github.com/google/osv-scanner/cmd/osv-scanner@latest"
            echo ""
        fi
    else
        echo -e "${YELLOW}✗${NC} Go is not installed"
        echo ""
        echo "osv-scanner requires Go. Install Go first:"
        echo "  - macOS:   brew install go"
        echo "  - Linux:   https://go.dev/doc/install"
        echo "  - Windows: https://go.dev/doc/install"
        echo ""
        echo "After installing Go, install osv-scanner:"
        echo "  go install github.com/google/osv-scanner/cmd/osv-scanner@latest"
        echo ""
    fi
fi

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
echo "     ./skills/sbom-analyzer/sbom-analyzer-claude.sh --help"
echo ""
