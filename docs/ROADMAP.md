<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Zero Roadmap

> Last updated: 2025-12-16 | Current version: 6.2.0

## Current State Assessment

### What's Working Well

| Component | Status | Coverage |
|-----------|--------|----------|
| 8 Super Scanners | ✅ Complete | All implemented and registered |
| SBOM Generation | ✅ Complete | CycloneDX 1.5 via cdxgen/syft |
| Package Analysis | ✅ Complete | 12 features (vulns, health, licenses, etc.) |
| Tech-ID + Semgrep | ✅ Complete | RAG patterns → Semgrep rules |
| Code Ownership | ✅ Complete | Adaptive periods, historical stats |
| Hydrate Command | ✅ Complete | Clone + scan with profiles |
| Agent Definitions | ✅ Complete | 12 specialist agents defined |

### Gaps Identified

| Area | Current State | Gap |
|------|---------------|-----|
| Test Coverage | 6-47% | Most packages at 0% |
| Agent Mode | Defined only | `/agent` not implemented |
| Report Command | Documented | Not implemented |
| MCP Integration | Scaffolded | Not functional |
| RAG Converter | Working | Edge cases, validation needed |

### Test Coverage Summary

```
pkg/config           47.2%  ✅ Good
pkg/scanner          11.9%  ⚠️  Needs improvement
pkg/scanners/tech-id  6.8%  ⚠️  Recently added tests
All other packages    0.0%  ❌ Critical gap
```

---

## Prioritized Roadmap

### Priority 1: Test Coverage (Critical)

**Rationale**: Low test coverage blocks confident releases and refactoring.

| Package | Priority | Complexity | Impact |
|---------|----------|------------|--------|
| `pkg/hydrate` | P1 | Medium | Core functionality |
| `pkg/scanners/sbom` | P1 | Medium | Source of truth |
| `pkg/scanners/package-analysis` | P1 | High | Supply chain |
| `pkg/scanners/code-security` | P2 | Medium | Security analysis |
| `pkg/scanners/crypto` | P2 | Medium | Crypto analysis |
| `pkg/terminal` | P2 | Low | Output display |
| `pkg/scanners/devops` | P3 | High | Multiple tools |
| `pkg/scanners/code-quality` | P3 | Medium | Quality metrics |
| `pkg/scanners/code-ownership` | P3 | Medium | Git analysis |

**Target**: 50% coverage across all packages

**Approach**:
1. Table-driven tests for scanner detection logic
2. Mock external tools (semgrep, osv-scanner, etc.)
3. Integration tests with test fixtures
4. Add test fixtures in `testdata/` directories

---

### Priority 2: Report Command

**Rationale**: Users need actionable output beyond raw JSON.

**Status**: Documented in architecture but not implemented

**Implementation**:
```
./zero report <org/repo>
./zero report expressjs/express --format html
./zero report expressjs/express --format markdown
./zero report expressjs/express --format json
```

**Features**:
- Executive summary with risk score
- Critical findings highlighted
- Remediation recommendations
- Export to HTML/Markdown/JSON

**Files to create/modify**:
- `cmd/zero/cmd/report.go` - New command
- `pkg/report/generator.go` - Report generation
- `pkg/report/templates/` - HTML/Markdown templates

---

### Priority 3: Agent Mode (`/agent`)

**Rationale**: Core value proposition - AI-assisted security analysis

**Status**: Agent definitions exist in CLAUDE.md, `/agent` command defined but not functional

**Implementation**:
1. `/agent` enters interactive mode with Zero orchestrator
2. Zero delegates to specialists via Task tool
3. Agents read cached analysis data
4. Context loading modes: summary, critical, full

**Files to create/modify**:
- `.claude/commands/agent.md` - Update with full implementation
- Agent context loading from `.zero/repos/<project>/analysis/`

---

### Priority 4: RAG Converter Improvements

**Rationale**: Tech-ID Semgrep integration is working but has edge cases

**Current Limitations**:
- Not all RAG pattern types converted
- Secret detection rules need validation
- AI/ML patterns need more coverage
- Error handling could be more graceful

**Improvements**:
1. Add support for all pattern types in `rag/technology-identification/`
2. Validate generated Semgrep rules syntax
3. Add fallback detection for unsupported patterns
4. Better error messages for rule generation failures

**Files**:
- `pkg/scanners/tech-id/rag_converter.go`
- `pkg/scanners/tech-id/rules.go`

---

### Priority 5: MCP Integration

**Rationale**: Enable Zero as MCP server for IDE integration

**Status**: Scaffolded in `pkg/mcp/`, not functional

**Implementation**:
1. MCP server exposing scanner results
2. Tool definitions for each scanner
3. Resource definitions for analysis data
4. Integration with Claude Desktop / VS Code

**Files**:
- `pkg/mcp/server.go`
- `pkg/mcp/tools.go`
- `pkg/mcp/resources.go`

---

### Priority 6: Documentation Cleanup

**Rationale**: Docs have legacy references and inconsistencies

**Tasks**:
1. Remove references to old scanner names in all docs
2. Update architecture diagrams with 8 scanner model
3. Add examples for each scanner
4. Document all CLI flags and options
5. Add troubleshooting guide

**Files**:
- `docs/scanners/*.md` - Scanner-specific docs
- `docs/GETTING_STARTED.md` - User guide
- `docs/architecture/*.md` - Technical docs

---

## Future Considerations

### Phase 2 (After Core Stabilization)

| Feature | Description | Complexity |
|---------|-------------|------------|
| Incremental Scanning | Only scan changed files | High |
| CI/CD Integration | GitHub Actions, GitLab CI | Medium |
| Dashboard | Web UI for results | High |
| Custom Rules | User-defined Semgrep rules | Medium |
| SARIF Export | Standard security format | Low |
| Multi-Repo Analysis | Compare across repos | Medium |

### Phase 3 (Long-term Vision)

| Feature | Description |
|---------|-------------|
| Cloud Service | Hosted Zero with API |
| Enterprise Features | SSO, teams, policies |
| Remediation Automation | Auto-fix PRs |
| Threat Intelligence | Real-time vuln feeds |

---

## Quick Wins (Can Do Now)

1. **Add tests for `pkg/hydrate`** - Core module, high value
2. **Add tests for `pkg/terminal`** - Low complexity, stable interface
3. **Clean up legacy docs** - Easy, improves user experience
4. **Add SARIF output** - Standard format, good for CI integration

---

## Recommended Next Steps

Based on current state and gaps, recommended order:

1. **Test Coverage Sprint** - Target 50% across packages
2. **Report Command** - Visible user value
3. **Agent Mode Polish** - Core differentiator
4. **RAG Converter Hardening** - Reliability
5. **MCP Integration** - IDE integration

---

*"Hack the planet!"*
