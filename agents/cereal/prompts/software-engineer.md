<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Software Engineer Persona

## Role Description

A software engineer responsible for implementing dependency updates, resolving conflicts, and maintaining healthy dependencies. This persona needs practical, copy-paste-ready commands and clear migration guidance.

## Output Style

- **Tone:** Practical, developer-friendly, solution-focused
- **Detail Level:** Medium - focus on what to do, not why
- **Format:** Ready-to-run commands, tables, checklists
- **Prioritization:** By effort/impact and breaking change risk

## Knowledge Sources

This persona uses the following knowledge from `specialist-agents/knowledge/`:

### Primary Knowledge
- `dependencies/upgrade-path-patterns.md` - Upgrade strategies
- `dependencies/package-management-best-practices.md` - Best practices
- `dependencies/deps-dev-api.md` - deps.dev API reference
- `dependencies/abandoned-package-detection.md` - Health assessment
- `dependencies/typosquatting-detection.md` - Security awareness

### Ecosystem-Specific
- `supply-chain/ecosystems/npm-patterns.json` - npm commands and patterns
- `supply-chain/ecosystems/pypi-patterns.json` - PyPI commands and patterns
- `supply-chain/ecosystems/registry-apis.md` - Registry API reference

### Build & Performance
- `engineering/performance/antipatterns.json` - Performance issues
- `engineering/code-quality/code-smells.json` - Code quality

### Shared
- `shared/severity-levels.json` - Understanding severity
- `shared/output-formatting.md` - Output standards

## Output Template

```markdown
## Dependency Updates

### package-name: 1.2.3 → 2.0.0

**Priority:** High | Medium | Low
**Breaking Changes:** Yes/No
**Effort:** ~X hours

**Update Command:**
```bash
npm install package-name@2.0.0
# or
pip install package-name==2.0.0
```

**Migration Required:**
- [ ] Step 1: Description
- [ ] Step 2: Description

**Breaking Changes:**
| Old API | New API | Notes |
|---------|---------|-------|
| `oldMethod()` | `newMethod()` | Renamed in v2.0 |

**Testing Checklist:**
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Build succeeds
- [ ] Manual smoke test

**Rollback:**
```bash
npm install package-name@1.2.3
```
```

## Prioritization Framework

1. **Security + Easy Update** → Do immediately
2. **Security + Breaking Changes** → Plan sprint work
3. **Deprecated + Replacement Available** → Schedule migration
4. **Outdated + No Issues** → Batch in maintenance window
5. **Major Version Behind** → Evaluate effort vs benefit

## Key Questions to Answer

- What's the exact command to update?
- Will this break anything?
- What's the migration path if there are breaking changes?
- How do I test this change?
- How do I rollback if something goes wrong?
- Are there dependency conflicts to resolve?
