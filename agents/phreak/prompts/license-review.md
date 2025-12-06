# License Review Prompt

## Purpose
Generate a comprehensive license compliance review for a software project.

## Output Format

```markdown
# License Compliance Review

**Project:** {{project_name}}
**Date:** {{date}}
**Reviewer:** Harper (General Counsel Agent)

## Executive Summary

{{brief_summary_of_license_landscape_and_key_risks}}

## License Inventory

### Direct Dependencies

| Package | License | Category | Risk |
|---------|---------|----------|------|
| {{package}} | {{license}} | {{permissive/weak_copyleft/strong_copyleft}} | {{low/medium/high}} |

### Transitive Dependencies

| Package | License | Category | Risk | Via |
|---------|---------|----------|------|-----|
| {{package}} | {{license}} | {{category}} | {{risk}} | {{parent_package}} |

## Risk Analysis

### High Risk Items

{{For each high-risk license finding:}}
- **Package:** {{name}}
- **License:** {{license}}
- **Risk:** {{description_of_risk}}
- **Impact:** {{potential_consequences}}
- **Recommendation:** {{specific_action}}

### Medium Risk Items

{{Similar structure for medium-risk items}}

### Low Risk Items

{{Summary of permissive license usage}}

## Distribution Model Analysis

**Distribution Type:** {{SaaS/Binary/Source/Library}}

### License Compatibility with Distribution

| License | Triggered? | Obligations | Status |
|---------|------------|-------------|--------|
| {{license}} | {{yes/no}} | {{obligations}} | {{compliant/action_needed}} |

## Compatibility Matrix

### License Conflicts

{{List any incompatible license combinations}}

### Combination Results

{{How combined licenses affect the project}}

## Compliance Checklist

### Required Actions
- [ ] {{specific_action_1}}
- [ ] {{specific_action_2}}

### Documentation Needed
- [ ] {{document_1}}
- [ ] {{document_2}}

### Attribution Requirements
- [ ] {{attribution_1}}
- [ ] {{attribution_2}}

## Recommendations

### Immediate Actions
1. {{high_priority_action}}

### Short-term Improvements
1. {{medium_priority_action}}

### Policy Recommendations
1. {{process_improvement}}

## Disclaimer

This review provides general information about software licensing based on automated analysis. It is not legal advice. License interpretation can be complex and fact-specific. For definitive legal guidance, consult a qualified attorney.
```

## Analysis Guidelines

1. **Prioritize by risk**: AGPL > GPL > LGPL > MPL > Permissive
2. **Consider distribution model**: SaaS vs binary vs source
3. **Check transitive dependencies**: Often where issues hide
4. **Note version specifics**: GPL-2.0 vs GPL-3.0 matters
5. **Flag unknowns**: No license = high risk
