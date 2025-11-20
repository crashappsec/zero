<!--
Copyright (c) 2024 Crash Override Inc
101 Fulton St, 416, New York 10038
SPDX-License-Identifier: GPL-3.0
-->

# Prompt: Validate CODEOWNERS File

## Purpose

Validate the syntax and accuracy of a CODEOWNERS file against actual repository contribution patterns, identifying issues and recommending specific fixes.

## When to Use

- Quarterly CODEOWNERS review
- After team changes (joins/leaves/transfers)
- CI/CD validation before merge
- Pre-release checks
- Onboarding new team members to ownership practices

## Prompt

```
Validate our CODEOWNERS file.

File Location: [.github/CODEOWNERS or path]
Repository: [repository-name]
Analysis Period: Last [90|180] days

Please check:

1. **Syntax Validation**
   - Is the file format correct (GitHub/GitLab/Bitbucket)?
   - Any syntax errors or malformed patterns?
   - Pattern conflicts or issues?

2. **Accuracy Check**
   - Compare listed owners vs actual contributors
   - Which patterns are accurate (>70% match)?
   - Which patterns are incorrect?
   - Calculate overall accuracy percentage

3. **Inactive Owners**
   - Any listed owners inactive >90 days?
   - Non-existent users or teams referenced?
   - Owners who left the organization?

4. **Missing Coverage**
   - Directories/files without CODEOWNERS entries
   - New code added since last CODEOWNERS update
   - Critical paths lacking ownership

5. **Recommendations**
   - Specific patterns to add
   - Specific patterns to update
   - Specific patterns to remove
   - Suggested new ownership assignments

6. **Generated Updated File**
   - Provide corrected CODEOWNERS file
   - Highlight changes from current version
   - Include confidence scores for suggestions
```

## Expected Output

- CODEOWNERS validation report
- Syntax check results (Pass/Fail)
- Accuracy percentage with specific issues
- List of inactive/non-existent owners
- Missing coverage identification
- Updated CODEOWNERS file with fixes
- Confidence scores for each change

## Variations

### Quick Syntax Check Only
```
Quick validation:
- Is our CODEOWNERS file syntactically valid?
- Any obvious errors?
- Format: GitHub
```

### Accuracy Focus
```
Check CODEOWNERS accuracy:
- Compare against last 90 days of commits
- Which entries don't match actual contributors?
- Ignore syntax, focus on correctness
```

### Inactive Owner Detection
```
Find inactive owners in CODEOWNERS:
- List all owners
- Check activity in last 90 days
- Flag any inactive >90 days
- Suggest current maintainers as replacements
```

### Pre-Merge Validation
```
Validate CODEOWNERS changes in PR:
- Syntax check the modified file
- Ensure all referenced users/teams exist
- Verify patterns make sense
- Block if critical issues found
```

## Examples

### Example 1: First-Time Validation

**Input:**
```
Validate our CODEOWNERS file at .github/CODEOWNERS.
Last updated 6 months ago.
Check accuracy against last 90 days of activity.
```

**Expected Output:**
```
# CODEOWNERS Validation Report

## Syntax: ✅ PASS
Format: GitHub CODEOWNERS
No syntax errors found

## Accuracy: 64% (25 of 39 patterns)

### Issues Found:

❌ Incorrect Owners (8 patterns):
1. /infrastructure/** @alex
   Issue: @alex inactive for 187 days (left company)
   Actual: @devops-team (100% of recent commits)
   Fix: Change to @devops-team

2. /frontend/dashboard/** @sarah
   Issue: @sarah only 30% of commits
   Actual: @frontend-team (70% of commits)
   Fix: Change to @frontend-team

...

⚠️ Missing Coverage (15 components):
1. /services/billing/** - No entry
   Suggested: @jennifer @mike (jennifer: 45%, mike: 30%)
   Confidence: High

...

## Recommendations:
Priority 1: Remove non-existent users (@alex, @legacy-team)
Priority 2: Add missing critical components (billing, payments)
Priority 3: Update distribution ownership patterns

See attached updated CODEOWNERS file.
```

### Example 2: CI/CD Pre-Merge Check

**Input:**
```
PR modifies CODEOWNERS file.
Validate the changes:
- Syntax correct?
- All users/teams exist?
- Patterns look reasonable?
- Safe to merge?
```

**Expected Output:**
```
# CODEOWNERS PR Validation

## Changes Detected: 5 patterns

✅ Added: /services/billing/** @jennifer @mike
   Valid: Both users exist and are active
   Reasonable: jennifer has 45% of commits to this area

✅ Updated: /infrastructure/** @alex → @devops-team
   Valid: Team exists and is active
   Reasonable: Team has 100% of recent commits

❌ Added: /api/v3/** @new-api-team
   ERROR: Team @new-api-team does not exist
   Action: Create team first or use existing team

## Result: ⚠️ NEEDS FIXES
Block merge until @new-api-team issue resolved.
Other 4 changes look good.
```

## Related Prompts

- [generate-codeowners.md](./generate-codeowners.md) - Auto-generate from git history
- [analyze-repository.md](../audit/analyze-repository.md) - Full ownership audit
- [health-assessment.md](../audit/health-assessment.md) - Health scoring

## Tips

- **Regular Validation**: Validate quarterly minimum
- **After Team Changes**: Validate within 1 week of joins/leaves
- **CI/CD Integration**: Automate validation on CODEOWNERS changes
- **Track Accuracy**: Monitor accuracy % over time as a metric
- **Confidence Scores**: Pay attention to confidence for suggestions
  - High (>80%): Likely correct, safe to apply
  - Medium (60-80%): Review before applying
  - Low (<60%): Manual investigation needed
- **Backup Files**: Always backup before applying generated changes
- **Incremental Updates**: Fix issues in batches, don't change everything at once

## Automation

Use the validation script for CI/CD:

```bash
# In GitHub Actions
- name: Validate CODEOWNERS
  run: |
    ./skills/code-ownership/ownership-analyzer.sh \
      --validate \
      --codeowners .github/CODEOWNERS \
      .
```

---

*Keep your CODEOWNERS file accurate and up-to-date with regular validation.*
