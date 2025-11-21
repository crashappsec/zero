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
done < <(find "$REPO_ROOT/utils" -type f -name "*.sh" -print0)

echo ""

# Check for Homebrew (macOS package manager)
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo -e "${BLUE}Checking for Homebrew...${NC}"
    if command -v brew &> /dev/null; then
        BREW_VERSION=$(brew --version | head -1)
        echo -e "${GREEN}✓${NC} Homebrew is installed: $BREW_VERSION"
        echo ""
    else
        echo -e "${YELLOW}⚠${NC} Homebrew is not installed"
        echo ""
        echo "Homebrew is recommended for managing dependencies on macOS."
        echo ""

        read -p "Would you like to install Homebrew now? (y/n) " -n 1 -r
        echo ""

        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo ""
            echo -e "${BLUE}Installing Homebrew...${NC}"
            /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

            if command -v brew &> /dev/null; then
                echo -e "${GREEN}✓${NC} Homebrew installed successfully"
            else
                echo -e "${YELLOW}⚠${NC} Homebrew installation may have failed"
                echo "You may need to add Homebrew to your PATH. Follow the instructions above."
            fi
            echo ""
        else
            echo ""
            echo -e "${YELLOW}Skipping Homebrew installation${NC}"
            echo "You'll need to install tools manually."
            echo ""
        fi
    fi
fi

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

    if command -v brew &> /dev/null; then
        read -p "Would you like to install syft via Homebrew? (y/n) " -n 1 -r
        echo ""

        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo ""
            echo -e "${BLUE}Installing syft...${NC}"
            if brew install syft; then
                echo -e "${GREEN}✓${NC} syft installed successfully"
                echo ""
            else
                echo -e "${YELLOW}✗${NC} Failed to install syft"
                echo ""
            fi
        else
            echo ""
            echo -e "${YELLOW}Skipping syft installation${NC}"
            echo "Note: SBOM generation will not work without syft"
            echo ""
        fi
    else
        echo "Install syft manually:"
        echo "  - macOS:   brew install syft"
        echo "  - Linux:   curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s"
        echo "  - Manual:  https://github.com/anchore/syft#installation"
        echo ""
    fi
fi

# Check for Go (required for osv-scanner)
echo -e "${BLUE}Checking for Go...${NC}"
if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    echo -e "${GREEN}✓${NC} Go is installed: $GO_VERSION"
    echo ""
else
    echo -e "${YELLOW}⚠${NC} Go is not installed"
    echo ""
    echo "Go is required to install osv-scanner."
    echo ""

    if command -v brew &> /dev/null; then
        read -p "Would you like to install Go via Homebrew? (y/n) " -n 1 -r
        echo ""

        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo ""
            echo -e "${BLUE}Installing Go...${NC}"
            if brew install go; then
                echo -e "${GREEN}✓${NC} Go installed successfully"
                echo ""
            else
                echo -e "${YELLOW}✗${NC} Failed to install Go"
                echo ""
            fi
        else
            echo ""
            echo -e "${YELLOW}Skipping Go installation${NC}"
            echo "Note: osv-scanner cannot be installed without Go"
            echo ""
        fi
    else
        echo "Install Go manually:"
        echo "  - macOS:   brew install go"
        echo "  - Linux:   https://go.dev/doc/install"
        echo "  - Windows: https://go.dev/doc/install"
        echo ""
    fi
fi

# Check for osv-scanner
echo -e "${BLUE}Checking for osv-scanner (vulnerability scanner)...${NC}"
if command -v osv-scanner &> /dev/null; then
    OSV_VERSION=$(osv-scanner --version 2>&1 | head -1 || echo "unknown")
    echo -e "${GREEN}✓${NC} osv-scanner is installed: $OSV_VERSION"
    echo ""
else
    echo -e "${YELLOW}⚠${NC} osv-scanner is not installed"
    echo ""
    echo "osv-scanner is required for vulnerability scanning and taint analysis."
    echo ""

    if command -v go &> /dev/null; then
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
            fi
        else
            echo ""
            echo -e "${YELLOW}Skipping osv-scanner installation${NC}"
            echo "Note: Vulnerability scanning will not work without osv-scanner"
            echo ""
        fi
    else
        echo -e "${YELLOW}Go is not installed - cannot install osv-scanner${NC}"
        echo "Install Go first, then install osv-scanner:"
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

# Summary of tool status
echo "Tool Status:"
echo "------------"

# Check each tool and report status
if [[ "$OSTYPE" == "darwin"* ]]; then
    if command -v brew &> /dev/null; then
        echo -e "${GREEN}✓${NC} Homebrew"
    else
        echo -e "${YELLOW}⚠${NC} Homebrew (not installed)"
    fi
fi

if command -v go &> /dev/null; then
    echo -e "${GREEN}✓${NC} Go"
else
    echo -e "${YELLOW}⚠${NC} Go (not installed - needed for osv-scanner)"
fi

if command -v syft &> /dev/null; then
    echo -e "${GREEN}✓${NC} syft"
else
    echo -e "${YELLOW}⚠${NC} syft (not installed - SBOM generation won't work)"
fi

if command -v osv-scanner &> /dev/null; then
    echo -e "${GREEN}✓${NC} osv-scanner"
else
    echo -e "${YELLOW}⚠${NC} osv-scanner (not installed - vulnerability scanning won't work)"
fi

echo ""
echo "Next steps:"
echo "  1. Copy .env.example to .env:"
echo "     cp .env.example .env"
echo ""
echo "  2. Add your Anthropic API key to .env:"
echo "     ANTHROPIC_API_KEY=sk-ant-xxx"
echo ""
echo "  3. Run any utility script:"
echo "     ./utils/code-ownership/ownership-analyzer-claude.sh --help"
echo "     ./utils/supply-chain/supply-chain-scanner.sh --help"
echo "     ./utils/supply-chain/vulnerability-analysis/vulnerability-analyzer.sh --help"
echo ""

# Warning if critical tools are missing
MISSING_TOOLS=false
if ! command -v syft &> /dev/null; then
    MISSING_TOOLS=true
fi
if ! command -v osv-scanner &> /dev/null; then
    MISSING_TOOLS=true
fi

if [[ "$MISSING_TOOLS" == "true" ]]; then
    echo -e "${YELLOW}⚠ Warning: Some required tools are not installed${NC}"
    echo "Some features may not work. Run ./bootstrap.sh again to install missing tools."
    echo ""
fi
