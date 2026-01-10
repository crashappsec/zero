# Zero Modernization Review: Aligning with Gen-AI Architecture Patterns

## Executive Summary

Zero is a well-architected engineering intelligence platform, but its terminology and messaging predate the modern gen-AI era. This review identifies strategic changes to align Zero with contemporary patterns (MCP, tool-use agents) while clarifying its positioning as an **engineering intelligence platform** rather than a security scanner.

**Key Recommendations:**
- Rename "scanners" to "analyzers" (NOT "tools" - that term is reserved for MCP-callable functions)
- Rebalance messaging from security-first to engineering intelligence across 6 dimensions
- Expand MCP tool coverage for all analysis domains
- Keep the Hackers character theme but default to professional voice
- Introduce use-case profiles (security-focused, engineering-health, compliance)

---

## Part 1: Terminology Modernization

### The "Scanner" Problem

**Current state:** Zero uses "scanners" as its primary term (7 super scanners).

**Issue:** In modern gen-AI systems, "tools" are callable functions for agents. Zero's scanners are NOT agent-callable - they produce artifacts that agents consume. Using "tools" would create confusion with the MCP layer.

### Recommendation: "Analyzers" (Not "Tools")

| Current Term | Proposed Term | Rationale |
|--------------|---------------|-----------|
| Scanner | **Analyzer** | Describes function; distinct from MCP tools |
| Super Scanner | **Analysis Domain** | Clearer abstraction |
| Feature | **Capability** | Modern product language |
| Scan | **Analyze** | Verb alignment |
| Hydrate | **Onboard** | More intuitive (keep hydrate as alias) |

### Why NOT "Tools"

```
User Query → Zero Agent → MCP Tools → Analysis Data
                              ↑
                        These are tools
                        (callable by agents)

Analyzers run in background, produce JSON artifacts.
MCP tools are how agents ACCESS those artifacts.
Conflating them creates confusion.
```

### Code Impact

| Location | Change |
|----------|--------|
| `pkg/scanner/` | Rename to `pkg/analyzer/` |
| `interface.go` | `Scanner` → `Analyzer` |
| `runner.go` | `NativeRunner` → `AnalysisRunner` |
| `zero.config.json` | `scanners` → `analyzers` |
| CLI commands | `zero scan` → `zero analyze` |

---

## Part 2: Positioning Clarification

### Current Problem

The tagline says "engineering intelligence" but agent prompts say "security analysis orchestrator". This creates mixed messaging.

**Evidence from codebase:**
- CLAUDE.md: "engineering intelligence platform" ✓
- orchestrator/agent.md: "security analysis orchestrator" ✗
- 7 analyzers cover 6 dimensions, only 2 are security-focused

### Recommended Positioning

**Primary tagline:** "Engineering Intelligence for Every Repository"

**Ten Dimensions of Intelligence:**

| Dimension | Analyzer | Focus |
|-----------|----------|-------|
| Security | code-security | Vulnerabilities, secrets, crypto |
| Supply Chain | code-packages | Dependencies, SBOMs, health |
| Quality | code-quality | Tech debt, complexity, coverage |
| DevOps | devops | IaC security, containers, DORA metrics |
| Technology | technology-identification | Stack detection, ML-BOM |
| Team | code-ownership, devx | Bus factor, onboarding |
| **Build** | **build** | CI/CD optimization, caching, cost, parallelization |
| **Tool Config** | **tool-config** | Linter, TypeScript, bundler, test config validation |
| **Infra Config** | **infra-config** | Docker, K8s, Terraform, Helm config validation |
| **Governance** | **repo-governance** | Branch protection, required reviews, security features |

### Agent Prompt Updates

**orchestrator/agent.md:**
```diff
- You are a security analysis orchestrator.
+ You are an engineering intelligence orchestrator.
```

**Non-security agents (Gibson, Nikon, Joey, etc.):**
- Remove security-heavy language
- Emphasize engineering/operations focus
- Keep security agents (Razor, Cereal, Gill) security-focused - that IS their domain

---

## Part 3: MCP Architecture Alignment

### Current MCP Tools (10)

```
list_projects, get_project_summary, get_vulnerabilities,
get_malcontent, get_technologies, get_package_health,
get_licenses, get_secrets, get_crypto_issues, get_analysis_raw
```

### Missing Coverage

| Missing Tool | Analyzer | Purpose |
|--------------|----------|---------|
| `get_devops_findings` | devops | IaC, containers, Actions |
| `get_code_quality` | code-quality | Tech debt, complexity |
| `get_ownership_metrics` | code-ownership | Bus factor, contributors |
| `get_dora_metrics` | devops | Deployment freq, MTTR |
| `get_devx_analysis` | devx | Onboarding, tooling |

### Architecture Model

```
┌─────────────────────────────────────────────────────┐
│                    User Query                        │
└─────────────────────┬───────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────┐
│              Zero (Orchestrator Agent)               │
│   "Engineering intelligence orchestrator"            │
└──────┬──────────────────┬───────────────────┬───────┘
       │                  │                   │
       ▼                  ▼                   ▼
┌──────────────┐  ┌───────────────┐  ┌───────────────┐
│  MCP Tools   │  │   Specialist  │  │   Standard    │
│  (Data API)  │  │    Agents     │  │    Tools      │
├──────────────┤  ├───────────────┤  ├───────────────┤
│ get_vulns    │  │ Cereal        │  │ Read          │
│ get_sbom     │  │ Razor         │  │ Grep          │
│ get_dora     │  │ Gibson        │  │ WebSearch     │
│ ...          │  │ ...           │  │ ...           │
└──────────────┘  └───────────────┘  └───────────────┘
       │                  │
       ▼                  ▼
┌─────────────────────────────────────────────────────┐
│           Analysis Artifacts (.zero/analysis/)       │
│  code-packages.json, code-security.json, etc.        │
└─────────────────────────────────────────────────────┘
       ▲
       │ (produced by)
┌─────────────────────────────────────────────────────┐
│                    Analyzers                         │
│  (Background processes, NOT agent-callable tools)    │
└─────────────────────────────────────────────────────┘
```

**Key Insight:** Analyzers are NOT MCP tools. They are background processes that produce artifacts. MCP tools are the interface for agents to ACCESS those artifacts.

---

## Part 4: Agent System Evolution

### Keep the Hackers Theme

**Recommendation:** Retain character names but modernize presentation.

**Rationale:**
- Unique brand differentiator
- Characters map well to domains
- VOICE modes already provide professional alternatives
- Default to `VOICE:minimal` in production

### Updated Agent Domain Mapping

| Agent | Character | Domain (Modernized) |
|-------|-----------|---------------------|
| Zero | Zero Cool | Engineering Intelligence Orchestration |
| Cereal | Cereal Killer | Supply Chain Intelligence |
| Razor | Razor | Security Analysis |
| Gill | Gill Bates | Cryptography |
| Hal | Hal | AI/ML Security |
| Blade | Blade | Compliance & Audit |
| Phreak | Phantom Phreak | Legal & Licensing |
| Acid | Acid Burn | Frontend Engineering |
| Flushot | Flu Shot | Backend Engineering |
| Nikon | Lord Nikon | Architecture |
| Joey | Joey | Build & CI/CD |
| Plague | The Plague | Infrastructure |
| Gibson | The Gibson | Engineering Metrics |

### Agent Definition Updates Required

For each agent in `agents/*/agent.md`:

1. **Update Role section** - use "engineering intelligence" framing
2. **Update Data Sources** - align with v4.0 analyzer naming
3. **Add MCP Tool References** - document which tools agent should use
4. **Rebalance VOICE sections** - reduce security-heavy language in non-security agents

---

## Part 5: Configuration & UX

### Profile Evolution

**Current profiles** are analyzer-centric. **Proposed profiles** are use-case-centric:

```json
{
  "profiles": {
    "quick": {
      "description": "Fast analysis across all dimensions"
    },
    "complete": {
      "description": "Full analysis with all capabilities"
    },
    "security-focused": {
      "description": "Deep security and supply chain analysis",
      "analyzers": ["code-packages", "code-security"]
    },
    "engineering-health": {
      "description": "Team productivity and code health",
      "analyzers": ["code-quality", "code-ownership", "developer-experience", "devops"]
    },
    "compliance": {
      "description": "Audit and compliance readiness"
    }
  }
}
```

### CLI Command Evolution

```bash
# Current → Proposed
zero hydrate <repo>    →  zero onboard <repo>   # (hydrate kept as alias)
zero scan <repo>       →  zero analyze <repo>
zero list              →  zero list             # (unchanged)

# New commands
zero query <repo> <domain>   # Query specific analysis domain
zero export <repo> [format]  # Export to SARIF, CSV, etc.
```

---

## Part 6: Implementation Roadmap

### Phase 1: Non-Breaking (Quick Wins)
- [ ] Update agent prompts for balanced messaging
- [ ] Add missing MCP tools (devops, quality, ownership, dora, devx)
- [ ] Update CLAUDE.md documentation
- [ ] Add CLI command aliases (onboard, analyze)
- [ ] Update README positioning

### Phase 2: Configuration
- [ ] Add use-case profiles (security-focused, engineering-health, compliance)
- [ ] Update profile descriptions
- [ ] Document migration path

### Phase 3: Breaking Changes
- [ ] Rename `pkg/scanner/` → `pkg/analyzer/`
- [ ] Update interface names
- [ ] Update CLI primary commands
- [ ] Create migration guide
- [ ] Deprecation warnings for old commands

### Phase 4: Web UI
- [ ] Update report navigation (6 dimensions, not security-first)
- [ ] Add "Engineering Health" dashboard
- [ ] Update branding/messaging

---

## Part 7: Trade-off Analysis

| Change | Benefit | Cost | Risk |
|--------|---------|------|------|
| Scanner → Analyzer | Clearer mental model, MCP distinction | Breaking change, code refactor | Medium |
| Balanced positioning | Broader market appeal | Documentation rewrite | Low |
| Expanded MCP tools | Better agent capabilities | Development effort | Low |
| Keep Hackers theme | Brand uniqueness | - | Low |
| Use-case profiles | Better UX | Config migration | Low |
| CLI renames | More intuitive | User adaptation | Medium |

---

## Part 8: Open Questions

### 1. Individual Analyzer Naming

Should individual analyzers also be renamed for brevity?

| Current | Option A (Shorter) | Option B (Clearer) |
|---------|--------------------|--------------------|
| code-security | security | security-analysis |
| code-packages | supply-chain | dependencies |
| code-quality | quality | quality-metrics |
| technology-identification | technology | tech-stack |
| code-ownership | ownership | team-health |
| developer-experience | devx | developer-experience |

### 2. Agent Consolidation

13 agents may be too many for users to remember. Consider consolidating:

| Current | Potential Merge |
|---------|-----------------|
| Acid (Frontend) + Flushot (Backend) | → Application Engineering |
| Joey (Build) + Plague (DevOps) | → Platform Engineering |
| Cereal (Supply Chain) + Phreak (Legal) | → Supply Chain & Compliance |

### 3. Backward Compatibility Timeline

How long to maintain deprecated commands?
- Option A: 1 release with warnings, then remove
- Option B: 2-3 releases with deprecation warnings
- Option C: Permanent aliases (no removal)

### 4. Existing GitHub Issues Alignment

How do open issues fit into this roadmap?

| Issue | Fits Where |
|-------|------------|
| #30 Incremental scanning | Phase 3 (could be "incremental analysis") |
| #29 Reachability analysis | New capability for code-packages analyzer |
| #40 Supabase backend | Independent track |
| #31-33 Refactoring | Phase 3 during rename |

---

## Summary

Zero is well-positioned as an "engineering intelligence platform" but needs terminology and messaging alignment:

| Priority | Change | Impact |
|----------|--------|--------|
| 1 | Rename scanners to analyzers | Distinguishes from MCP tools |
| 2 | Rebalance messaging | Engineering intelligence, not security-first |
| 3 | Expand MCP coverage | Add missing domain tools |
| 4 | Keep Hackers theme | Unique differentiator, use minimal voice |
| 5 | Use-case profiles | security-focused, engineering-health, compliance |
| 6 | Phased rollout | Non-breaking first, then breaking with migration |

---

## Part 9: Hybrid Cache Architecture

### Current Problem: "Hydrate" is Confusing

The term "hydrate" doesn't clearly communicate what's happening. Users must:
1. Run `zero hydrate` first (mandatory)
2. Wait for all analyzers to complete
3. Then ask questions

This is a poor UX for quick queries and doesn't align with how modern AI tools work.

### Recommendation: Cache-Centric Architecture

**Rename terminology:**
| Current | Proposed | Rationale |
|---------|----------|-----------|
| hydrate | `cache generate` | Describes what it does |
| .zero/repos/.../analysis/ | `.zero/cache/` | It's a cache |
| "hydrated project" | "cached project" | Clearer mental model |

**New CLI commands:**
```bash
zero cache generate <repo>     # Generate full cache (was: hydrate)
zero cache status              # Show cache freshness (was: status)
zero cache invalidate <repo>   # Clear cache for repo
zero cache gc                  # Garbage collect old caches
```

### Hybrid Execution Model

Instead of "generate everything upfront", use lazy evaluation with TTL-based freshness:

```
┌─────────────────────────────────────────────────────────────┐
│                   Agent/MCP Tool Request                     │
│              "Get vulnerabilities for express"               │
└───────────────────────────┬─────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Cache Lookup (Smart)                       │
│  1. Does cache exist for this repo + analyzer?               │
│  2. Is it fresh enough? (TTL per analyzer type)              │
└───────────┬─────────────────────────────────────┬───────────┘
            │                                     │
     Cache Hit (Fresh)                    Cache Miss (Stale/None)
            │                                     │
            ▼                                     ▼
┌───────────────────────┐          ┌──────────────────────────┐
│  Return cached data   │          │  Run analyzer on-demand  │
│  (fast: ~10ms)        │          │  Update cache            │
└───────────────────────┘          │  Return fresh data       │
                                   │  (slower: 5-60s)         │
                                   └──────────────────────────┘
```

### Freshness TTLs by Analyzer Type

| Analyzer | Default TTL | Rationale |
|----------|-------------|-----------|
| vulnerabilities | 1 hour | CVE data changes frequently |
| sbom | 24 hours | Dependencies change with commits |
| secrets | 24 hours | Secrets change with commits |
| licenses | 7 days | Licenses rarely change |
| ownership | 7 days | Contributor patterns stable |
| dora | 1 hour | Deployment data changes |
| technology | 7 days | Tech stack rarely changes |

### Benefits of Hybrid Model

1. **No mandatory "hydrate" step** - Just ask questions, cache builds lazily
2. **Always fresh-enough data** - TTL ensures relevance
3. **Fast repeated queries** - Cache hits are instant
4. **Efficient compute** - Only analyze what's requested
5. **Optional warm-up** - `cache generate` pre-populates everything

### Implementation Notes

```go
// MCP tool with hybrid cache
func (s *Server) handleGetVulnerabilities(ctx context.Context, input VulnsInput) (*Result, error) {
    // Check cache with TTL
    if cached := s.cache.Get(input.Project, "vulns"); cached.IsFresh(1 * time.Hour) {
        return cached.Data, nil
    }

    // Run analyzer on-demand (only the vulns capability)
    result, err := s.runAnalyzer(ctx, input.Project, "vulnerabilities")
    if err != nil {
        return nil, err
    }

    // Update cache
    s.cache.Set(input.Project, "vulns", result)
    return result, nil
}
```

---

## Part 10: Flattened Analyzer Architecture

### Current Problem: "Super Analyzers" Hide Complexity

The current architecture has 7 "super analyzers" with 45+ internal "features":

```
pkg/scanner/
├── code-packages/           # Super analyzer
│   ├── scanner.go           # 2000+ lines
│   ├── config.go
│   ├── generation.go        # Feature: SBOM generation
│   ├── vulns.go             # Feature: Vulnerability scanning
│   ├── health.go            # Feature: Package health
│   ├── malcontent.go        # Feature: Malware detection
│   ├── licenses.go          # Feature: License analysis
│   └── ... (9 more features)
```

**Problems:**
1. **Hard to find code** - Where's the typosquatting logic? Buried in code-packages.
2. **No knowledge co-location** - RAG patterns in `/rag/`, agent knowledge in `/agents/`, code in `/pkg/scanner/`
3. **Monolithic output** - One JSON file per super analyzer, hard to cache granularly
4. **Feature dependencies unclear** - Which features need SBOM first?

### Recommendation: Individual Analyzers with Co-located Knowledge

Flatten the hierarchy so each capability is a self-contained analyzer:

```
pkg/analyzers/
├── sbom/                        # Individual analyzer
│   ├── analyzer.go              # Implementation (~200 lines)
│   ├── analyzer_test.go         # Tests
│   ├── config.go                # Configuration
│   ├── knowledge/               # Co-located knowledge
│   │   ├── patterns/            # Detection patterns (from RAG)
│   │   │   └── ecosystems.json
│   │   └── guidance/            # Interpretation guidance
│   │       └── sbom-analysis.md
│   ├── prompts/                 # Agent prompts for this analyzer
│   │   └── interpret-sbom.md
│   └── README.md                # Documentation
│
├── vulnerabilities/
│   ├── analyzer.go
│   ├── knowledge/
│   │   ├── patterns/
│   │   │   └── cwe-patterns.json
│   │   └── guidance/
│   │       └── cvss-scoring.md
│   ├── prompts/
│   │   └── triage-vulns.md
│   └── README.md
│
├── malcontent/
│   ├── analyzer.go
│   ├── knowledge/
│   │   ├── patterns/
│   │   │   ├── behavioral-signals.json
│   │   │   └── obfuscation-patterns.json
│   │   └── guidance/
│   │       └── threat-assessment.md
│   ├── prompts/
│   │   └── investigate-malcontent.md
│   └── README.md
│
├── secrets/
├── licenses/
├── ownership/
├── dora-metrics/
├── iac-security/
├── container-security/
├── ... (total: ~25-30 individual analyzers)
```

### Benefits of Flattened Structure

| Benefit | Description |
|---------|-------------|
| **Discoverability** | `ls pkg/analyzers/` shows all capabilities |
| **Co-location** | Code + knowledge + prompts in one place |
| **Granular caching** | Cache each analyzer independently |
| **Easier testing** | Test one analyzer in isolation |
| **Clear dependencies** | `vulnerabilities` depends on `sbom` (explicit) |
| **Agent alignment** | Each agent maps to specific analyzers |

### Analyzer Interface (Updated)

```go
// pkg/analyzers/interface.go
type Analyzer interface {
    // Identity
    Name() string                    // e.g., "vulnerabilities"
    Domain() string                  // e.g., "security" | "supply-chain" | "devops"
    Description() string

    // Execution
    Run(ctx context.Context, opts *Options) (*Result, error)

    // Dependencies
    Dependencies() []string          // Other analyzers that must run first

    // Knowledge (for agents)
    KnowledgePath() string           // Path to knowledge/ directory
    PromptsPath() string             // Path to prompts/ directory

    // Caching
    DefaultTTL() time.Duration       // How long results stay fresh
}
```

### Dependency Graph (Explicit)

```
sbom ─────────────────┬──► vulnerabilities
                      ├──► licenses
                      ├──► health
                      ├──► malcontent
                      └──► typosquats

git-history ──────────┬──► ownership
                      ├──► dora-metrics
                      └──► churn

technology-detection ─┬──► ml-models
                      ├──► frameworks
                      └──► devx
```

### Migration from Super Analyzers

| Super Analyzer | Becomes Individual Analyzers |
|----------------|------------------------------|
| code-packages | sbom, vulnerabilities, licenses, health, malcontent, typosquats, confusion, deprecations, duplicates, reachability, provenance, bundle |
| code-security | sast, secrets, api-security, ciphers, keys, random, tls, certificates |
| code-quality | tech-debt, complexity, test-coverage, documentation |
| devops | iac, containers, github-actions, dora-metrics, git-analysis |
| technology-identification | technology-detection, ml-models, frameworks, datasets, ai-security, ai-governance |
| code-ownership | contributors, bus-factor, codeowners, orphans, churn, patterns |
| developer-experience | onboarding, tool-sprawl, workflow |

### Agent-Analyzer Mapping (Clear)

Each agent explicitly references which analyzers it uses:

```yaml
# agents/supply-chain/agent.yaml (new format)
name: cereal
domain: Supply Chain Intelligence
character: Cereal Killer

analyzers:
  primary:
    - sbom
    - vulnerabilities
    - malcontent
    - health
  secondary:
    - licenses
    - typosquats

knowledge_sources:
  - pkg/analyzers/sbom/knowledge/
  - pkg/analyzers/vulnerabilities/knowledge/
  - pkg/analyzers/malcontent/knowledge/
```

### Output Structure (Granular)

Instead of one big JSON per super analyzer:

```
.zero/cache/expressjs/express/
├── sbom.json                    # CycloneDX SBOM
├── vulnerabilities.json         # CVEs with CVSS
├── licenses.json                # License analysis
├── malcontent.json              # Behavioral findings
├── secrets.json                 # Detected secrets
├── ownership.json               # Contributor analysis
├── dora-metrics.json            # DORA calculations
└── _meta.json                   # Cache metadata (TTLs, timestamps)
```

### MCP Tool Alignment

Each analyzer maps 1:1 to an MCP tool:

| Analyzer | MCP Tool | Returns |
|----------|----------|---------|
| sbom | `get_sbom` | CycloneDX JSON |
| vulnerabilities | `get_vulnerabilities` | CVE list |
| malcontent | `get_malcontent` | Behavioral findings |
| licenses | `get_licenses` | License inventory |
| secrets | `get_secrets` | Secret detections |
| ownership | `get_ownership` | Contributor stats |
| dora-metrics | `get_dora_metrics` | DORA calculations |

---

## Part 11: Knowledge Architecture

### Current Problem: Knowledge is Scattered

```
Current layout:
├── rag/                         # RAG patterns (detection)
│   └── technology-identification/
├── agents/                      # Agent knowledge (interpretation)
│   └── supply-chain/
│       └── knowledge/
└── pkg/scanner/                 # Code (no knowledge)
```

Developers must look in 3 places to understand one capability.

### Recommendation: Co-located Knowledge

Move knowledge INTO each analyzer:

```
pkg/analyzers/malcontent/
├── analyzer.go                  # Code
├── knowledge/
│   ├── patterns/                # Detection patterns (was in /rag/)
│   │   ├── behavioral-signals.json
│   │   ├── obfuscation.json
│   │   └── network-indicators.json
│   └── guidance/                # Interpretation (was in /agents/)
│       ├── threat-levels.md
│       ├── investigation-steps.md
│       └── false-positive-patterns.md
├── prompts/                     # Agent prompts for this analyzer
│   ├── triage.md
│   ├── investigate.md
│   └── report.md
└── README.md                    # Human documentation
```

### Knowledge Types

| Type | Purpose | Format | Used By |
|------|---------|--------|---------|
| **patterns/** | Detection rules | JSON | Analyzer code |
| **guidance/** | Interpretation frameworks | Markdown | Agents |
| **prompts/** | Task-specific prompts | Markdown | Agents |
| **README.md** | Human documentation | Markdown | Developers |

### Agent Knowledge Loading

Agents dynamically load knowledge from their assigned analyzers:

```go
// Agent loads knowledge from analyzers it uses
func (a *Agent) LoadKnowledge() error {
    for _, analyzerName := range a.Definition.Analyzers.Primary {
        analyzer := registry.Get(analyzerName)

        // Load guidance into agent context
        guidance := loadMarkdownFiles(analyzer.KnowledgePath() + "/guidance/")
        a.Context.AddKnowledge(analyzerName, guidance)

        // Load prompts for task-specific responses
        prompts := loadMarkdownFiles(analyzer.PromptsPath())
        a.Prompts[analyzerName] = prompts
    }
    return nil
}
```

---

## Part 12: RAG Data Versioning & Freshness

### Current Problem: No Version Control on Knowledge

The RAG data in `/rag/` has no versioning or freshness tracking:

```
rag/
├── ai-ml/patterns/api-provider-patterns.md     # When was OpenAI API last updated?
├── supply-chain/package-health/deps-dev-api.md # Is deps.dev API still current?
├── technology-identification/...               # Are these patterns current?
```

**Issues:**
1. **No source tracking** - Where did this pattern come from?
2. **No version info** - What version of the API/library does this cover?
3. **No freshness check** - Is this pattern still accurate?
4. **No update mechanism** - How do we know when to refresh?

### Recommendation: Knowledge Provenance System

Add metadata to each knowledge file:

```yaml
# knowledge/_meta.yaml (in each analyzer's knowledge/ dir)
provenance:
  sources:
    - url: "https://docs.openai.com/api-reference"
      last_checked: "2024-01-15"
      version: "v1"
    - url: "https://github.com/openai/openai-python"
      last_checked: "2024-01-15"
      version: "1.6.0"

  patterns:
    - file: "api-patterns.json"
      source: "https://docs.openai.com/api-reference"
      extracted: "2024-01-15"

freshness:
  check_interval: "30d"          # How often to verify
  last_verified: "2024-01-15"
  next_check: "2024-02-15"
  auto_update: false             # Manual review required
```

### Freshness Check Command

```bash
zero knowledge check              # Check all knowledge freshness
zero knowledge check malcontent   # Check specific analyzer
zero knowledge update             # Interactive update workflow
zero knowledge sources            # List all external sources
```

### Automated Freshness Monitoring

```go
// pkg/knowledge/freshness.go
type FreshnessChecker struct {
    analyzers []string
}

func (fc *FreshnessChecker) Check(ctx context.Context) (*FreshnessReport, error) {
    report := &FreshnessReport{}

    for _, analyzer := range fc.analyzers {
        meta := loadMeta(analyzer)

        for _, source := range meta.Provenance.Sources {
            // Check if source has been updated
            currentVersion, err := fetchCurrentVersion(source.URL)
            if err != nil {
                report.Errors = append(report.Errors, err)
                continue
            }

            if currentVersion != source.Version {
                report.Stale = append(report.Stale, StaleKnowledge{
                    Analyzer:       analyzer,
                    Source:         source.URL,
                    OurVersion:     source.Version,
                    CurrentVersion: currentVersion,
                })
            }
        }
    }

    return report, nil
}
```

### Integration with CI/CD

```yaml
# .github/workflows/knowledge-freshness.yaml
name: Knowledge Freshness Check
on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go run ./cmd/zero knowledge check --ci
      - run: |
          if [ -f stale-knowledge.json ]; then
            gh issue create --title "Stale knowledge detected" \
              --body-file stale-knowledge.json
          fi
```

---

## Part 13: Semgrep - Hybrid Approach (Keep + Remove)

### TL;DR: What Changes

| Component | Action | Rationale |
|-----------|--------|-----------|
| **Community SAST rules** | **KEEP** | Battle-tested, CVE-linked, maintained by security community |
| **`zero feeds semgrep`** | **KEEP** | Sync community rules for code-security |
| **`rag_converter.go`** | **REMOVE** | 978 lines just to use Semgrep as pattern engine |
| **`.zero/rules/generated/`** | **REMOVE** | Zero-specific rules don't need Semgrep |
| **`zero feeds rag`** | **REMOVE** | No longer needed |

### Current Architecture (The Problem)

```
RAG Markdown → rag_converter.go → Semgrep YAML → semgrep binary → Findings
    (our patterns)    (978 lines)     (our rules)    (overkill for this)
```

We're using Semgrep as a generic pattern execution engine, which is overkill. The conversion layer is complex and loses context.

### Proposed Architecture (Hybrid)

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Detection Layer                               │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────────────────────┐    ┌─────────────────────────────┐ │
│  │    Custom Go Analyzers      │    │    Semgrep (Community)      │ │
│  │         (NEW)               │    │         (KEEP)              │ │
│  │                             │    │                             │ │
│  │  • Technology detection     │    │  • SAST security rules      │ │
│  │  • Config file parsing      │    │  • CVE-specific patterns    │ │
│  │  • Secret detection (regex) │    │  • Taint tracking           │ │
│  │  • Package analysis         │    │  • Data flow analysis       │ │
│  │  • Build optimization       │    │                             │ │
│  │  • Tool config validation   │    │                             │ │
│  │                             │    │                             │ │
│  │  Source: knowledge/patterns │    │  Source: community rules    │ │
│  │  Format: JSON               │    │  Format: Semgrep YAML       │ │
│  │  Speed: Fast (native Go)    │    │  Speed: Slower (process)    │ │
│  │  Context: Full (guidance)   │    │  Context: Limited           │ │
│  └─────────────────────────────┘    └─────────────────────────────┘ │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### What We KEEP (Semgrep)

Semgrep remains valuable for security scanning with community rules:

| Use Case | Why Semgrep |
|----------|-------------|
| **SAST scanning** | Thousands of tested rules |
| **CVE detection** | Rules linked to specific CVEs |
| **Taint tracking** | Complex data flow analysis |
| **SARIF output** | Industry-standard format |

```bash
zero feeds semgrep    # Still works - syncs community rules
```

### What We REMOVE

| Component | Lines | Why Remove |
|-----------|-------|------------|
| `rag_converter.go` | 978 | Complex conversion layer |
| `.zero/rules/generated/*.yaml` | ~2000 | Generated rules we don't need |
| `zero feeds rag` | - | Command no longer needed |

### Why This Matters

**Before (complex):**
```
Technology detection: RAG → rag_converter → Semgrep YAML → semgrep → parse output
                      (markdown)  (978 lines)    (YAML)      (spawn)   (JSON)
```

**After (simple):**
```
Technology detection: JSON patterns → Go analyzer → findings
                      (versioned)     (native)      (direct)
```

### Recommendation: Hybrid Approach

```
┌─────────────────────────────────────────────────────────────┐
│                    Detection Layer                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────┐    ┌─────────────────────────────┐ │
│  │   Custom Analyzers  │    │      Semgrep (Optional)     │ │
│  │                     │    │                             │ │
│  │  • Tech detection   │    │  • Community SAST rules     │ │
│  │  • Config files     │    │  • Complex taint tracking   │ │
│  │  • Package parsing  │    │  • CVE-specific patterns    │ │
│  │  • Simple patterns  │    │                             │ │
│  │  • Secrets (regex)  │    │                             │ │
│  │                     │    │                             │ │
│  │  Uses: Native Go    │    │  Uses: semgrep binary       │ │
│  │  Speed: Fast        │    │  Speed: Slower              │ │
│  │  Knowledge: Direct  │    │  Knowledge: Limited         │ │
│  └─────────────────────┘    └─────────────────────────────┘ │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Migration Strategy

**Phase 1: Keep Semgrep for SAST**
- Keep community Semgrep rules for security scanning
- These are well-tested, CVE-linked, maintained by community

**Phase 2: Custom Analyzers for Zero-Specific Detection**
- Technology identification → Custom analyzer (no Semgrep)
- Config file detection → Custom analyzer (glob + parse)
- Secret detection → Custom analyzer (regex, faster than Semgrep)
- Package analysis → Custom analyzer (already is)

**Phase 3: Eliminate RAG → Semgrep Conversion**
- Move RAG patterns directly into analyzer knowledge/
- Delete `rag_converter.go` (978 lines)
- Patterns become JSON in `knowledge/patterns/`
- Interpretation stays as markdown in `knowledge/guidance/`

### New Pattern Format

Instead of RAG markdown → Semgrep YAML:

```json
// pkg/analyzers/technology-detection/knowledge/patterns/openai.json
{
  "name": "OpenAI",
  "version": "1.0.0",
  "source": "https://docs.openai.com",
  "last_updated": "2024-01-15",

  "detection": {
    "packages": {
      "npm": ["openai"],
      "pypi": ["openai"],
      "go": ["github.com/sashabaranov/go-openai"]
    },
    "imports": {
      "python": ["^import openai", "^from openai import"],
      "javascript": ["from ['\"]openai['\"]", "require\\(['\"]openai['\"]\\)"]
    },
    "config_files": [".openai", "openai.yaml"],
    "env_vars": ["OPENAI_API_KEY"]
  },

  "confidence": {
    "package": 95,
    "import": 90,
    "config_file": 85,
    "env_var": 70
  }
}
```

### Custom Analyzer Implementation

```go
// pkg/analyzers/technology-detection/analyzer.go
func (a *TechnologyDetector) Run(ctx context.Context, opts *Options) (*Result, error) {
    result := &Result{}

    // Load patterns from co-located knowledge
    patterns := a.loadPatterns()  // From knowledge/patterns/*.json

    for _, pattern := range patterns {
        // Check packages (fast: parse package.json, go.mod, etc.)
        if matches := a.checkPackages(opts.RepoPath, pattern); len(matches) > 0 {
            result.AddFinding(pattern.Name, "package", matches, pattern.Confidence.Package)
        }

        // Check imports (medium: grep-based)
        if matches := a.checkImports(opts.RepoPath, pattern); len(matches) > 0 {
            result.AddFinding(pattern.Name, "import", matches, pattern.Confidence.Import)
        }

        // Check config files (fast: glob)
        if matches := a.checkConfigFiles(opts.RepoPath, pattern); len(matches) > 0 {
            result.AddFinding(pattern.Name, "config", matches, pattern.Confidence.ConfigFile)
        }
    }

    return result, nil
}
```

### Benefits of Custom Approach

| Benefit | Description |
|---------|-------------|
| **No external dependency** | Pure Go, no semgrep binary needed |
| **Faster execution** | No process spawn, direct file access |
| **Knowledge preserved** | Patterns + guidance stay together |
| **Easier debugging** | Pattern → finding is direct |
| **Version tracked** | Patterns have explicit versions |
| **Agent integration** | Agents can access guidance directly |

### What to Keep in Semgrep

- `pkg/scanner/code-security/semgrep.go` - For SAST with community rules
- Community rule sync via `zero feeds semgrep`
- CVE-specific detection rules
- Complex taint tracking (if needed)

### What to Remove

- `rag_converter.go` - 978 lines of conversion code
- `.zero/rules/generated/` - Generated Semgrep rules from RAG
- `zero feeds rag` command - No longer needed

---

## Part 14: New Analyzer Categories

### Current Gap Analysis

The current 7 super scanners are missing two important dimensions:

| Missing | Current State | Impact |
|---------|---------------|--------|
| **Build Analysis** | Partially in `devops` (github_actions) | No optimization insights, no cost analysis |
| **Tool Configuration** | Not covered | No validation of ESLint, Prettier, TypeScript configs |

### New Analyzer: `build`

**Purpose:** Analyze CI/CD pipelines for optimization opportunities, cost reduction, and best practices.

**Capabilities:**

| Capability | Description | Example Findings |
|------------|-------------|------------------|
| **Cost Analysis** | GitHub Actions minutes, runner costs | "macOS runners cost 10x Linux - 847 minutes/month" |
| **Caching Optimization** | Detect missing or inefficient caches | "npm cache not configured - adds 2min to each build" |
| **Parallelization** | Find sequential jobs that could parallelize | "Tests and lint run sequentially - could save 5min" |
| **Duplicate Work** | Same dependencies installed multiple times | "node_modules installed in 3 jobs - use cache" |
| **Flaky Tests** | Detect retry patterns, intermittent failures | "test-integration has 15% retry rate" |
| **Build Time Trends** | Track build duration over time | "Average build time increased 40% this month" |
| **Runner Optimization** | Right-size runners, spot instances | "Large runner used for 30s job - use small" |
| **Matrix Optimization** | Reduce redundant matrix combinations | "Testing 12 Node versions - consider reducing" |

**Output Structure:**
```json
{
  "build": {
    "summary": {
      "estimated_monthly_cost": 145.00,
      "total_workflows": 8,
      "optimization_opportunities": 12,
      "potential_savings_percent": 35
    },
    "findings": {
      "cost": [...],
      "caching": [...],
      "parallelization": [...],
      "flaky_tests": [...]
    }
  }
}
```

**Agent Mapping:** Joey (Build & CI/CD specialist)

**MCP Tool:** `get_build_analysis`

---

### New Analyzer: `tool-config`

**Purpose:** Validate configuration files for developer tools. Detect misconfigurations, conflicts, and suggest best practices.

**Capabilities:**

| Capability | Description | Example Findings |
|------------|-------------|------------------|
| **Linter Config** | ESLint, Prettier, Biome validation | "ESLint extends deprecated config" |
| **TypeScript Config** | tsconfig.json best practices | "strict mode disabled - consider enabling" |
| **Bundler Config** | Webpack, Vite, esbuild settings | "Source maps in production build" |
| **Test Config** | Jest, Vitest, pytest settings | "No coverage threshold configured" |
| **Package Manager** | npm, yarn, pnpm settings | "No lockfile - builds not reproducible" |
| **Editor Config** | .editorconfig, VS Code settings | "Inconsistent indent settings across configs" |
| **Config Conflicts** | Conflicting settings across tools | "Prettier and ESLint have conflicting rules" |
| **Deprecated Options** | Outdated configuration options | "webpack.config uses deprecated 'node' option" |

**Output:** `tool-config.json`

**Agent Mapping:** Acid (Frontend) + Flu Shot (Backend)

**MCP Tool:** `get_tool_config`

---

### New Analyzer: `infra-config`

**Purpose:** Parse and validate infrastructure configuration files for best practices and security.

**Capabilities:**

| Capability | Description | Example Findings |
|------------|-------------|------------------|
| **Dockerfile** | Multi-stage builds, layer optimization, security | "Running as root - use non-root user" |
| **Docker Compose** | Service configuration, networking, volumes | "No healthcheck defined for database service" |
| **Kubernetes** | Manifests, resource limits, security contexts | "No resource limits - could starve cluster" |
| **Helm Charts** | Template validation, values best practices | "No default resource limits in values.yaml" |
| **Terraform** | Module structure, state management, providers | "No backend configured - state is local only" |
| **CloudFormation** | Template validation, IAM best practices | "IAM policy too permissive (*)" |
| **Nginx/Apache** | Web server configuration, TLS settings | "TLS 1.0/1.1 still enabled" |
| **Database Configs** | Connection pooling, timeouts, security | "No connection timeout configured" |

**Output:** `infra-config.json`

**Agent Mapping:** Plague (Infrastructure/DevOps)

**MCP Tool:** `get_infra_config`

---

### New Analyzer: `repo-governance`

**Purpose:** Check repository settings and policies via GitHub API. Ensure security and compliance best practices.

**Capabilities:**

| Capability | Description | Example Findings |
|------------|-------------|------------------|
| **Branch Protection** | Main/master protection rules | "main branch has no protection rules" |
| **Required Reviews** | PR approval requirements | "No required reviewers configured" |
| **Status Checks** | Required CI checks before merge | "CI checks not required for merge" |
| **Signed Commits** | Commit signature requirements | "Signed commits not required" |
| **Force Push** | Prevent force push to protected branches | "Force push allowed on main branch" |
| **Delete Protection** | Prevent branch deletion | "Branch deletion not restricted" |
| **CODEOWNERS** | Code ownership enforcement | "CODEOWNERS file missing or incomplete" |
| **Dependabot** | Security update configuration | "Dependabot not enabled for security updates" |
| **Secret Scanning** | GitHub secret scanning status | "Secret scanning not enabled" |
| **Vulnerability Alerts** | Dependency vulnerability alerts | "Vulnerability alerts disabled" |
| **Actions Permissions** | Workflow permissions and OIDC | "Actions can write to repo without approval" |
| **Deploy Keys** | SSH key management | "Deploy key has write access - should be read-only" |

**Data Source:** GitHub API (requires repo admin access for some checks)

**Output:** `repo-governance.json`

**Agent Mapping:** Gibson (Engineering Leader) + Blade (Compliance)

**MCP Tool:** `get_repo_governance`

---

### Where These Fit in Architecture

**Option A: Separate Analyzers (Recommended)**

```
pkg/analyzers/
├── build/                    # NEW
│   ├── analyzer.go
│   ├── knowledge/
│   │   ├── patterns/
│   │   │   ├── github-actions.json
│   │   │   ├── caching-patterns.json
│   │   │   └── cost-factors.json
│   │   └── guidance/
│   │       ├── optimization-strategies.md
│   │       └── cost-reduction.md
│   └── prompts/
│       └── optimize-build.md
│
├── tool-config/              # NEW
│   ├── analyzer.go
│   ├── knowledge/
│   │   ├── patterns/
│   │   │   ├── eslint-rules.json
│   │   │   ├── typescript-options.json
│   │   │   ├── bundler-settings.json
│   │   │   └── deprecated-options.json
│   │   └── guidance/
│   │       ├── linter-best-practices.md
│   │       └── typescript-strict-mode.md
│   └── prompts/
│       └── fix-config.md
```

**Option B: Fold into Existing**

| New Capability | Fold Into | Rationale |
|----------------|-----------|-----------|
| Build analysis | `devops` | Already has github_actions |
| Tool config | `developer-experience` | Related to DX |

**Recommendation:** Option A (separate analyzers) because:
- Build optimization is distinct from infrastructure security
- Tool configuration is cross-cutting (frontend + backend)
- Clearer agent mapping
- Better knowledge co-location

---

### Updated Analyzer Count

| Category | Current | Proposed |
|----------|---------|----------|
| Super analyzers | 7 | - |
| Individual analyzers | ~45 features | ~30 analyzers |
| **New analyzers** | - | **+4 (build, tool-config, infra-config, repo-governance)** |

### New Analyzer Summary

| Analyzer | Purpose | Agent | MCP Tool |
|----------|---------|-------|----------|
| `build` | CI/CD optimization, cost, caching | Joey | `get_build_analysis` |
| `tool-config` | Linter, TS, bundler configs | Acid, Flu Shot | `get_tool_config` |
| `infra-config` | Docker, K8s, Terraform configs | Plague | `get_infra_config` |
| `repo-governance` | GitHub settings, branch protection | Gibson, Blade | `get_repo_governance` |

---

## Part 15: Finding Validation & False Positive Reduction

### The Problem

Analyzers (especially technology identification) generate false positives:
- Pattern matches that aren't real usage
- Outdated detections (library was removed)
- Misidentified technologies
- Context-blind matches (test files, comments, examples)

False positives erode trust. Users stop paying attention to findings.

### Strategy 1: Multi-Signal Correlation

**Concept:** Require multiple independent signals before reporting a finding with high confidence.

```
Single signal (low confidence):
  - Import pattern matches "openai" → 70% confidence

Multiple signals (high confidence):
  - Import pattern matches "openai" → +30%
  - Package in package.json → +40%
  - Config file exists (.openai) → +20%
  - Environment variable referenced → +10%
  = 100% confidence (capped)
```

**Implementation:**
```json
// Finding with correlation
{
  "technology": "OpenAI",
  "confidence": 95,
  "signals": [
    {"type": "package", "source": "package.json", "weight": 40},
    {"type": "import", "source": "src/ai.ts:5", "weight": 30},
    {"type": "env_var", "source": ".env.example", "weight": 15}
  ],
  "correlation_score": 85
}
```

### Strategy 2: Context-Aware Filtering

**Concept:** Exclude matches in contexts that are likely false positives.

| Context | Action | Example |
|---------|--------|---------|
| Test files | Lower confidence or exclude | `__tests__/`, `*.test.js` |
| Comments | Exclude | `// TODO: add openai` |
| Documentation | Exclude | `README.md`, `docs/` |
| Examples | Lower confidence | `examples/`, `samples/` |
| Vendor/deps | Exclude | `node_modules/`, `vendor/` |
| Lock files | Exclude | `package-lock.json` |
| Generated code | Exclude | Files with `@generated` |

**Implementation:**
```go
type ContextFilter struct {
    ExcludePaths    []string  // node_modules, vendor, etc.
    ExcludePatterns []string  // *_test.go, *.test.js
    LowerConfidence []string  // examples/, docs/
    ExcludeInline   []string  // Comments, strings
}
```

### Strategy 3: User Feedback Loop

**Concept:** Allow users to mark findings as false positives, learn from feedback.

**Feedback Storage:**
```json
// .zero/feedback.json
{
  "false_positives": [
    {
      "analyzer": "technology-detection",
      "finding_id": "openai-import-src/legacy.ts:10",
      "reason": "commented_out",
      "timestamp": "2024-01-15T10:30:00Z",
      "user": "developer@example.com"
    }
  ],
  "confirmed": [
    {
      "analyzer": "technology-detection",
      "finding_id": "react-package",
      "timestamp": "2024-01-15T10:31:00Z"
    }
  ]
}
```

**CLI Commands:**
```bash
zero feedback false-positive <finding-id>   # Mark as false positive
zero feedback confirm <finding-id>          # Confirm finding is valid
zero feedback list                          # Show feedback history
zero feedback export                        # Export for pattern improvement
```

**Learning from Feedback:**
- Suppress same finding in future scans
- Aggregate feedback across users (opt-in)
- Identify patterns that generate most false positives
- Auto-adjust confidence based on feedback rate

### Strategy 4: LLM-Assisted Validation

**Concept:** Use Claude to validate uncertain findings by examining context.

**When to Use:**
- Low confidence findings (< 70%)
- Conflicting signals
- User requests validation
- High-value findings (critical severity)

**Implementation:**
```go
type LLMValidator struct {
    client *anthropic.Client
}

func (v *LLMValidator) Validate(finding Finding, context CodeContext) (*Validation, error) {
    prompt := fmt.Sprintf(`
        Analyze this potential technology detection:

        Finding: %s detected at %s
        Code context:
        %s

        Is this a genuine usage of %s or a false positive?
        Consider: Is it in a comment? Test file? Actually used?

        Return: CONFIRMED, FALSE_POSITIVE, or UNCERTAIN with explanation.
    `, finding.Technology, finding.Location, context.Code, finding.Technology)

    response, err := v.client.Message(prompt)
    // Parse and return validation result
}
```

**Cost Control:**
- Only validate uncertain findings
- Batch validation requests
- Cache validation results
- User opt-in for LLM validation

### Strategy 5: Test Corpus Validation

**Concept:** Maintain a corpus of repositories with known expected results.

**Test Corpus Structure:**
```
test-corpus/
├── react-app/                    # Known: React, TypeScript, Jest
│   ├── expected.json             # Expected findings
│   └── repo/                     # Actual code
├── python-ml/                    # Known: PyTorch, FastAPI, OpenAI
│   ├── expected.json
│   └── repo/
├── false-positive-cases/         # Known false positive patterns
│   ├── commented-imports/
│   ├── test-only-deps/
│   └── docs-examples/
```

**CI Integration:**
```yaml
# .github/workflows/pattern-validation.yaml
name: Pattern Validation
on: [push]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go test ./pkg/analyzers/... -corpus=./test-corpus
      - run: |
          # Check precision/recall metrics
          zero validate-patterns --corpus=./test-corpus --min-precision=0.95
```

**Metrics Tracked:**
- Precision: True positives / (True positives + False positives)
- Recall: True positives / (True positives + False negatives)
- F1 Score: Balance of precision and recall

### Strategy 6: Confidence Decay

**Concept:** Reduce confidence for findings that haven't been re-validated.

```
Initial detection: 90% confidence
After 30 days without code change: 85%
After 90 days: 75%
After 180 days: Mark as "stale - needs revalidation"
```

**Use Case:** Technology detected 6 months ago but package was removed.

### Strategy 7: Community Pattern Validation

**Concept:** Crowdsource pattern effectiveness data (opt-in).

**Anonymous Telemetry:**
```json
{
  "pattern_id": "zero.openai.import.python",
  "executions": 1000,
  "false_positive_rate": 0.05,
  "common_fp_contexts": ["test_files", "comments"]
}
```

**Benefits:**
- Identify problematic patterns across many repos
- Prioritize pattern improvements
- Share learnings without sharing code

### Strategy 8: Hierarchical Validation

**Concept:** Different validation levels for different use cases.

| Level | Validation | Use Case |
|-------|------------|----------|
| **Quick** | Pattern match only | CI/CD, fast feedback |
| **Standard** | Multi-signal correlation | Default analysis |
| **Thorough** | + Context filtering + LLM spot-check | Compliance, audits |
| **Verified** | + User confirmation | High-stakes decisions |

```bash
zero analyze repo --validation=quick      # Fast, more FPs
zero analyze repo --validation=standard   # Default
zero analyze repo --validation=thorough   # Slower, fewer FPs
```

### Implementation Recommendation

**Phase 1: Foundation**
- [ ] Add confidence scoring to all findings
- [ ] Implement context-aware filtering (exclude test files, comments)
- [ ] Add multi-signal correlation for technology detection

**Phase 2: Feedback**
- [ ] Add `zero feedback` CLI commands
- [ ] Store feedback in `.zero/feedback.json`
- [ ] Suppress known false positives in future scans

**Phase 3: Validation**
- [ ] Create test corpus with expected results
- [ ] Add precision/recall metrics to CI
- [ ] Track pattern effectiveness over time

**Phase 4: Intelligence**
- [ ] Optional LLM validation for uncertain findings
- [ ] Community pattern telemetry (opt-in)
- [ ] Confidence decay for stale findings

### Finding Schema (Updated)

```json
{
  "id": "tech-openai-src/ai.ts:5",
  "analyzer": "technology-detection",
  "type": "technology",
  "value": "OpenAI",

  "confidence": {
    "score": 85,
    "level": "high",
    "signals": [
      {"type": "package", "weight": 40, "source": "package.json"},
      {"type": "import", "weight": 30, "source": "src/ai.ts:5"},
      {"type": "env_var", "weight": 15, "source": ".env.example"}
    ]
  },

  "validation": {
    "status": "unvalidated",
    "llm_check": null,
    "user_feedback": null,
    "last_validated": null
  },

  "context": {
    "file_type": "source",
    "in_test": false,
    "in_comment": false,
    "in_docs": false
  },

  "location": {
    "file": "src/ai.ts",
    "line": 5,
    "snippet": "import OpenAI from 'openai'"
  }
}
```

---

## Updated Implementation Roadmap

### Phase 1: Non-Breaking (Quick Wins)
- [ ] Update agent prompts for balanced messaging (engineering intelligence)
- [ ] Add missing MCP tools (devops, quality, ownership, dora, devx)
- [ ] Add CLI command aliases (`cache`, `analyze`, `onboard`)
- [ ] Update CLAUDE.md and README positioning
- [ ] Add confidence scoring to all findings
- [ ] Implement context-aware filtering (exclude test files, comments)

### Phase 2: Cache Architecture
- [ ] Implement hybrid cache with TTLs
- [ ] Add `zero cache` commands (generate, status, invalidate, gc)
- [ ] Lazy analyzer execution in MCP tools
- [ ] Cache invalidation on git changes

### Phase 3: Finding Validation System
- [ ] Add multi-signal correlation for technology detection
- [ ] Add `zero feedback` CLI commands
- [ ] Store feedback in `.zero/feedback.json`
- [ ] Suppress known false positives in future scans
- [ ] Create test corpus with expected results
- [ ] Add precision/recall metrics to CI

### Phase 4: Flatten Analyzers + Knowledge Co-location
- [ ] Create `pkg/analyzers/` structure
- [ ] Migrate first analyzer (technology-detection) as proof of concept
- [ ] Move RAG patterns to `knowledge/patterns/` (JSON format)
- [ ] Move agent guidance to `knowledge/guidance/`
- [ ] Add `_meta.yaml` provenance tracking
- [ ] Update agent knowledge loading

### Phase 5: New Analyzers
- [ ] Implement `build` analyzer (CI/CD optimization, cost)
- [ ] Implement `tool-config` analyzer (linter, TS, bundler configs)
- [ ] Implement `infra-config` analyzer (Docker, K8s, Terraform)
- [ ] Implement `repo-governance` analyzer (GitHub settings)
- [ ] Add corresponding MCP tools
- [ ] Add corresponding agent mappings

### Phase 6: Eliminate RAG → Semgrep Pipeline
- [ ] Convert RAG markdown patterns to JSON format
- [ ] Delete `rag_converter.go` (978 lines)
- [ ] Remove `.zero/rules/generated/` directory
- [ ] Keep Semgrep only for community SAST rules
- [ ] Update `zero feeds` commands

### Phase 7: Full Migration
- [ ] Migrate remaining analyzers from super scanners
- [ ] Remove old `pkg/scanner/` structure
- [ ] Update all MCP tools
- [ ] Add `zero knowledge check` command
- [ ] Migration guide for config files
- [ ] Terminology rename (scanner→analyzer throughout)

### Phase 8: Web UI & Polish
- [ ] Update report navigation (10 dimensions)
- [ ] Real-time cache status display
- [ ] Per-analyzer freshness indicators
- [ ] Knowledge freshness dashboard
- [ ] Finding validation UI (confirm/reject)
- [ ] Optional LLM validation integration

---

## Summary (Updated)

| Priority | Change | Impact |
|----------|--------|--------|
| 1 | Hybrid cache architecture | Better UX, lazy evaluation, no mandatory hydrate |
| 2 | Flatten analyzers | Discoverability, ~34 individual analyzers |
| 3 | Co-locate knowledge | Code + patterns + guidance + prompts together |
| 4 | Knowledge provenance | Track sources, versions, freshness |
| 5 | Semgrep hybrid approach | Keep for SAST, remove RAG→Semgrep pipeline |
| 6 | Custom analyzers for detection | Faster, no external deps, better agent integration |
| 7 | **New: `build` analyzer** | CI/CD optimization, cost analysis, caching |
| 8 | **New: `tool-config` analyzer** | Linter, TypeScript, bundler config validation |
| 9 | **New: `infra-config` analyzer** | Docker, K8s, Terraform, Helm config validation |
| 10 | **New: `repo-governance` analyzer** | Branch protection, required reviews, security features |
| 11 | **Finding validation system** | Multi-signal correlation, feedback loops, LLM validation |
| 12 | Terminology (scanner→analyzer, hydrate→cache) | Clarity |
| 13 | Rebalance messaging | 10 dimensions of engineering intelligence |

---

## Next Steps

1. Review this document and provide feedback on open questions
2. Create GitHub issues for each phase
3. Prioritize Phase 1 (non-breaking) changes
4. Plan Phase 3 (breaking changes) for a major version release
