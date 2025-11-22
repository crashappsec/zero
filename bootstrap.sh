#!/bin/bash
# Copyright (c) 2024 Gibson Powers Contributors
# 
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

# Track tools to install
TOOLS_TO_INSTALL=()

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

# Check for core utilities
echo -e "${BLUE}Checking for core utilities...${NC}"

# jq
if command -v jq &> /dev/null; then
    echo -e "${GREEN}✓${NC} jq is installed"
else
    echo -e "${YELLOW}⚠${NC} jq is not installed (required)"
    TOOLS_TO_INSTALL+=("jq")
fi

# curl
if command -v curl &> /dev/null; then
    echo -e "${GREEN}✓${NC} curl is installed"
else
    echo -e "${YELLOW}⚠${NC} curl is not installed (required)"
    TOOLS_TO_INSTALL+=("curl")
fi

# git
if command -v git &> /dev/null; then
    echo -e "${GREEN}✓${NC} git is installed"
else
    echo -e "${YELLOW}⚠${NC} git is not installed (required)"
    TOOLS_TO_INSTALL+=("git")
fi

# openssl
if command -v openssl &> /dev/null; then
    echo -e "${GREEN}✓${NC} openssl is installed"
else
    echo -e "${YELLOW}⚠${NC} openssl is not installed (required for certificate analyser)"
    TOOLS_TO_INSTALL+=("openssl")
fi

# bc (for cost calculations)
if command -v bc &> /dev/null; then
    echo -e "${GREEN}✓${NC} bc is installed"
else
    echo -e "${YELLOW}⚠${NC} bc is not installed (required for cost tracking)"
    TOOLS_TO_INSTALL+=("bc")
fi

echo ""

# Check for syft (SBOM generator)
echo -e "${BLUE}Checking for syft (SBOM generator)...${NC}"
if command -v syft &> /dev/null; then
    SYFT_VERSION=$(syft version 2>&1 | head -1 || echo "unknown")
    echo -e "${GREEN}✓${NC} syft is installed: $SYFT_VERSION"
else
    echo -e "${YELLOW}⚠${NC} syft is not installed (required for SBOM generation)"
    TOOLS_TO_INSTALL+=("syft")
fi
echo ""

# Check for Go (required for osv-scanner)
echo -e "${BLUE}Checking for Go...${NC}"
if command -v go &> /dev/null; then
    GO_VERSION=$(go version)
    echo -e "${GREEN}✓${NC} Go is installed: $GO_VERSION"
else
    echo -e "${YELLOW}⚠${NC} Go is not installed (required for osv-scanner)"
    TOOLS_TO_INSTALL+=("go")
fi
echo ""

# Check for osv-scanner
echo -e "${BLUE}Checking for osv-scanner (vulnerability scanner)...${NC}"
if command -v osv-scanner &> /dev/null; then
    OSV_VERSION=$(osv-scanner --version 2>&1 | head -1 || echo "unknown")
    echo -e "${GREEN}✓${NC} osv-scanner is installed: $OSV_VERSION"
else
    echo -e "${YELLOW}⚠${NC} osv-scanner is not installed (will install via go if available)"
fi
echo ""

# Batch install missing tools
if [[ ${#TOOLS_TO_INSTALL[@]} -gt 0 ]] && command -v brew &> /dev/null; then
    echo ""
    echo "========================================="
    echo -e "${YELLOW}Missing Tools Summary${NC}"
    echo "========================================="
    echo ""
    echo "The following tools need to be installed:"
    for tool in "${TOOLS_TO_INSTALL[@]}"; do
        echo "  - $tool"
    done
    echo ""

    read -p "Would you like to install all missing tools via Homebrew? (y/n) " -n 1 -r
    echo ""

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo ""
        echo -e "${BLUE}Installing missing tools...${NC}"
        for tool in "${TOOLS_TO_INSTALL[@]}"; do
            echo -e "${BLUE}Installing $tool...${NC}"
            if brew install "$tool" 2>&1 | grep -v "Warning"; then
                echo -e "${GREEN}✓${NC} $tool installed successfully"
            else
                echo -e "${YELLOW}⚠${NC} $tool installation may have issues"
            fi
        done
        echo ""
    fi
fi

# Install osv-scanner if Go is available
if ! command -v osv-scanner &> /dev/null && command -v go &> /dev/null; then
    read -p "Would you like to install osv-scanner via Go? (y/n) " -n 1 -r
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
            fi
        else
            echo -e "${YELLOW}✗${NC} Failed to install osv-scanner"
        fi
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

if command -v cosign &> /dev/null; then
    echo -e "${GREEN}✓${NC} cosign"
else
    echo -e "${YELLOW}⚠${NC} cosign (not installed - provenance verification won't work)"
fi

if command -v rekor-cli &> /dev/null; then
    echo -e "${GREEN}✓${NC} rekor-cli"
else
    echo -e "${YELLOW}⚠${NC} rekor-cli (not installed - transparency log checks won't work)"
fi

echo ""

# Check for .env file and API key
echo -e "${BLUE}Checking for .env file and API configuration...${NC}"
if [ -f "$REPO_ROOT/.env" ]; then
    echo -e "${GREEN}✓${NC} .env file exists"

    # Check if ANTHROPIC_API_KEY is set in the file
    if grep -q "^ANTHROPIC_API_KEY=sk-ant-" "$REPO_ROOT/.env" 2>/dev/null; then
        echo -e "${GREEN}✓${NC} ANTHROPIC_API_KEY is configured"
    else
        echo -e "${YELLOW}⚠${NC} ANTHROPIC_API_KEY is not set or invalid in .env"
        echo ""
        echo "Claude-enabled scripts require an Anthropic API key."
        echo "Edit .env and add:"
        echo "  ANTHROPIC_API_KEY=sk-ant-xxx..."
        echo ""
        echo "Get your API key from: https://console.anthropic.com/"
        echo ""
    fi
else
    echo -e "${YELLOW}⚠${NC} .env file does not exist"
    echo ""

    if [ -f "$REPO_ROOT/.env.example" ]; then
        read -p "Would you like to create .env from .env.example? (y/n) " -n 1 -r
        echo ""

        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo ""
            echo -e "${BLUE}Creating .env file...${NC}"
            if cp "$REPO_ROOT/.env.example" "$REPO_ROOT/.env"; then
                echo -e "${GREEN}✓${NC} Created .env file from template"
                echo ""
                echo -e "${YELLOW}Important:${NC} Edit .env and add your Anthropic API key:"
                echo "  ANTHROPIC_API_KEY=sk-ant-xxx..."
                echo ""
                echo "Get your API key from: https://console.anthropic.com/"
                echo ""
            else
                echo -e "${YELLOW}✗${NC} Failed to create .env file"
                echo ""
            fi
        else
            echo ""
            echo -e "${YELLOW}Skipping .env creation${NC}"
            echo "Create it manually:"
            echo "  cp .env.example .env"
            echo "Then add your Anthropic API key to .env"
            echo ""
        fi
    else
        echo "Create a .env file with your Anthropic API key:"
        echo "  echo 'ANTHROPIC_API_KEY=sk-ant-xxx...' > .env"
        echo ""
        echo "Get your API key from: https://console.anthropic.com/"
        echo ""
    fi
fi

# Configure RAG paths
echo -e "${BLUE}Configuring RAG (Retrieval-Augmented Generation) paths...${NC}"
if [ -d "$REPO_ROOT/rag" ]; then
    echo -e "${GREEN}✓${NC} RAG folder found at: $REPO_ROOT/rag"

    # Check if global config exists
    if [ -f "$REPO_ROOT/utils/lib/config.sh" ]; then
        # RAG_DIR is already configured in config.sh
        echo -e "${GREEN}✓${NC} RAG paths configured in utils/lib/config.sh"

        # Count RAG files
        RAG_FILE_COUNT=$(find "$REPO_ROOT/rag" -type f -name "*.md" | wc -l | tr -d ' ')
        echo -e "${GREEN}✓${NC} Found $RAG_FILE_COUNT RAG reference documents"
    else
        echo -e "${YELLOW}⚠${NC} Global config not found (should have been created during setup)"
    fi
else
    echo -e "${YELLOW}⚠${NC} RAG folder not found - reference documentation unavailable"
fi
echo ""

# Export skills to templates
echo -e "${BLUE}Exporting Claude Code skills to portable templates...${NC}"
if [ -d "$REPO_ROOT/skills" ]; then
    SKILL_COUNT=$(find "$REPO_ROOT/skills" -name "*.skill" -o -name "skill.md" | wc -l | tr -d ' ')
    echo -e "${GREEN}✓${NC} Found $SKILL_COUNT skills in repository"

    if [ -x "$REPO_ROOT/export-skills-to-templates.sh" ]; then
        if "$REPO_ROOT/export-skills-to-templates.sh" > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC} Skills exported to ~/claude-templates"
        else
            echo -e "${YELLOW}⚠${NC} Failed to export skills (non-fatal)"
        fi
    else
        echo -e "${YELLOW}⚠${NC} export-skills-to-templates.sh not executable"
    fi
else
    echo -e "${YELLOW}⚠${NC} No skills directory found"
fi
echo ""

echo ""
echo "========================================="
echo -e "${BLUE}Configuration Summary${NC}"
echo "========================================="
echo ""

# Count configured features
CONFIGURED_COUNT=0
TOTAL_FEATURES=5

# Check .env
if [ -f "$REPO_ROOT/.env" ] && grep -q "^ANTHROPIC_API_KEY=sk-ant-" "$REPO_ROOT/.env" 2>/dev/null; then
    echo -e "${GREEN}✓${NC} API Key configured"
    ((CONFIGURED_COUNT++))
else
    echo -e "${YELLOW}⚠${NC} API Key not configured"
fi

# Check RAG
if [ -d "$REPO_ROOT/rag" ] && [ -f "$REPO_ROOT/utils/lib/config.sh" ]; then
    echo -e "${GREEN}✓${NC} RAG paths configured"
    ((CONFIGURED_COUNT++))
else
    echo -e "${YELLOW}⚠${NC} RAG paths not configured"
fi

# Check skills
if [ -d "$REPO_ROOT/skills" ]; then
    echo -e "${GREEN}✓${NC} Skills available"
    ((CONFIGURED_COUNT++))
else
    echo -e "${YELLOW}⚠${NC} Skills not found"
fi

# Check global config
if [ -f "$REPO_ROOT/utils/lib/config.sh" ]; then
    echo -e "${GREEN}✓${NC} Global configuration loaded"
    ((CONFIGURED_COUNT++))
else
    echo -e "${YELLOW}⚠${NC} Global configuration missing"
fi

# Check GitHub token
if [ -n "${GITHUB_TOKEN:-}" ]; then
    echo -e "${GREEN}✓${NC} GitHub token configured"
    ((CONFIGURED_COUNT++))
else
    echo -e "${YELLOW}⚠${NC} GitHub token not set (optional, for private repos)"
fi

echo ""
echo "Configuration: $CONFIGURED_COUNT/$TOTAL_FEATURES features ready"
echo ""

echo "========================================="
echo -e "${BLUE}Next Steps${NC}"
echo "========================================="
echo ""

# Only show .env setup if not configured
if [ ! -f "$REPO_ROOT/.env" ] || ! grep -q "^ANTHROPIC_API_KEY=sk-ant-" "$REPO_ROOT/.env" 2>/dev/null; then
    echo -e "${YELLOW}1. Configure your API key:${NC}"
    if [ ! -f "$REPO_ROOT/.env" ]; then
        echo "   cp .env.example .env"
    fi
    echo "   Edit .env and add: ANTHROPIC_API_KEY=sk-ant-xxx..."
    echo "   Get your key from: https://console.anthropic.com/"
    echo ""
    echo -e "${BLUE}2. Try an analyser:${NC}"
else
    echo -e "${BLUE}1. Try an analyser:${NC}"
fi

echo "   # Code ownership analysis"
echo "   ./utils/code-ownership/ownership-analyser.sh --repo owner/repo"
echo ""
echo "   # Package health analysis"
echo "   ./utils/supply-chain/package-health-analysis/package-health-analyser.sh --repo owner/repo"
echo ""
echo "   # Certificate analysis"
echo "   ./utils/certificate-analyser/cert-analyser.sh example.com"
echo ""
echo "   # See all analysers:"
echo "   find utils -name '*-analyser.sh' -type f"
echo ""

# Warning if critical tools or config are missing
MISSING_TOOLS=false
MISSING_CONFIG=false

if ! command -v syft &> /dev/null; then
    MISSING_TOOLS=true
fi
if ! command -v osv-scanner &> /dev/null; then
    MISSING_TOOLS=true
fi

if [ ! -f "$REPO_ROOT/.env" ] || ! grep -q "^ANTHROPIC_API_KEY=sk-ant-" "$REPO_ROOT/.env" 2>/dev/null; then
    MISSING_CONFIG=true
fi

if [[ "$MISSING_TOOLS" == "true" ]]; then
    echo -e "${YELLOW}⚠ Warning: Some required tools are not installed${NC}"
    echo "Some features may not work. Run ./bootstrap.sh again to install missing tools."
    echo ""
fi

if [[ "$MISSING_CONFIG" == "true" ]]; then
    echo -e "${YELLOW}⚠ Warning: ANTHROPIC_API_KEY is not configured${NC}"
    echo "Claude-enabled scripts will not work without an API key."
    echo "Edit .env and add your API key from https://console.anthropic.com/"
    echo ""
fi
