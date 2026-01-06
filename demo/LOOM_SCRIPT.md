# Zero Demo Script (2-3 minutes)

**Audience:** Developers
**Duration:** 2-3 minutes
**Focus:** Engineering intelligence - ownership, package health, security
**Demo Target:** zero-test-org (GitHub organization with multiple repos)

---

## Pre-Recording Checklist

- [ ] Terminal with dark theme, large font (16pt+)
- [ ] Clear terminal history: `clear`
- [ ] Ensure `GITHUB_TOKEN` and `ANTHROPIC_API_KEY` are set
- [ ] Pre-hydrate the demo org: `./zero hydrate zero-test-org all-quick`
- [ ] Close unnecessary apps/notifications
- [ ] Have the Zero codebase open in your editor for the tour

---

## Scene 1: Hook (10 seconds)

**SAY:**
> "Zero gives you engineering intelligence for any repository - code ownership, package health, security, and more. Let me show you."

---

## Scene 2: Building from Source (20 seconds)

**SAY:**
> "Zero is a Go project. Building it is one command."

**[Show terminal at repo root]**

**TYPE:**
```bash
# Clone the repo (if not already done)
git clone https://github.com/crashappsec/zero.git
cd zero
```

**SAY:**
> "From the repo root, just run go build."

**TYPE:**
```bash
go build -o zero ./cmd/zero
```

**[Build completes quickly]**

**SAY:**
> "That's it. You now have the zero binary. Let's verify it works."

**TYPE:**
```bash
./zero --help
```

**[Shows help output with available commands]**

---

## Scene 3: Configure Credentials (20 seconds)

**SAY:**
> "Zero needs a GitHub token to clone repos and an Anthropic key for AI agents. The config command makes this easy."

**TYPE:**
```bash
./zero config
```

**[Shows current credential status - likely empty]**

**SAY:**
> "Let's add a GitHub token. Use a fine-grained token scoped to the repos you want to analyze."

**TYPE:**
```bash
./zero config set github_token
```

**[Prompts for token, enter it - shown masked]**

**SAY:**
> "And the Anthropic key for AI agent features."

**TYPE:**
```bash
./zero config set anthropic_key
```

**[Prompts for key, enter it - shown masked]**

**SAY:**
> "Credentials are stored securely in ~/.zero/credentials.json with restricted permissions. You can also use environment variables GITHUB_TOKEN and ANTHROPIC_API_KEY instead."

---

## Scene 4: Check Prerequisites (20 seconds)

**SAY:**
> "Zero needs a few external tools. The checkup command tells you exactly what's configured and what's missing."

**[Show terminal]**

**TYPE:**
```bash
./zero checkup
```

**[Shows Zero logo, prerequisites, GitHub token status, accessible repos, external tools, scanner compatibility]**

**SAY:**
> "It checks your tools, validates your GitHub token, shows which repos you can access, and which scanners are ready. Green means good to go. If anything's missing, run checkup with --fix to install it."

**TYPE:**
```bash
./zero checkup --fix
```

**[Optional: show it offering to install missing tools]**

---

## Scene 5: Quick Codebase Tour (50 seconds)

**[Show editor or terminal with tree view]**

**SAY:**
> "Zero is organized around scanners, profiles, agents, and a RAG knowledge base. Let me show you."

**TYPE:**
```bash
ls -la pkg/scanner/
```

**SAY:**
> "Seven super scanners - packages, security, quality, devops, technology detection, ownership, and developer experience."

**TYPE:**
```bash
cat config/zero.config.json
```

**SAY:**
> "Profiles control which scanners run and how. 'all-quick' runs everything with fast defaults. 'all-complete' enables every feature. You can create custom profiles for your workflow."

**TYPE:**
```bash
ls -la rag/
```

**SAY:**
> "The RAG directory is the brain. It contains structured knowledge that powers everything - technology detection patterns, supply chain specs, DORA metrics definitions, and more."

**TYPE:**
```bash
cat rag/technology-identification/web-frameworks/frontend/react/patterns.md | head -30
```

**SAY:**
> "Here's a React detection pattern. These get converted to Semgrep rules at runtime - so adding new technology detection is just adding a markdown file."

**TYPE:**
```bash
ls -la agents/
```

**SAY:**
> "Specialist AI agents - each named after characters from Hackers. They use the RAG knowledge to provide deep domain expertise."

---

## Scene 6: Hydrate an Organization (30 seconds)

**SAY:**
> "Let's analyze an entire GitHub organization. One command scans every repo."

**TYPE:**
```bash
./zero hydrate zero-test-org all-quick
```

**[Show repos being cloned and scanned in parallel]**

**SAY:**
> "Zero clones each repo, then runs all seven scanners in parallel across the entire org. This is how you get visibility at scale - not repo by repo, but your whole portfolio at once."

**[If already hydrated, show the cached result with multiple repos]**

---

## Scene 7: View Results (45 seconds)

**SAY:**
> "Now let's explore what we found."

**TYPE:**
```bash
./zero status
```

**[Show status with multiple repos and freshness indicators]**

**SAY:**
> "Status shows all scanned repos with freshness indicators. Green is fresh, yellow is stale."

**TYPE:**
```bash
./zero serve
```

**[Browser opens web UI at localhost:3001]**

**SAY (while navigating):**
> "The web UI shows all your projects and their analysis data."

**[Click on a project]**

> "Code ownership - bus factor, top contributors, CODEOWNERS status. Critical for maintainability."

**[Navigate to packages/dependencies]**

> "Full SBOM - every package, version, and license. Essential for compliance."

**[Navigate to security findings]**

> "Package health scores, vulnerabilities, and security findings."

**[Press Ctrl+C to stop server]**

---

## Scene 8: Agent Mode (30 seconds)

**SAY:**
> "But here's where it gets powerful. Zero has AI agents for deeper analysis."

**TYPE:**
```bash
# In Claude Code:
/agent
```

**[Zero greets you]**

**SAY:**
> "Meet Zero - the orchestrator who delegates to specialists."

**TYPE (in agent mode):**
```
Which repos in zero-test-org have the worst bus factor?
```

**[Let Zero/Gibson respond with org-wide analysis]**

**SAY:**
> "Gibson analyzes engineering metrics across the whole org. Cereal handles supply chain. Each agent has deep domain expertise and can work at scale."

---

## Scene 9: Wrap Up (10 seconds)

**SAY:**
> "That's Zero - engineering intelligence from your terminal. Clone, scan, understand. Check out the GitHub repo to get started."

**[Show GitHub URL or end screen]**

---

## Backup Commands

```bash
./zero status
./zero serve --port 8080
./zero checkup  # Verify installation
```

---

## Post-Production Notes

1. **Speed up** hydrate if needed (2x-4x)
2. **Cut** loading spinners > 3 seconds
3. Add **callouts** pointing to:
   - RAG pattern structure (tiered detection)
   - "Converted to Semgrep rules" moment
   - Bus factor indicator
   - Package health scores
   - SBOM/license distribution
   - Contributor breakdown

---

## Key Talking Points

| Topic | What to Highlight |
|-------|-------------------|
| **RAG Knowledge** | Structured patterns, converted to Semgrep rules, easy to extend |
| **Ownership** | Bus factor, CODEOWNERS, contributor distribution |
| **Package Health** | Dependency freshness, maintenance status, health scores |
| **SBOM** | Complete inventory, license distribution |
| **Security** | Vulnerabilities, secrets (mention but don't dwell) |
| **Agents** | Domain specialists, powered by RAG knowledge |

---

## Alternative Demo Targets

| Target | Why |
|--------|-----|
| `zero-test-org` | Multiple repos, shows scale (recommended) |
| `strapi/strapi` | Single repo, popular CMS, lots of findings |
| `calcom/cal.com` | Single repo, modern Next.js stack |
