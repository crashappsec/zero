# Persona Overlays

Overlays add agent-specific customizations to universal personas without duplicating the base persona definition.

## How Overlays Work

When an agent is invoked with a persona:

1. Load the universal persona from `rag/personas/[persona].md`
2. Check for overlay at `rag/personas/overlays/[agent]/[persona]-overlay.md`
3. If overlay exists, merge it with the base persona
4. Apply combined persona to agent output

## Overlay Structure

Overlays should ONLY contain:

1. **Additional Knowledge Sources** - Agent-specific knowledge files to reference
2. **Domain-Specific Examples** - Examples tailored to the agent's domain
3. **Specialized Prioritization** - Domain-specific prioritization rules

Overlays should NOT contain:
- Complete persona definitions (use base persona)
- Duplicated content from base persona
- Output templates (use base persona templates)

## Example Overlay

```markdown
# Security Engineer Overlay for Cereal (Supply Chain)

## Additional Knowledge Sources

### Vulnerability Assessment
- `knowledge/patterns/ecosystems/*.json` - Ecosystem-specific patterns
- `knowledge/guidance/vulnerability-scoring.md` - CVSS/EPSS interpretation

### Supply Chain Context
- `knowledge/patterns/health/abandonment-signals.json` - Package health
- `knowledge/patterns/health/typosquat-patterns.json` - Typosquatting detection

## Domain-Specific Examples

When reporting supply chain vulnerabilities:
- Include CISA KEV status for CVEs
- Add EPSS scores for exploitability context
- Reference affected package versions
- Note transitive vs direct dependency

## Specialized Prioritization

For supply chain findings, also consider:
1. **CISA KEV listed** - Immediate action
2. **High EPSS (>0.5)** - Within 24 hours
3. **Direct dependency** - Higher priority than transitive
```

## Directory Structure

```
overlays/
├── README.md           # This file
├── cereal/             # Supply chain agent overlays
│   ├── security-engineer-overlay.md
│   └── auditor-overlay.md
├── razor/              # Code security agent overlays
│   └── security-engineer-overlay.md
└── blade/              # Compliance agent overlays
    └── auditor-overlay.md
```

## Creating an Overlay

1. Identify agent-specific knowledge that enhances the persona
2. Create overlay file: `overlays/[agent]/[persona]-overlay.md`
3. Include ONLY additive content (knowledge refs, examples, prioritization)
4. Test that overlay merges correctly with base persona

## When to Create Overlays

Create an overlay when:
- Agent has domain-specific knowledge files to reference
- Domain requires specialized examples in output
- Prioritization rules differ for the domain

Do NOT create an overlay when:
- Base persona already covers the use case
- You would duplicate base persona content
- The customization is minor
