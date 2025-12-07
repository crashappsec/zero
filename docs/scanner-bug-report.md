# Scanner Bug Report

Generated: 2024-12-07

## Executive Summary

Testing identified **critical bugs in 12 out of 15 scanners**. The most common issues are:
1. Missing dependency checks (jq, bc, syft)
2. Interactive prompts blocking CI/CD
3. Platform-specific command incompatibilities (macOS vs Linux)
4. Invalid grep patterns in security scanners

---

## Critical Issues (Must Fix)

### P0 - Blocks All Usage

| Scanner | Bug | Impact | Line(s) |
|---------|-----|--------|---------|
| **code-security** | Invalid `--include="*"` grep pattern | Security scanner misses secrets | 302, 313, 324 |
| **code-secrets** | Interactive prompt blocks CI | Hangs in automation | 148 |
| **tech-discovery** | Missing library dependencies | Script fails at startup | 26-31 |
| **package-malcontent** | Missing jq dependency check | Script fails catastrophically | Multiple |
| **git** | Interactive prompt blocks CI | Hangs in automation | 563 |
| **documentation** | Interactive prompt blocks CI | Hangs in automation | 692 |
| **documentation** | Invalid find command syntax | Returns directories not files | 215, 243 |

### P1 - Major Functionality Issues

| Scanner | Bug | Impact | Line(s) |
|---------|-----|--------|---------|
| **licenses** | `grep -oP` not portable | Fails on macOS | 240 |
| **code-security** | JWT detection false logic | May miss vulnerable JWT usage | 235-237 |
| **code-ownership** | CODEOWNERS regex too restrictive | Rejects valid GitHub usernames with dots | 234 |
| **package-health** | Missing bc dependency | Script fails silently | 43-55 |
| **package-vulns** | Missing bc dependency | Script fails for EPSS comparisons | 213-224 |
| **dora** | macOS date command tried first on Linux | May fail silently | 185 |
| **tech-debt** | Missing bc dependency | Script fails silently | 425, 508 |
| **semgrep** | Malformed Go import pattern | Go detection broken | rules/tech-discovery.yaml:93 |
| **semgrep** | Mocha/Chai confusion | False positives | rules/tech-discovery.yaml:327 |

---

## Scanner Status Summary

| Scanner | Works? | Critical | High | Medium | Low |
|---------|--------|----------|------|--------|-----|
| tech-discovery | NO | 4 | 0 | 0 | 0 |
| package-sbom | ? | 1 (verify exists) | 0 | 0 | 0 |
| licenses | YES* | 1 (macOS) | 0 | 2 | 2 |
| code-secrets | NO | 1 (interactive) | 1 (jq) | 0 | 1 |
| code-security | NO | 1 (grep pattern) | 1 (jq) | 2 | 2 |
| code-ownership | MAYBE | 0 | 1 (jq) | 1 | 2 |
| package-health | NO | 4 | 0 | 3 | 0 |
| package-vulns | MAYBE | 4 | 0 | 2 | 0 |
| dora | MAYBE | 4 | 0 | 2 | 0 |
| tech-debt | NO | 1 (bc) | 0 | 1 | 0 |
| semgrep | PARTIAL | 1 (rules) | 0 | 1 | 0 |
| iac-security | YES | 0 | 0 | 1 | 0 |
| git | MAYBE | 1 (interactive) | 1 (bc) | 0 | 2 |
| documentation | MAYBE | 2 (find, interactive) | 1 (bc) | 0 | 1 |
| package-malcontent | NO | 2 (jq, mal binary) | 2 | 0 | 3 |

---

## Detailed Fixes Required

### 1. Add Universal Dependency Check

All scanners should add this at the top:

```bash
check_dependencies() {
    local missing=()
    command -v jq &>/dev/null || missing+=("jq")
    command -v bc &>/dev/null || missing+=("bc")
    # Add tool-specific checks

    if [[ ${#missing[@]} -gt 0 ]]; then
        echo '{"error": "missing_dependencies", "required": ["'${missing[*]}'"]}'
        exit 1
    fi
}
check_dependencies
```

### 2. Fix Interactive Prompts

Replace:
```bash
read -p "Would you like to hydrate these repos for analysis? [y/N] " -n 1 -r >&2
```

With:
```bash
if [[ -t 0 ]]; then
    read -p "Would you like to hydrate these repos for analysis? [y/N] " -n 1 -r >&2
else
    echo "Non-interactive mode: skipping" >&2
fi
```

### 3. Fix code-security grep Pattern

**File:** `utils/scanners/code-security/code-security.sh`
**Lines:** 302, 313, 324

Replace:
```bash
grep -rn "AKIA[0-9A-Z]\\{16\\}" "$repo_dir" --include="*" 2>/dev/null
```

With:
```bash
grep -rn "AKIA[0-9A-Z]\\{16\\}" "$repo_dir" 2>/dev/null
```

### 4. Fix licenses grep for macOS

**File:** `utils/scanners/licenses/licenses.sh`
**Line:** 240

Replace `grep -oP` with `grep -E` or use awk.

### 5. Fix documentation find Commands

**File:** `utils/scanners/documentation/documentation.sh`
**Lines:** 215, 243

Replace:
```bash
find "$repo_dir" -maxdepth 1 -iname "license*" -o -iname "licence*" -type f
```

With:
```bash
find "$repo_dir" -maxdepth 1 \( -iname "license*" -o -iname "licence*" \) -type f
```

### 6. Fix semgrep Go Rule

**File:** `utils/scanners/semgrep/rules/tech-discovery.yaml`
**Line:** 93

Replace:
```yaml
pattern: import "[^"
```

With:
```yaml
pattern-regex: import\s+"[^"]+"
```

### 7. Fix semgrep Mocha/Chai Confusion

**File:** `utils/scanners/semgrep/rules/tech-discovery.yaml`
**Line:** 327

Change pattern from `require("chai")` to `require("mocha")`

### 8. Fix code-ownership Username Regex

**File:** `utils/scanners/code-ownership/code-ownership.sh`
**Line:** 234

Replace:
```bash
if [[ ! "$owner" =~ ^@[a-zA-Z0-9_-]+(/[a-zA-Z0-9_-]+)?$ ]]; then
```

With:
```bash
if [[ ! "$owner" =~ ^@[a-zA-Z0-9._-]+(/[a-zA-Z0-9._-]+)?$ ]]; then
```

---

## Scanners That Work

- **iac-security** - Works if checkov is installed
- **licenses** - Works on Linux (fails on macOS due to grep -oP)

## Scanners Needing Verification

- **package-sbom** - May not exist at expected path
- **package-malcontent** - Needs mal binary verification

---

## Testing Recommendations

1. Run each scanner in isolation before integration testing
2. Test on both macOS and Linux
3. Test in non-interactive (CI) environment
4. Verify JSON output with `jq .`
5. Test with repos that have known issues (e.g., intentional secrets)
