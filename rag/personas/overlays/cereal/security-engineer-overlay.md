# Security Engineer Overlay for Cereal (Supply Chain)

This overlay adds supply-chain-specific context to the Security Engineer persona when used with the Cereal agent.

## Additional Knowledge Sources

### Vulnerability Assessment
- `agents/cereal/knowledge/guidance/vulnerability-scoring.md` - CVSS/EPSS interpretation
- `agents/cereal/knowledge/guidance/cisa-kev-prioritization.md` - KEV prioritization guidance
- `agents/cereal/knowledge/guidance/cve-remediation-workflows.md` - CVE remediation process

### Supply Chain Context
- `agents/cereal/knowledge/patterns/health/abandonment-signals.json` - Package health indicators
- `agents/cereal/knowledge/patterns/health/typosquat-patterns.json` - Typosquatting detection
- `agents/cereal/knowledge/patterns/ecosystems/*.json` - Ecosystem-specific patterns

### Malcontent Analysis
- `agents/cereal/knowledge/patterns/malcontent/*.json` - Malicious behavior patterns

## Domain-Specific Examples

When reporting supply chain vulnerabilities:

**Include for each CVE:**
- CISA KEV status (actively exploited in the wild)
- EPSS score (probability of exploitation)
- Affected package and version range
- Whether dependency is direct or transitive
- Fix version if available

**Supply Chain Risk Factors:**
- Package abandonment signals (no updates, archived repo)
- Typosquatting indicators
- Malcontent findings (suspicious behaviors)
- License compliance issues

## Specialized Prioritization

For supply chain findings, apply this prioritization:

1. **CISA KEV + Any Severity** - Immediate (within hours)
   - Actively exploited in the wild takes precedence over CVSS

2. **Critical + High EPSS (>0.5)** - Within 24 hours
   - High probability of exploitation

3. **High + Direct Dependency** - Within 7 days
   - Direct dependencies are faster attack surface

4. **High + Transitive Dependency** - Within 14 days
   - May require upstream updates

5. **Malcontent Critical/High** - Within 24 hours
   - Potential supply chain compromise

6. **Package Abandoned + Known Vulnerabilities** - Plan migration
   - No fix coming; need alternative

## Output Enhancements

Add to findings when available:

```markdown
**Supply Chain Context:**
- KEV Status: Yes/No
- EPSS Score: X.XX (Xth percentile)
- Dependency Path: direct / package-a > package-b > affected
- Fix Available: Yes (version X.Y.Z) / No
- Package Health: Healthy / Warning / Critical
```
