# Universal Personas

This directory contains persona definitions and reasoning frameworks that are shared across ALL scanners (supply-chain, code-security, legal, certificate, etc.).

## Directory Structure

```
rag/personas/
├── README.md                       # This file
├── definitions/                    # Persona profiles
│   ├── security-engineer.md        # Security-focused analysis
│   ├── software-engineer.md        # Developer-focused analysis
│   ├── engineering-leader.md       # Executive/strategic analysis
│   └── auditor.md                  # Compliance-focused analysis
├── reasoning/                      # Analysis frameworks
│   └── analysis-framework.md       # Chain of reasoning methodology
└── output-formats/                 # Output templates (optional)
    └── (persona-specific output templates)
```

## Available Personas

| Persona | Description | Primary Focus |
|---------|-------------|---------------|
| `security-engineer` | Technical security professional | CVEs, CVSS, remediation, attack surface |
| `software-engineer` | Developer/engineer | Commands, versions, breaking changes, effort |
| `engineering-leader` | Engineering manager/director | Metrics, trends, resources, strategic decisions |
| `auditor` | Compliance/audit professional | Controls, frameworks, evidence, findings |

## How Personas Work

### Chain of Reasoning Framework

Each persona report is generated using a 3-phase reasoning process:

1. **Phase 1: Understand Your Audience** - Load the persona definition to understand who the report is for, what they care about, and how they communicate.

2. **Phase 2: Apply Domain Knowledge** - Load scanner-specific RAG content (e.g., supply-chain security best practices) and apply it through the persona's lens.

3. **Phase 3: Generate Output** - Transform scan data into a report that serves the persona's specific needs, using appropriate filtering, formatting, and framing.

### Persona Definitions

Each persona definition in `definitions/` includes:

- **Identity** - Who this person is
- **Profile** - Role, responsibilities, reporting structure
- **What They Care About** - High/medium/low priority items
- **Language Style** - Terminology and communication preferences
- **Decision Context** - How they'll use the report
- **What Success Looks Like** - Desired outcomes

## Usage

### In Shell Scripts

```bash
source "$REPO_ROOT/lib/universal-persona-loader.sh"

# Check if persona is valid
if is_valid_universal_persona "$persona"; then
    # Build a complete persona prompt
    prompt=$(build_persona_prompt "$persona" "$rag_content" "$scan_data" "Scanner Name")
fi
```

### Default Behavior

When no persona is specified and Claude AI is enabled:
- **Default**: Generates reports for ALL personas (4 reports)
- **Single persona**: Use `--persona security-engineer` for a single report

## Adding New Personas

1. Create `definitions/<persona-name>.md` following the existing format
2. Add the persona name to `UNIVERSAL_PERSONAS` array in `lib/universal-persona-loader.sh`
3. Optionally create `output-formats/<persona-name>.md` for custom output templates

## Scanner-Specific Domain Knowledge

While personas are universal, each scanner can have domain-specific RAG content that gets applied through the persona's lens:

- `rag/supply-chain/` - Vulnerability, SBOM, SLSA knowledge
- `rag/code-security/` - Static analysis, security patterns
- `rag/legal-review/` - Licensing, compliance, secrets

These are combined with the persona definition during report generation.
