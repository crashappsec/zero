# Zero Master Plan

**Version:** 5.0.0
**Last Updated:** 2026-01-07
**Status:** Active

This document consolidates all roadmaps into a single source of truth.

---

## Current State

### What's Complete

| Component | Status | Notes |
|-----------|--------|-------|
| 7 Super Scanners | ✅ Complete | code-packages, code-security, code-quality, devops, technology-identification, code-ownership, devx |
| 12 Specialist Agents | ✅ Complete | Cereal, Razor, Gill, Turing, Blade, Phreak, Acid, Dade, Nikon, Joey, Plague, Gibson |
| Agent CLI (`./zero agent`) | ✅ Complete | Interactive agent mode with Zero orchestrator |
| GetSystemInfo Tool | ✅ Complete | Agents can query RAG patterns, scanners, feeds, agents, config |
| RAG Pattern System | ✅ Complete | 23 categories, 400+ patterns |
| Evidence Reports | ✅ Complete | HTML reports via Evidence.dev |
| Hydrate Command | ✅ Complete | Clone + scan with profiles |
| Freshness Tracking | ✅ Complete | Fresh/stale/expired indicators |
| Feed Sync | ✅ Complete | Semgrep rules, RAG rule generation |

### Gaps to Address

| Area | Current State | Priority |
|------|---------------|----------|
| Test Coverage | 6-47% across packages | P1 |
| Web UI Performance | 500-2000ms response times | P2 |
| MCP Integration | Scaffolded, not functional | P3 |
| Reachability Analysis | Not implemented | P3 |

---

## Priority 1: Test Coverage

**Rationale:** Low coverage blocks confident releases and refactoring.

| Package | Current | Target | Complexity |
|---------|---------|--------|------------|
| `pkg/api/handlers` | 0% | 70% | Medium |
| `pkg/scanner/code-packages` | 8% | 70% | High |
| `pkg/scanner/code-security` | 28% | 70% | Medium |
| `pkg/core/scoring` | 0% | 70% | Low |
| `pkg/workflow/hydrate` | 17% | 70% | Medium |

**Approach:**
1. Table-driven tests for scanner detection logic
2. Mock external tools (semgrep, osv-scanner, cdxgen)
3. Integration tests with test fixtures in `testdata/`

---

## Priority 2: Web UI Performance (In Progress)

**Rationale:** Dashboard unusable with current response times.

**Plan:** docs/PERFORMANCE-IMPLEMENTATION-PLAN.md

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1 | SQLite storage layer | In Progress |
| Phase 2 | In-memory caching | Planned |
| Phase 3 | SWR request deduplication | Planned |
| Phase 4 | Incremental sync | Planned |

**Target:** <50ms API responses, <500ms dashboard load

---

## Priority 3: MCP Integration

**Rationale:** Enable Zero as MCP server for IDE integration.

**Current State:** Scaffolded in `pkg/mcp/`, not functional.

**Implementation:**
1. MCP server exposing scanner results
2. Tool definitions for each scanner
3. Resource definitions for analysis data
4. Integration with Claude Desktop / VS Code

**Files:**
- `pkg/mcp/server.go`
- `pkg/mcp/tools.go`
- `pkg/mcp/resources.go`

---

## Priority 4: Reachability Analysis

**Rationale:** Prioritize actually-reachable vulnerabilities.

**Features:**
- Vulnerable code path detection
- Call graph analysis
- Risk prioritization based on reachability

**Complexity:** High - requires static analysis tooling integration

---

## Future Roadmap

### Phase 2 (After Core Stabilization)

| Feature | Description |
|---------|-------------|
| Dependency Graph Visualization | Interactive dependency explorer |
| Cloud Asset Inventory | AWS/Azure/GCP resource discovery |
| Ocular Integration | Code sync and orchestration |
| Chalk Integration | Build-time attestation |
| GitHub/GitLab Org Analysis | Repository security audit |

### Phase 3 (Long-term Vision)

| Feature | Description |
|---------|-------------|
| Database Backend | SQLite/DuckDB/PostgreSQL for multi-user |
| CI/CD Integration | GitHub Actions, GitLab CI templates |
| Remediation Automation | Auto-fix PRs |
| Cloud Service | Hosted Zero with API |

---

## Documentation Cleanup

The following docs should be archived or merged:

| File | Action | Reason |
|------|--------|--------|
| `ENGINEERING_ROADMAP.md` | Archive | Outdated (v3.7.0, Dec 2024), content merged here |
| `docs/ROADMAP.md` | Archive | Content merged here |
| `docs/IMPLEMENTATION-PLAN.md` | Keep | Reference for completed work |
| `docs/PERFORMANCE-IMPLEMENTATION-PLAN.md` | Keep | Active implementation plan |

---

## Quick Reference

### CLI Commands

```bash
# Hydrate a repository
./zero hydrate owner/repo

# Check status
./zero status

# Generate report
./zero report owner/repo

# Enter agent mode
./zero agent

# Sync feeds
./zero feeds semgrep
./zero feeds rag
```

### Agent Invocation

```bash
# From CLI
./zero agent

# Agents available
cereal, razor, gill, turing, blade, phreak, acid, dade, nikon, joey, plague, gibson
```

---

*"Hack the planet!"*
