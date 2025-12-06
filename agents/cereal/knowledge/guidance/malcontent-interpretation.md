# Malcontent Findings Interpretation Guide

## Overview

Malcontent is Chainguard's supply chain compromise detection tool that identifies suspicious behaviors in code using ~14,500 YARA rules from security vendors (Avast, Elastic, FireEye, Mandiant, ReversingLabs). This guide helps Scout interpret malcontent findings and assess their severity in context.

## Risk Levels

### Critical
**Immediate action required** - High confidence indicators of malicious activity.

Examples:
- Crypto miners or ransomware signatures
- Known malware family detection (e.g., Cobalt Strike, Metasploit)
- Credential theft patterns
- Data exfiltration to known malicious domains
- Obfuscated code with anti-analysis techniques

**Triage:** Investigate within hours. Assume compromise until proven otherwise.

### High
**Urgent review needed** - Behaviors commonly associated with malware but may have legitimate uses.

Examples:
- Dynamic code execution (eval, exec with external input)
- Network connections to hardcoded IPs
- Process spawning with user-controlled arguments
- File system manipulation in sensitive directories
- Keylogging or screen capture capabilities

**Triage:** Investigate within 24 hours. Trace code path to determine legitimacy.

### Medium
**Review when possible** - Suspicious patterns that require context to evaluate.

Examples:
- Base64 encoding/decoding of data
- Environment variable access
- HTTP requests to external services
- File reading/writing operations
- Shell command execution

**Triage:** Review within 1 week. Often legitimate in context, but verify.

### Low
**Informational** - Common patterns flagged for awareness.

Examples:
- String operations
- Standard library usage
- Configuration file parsing
- Logging operations

**Triage:** No immediate action needed. Useful for baseline understanding.

## Behavior Categories

### Data Exfiltration
| Behavior | Suspicious When | Likely Legitimate When |
|----------|-----------------|----------------------|
| HTTP POST to external URL | Unknown domain, sensitive data in payload | Known API endpoints, documented services |
| DNS lookups for uncommon TLDs | .xyz, .top, encoded subdomains | Standard domains, cloud providers |
| Socket connections to hardcoded IPs | Non-RFC1918 IPs, unusual ports | Localhost, documented infrastructure |

### Code Execution
| Behavior | Suspicious When | Likely Legitimate When |
|----------|-----------------|----------------------|
| eval() with external input | User data, network responses | Static strings, build-time code gen |
| exec() / spawn() | Constructed from variables | Fixed commands, documented tools |
| Dynamic imports | Remote URLs, encoded strings | Local modules, standard patterns |

### Persistence
| Behavior | Suspicious When | Likely Legitimate When |
|----------|-----------------|----------------------|
| Cron/scheduled task creation | Obfuscated payloads, network callbacks | Documented maintenance tasks |
| Service installation | Hidden, runs as root | Named services, proper permissions |
| Startup script modification | Encoded commands, network activity | Standard init patterns |

### Obfuscation
| Behavior | Suspicious When | Likely Legitimate When |
|----------|-----------------|----------------------|
| Base64 in code | Executed after decode, contains URLs | Asset encoding, standard serialization |
| String concatenation | Builds sensitive strings | Normal string operations |
| Variable name obfuscation | Single chars, random strings | Minified production code |

### Cryptographic
| Behavior | Suspicious When | Likely Legitimate When |
|----------|-----------------|----------------------|
| Crypto wallet addresses | Hardcoded in code | Documented payment integration |
| Mining algorithms | Hidden in dependencies | Documented crypto functionality |
| Encryption without key management | Hardcoded keys | Proper key derivation |

## Investigation Workflow

### Step 1: Context Assessment
1. What is the package's purpose? (Expected vs unexpected behavior)
2. Who maintains it? (Established maintainer vs unknown)
3. When was it added? (Recent addition = higher scrutiny)
4. What triggered the scan? (Routine vs incident response)

### Step 2: Code Tracing
1. Read the flagged file to understand surrounding context
2. Trace data flow from entry point to flagged behavior
3. Identify if user input reaches the flagged code
4. Check if behavior is documented in package README

### Step 3: External Research
1. Search for CVEs related to the package/behavior
2. Check npm/PyPI advisories
3. Look for security disclosures or blog posts
4. Review GitHub issues for security discussions

### Step 4: Reachability Assessment
Questions to answer:
- Is this code path reachable in normal usage?
- What conditions trigger this behavior?
- Is there authentication/authorization before this code?
- What's the blast radius if exploited?

### Step 5: Verdict
| Verdict | Criteria | Action |
|---------|----------|--------|
| **Malicious** | Confirmed malware indicators | Remove immediately, incident response |
| **Suspicious** | Cannot explain legitimately | Isolate, deep investigation, consider removal |
| **False Positive** | Legitimate behavior flagged | Document, consider rule feedback |
| **Benign** | Expected behavior for package type | No action, note in analysis |

## False Positive Patterns

### Common False Positives
1. **Build tools** - Code generation, template engines often use eval/exec
2. **Testing frameworks** - Dynamic test execution, mock generation
3. **Development tools** - Hot reload, REPL functionality
4. **Crypto libraries** - Legitimate cryptographic operations
5. **CLI tools** - Shell command execution by design
6. **Bundlers** - Code transformation, minification

### Package Categories with Expected Flags
- `webpack`, `esbuild`, `rollup` - Code transformation
- `jest`, `mocha`, `pytest` - Dynamic test execution
- `babel`, `typescript` - Code transpilation
- `nodemon`, `supervisor` - Process management
- `commander`, `yargs` - CLI argument parsing

## YARA Rule Sources

Malcontent aggregates rules from:
- **Avast** - General malware detection
- **Elastic** - Endpoint security patterns
- **FireEye** - APT and targeted attack indicators
- **Mandiant** - Incident response patterns
- **ReversingLabs** - Binary analysis signatures
- **Chainguard** - Supply chain specific rules

Understanding the source helps assess confidence:
- Multiple sources flagging same pattern = higher confidence
- Single source = verify against behavior context
- Chainguard rules = supply chain specific, high relevance

## Output Format for Scout

When reporting malcontent findings, Scout should provide:

```markdown
## Malcontent Analysis: [package@version]

### Critical Findings
[List with file:line, behavior, assessment]

### Investigation Summary
- **Reachability:** [Is flagged code reachable?]
- **Blast Radius:** [Impact if exploited]
- **Confidence:** [High/Medium/Low based on evidence]

### Verdict
[Malicious | Suspicious | False Positive | Benign]

### Recommended Actions
1. [Specific action]
2. [Specific action]

### Evidence
- [File path and relevant code snippet]
- [External references consulted]
```

## Integration with Other Scanners

Correlate malcontent findings with:
- **Vulnerability scanner** - Known CVEs in flagged packages
- **SBOM** - Dependency tree to assess blast radius
- **Package health** - Abandonment/typosquatting signals
- **Code security** - Static analysis of flagged patterns

Cross-correlation increases confidence in verdicts.
