# Zero Implementation Plan

> Generated from ROADMAP.md and code review findings
> Last updated: 2026-01-08

This document contains prioritized, actionable items ready for GitHub Issues.

---

## P0: Critical Bugs (Fix Immediately)

### Bug: Web Chat UI Freezing
- **Description**: Web chat interface freezes when asking questions - SSE streaming works but UI doesn't update
- **Root cause**: Likely React state update issue or stale closure in useChat hook
- **Files**: `web/src/hooks/useApi.ts`, `web/src/app/chat/page.tsx`
- **Labels**: `bug`, `priority:critical`, `component:web`

### Bug: Unchecked Errors in Package Scanner
- **Description**: 8+ instances of `io.ReadAll()` errors ignored in code-packages scanner
- **Impact**: Network errors cause silent failures, incorrect analysis results
- **Files**: `pkg/scanner/code-packages/code-packages.go` (lines 1855, 2270, 2290, 2634, 2732, 2754, 2802, 2835)
- **Labels**: `bug`, `priority:high`, `component:scanner`

### Bug: WebSocket CORS Allows All Origins
- **Description**: WebSocket upgrader uses `CheckOrigin: func(r *http.Request) bool { return true }`
- **Impact**: Potential CSRF/cross-site WebSocket hijacking in production
- **Files**: `pkg/api/agent/handler.go:29-31`, `pkg/api/ws/hub.go:33-36`
- **Labels**: `bug`, `security`, `priority:high`

---

## P1: High Priority Features

### Feature: Complete Test Coverage
- **Description**: Increase test coverage to 70% across critical packages
- **Current state**: 0-47% coverage in most packages
- **Scope**:
  - [ ] `pkg/api/handlers` (0% → 70%)
  - [ ] `pkg/scanner/code-packages` (8% → 70%)
  - [ ] `pkg/scanner/code-security` (28% → 70%)
  - [ ] `pkg/core/scoring` (0% → 70%)
  - [ ] `pkg/workflow/hydrate` (17% → 70%)
- **Labels**: `enhancement`, `testing`, `priority:high`

### Feature: MCP Server Integration
- **Description**: Enable Zero as MCP server for IDE integration (Claude Desktop, VS Code)
- **Current state**: Scaffolded in `pkg/mcp/`, not functional
- **Scope**:
  - [ ] MCP server exposing scanner results
  - [ ] Tool definitions for each scanner
  - [ ] Resource definitions for analysis data
  - [ ] Integration testing with Claude Desktop
- **Labels**: `enhancement`, `priority:high`, `component:mcp`

### Feature: Fix Web UI Chat Streaming
- **Description**: Complete debugging and fix web chat streaming
- **Current issues**:
  - State updates not triggering re-renders
  - Potential stale closure in useChat hook
  - Missing cleanup on component unmount
- **Files**: `web/src/hooks/useApi.ts`, `web/src/lib/api.ts`
- **Labels**: `enhancement`, `priority:high`, `component:web`

---

## P2: Medium Priority Features

### Feature: Reachability Analysis
- **Description**: Trace calls to vulnerable functions to prioritize truly-reachable vulns
- **Scope**:
  - [ ] Call graph analysis
  - [ ] Vulnerable code path detection
  - [ ] Risk prioritization based on reachability
- **Labels**: `enhancement`, `priority:medium`, `component:scanner`

### Feature: CI/CD Integration
- **Description**: GitHub Actions workflow for automated scanning
- **Scope**:
  - [ ] GitHub Action for Zero scanning
  - [ ] PR comment integration
  - [ ] SARIF export for GitHub Security tab
- **Labels**: `enhancement`, `priority:medium`, `component:ci`

### Feature: SQLite Storage Layer
- **Description**: Add SQLite backend for better performance
- **Current state**: JSON file-based storage
- **Target**: <500ms load times (currently 500-2000ms)
- **Labels**: `enhancement`, `priority:medium`, `component:storage`

### Feature: Incremental Scanning
- **Description**: Only scan changed files for faster repeat scans
- **Scope**:
  - [ ] Git diff detection
  - [ ] Partial scanner execution
  - [ ] Result merging
- **Labels**: `enhancement`, `priority:medium`, `component:scanner`

---

## P3: Code Quality Improvements

### Refactor: Error Handling in Scanners
- **Description**: Add proper error handling for all ignored errors
- **Files affected**:
  - `pkg/scanner/code-packages/code-packages.go`
  - `pkg/scanner/code-security/security.go`
  - `pkg/scanner/technology-identification/technology.go`
  - `pkg/scanner/devops/devops.go`
- **Labels**: `refactor`, `code-quality`, `priority:medium`

### Refactor: Resource Leak Fixes
- **Description**: Fix file descriptor leaks with proper defer patterns
- **Files**:
  - `pkg/scanner/devops/devops.go` (lines 757, 897)
  - `pkg/scanner/code-packages/code-packages.go` (line 2242)
  - `pkg/storage/sqlite/store.go` (lines 387-397)
- **Labels**: `bug`, `code-quality`, `priority:medium`

### Refactor: Extract Duplicate Package Check Functions
- **Description**: `checkNPMPackage` and `checkPyPIPackage` have identical patterns
- **File**: `pkg/scanner/code-packages/code-packages.go` (lines 2260-2298)
- **Solution**: Extract to generic `fetchPackageVersion` function
- **Labels**: `refactor`, `code-quality`, `priority:low`

### Refactor: React Hook Cleanup
- **Description**: Fix useChat hook to properly cleanup streaming on unmount
- **Issues**:
  - Missing cleanup function exposure
  - Stale closure in tool call tracking
  - useWebSocket infinite reconnect potential
- **Labels**: `refactor`, `code-quality`, `component:web`

---

## P4: Future Features

### Feature: Agent Idle Banter (Full Personality Mode)
- **Description**: Agents chat among themselves when idle in "full personality mode"
- **Behavior**:
  - Agents engage in conversations based on their Hackers (1995) movie personas
  - Discuss the Hackers movie, hacking culture, and development topics
  - Reference actual code and findings from current analysis as conversation topics
  - Generate puns and witty banter related to security, coding, and their specialties
  - Cereal makes supply chain jokes, Razor cuts with sharp security puns, etc.
- **Scope**:
  - [ ] Implement "full personality mode" toggle in settings
  - [ ] Create persona-specific banter templates for each agent
  - [ ] Generate context-aware jokes based on current scan results
  - [ ] Agent-to-agent conversation system during idle periods
  - [ ] Hackers movie quotes and references database
  - [ ] Pun generation for security/code topics
- **Example banter**:
  - Cereal: "Hey Razor, saw you found 47 hardcoded secrets. Talk about an open door policy!"
  - Razor: "At least they're not as exposed as Plague's firewall configs. That's what I call 'security theater'!"
  - Acid: "Speaking of theater, remember when we crashed the Gibson? Good times."
  - Joey: "I crashed a build once. Does that count?"
- **Labels**: `enhancement`, `priority:low`, `fun`, `component:agents`

### Feature: Multi-Repo Analysis
- **Description**: Compare security posture across multiple repositories
- **Labels**: `enhancement`, `priority:low`

### Feature: Remediation Automation
- **Description**: Auto-fix PRs for common issues
- **Labels**: `enhancement`, `priority:low`

### Feature: Cloud SBOM Generation
- **Description**: CycloneDX for AWS/Azure/GCP resources
- **Labels**: `enhancement`, `priority:low`

### Feature: PDF Export
- **Description**: Executive summary reports in PDF format
- **Labels**: `enhancement`, `priority:low`

### Feature: Trend Analysis
- **Description**: Track security posture over time
- **Labels**: `enhancement`, `priority:low`

---

## Issue Labels

| Label | Description |
|-------|-------------|
| `priority:critical` | Must fix immediately |
| `priority:high` | Next sprint |
| `priority:medium` | This quarter |
| `priority:low` | Backlog |
| `bug` | Something isn't working |
| `enhancement` | New feature or request |
| `refactor` | Code improvement without behavior change |
| `security` | Security-related issue |
| `testing` | Test coverage improvements |
| `code-quality` | Code quality improvements |
| `component:web` | Web UI related |
| `component:scanner` | Scanner related |
| `component:api` | API server related |
| `component:mcp` | MCP integration |
| `component:ci` | CI/CD integration |
| `component:storage` | Storage layer |
| `component:agents` | Agent system |
| `fun` | Fun/personality features |

---

## Milestones

### v1.0-alpha (Current)
- 7 super scanners working
- 12 specialist agents
- CLI functional
- Web UI experimental

### v1.0-beta (Target: Q1 2026)
- [ ] All P0 bugs fixed
- [ ] Test coverage >50%
- [ ] Web UI stable
- [ ] MCP integration working

### v1.0 (Target: Q2 2026)
- [ ] Test coverage >70%
- [ ] Web UI production-ready
- [ ] CI/CD integration
- [ ] SQLite storage layer
- [ ] Documentation complete
