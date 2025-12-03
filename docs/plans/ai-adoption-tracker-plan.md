# AI Adoption Tracker - Implementation Plan

## Overview

Create a new **report type** that combines data from existing scanners (**tech-discovery** + **code-ownership**) to help organizations understand AI adoption across their codebase.

**Key Principle**: No new scanner needed - leverage existing data, create new report correlation.

## Business Value

Organizations need visibility into:
- **What AI** is being adopted (APIs, frameworks, coding assistants)
- **Where** AI is being used (which parts of the codebase)
- **Who** is driving adoption (teams/individuals)
- **Trends** over time (is AI adoption accelerating?)

This enables:
- Governance and policy enforcement
- Security review prioritization
- Training needs identification
- Budget forecasting for AI tools
- Risk assessment for AI dependencies

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      EXISTING SCANNERS                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  tech-discovery.sh          code-ownership.sh / bus-factor.sh  │
│  ─────────────────          ─────────────────────────────────  │
│  • Detects technologies     • Git blame analysis               │
│  • Categories (ai-ml, etc)  • CODEOWNERS parsing               │
│  • Package detection        • Contributor stats                │
│  • Config file detection    • Bus factor calculation           │
│                                                                 │
│         ↓                              ↓                        │
│  tech-discovery.json            code-ownership.json            │
│                                 bus-factor.json                │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                    NEW: AI ADOPTION REPORT TYPE                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  utils/phantom/lib/report-types/ai-adoption.sh                 │
│  ─────────────────────────────────────────────                 │
│  • Filter tech-discovery for AI categories only                │
│  • Correlate AI tech files with ownership data                 │
│  • Aggregate by adopter (who uses what AI)                     │
│  • Generate adoption metrics & visualizations                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────────┐
│                        OUTPUT FORMATS                           │
├─────────────────────────────────────────────────────────────────┤
│  Terminal  │  Markdown  │  HTML  │  CSV  │  JSON               │
└─────────────────────────────────────────────────────────────────┘
```

## Data Sources (Existing)

### 1. tech-discovery.json

From `utils/scanners/tech-discovery/tech-discovery.sh`:

```json
{
  "technologies": [
    {
      "name": "OpenAI",
      "category": "ai-ml/llm-apis",
      "confidence": 95,
      "detection_method": "sbom-package",
      "evidence": ["package dependency: openai@4.20.0"]
    }
  ]
}
```

**AI Categories in RAG** (`rag/technology-identification/`):
- `ai-ml/apis/` - Anthropic, OpenAI, Google AI, Cohere, Mistral, etc.
- `ai-ml/frameworks/` - LangChain, LlamaIndex
- `ai-ml/vectordb/` - Pinecone, Weaviate, Qdrant, ChromaDB
- `ai-ml/mlops/` - Hugging Face, Weights & Biases
- `genai-tools/` - Copilot, Cursor, Codeium, Tabnine, etc.

### 2. code-ownership.json / bus-factor.json

From `utils/scanners/code-ownership/`:

```json
{
  "contributors": [
    {
      "name": "alice@company.com",
      "commits": 145,
      "lines_added": 12500,
      "percentage": 35.2
    }
  ],
  "codeowners": {
    "exists": true,
    "patterns": [...]
  }
}
```

## Implementation

### Step 1: Implementation Options

**Current State:**
- `technology.json` has detected technologies but evidence is text ("Dockerfile found"), not file paths
- `ownership.json` has contributors at project level, not per-file
- `patterns.md` files have import patterns (e.g., `import.*from ['"]openai['"]`) that could locate files

**Option A: Minimal MVP (Phase 1)**
- Filter existing tech-discovery for AI categories
- Show AI technologies + top contributors in separate sections
- No direct "who uses which AI" correlation
- Fast to implement, still valuable

**Option B: Enhanced with File-Level Scanning (Phase 2)**
- Add import/env scanning to tech-discovery using patterns.md
- Output includes file paths per technology
- Correlate files with git blame for ownership
- Full "alice uses openai in src/ai/chat.ts" correlation

**Recommended approach**: Start with Option A, then enhance to Option B

### Step 2: Create AI category filter

Define which categories count as "AI" in `rag/ai-adoption/ai-categories.json`:

```json
{
  "ai_categories": [
    "ai-ml/*",
    "genai-tools/*"
  ],
  "category_labels": {
    "ai-ml/apis": "LLM APIs",
    "ai-ml/frameworks": "AI Frameworks",
    "ai-ml/vectordb": "Vector Databases",
    "ai-ml/mlops": "MLOps Tools",
    "genai-tools": "AI Coding Assistants"
  }
}
```

### Step 3: Create report type

`utils/phantom/lib/report-types/ai-adoption.sh`:

```bash
generate_report_data() {
    local project_id="$1"
    local analysis_path="$2"

    # Load existing scanner data
    local tech_data=$(load_scanner_data "$analysis_path" "tech-discovery")
    local ownership_data=$(load_scanner_data "$analysis_path" "code-ownership")

    # Filter for AI technologies only
    local ai_techs=$(filter_ai_technologies "$tech_data")

    # Correlate with ownership
    local adoption_by_tech=$(correlate_ownership "$ai_techs" "$ownership_data")
    local adoption_by_person=$(aggregate_by_adopter "$adoption_by_tech")

    # Output combined data
    jq -n \
        --argjson techs "$ai_techs" \
        --argjson by_tech "$adoption_by_tech" \
        --argjson by_person "$adoption_by_person" \
        '{...}'
}
```

### Step 4: Add report format renderers

In `utils/phantom/lib/report-formats/`:
- `terminal.sh` - Add `render_ai_adoption()` function
- `markdown.sh` - Add markdown rendering
- `html.sh` - Add HTML rendering
- `csv.sh` - Add CSV export

### Step 5: Register report type

Update `utils/phantom/lib/report-common.sh`:

```bash
REPORT_TYPES=("summary" "security" "licenses" "compliance" "sbom" "supply-chain" "dora" "code-ownership" "ai-adoption" "full")
```

## Output Design

### Terminal Report

```
AI ADOPTION REPORT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

SUMMARY
  AI Technologies     5 detected
  AI Adopters         8 contributors
  Categories          4 (LLM APIs, Frameworks, Vector DBs, Coding Tools)

TECHNOLOGIES BY CATEGORY
──────────────────────────────────────────────────────────────────
  LLM APIs            ████████████████████  3 techs
                      OpenAI, Anthropic, Google AI

  AI Frameworks       ████████████          2 techs
                      LangChain, LlamaIndex

  Vector Databases    ██████                1 tech
                      Pinecone

  Coding Assistants   ██████                1 tech
                      GitHub Copilot

TOP AI ADOPTERS
──────────────────────────────────────────────────────────────────
  Rank  Contributor              Commits  Technologies
  ─────────────────────────────────────────────────────────────────
  1     alice@company.com        45       openai, langchain, pinecone
  2     bob@company.com          32       anthropic, openai
  3     carol@company.com        18       github-copilot

AI TECHNOLOGY DETAILS
──────────────────────────────────────────────────────────────────
  OpenAI (ai-ml/llm-apis)
    Confidence: 95%  │  Detection: sbom-package
    Evidence: openai@4.20.0 in package.json

  Anthropic (ai-ml/llm-apis)
    Confidence: 90%  │  Detection: env-variable
    Evidence: ANTHROPIC_API_KEY in .env.example

──────────────────────────────────────────────────────────────────
Generated by Phantom Report v1.0.0
```

### Markdown Report

```markdown
# AI Adoption Report

## Summary

| Metric | Value |
|--------|-------|
| AI Technologies | 5 |
| AI Adopters | 8 |
| Categories | 4 |

## Technologies by Category

### LLM APIs (3)
- OpenAI
- Anthropic
- Google AI

### AI Frameworks (2)
- LangChain
- LlamaIndex

...

## Top Adopters

| Rank | Contributor | Commits | Technologies |
|------|-------------|---------|--------------|
| 1 | alice@company.com | 45 | openai, langchain, pinecone |
| 2 | bob@company.com | 32 | anthropic, openai |
```

## File Structure

```
utils/phantom/lib/report-types/
└── ai-adoption.sh                    # NEW: Report type logic

utils/phantom/lib/report-formats/
├── terminal.sh                       # UPDATE: Add render_ai_adoption()
├── markdown.sh                       # UPDATE: Add markdown rendering
├── html.sh                           # UPDATE: Add HTML rendering
└── csv.sh                            # UPDATE: Add CSV export

rag/ai-adoption/
├── ai-categories.json                # NEW: Category definitions
└── README.md                         # NEW: Documentation
```

## Implementation Steps

1. **Create `rag/ai-adoption/ai-categories.json`**
   - Define which tech categories are "AI"
   - Map category paths to friendly labels

2. **Create `utils/phantom/lib/report-types/ai-adoption.sh`**
   - `generate_report_data()` function
   - Filter tech-discovery for AI categories
   - Correlate with ownership data
   - Aggregate metrics

3. **Update report format renderers**
   - Add `render_ai_adoption()` to each format
   - Terminal, Markdown, HTML, CSV support

4. **Register new report type**
   - Add to REPORT_TYPES in report-common.sh
   - Update report.sh to handle ai-adoption type

5. **Test with phantom**
   - Run on test repository with AI dependencies
   - Verify all formats render correctly

## Future Enhancements

1. **File-level correlation**: Track exactly which files use each AI tech and who owns them
2. **Timeline analysis**: Show when AI technologies were first introduced
3. **Policy enforcement**: Flag unapproved AI vendors
4. **Cost estimation**: Estimate API costs based on usage patterns
5. **Trend reporting**: Track adoption velocity over time

## Success Criteria

- Report correctly identifies all AI technologies from tech-discovery data
- Correlates AI usage with correct code owners
- Renders cleanly in all output formats (terminal, markdown, html, csv)
- Integrates with `phantom report --type ai-adoption`
