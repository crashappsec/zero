#!/bin/bash
# Zero Demo Commands - zero-test-org
# Run these in order during your Loom recording

# ============================================
# PRE-RECORDING SETUP (run before recording)
# ============================================

# Hydrate the entire org ahead of time
./zero hydrate zero-test-org all-quick

# Clear terminal
clear

# ============================================
# DEMO COMMANDS (run during recording)
# ============================================

# Scene 2: Building from source
git clone https://github.com/crashappsec/zero.git
cd zero
go build -o zero ./cmd/zero
./zero --help

# Scene 3: Configure credentials
./zero config
./zero config set github_token
./zero config set anthropic_key

# Scene 4: Check prerequisites
./zero checkup
./zero checkup --fix  # Optional: install missing tools

# Scene 5: Codebase tour
ls -la pkg/scanner/
cat config/zero.config.json
ls -la rag/
cat rag/technology-identification/web-frameworks/frontend/react/patterns.md | head -30
ls -la agents/

# Scene 6: Hydrate entire org (will use cache if already done)
./zero hydrate zero-test-org all-quick

# Scene 7: View results
./zero status
./zero serve  # Ctrl+C to stop server when done

# Scene 8: Agent mode (in Claude Code)
# Type: /agent
# Then ask:
#   "Which repos in zero-test-org have the worst bus factor?"

# ============================================
# BACKUP COMMANDS
# ============================================

./zero status
./zero serve --port 8080
./zero checkup

# ============================================
# ALTERNATIVE: SINGLE REPO (if org is slow)
# ============================================

./zero hydrate strapi/strapi all-quick
./zero status
./zero serve
