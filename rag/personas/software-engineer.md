# Software Engineer Persona

## Role Description

A software engineer responsible for implementing fixes, resolving issues, and maintaining healthy systems. This persona needs practical, copy-paste-ready commands and clear step-by-step guidance.

**What they care about:**
- Exact commands to run
- Breaking changes and migration steps
- Testing checklists
- Rollback procedures

## Output Style

- **Tone:** Practical, developer-friendly, solution-focused
- **Detail Level:** Medium - focus on what to do, not deep analysis of why
- **Format:** Ready-to-run commands, tables, checklists
- **Prioritization:** By effort/impact and breaking change risk

## Output Template

```markdown
## Remediation Guide

### Summary

**Total Items:** X
**Estimated Effort:** Y hours
**Breaking Changes:** Yes/No

### Quick Wins (< 1 hour)

| Item | Command/Action | Risk |
|------|----------------|------|
| [Item 1] | `command here` | Low |
| [Item 2] | `command here` | Low |

### Detailed Remediation

#### Item: [Name]

**Priority:** High/Medium/Low
**Effort:** ~X hours
**Breaking Changes:** Yes/No

**Fix Command:**
```bash
[exact command to run]
```

**Steps:**
1. [ ] Step 1 description
2. [ ] Step 2 description
3. [ ] Step 3 description

**Breaking Changes:**
| Before | After | Migration |
|--------|-------|-----------|
| `oldApi()` | `newApi()` | Update all call sites |

**Testing Checklist:**
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Build succeeds
- [ ] Manual smoke test completed

**Rollback:**
```bash
[command to revert if needed]
```

---
```

## Prioritization Framework

1. **Security Fix + Easy Update** - Do immediately
2. **Security Fix + Breaking Changes** - Plan sprint work
3. **Deprecated + Replacement Available** - Schedule migration
4. **Outdated + No Issues** - Batch in maintenance window
5. **Major Version Behind** - Evaluate effort vs benefit

## Key Questions to Answer

- What's the exact command to fix this?
- Will this break anything?
- What's the migration path if there are breaking changes?
- How do I test this change?
- How do I rollback if something goes wrong?
- Are there dependency conflicts to resolve?
