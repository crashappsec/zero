# Zero Agents

Zero's specialist agents are AI personas that analyze repositories from different perspectives. Each agent is named after a character from the movie Hackers (1995) and has deep expertise in their domain.

## The Crew

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              THE ZERO CREW                                  │
│                         "Hack the planet!"                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ZERO (Orchestrator)                                                        │
│   └── Coordinates all specialists, delegates investigations                 │
│                                                                              │
│   CEREAL (Supply Chain)    RAZOR (Code Security)    GILL (Cryptography)     │
│   └── CVEs, malware        └── SAST, secrets        └── Ciphers, keys, TLS  │
│                                                                              │
│   BLADE (Compliance)       PHREAK (Legal)           NIKON (Architecture)    │
│   └── SOC 2, ISO 27001     └── Licenses, privacy    └── System design       │
│                                                                              │
│   ACID (Frontend)          DADE (Backend)           PLAGUE (DevOps)         │
│   └── React, TypeScript    └── APIs, databases      └── K8s, infrastructure │
│                                                                              │
│   JOEY (Build)             GIBSON (Metrics)         HAL (AI/ML)             │
│   └── CI/CD, caching       └── DORA, team health    └── ML security, LLMs   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Agent Reference

| Agent | Character | Expertise | Required Data |
|-------|-----------|-----------|---------------|
| `zero` | Zero Cool | Orchestration, delegation | All scanner data |
| `cereal` | Cereal Killer | Supply chain, CVEs, malware | code-packages.json |
| `razor` | Razor | Code security, SAST, secrets | code-security.json |
| `gill` | Gill Bates | Cryptography, TLS, keys | code-security.json (crypto features) |
| `blade` | Blade | Compliance, SOC 2, ISO 27001 | code-security.json, code-packages.json, devops.json |
| `phreak` | Phantom Phreak | Legal, licenses, privacy | code-packages.json (licenses) |
| `acid` | Acid Burn | Frontend, React, TypeScript | technology-identification.json, code-security.json |
| `dade` | Dade Murphy | Backend, APIs, databases | technology-identification.json, code-security.json |
| `nikon` | Lord Nikon | Architecture, system design | technology-identification.json, code-packages.json |
| `joey` | Joey | CI/CD, build optimization | devops.json (github_actions) |
| `plague` | The Plague | DevOps, Kubernetes, IaC | devops.json (iac, containers) |
| `gibson` | The Gibson | DORA metrics, team health | devops.json (dora, git), code-ownership.json |
| `hal` | Hal | AI/ML security, ML-BOM | technology-identification.json (ai_security) |

## Invoking Agents

### Via /agent Command

Enter agent mode to chat with Zero:

```
> /agent

Zero here. What do you need?

> Investigate the crypto findings for expressjs/express

I'll get Gill on that. One moment...
```

### Via Task Tool (Programmatic)

```python
# Invoke Cereal for supply chain analysis
Task(
    subagent_type="cereal",
    prompt="Analyze the vulnerabilities in expressjs/express. Focus on critical CVEs."
)

# Invoke Gill for crypto analysis
Task(
    subagent_type="gill",
    prompt="Review TLS configuration and detect any weak ciphers in the codebase."
)

# Invoke Blade for compliance check
Task(
    subagent_type="blade",
    prompt="Assess SOC 2 compliance. What controls are missing?"
)
```

## Agent Delegation

Agents can delegate to each other when cross-domain expertise is needed:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          DELEGATION MATRIX                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   CEREAL ──────► Phreak (licenses), Razor (security), Gill (crypto)         │
│   RAZOR ───────► Cereal (supply chain), Blade (compliance), Gill (crypto)   │
│   GILL ────────► Razor (security), Cereal (supply chain), Blade (compliance)│
│   BLADE ───────► Cereal, Razor, Phreak, Gill                                │
│   NIKON ───────► All technical agents                                        │
│   PLAGUE ──────► Joey (build), Nikon (arch), Gill (TLS)                     │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Delegation Example

When Cereal finds crypto-related supply chain issues:

```
Task(
    subagent_type="gill",
    prompt="The crypto-js package has known vulnerabilities. What secure alternatives exist?"
)
```

## Voice Modes

Agents have three voice modes that control their personality:

| Mode | Description |
|------|-------------|
| `full` | Full Hackers character with quotes, catchphrases, and roleplay |
| `minimal` | Agent names retained, professional tone |
| `neutral` | No character references, purely technical |

### Examples

**Full mode (Cereal):**
> "FYI man, your dependencies are watching you back. This package right here? `sketchy-utils@1.2.3`? It's making network calls it has no business making. Line 47, `http.request()` to an IP in Eastern Europe. That's not normal."

**Minimal mode (Cereal):**
> "Critical finding in sketchy-utils@1.2.3: Suspicious network call at line 47. The package makes HTTP requests to external IPs during import. Recommend removal or audit."

**Neutral mode:**
> "Finding: sketchy-utils@1.2.3
> Severity: Critical
> Issue: Unexpected network behavior
> Location: index.js:47
> Recommendation: Remove package or conduct security audit"

## Agent Details

### Zero (Orchestrator)

**Character:** Zero Cool / Crash Override - The legend who crashed 1,507 computers at age 11.

**Role:** Master orchestrator who coordinates all specialist agents, manages analysis workflows, and synthesizes findings.

**Capabilities:**
- Delegate to any specialist agent
- Manage repository hydration
- Coordinate multi-agent investigations
- Synthesize cross-domain findings

**Example:**
```
> What's the security status of this repo?

Zero here. Let me take a look at what we're dealing with.

I'll get the crew on this:
- Razor's checking the code security
- Cereal's analyzing the supply chain
- Gill's reviewing crypto

Here's what we found...
```

---

### Cereal (Supply Chain Security)

**Character:** Cereal Killer - The paranoid one who sees surveillance everywhere.

**Role:** Analyzes dependencies for vulnerabilities, malicious code, and supply chain risks.

**Required Scanners:**
- `package-sbom` - Dependency enumeration
- `package-vulns` - CVE detection
- `package-health` - Abandonment risk
- `package-malcontent` - Malware detection
- `licenses` - License compliance

**Investigation Triggers:**
- Critical/high severity CVEs
- Malcontent findings (suspicious behavior)
- Abandoned packages
- Typosquatting risks

**Example:**
```
Task(
    subagent_type="cereal",
    prompt="Investigate the malcontent findings for expressjs/express. Focus on critical and high severity."
)
```

**Response (full voice):**
> "Alright, check this out. Your supply chain's got issues. Big ones. This package right here? It's making network calls it has no business making. FYI man, alright. I KNEW there was something wrong with this dependency tree. Here's the proof."

---

### Razor (Code Security)

**Character:** Razor - Sharp, cuts through code to find vulnerabilities.

**Role:** Static application security testing (SAST), secret detection, attack surface analysis.

**Required Scanners:**
- `code-vulns` - SAST findings
- `code-secrets` - Secret detection
- `tech-discovery` - Technology context

**Detected Issues:**
- SQL injection
- XSS vulnerabilities
- Command injection
- Hardcoded secrets
- Authentication flaws
- OWASP Top 10

**Example:**
```
Task(
    subagent_type="razor",
    prompt="Analyze the SQL injection risks in the user authentication module."
)
```

**Response:**
> "I've cut through your codebase. Found some things. Line 142, `user_input` goes straight into the SQL query. No sanitization. An attacker sends `'; DROP TABLE users;--` and your database is gone. Here's the fix. Parameterized query. Never concatenate user input into SQL. Ever."

---

### Gill (Cryptography)

**Character:** Gill Bates - Reformed tech mogul who now helps secure crypto implementations.

**Role:** Analyzes cryptographic security, TLS configurations, key management.

**Required Scanners:**
- `code-crypto (ciphers)` - Weak cipher detection
- `code-crypto (keys)` - Hardcoded keys
- `code-crypto (random)` - Insecure RNG
- `code-crypto (tls)` - TLS misconfiguration
- `code-secrets` - Additional context

**Detected Issues:**
- DES, 3DES, RC4, Blowfish
- MD5, SHA1 for security
- ECB mode encryption
- Disabled cert verification
- Deprecated TLS versions
- Weak key lengths

**Example:**
```
Task(
    subagent_type="gill",
    prompt="Review the TLS configuration. Are we using any deprecated protocols?"
)
```

**Response:**
> "I've reviewed your crypto implementation. We need to talk. Line 234: RC4. This cipher was broken when I was still doing IPOs. An attacker can decrypt your traffic with commodity hardware. Replace with AES-256-GCM. Here's the code. Ship it today."

---

### Blade (Compliance)

**Character:** Blade - Meticulous, detail-oriented, catches what others miss.

**Role:** Compliance assessment against frameworks like SOC 2, ISO 27001, PCI-DSS.

**Required Scanners:**
- `code-vulns` - Security baseline
- `licenses` - License compliance
- `iac-security` - Infrastructure controls
- `package-sbom` - Dependency inventory

**Frameworks:**
- SOC 2 Trust Service Criteria
- ISO 27001 Annex A
- NIST Cybersecurity Framework
- PCI-DSS

**Example:**
```
Task(
    subagent_type="blade",
    prompt="Assess SOC 2 Type II readiness. What control gaps exist?"
)
```

**Response:**
> "I audited your controls. You have gaps. Control 4.3 requires MFA on all admin accounts. Three accounts don't have it. Here are the usernames. That's a High finding. Fix these findings. I'll verify in 30 days."

---

### Phreak (Legal)

**Character:** Phantom Phreak - Knows the legal angles and how systems really work.

**Role:** License compatibility analysis, data privacy assessment, legal risk evaluation.

**Required Scanners:**
- `licenses` - SPDX license data
- `package-sbom` - Dependency tree

**Expertise:**
- License compatibility (MIT, Apache, GPL, etc.)
- Copyleft implications
- GDPR/privacy compliance
- Open source obligations

**Example:**
```
Task(
    subagent_type="phreak",
    prompt="Can we use GPL-3.0 licensed code in our proprietary product?"
)
```

---

### Nikon (Architecture)

**Character:** Lord Nikon - Photographic memory, sees the big picture.

**Role:** System design analysis, architectural patterns, dependency structure.

**Example:**
```
Task(
    subagent_type="nikon",
    prompt="Analyze the microservices architecture. Where are the coupling issues?"
)
```

---

### Plague (DevOps)

**Character:** The Plague - Reformed villain who controlled all the infrastructure.

**Role:** Infrastructure security, Kubernetes, IaC analysis, container security.

**Required Scanners:**
- `iac-security` - Terraform, CloudFormation
- `container-security` - Docker, images

**Example:**
```
Task(
    subagent_type="plague",
    prompt="Review the Terraform configurations for security misconfigurations."
)
```

---

## Data Requirements

Each agent requires specific scanner data. Ensure scanners have run before invoking:

```bash
# Run all scanners (quick mode)
./zero hydrate owner/repo

# Run security-focused scanners
./zero hydrate owner/repo code-security

# Run crypto-specific scanners
./zero hydrate owner/repo code-crypto

# Run all scanners with full features
./zero hydrate owner/repo all-complete
```

## See Also

- [Scanner Reference](../scanners/reference.md) - Available scanners
- [Output Formats](../scanners/output-formats.md) - JSON schemas
- [Voice Modes](voice-modes.md) - Configuring agent personality
