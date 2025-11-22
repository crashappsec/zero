<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Build Comprehensive Legal Review Analyser

Create a comprehensive legal review analyser for source code that detects license violations, secrets, inappropriate content, and legal risks.

## Overview

The legal analyser should combine license compliance checking, secret detection, and content policy enforcement into a single unified tool that integrates with Gibson Powers analysers.

## Requirements

### 1. License Analysis
- Detect licenses in source files (SPDX identifiers, license text)
- Scan dependencies for licenses (via SBOM)
- Check license compatibility
- Identify denied/restricted licenses
- Verify attribution requirements
- Generate NOTICE/ATTRIBUTION files

### 2. Secret Detection
- Scan for hardcoded credentials (API keys, passwords, tokens)
- Detect high-entropy strings (potential secrets)
- Check for private keys and certificates
- Flag PII (SSN, credit cards, emails in inappropriate contexts)
- Support custom secret patterns

### 3. Content Policy
- Detect profanity in identifiers and comments
- Check for non-inclusive language
- Flag offensive or discriminatory content
- Identify export control concerns
- Detect trademark violations

### 4. Reporting
- JSON and Markdown output formats
- Severity classification (critical, high, medium, low)
- Actionable remediation recommendations
- SPDX/CycloneDX license reports
- Executive summary for stakeholders

## Architecture

```
legal-analyser.sh
├── License scanning (via ScanCode/Licensee or custom)
├── Secret detection (via TruffleHog/GitLeaks or custom)
├── Content policy (custom regex + patterns)
├── SBOM integration (utils/lib/sbom.sh)
└── Claude AI analysis (optional)
```

## Integration Points

### Global Libraries
```bash
source "$REPO_ROOT/utils/lib/sbom.sh"
source "$REPO_ROOT/utils/lib/github.sh"
source "$REPO_ROOT/utils/lib/config.sh"
source "$REPO_ROOT/utils/lib/claude-cost.sh"
```

### Configuration
```json
{
  "legal_review": {
    "licenses": {
      "allowed": ["MIT", "Apache-2.0", "BSD-2-Clause", "BSD-3-Clause"],
      "review_required": ["MPL-2.0", "LGPL-3.0"],
      "denied": ["GPL-3.0", "AGPL-3.0"]
    },
    "secrets": {
      "entropy_threshold": 4.5,
      "custom_patterns": []
    },
    "content_policy": {
      "check_profanity": true,
      "check_inclusive_language": true,
      "severity_threshold": "medium"
    }
  }
}
```

## Implementation Phases

### Phase 1: License Scanner
- Detect license files (LICENSE, COPYING, etc.)
- Extract SPDX identifiers from headers
- Parse package manifests (package.json, pom.xml, etc.)
- Generate SBOM and extract licenses
- Check against policy

### Phase 2: Secret Detection
- Regex patterns for common secrets
- Entropy calculation for random strings
- Git history scanning
- False positive filtering
- Severity scoring

### Phase 3: Content Policy
- Profanity word lists
- Inclusive language checks
- Context-aware scanning
- Whitelist management

### Phase 4: Integration
- Unified reporting
- Claude AI enhancement
- CI/CD hooks
- Dashboard generation

## Expected Output

### JSON Format
```json
{
  "scan_metadata": {
    "timestamp": "2025-01-01T00:00:00Z",
    "repository": "owner/repo",
    "analyser_version": "1.0.0"
  },
  "licenses": {
    "total_files": 150,
    "violations": [
      {
        "file": "src/util.js",
        "license": "GPL-3.0",
        "severity": "high",
        "reason": "GPL-3.0 is on denied list"
      }
    ],
    "summary": {
      "allowed": 145,
      "denied": 1,
      "unknown": 4
    }
  },
  "secrets": {
    "total_findings": 3,
    "high_confidence": [
      {
        "type": "AWS Access Key",
        "file": "config.py",
        "line": 42,
        "severity": "critical"
      }
    ]
  },
  "content_policy": {
    "violations": [
      {
        "type": "profanity",
        "file": "test.js",
        "line": 15,
        "text": "shitty_implementation",
        "severity": "medium",
        "recommendation": "Rename to 'poor_implementation'"
      }
    ]
  }
}
```

### Markdown Format
```markdown
# Legal Review Report

## Executive Summary
- 1 critical issue (hardcoded secret)
- 1 high severity license violation
- 3 medium content policy violations

## License Compliance
✅ 145 files with approved licenses
❌ 1 file with denied license (GPL-3.0)
⚠️  4 files with unknown licenses

### Violations
1. **src/util.js** - GPL-3.0 (denied)
   - Recommendation: Replace with MIT-licensed alternative

## Secret Detection
❌ Found 1 hardcoded AWS key in config.py:42

### Critical Findings
1. **AWS Access Key** in config.py
   - Line 42: `AWS_KEY = "AKIA..."`
   - Action: Remove immediately and rotate key

## Content Policy
⚠️  3 naming violations

1. test.js:15 - Profanity in function name
2. utils.py:87 - Non-inclusive term "whitelist"
3. README.md:12 - Non-inclusive term "master"
```

## Testing

Create test cases covering:
- MIT licensed repository (should pass)
- Repository with GPL dependency (should flag)
- Code with hardcoded API key (should detect)
- Code with profanity (should flag)
- Multi-license repository (compatibility check)

## Integration Examples

### Pre-commit Hook
```bash
#!/bin/bash
./utils/legal-review/legal-analyser.sh --quick --staged
```

### CI/CD
```yaml
- name: Legal Review
  run: ./utils/legal-review/legal-analyser.sh --repo . --fail-on-critical
```

### Full Analysis
```bash
./utils/legal-review/legal-analyser.sh \
  --repo owner/repo \
  --claude \
  --output legal-report.md
```

## Claude AI Integration

The analyser should support Claude AI enhancement for:
- License compatibility analysis
- Risk assessment and prioritization
- Remediation strategy recommendations
- Policy exception evaluation
- Compliance documentation generation

Include RAG context from:
- rag/legal-review/license-compliance-guide.md
- rag/legal-review/content-policy-guide.md
- rag/legal-review/legal-review-tools.md

## Success Criteria

- [ ] Detects all common open source licenses
- [ ] Identifies hardcoded secrets with <5% false positive rate
- [ ] Flags inappropriate content in code and comments
- [ ] Generates actionable reports
- [ ] Integrates with existing Gibson Powers infrastructure
- [ ] Supports both local and remote repositories
- [ ] Provides Claude AI-enhanced analysis
- [ ] Runs in <2 minutes for typical repository
- [ ] Zero false negatives for critical issues

## References

- [License Compliance Guide](../../rag/legal-review/license-compliance-guide.md)
- [Content Policy Guide](../../rag/legal-review/content-policy-guide.md)
- [Legal Review Tools](../../rag/legal-review/legal-review-tools.md)
- [SBOM Library](../../utils/lib/sbom.sh)
