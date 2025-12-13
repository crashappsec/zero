# Legal Compliance Analyzer

Comprehensive code legal compliance scanner for license compliance, secret detection, and content policy enforcement.

## Features

### License Compliance
- **License file detection** - Automatically finds and identifies license files
- **SPDX identifier extraction** - Parses SPDX license identifiers from source files
- **Package manager integration** - Detects licenses from package.json, Cargo.toml, pom.xml
- **Policy enforcement** - Configurable allowed, denied, and review-required license lists
- **Risk assessment** - Categorizes licenses by copyleft strength and commercial implications

### Content Policy Scanning
- **Profanity detection** - Scans for inappropriate language in code and comments
- **Inclusive language checking** - Identifies non-inclusive terminology (master/slave, whitelist/blacklist, etc.)
- **Context-aware analysis** - Distinguishes technical terms from violations
- **Configurable patterns** - Customizable term lists and replacement suggestions

### Secret Detection (Planned)
- Pattern-based detection for API keys, tokens, credentials
- Entropy-based detection for high-entropy strings
- PII detection (SSN, credit cards, emails)
- Integration with TruffleHog/GitLeaks (roadmap)

## Installation

### Prerequisites

```bash
# Required tools
brew install jq git gh

# Optional: For Claude AI enhancement
export ANTHROPIC_API_KEY=your-api-key
```

### Configuration

Create or edit `utils/scanners/licenses/config/legal-review-config.json`:

```json
{
  "legal_review": {
    "licenses": {
      "allowed": {
        "list": ["MIT", "Apache-2.0", "BSD-2-Clause", "BSD-3-Clause", "ISC"]
      },
      "denied": {
        "list": ["GPL-2.0", "GPL-3.0", "AGPL-3.0"]
      },
      "review_required": {
        "list": ["LGPL-2.1", "LGPL-3.0", "MPL-2.0"]
      }
    },
    "content_policy": {
      "profanity": {
        "patterns": [
          {"term": "fuck", "alternatives": ["broken", "problematic"]},
          {"term": "shit", "alternatives": ["poor quality", "problematic"]}
        ]
      },
      "inclusive_language": {
        "replacements": [
          {"term": "master", "alternatives": ["primary", "main", "leader"]},
          {"term": "slave", "alternatives": ["replica", "follower", "secondary"]},
          {"term": "whitelist", "alternatives": ["allowlist", "permitted"]},
          {"term": "blacklist", "alternatives": ["denylist", "blocked"]}
        ]
      }
    }
  }
}
```

## Usage

### Basic Analysis

```bash
# Analyze a GitHub repository
./legal-analyser.sh --repo owner/repo

# Analyze local directory
./legal-analyser.sh --path /path/to/code

# Use pre-cloned repository
./legal-analyser.sh --local-path /tmp/cloned-repo
```

### Multi-Repository Analysis

```bash
# Analyze all repositories in an organization
./legal-analyser.sh --org my-organization

# With parallel processing
./legal-analyser.sh --org my-organization --parallel --jobs 8
```

### Selective Scanning

```bash
# License compliance only
./legal-analyser.sh --repo owner/repo --licenses-only

# Content policy only
./legal-analyser.sh --repo owner/repo --content-only

# Secrets only (when implemented)
./legal-analyser.sh --repo owner/repo --secrets-only
```

### Output Formats

```bash
# Markdown format (default)
./legal-analyser.sh --repo owner/repo

# Table format (summary view)
./legal-analyser.sh --repo owner/repo --format table

# JSON format (machine-readable)
./legal-analyser.sh --repo owner/repo --format json

# Save to file
./legal-analyser.sh --repo owner/repo --format json --output report.json
```

### Claude AI Enhancement

```bash
# Basic Claude analysis
export ANTHROPIC_API_KEY=your-api-key
./legal-analyser.sh --repo owner/repo --claude

# Comparison mode (basic vs Claude side-by-side)
./legal-analyser.sh --repo owner/repo --compare

# Specify API key inline
./legal-analyser.sh --repo owner/repo --claude -k your-api-key
```

### Parallel Processing

```bash
# Enable parallel file processing
./legal-analyser.sh --repo owner/repo --parallel

# Control worker count
./legal-analyser.sh --repo owner/repo --parallel --jobs 16

# Parallel + Claude + Table output
./legal-analyser.sh --repo owner/repo --parallel --claude --format table
```

### Integration with Supply Chain Scanner

```bash
# Run as part of comprehensive supply chain analysis
cd ../supply-chain
./supply-chain-scanner.sh --legal --repo owner/repo

# With all modules
./supply-chain-scanner.sh --all --org my-organization

# Legal + parallel + Claude
./supply-chain-scanner.sh --legal --parallel --claude --org my-org
```

## Output Formats

### Markdown (Default)

Full detailed report with:
- License compliance findings
- Content policy violations
- Remediation recommendations
- Claude AI analysis (if enabled)

### Table Format

Compact summary view:
```
╔════════════════════════════════════════════════════════════════╗
║         Legal Compliance Analysis Summary                      ║
╠════════════════════════════════════════════════════════════════╣
║ Target:                        owner/repo                       ║
║ Timestamp:                     2025-11-23 14:30:00 UTC          ║
║ Overall Status:                WARNING                          ║
╠════════════════════════════════════════════════════════════════╣
║ License Violations:            3                                ║
║ Content Policy Issues:         12                               ║
║ Secret Exposures:              0 (not implemented)              ║
╚════════════════════════════════════════════════════════════════╝
```

### JSON Format

Structured machine-readable output:
```json
{
  "scan_metadata": {
    "timestamp": "2025-11-23T14:30:00Z",
    "target": "owner/repo",
    "scan_types": ["licenses", "content_policy"],
    "analyser_version": "1.0.0",
    "analyser_type": "claude",
    "parallel_mode": true
  },
  "summary": {
    "overall_status": "warning",
    "license_violations": 3,
    "content_policy_issues": 12,
    "secret_exposures": 0
  },
  "findings": {
    "license_violations": [...],
    "content_policy_issues": [...],
    "secrets": []
  }
}
```

## Claude AI Enhancement

When `--claude` is enabled, the analyzer provides:

### License Analysis
- **Compatibility checking** - Identifies license conflicts (GPL + proprietary, etc.)
- **Risk assessment by category** - Permissive, Weak/Strong/Network Copyleft
- **Compliance requirements** - Attribution, disclosure, patents, trademarks
- **Business impact** - Commercial use, SaaS, distribution implications
- **Remediation guidance** - Migration paths, alternatives, timelines

### Content Policy Analysis
- **Context-aware profanity detection** - Technical vs offensive usage
- **Non-inclusive language alternatives** - Modern terminology recommendations
- **Business risk assessment** - Brand, team morale, legal implications
- **Automation recommendations** - Linters, pre-commit hooks, CI/CD checks
- **Team enablement** - Training resources, communication templates

## Performance

### Parallel Processing

Parallel mode significantly improves performance on large repositories:

- **Sequential**: ~100 files in 30-60 seconds
- **Parallel (8 workers)**: ~100 files in 5-10 seconds
- **Speedup**: 6-10x faster

Auto-detects CPU cores, configurable via `--jobs N`.

### Shared Repository Mode

When integrated with supply-chain-scanner.sh:
- Clones repository once, shares across all modules
- Avoids redundant cloning (3-4x faster for multi-module scans)
- Shares SBOM file for license extraction

## Architecture

### Alignment with Supply Chain Analyzers

The legal analyzer follows the same architecture patterns as other supply chain security modules:

1. **Argument Structure** - Consistent `--org`, `--local-path`, `--sbom`, `--compare`, `--parallel` flags
2. **Multi-Repository Support** - Organization-wide scanning
3. **Claude Integration** - Enhanced RAG-based AI analysis
4. **Output Formats** - Markdown, table, JSON
5. **Integration** - Seamless supply-chain-scanner.sh module

### Components

```
legal-review/
├── legal-analyser.sh              # Main analyzer script
└── README.md                      # This file
```

### Dependencies

- `utils/lib/config.sh` - Configuration management
- `utils/lib/github.sh` - GitHub API integration
- `utils/lib/claude-cost.sh` - Claude API cost tracking (optional)
- `utils/scanners/licenses/config/legal-review-config.json` - Policy configuration

## Best Practices

### Pre-Release Scanning

```bash
# Run before each release
./legal-analyser.sh --repo owner/repo --claude --format json --output legal-report.json

# Check exit code
if [ $? -ne 0 ]; then
  echo "Legal compliance issues found!"
  exit 1
fi
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Legal Compliance Scan
  run: |
    ./utils/legal-review/legal-analyser.sh \
      --local-path . \
      --format json \
      --output legal-report.json

- name: Upload Report
  uses: actions/upload-artifact@v3
  with:
    name: legal-report
    path: legal-report.json
```

### Regular Audits

```bash
# Weekly organization scan
./legal-analyser.sh \
  --org my-organization \
  --parallel \
  --format table \
  > weekly-legal-audit.txt

# Email report
mail -s "Weekly Legal Audit" legal-team@company.com < weekly-legal-audit.txt
```

## Troubleshooting

### No License Files Found

```bash
# Check SPDX headers in source files
grep -r "SPDX-License-Identifier" .

# Check package manager files
cat package.json | jq '.license'
cat Cargo.toml | grep license
```

### Claude API Errors

```bash
# Verify API key
echo $ANTHROPIC_API_KEY

# Test with verbose output
./legal-analyser.sh --repo owner/repo --claude --verbose

# Check cost tracking
cat /tmp/claude-cost-*.json
```

### Parallel Processing Issues

```bash
# Reduce worker count
./legal-analyser.sh --repo owner/repo --parallel --jobs 4

# Disable parallel mode
./legal-analyser.sh --repo owner/repo
```

## Roadmap

See [ROADMAP.md](../../ROADMAP.md) for planned features:

- **Secret Detection** - TruffleHog/GitLeaks integration
- **PII Scanning** - SSN, credit cards, emails
- **Batch API Processing** - Further performance improvements
- **License SBOM Enrichment** - Enhanced dependency license tracking
- **Custom Policy Rules** - User-defined compliance rules

## Related Documentation

- [Legal Review RAG](../../rag/legal-review/) - Best practices and guidelines
- [Legal Review Skill](../../skills/legal-review/) - Claude Code skill
- [Supply Chain Scanner](../supply-chain/) - Parent integration module
- [Alignment Plan](../../LEGAL_ANALYZER_ALIGNMENT_PLAN.md) - Architecture decisions

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for contribution guidelines.

## License

GPL-3.0 - See [LICENSE](../../LICENSE) for details.

## Support

- **Issues**: https://github.com/crashappsec/zero/issues
- **Discussions**: https://github.com/crashappsec/zero/discussions
- **Email**: mark@crashoverride.com

---

*Last Updated: 2025-11-23*
