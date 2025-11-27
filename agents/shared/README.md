# Shared Knowledge

Cross-agent knowledge that ensures consistency across all agents.

## Contents

| File | Purpose |
|------|---------|
| `severity-levels.json` | Universal severity definitions (Critical, High, Medium, Low, Info) |
| `confidence-levels.json` | Confidence scoring for findings |
| `output-formatting.md` | Output format conventions |

## Usage

Agents reference shared knowledge using relative paths:

```markdown
### Shared
- `../shared/severity-levels.json` - Severity definitions
- `../shared/confidence-levels.json` - Confidence scoring
```

## Severity Levels

Standardized severity classification:

| Level | CVSS Range | Response Time |
|-------|------------|---------------|
| Critical | 9.0-10.0 | Immediate (hours) |
| High | 7.0-8.9 | 1-7 days |
| Medium | 4.0-6.9 | 30 days |
| Low | 0.1-3.9 | 90 days |
| Info | 0.0 | No deadline |

## Confidence Levels

Finding confidence classification:

| Level | Certainty | Evidence Required |
|-------|-----------|-------------------|
| Confirmed | 95-100% | Direct verification |
| High | 80-94% | Strong pattern match |
| Medium | 50-79% | Reasonable evidence |
| Low | 20-49% | Limited evidence |
| Speculative | 0-19% | Minimal evidence |

## Extending Shared Knowledge

When adding new shared definitions:

1. Ensure it's truly cross-agent (used by 2+ agents)
2. Follow existing JSON schema conventions
3. Update this README
4. Update agents that should use it
