<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com
SPDX-License-Identifier: GPL-3.0
-->

# Remediation Techniques for Supply Chain Vulnerabilities

## Upgrade Strategies

### Direct Dependency Upgrades

**NPM/Node.js:**
```bash
# Check for updates
npm outdated

# Update to specific version
npm install package@1.2.3

# Update to latest (within semver)
npm update package

# Force update (including breaking)
npm install package@latest

# Fix audit issues automatically
npm audit fix
npm audit fix --force  # Include breaking changes
```

**Python/pip:**
```bash
# Check for updates
pip list --outdated

# Upgrade specific package
pip install package==1.2.3
pip install --upgrade package

# Fix vulnerabilities
pip-audit --fix
```

**Go:**
```bash
# Update specific dependency
go get package@v1.2.3

# Update all dependencies
go get -u ./...

# Tidy up
go mod tidy
```

**Rust/Cargo:**
```bash
# Update dependencies
cargo update

# Check for vulnerabilities
cargo audit
cargo audit fix
```

### Transitive Dependency Handling

**NPM - Override transitive dependencies:**
```json
// package.json
{
  "overrides": {
    "vulnerable-package": "2.0.0",
    "parent-package>vulnerable-package": "2.0.0"
  }
}
```

**Yarn - Resolutions:**
```json
// package.json
{
  "resolutions": {
    "vulnerable-package": "2.0.0",
    "**/vulnerable-package": "2.0.0"
  }
}
```

**Python - Constraints:**
```bash
# constraints.txt
vulnerable-package==2.0.0

# Install with constraints
pip install -c constraints.txt -r requirements.txt
```

## Patching Without Upgrading

### When to Use Manual Patches

- Vendor hasn't released fix yet
- Upgrade would require significant refactoring
- Critical vulnerability needs immediate mitigation

### Applying Patches

**Using patch files:**
```bash
# Download security patch
curl -O https://example.com/security-patch.diff

# Apply patch
patch -p1 < security-patch.diff

# Verify patch applied
git diff
```

**NPM patch-package:**
```bash
# Make local fix
vim node_modules/package/vulnerable-file.js

# Create patch
npx patch-package package-name

# Patch auto-applies on npm install
```

## Compensating Controls

### Network-Level Controls

**WAF Rules (example for Log4Shell):**
```yaml
# AWS WAF Rule
Rules:
  - Name: BlockLog4Shell
    Priority: 1
    Action:
      Block: {}
    Statement:
      OrStatement:
        Statements:
          - ByteMatchStatement:
              FieldToMatch:
                AllQueryArguments: {}
              SearchString: "${jndi:"
              TextTransformations:
                - Priority: 1
                  Type: URL_DECODE
```

**Network Segmentation:**
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Internet   │────▶│   WAF/LB    │────▶│  DMZ Zone   │
└─────────────┘     └─────────────┘     └─────────────┘
                                              │
                                              ▼
                                        ┌─────────────┐
                                        │ App Servers │
                                        │(vulnerable) │
                                        └─────────────┘
                                              │
                                              ▼ (blocked)
                                        ┌─────────────┐
                                        │  Internal   │
                                        │  Network    │
                                        └─────────────┘
```

### Application-Level Controls

**Disable vulnerable feature:**
```java
// Log4j mitigation
System.setProperty("log4j2.formatMsgNoLookups", "true");
```

```bash
# Environment variable
export LOG4J_FORMAT_MSG_NO_LOOKUPS=true
```

**Input validation:**
```python
# Block malicious patterns
import re

BLOCKED_PATTERNS = [
    r'\$\{jndi:',
    r'\$\{lower:',
    r'\$\{upper:',
]

def sanitize_input(user_input):
    for pattern in BLOCKED_PATTERNS:
        if re.search(pattern, user_input, re.IGNORECASE):
            raise ValueError("Blocked input pattern detected")
    return user_input
```

### Detection Controls

**Enhanced logging:**
```yaml
# Log suspicious patterns
- pattern: '\$\{jndi:'
  action: alert
  severity: critical
  notify: security-team
```

**IDS/IPS signatures:**
```
alert http any any -> any any (
  msg:"Potential Log4j Exploitation Attempt";
  content:"${jndi:";
  nocase;
  sid:2021001;
  rev:1;
)
```

## Dependency Replacement

### When to Replace vs Upgrade

Replace when:
- Package is deprecated/abandoned
- Multiple recurring vulnerabilities
- Better maintained alternative exists
- License concerns

### Common Replacements

| Vulnerable | Replacement | Migration Effort |
|------------|-------------|------------------|
| moment.js | date-fns, luxon, dayjs | 2-4 hours |
| request | axios, node-fetch, got | 1-2 hours |
| lodash (full) | lodash-es (tree-shake) | 30 min |
| crypto-js | Web Crypto API | 2-4 hours |
| colors | chalk, picocolors | 30 min |

### Migration Process

1. **Identify all usage:**
   ```bash
   grep -r "require('vulnerable-package')" src/
   grep -r "import.*from 'vulnerable-package'" src/
   ```

2. **Create adapter layer:**
   ```javascript
   // lib/date-adapter.js
   // Abstracts date library for easier future changes
   import { format, parse } from 'date-fns';

   export const formatDate = (date, formatStr) => {
     // Map old format strings to new
     return format(date, convertFormat(formatStr));
   };
   ```

3. **Gradual migration:**
   - Replace in non-critical paths first
   - Add comprehensive tests
   - Roll out incrementally

## Container-Specific Remediation

### Base Image Updates

```dockerfile
# Before: vulnerable base
FROM node:14-alpine

# After: patched base
FROM node:14.21-alpine3.18
```

### Multi-stage builds to reduce attack surface

```dockerfile
# Build stage
FROM node:18 AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

# Production stage - minimal image
FROM gcr.io/distroless/nodejs18-debian11
COPY --from=builder /app/node_modules ./node_modules
COPY . .
CMD ["server.js"]
```

### Runtime mitigation

```yaml
# Kubernetes security context
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
```

## Emergency Response Procedures

### Critical Vulnerability Response

```
T+0h   : Discovery/Alert
T+1h   : Initial triage complete
T+4h   : Impact assessment complete
T+8h   : Remediation plan approved
T+24h  : Production patched or compensating control active
T+48h  : All environments remediated
T+72h  : Post-incident review scheduled
```

### Communication Template

```markdown
## Security Advisory: [CVE-ID]

**Status:** [Investigating/Mitigated/Resolved]
**Severity:** [Critical/High/Medium/Low]
**Affected Systems:** [List]

### Summary
[Brief description of vulnerability and impact]

### Current Status
[What has been done]

### Action Required
[What recipients need to do]

### Timeline
- [Date/Time]: [Action]

### Contact
Security Team: security@example.com
```
