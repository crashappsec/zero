# License Compliance Guide

## Overview

Open source license compliance ensures your use of third-party software meets the legal obligations of those licenses. Non-compliance can result in:
- Injunctions requiring removal of infringing code
- Requirement to release proprietary source code
- Damages and legal fees
- Reputational harm

## License Categories

### Permissive Licenses

**Examples:** MIT, BSD, Apache 2.0, ISC

**Key characteristics:**
- Allow proprietary use
- Minimal obligations (typically attribution)
- Can combine with most other licenses

**Typical obligations:**
1. Include license text in distribution
2. Include copyright notice
3. (Apache) State changes made
4. (Apache) Include NOTICE file

### Weak Copyleft

**Examples:** LGPL, MPL

**Key characteristics:**
- Changes to the library must be shared
- Proprietary code can link to the library
- File-level (MPL) or library-level (LGPL) scope

**LGPL obligations:**
1. Provide library source code (modified)
2. Allow replacement of the library
3. Permit reverse engineering for debugging

### Strong Copyleft

**Examples:** GPL v2, GPL v3, AGPL

**Key characteristics:**
- Derivative works must use same license
- "Viral" effect on combined works
- (AGPL) Network use triggers obligations

**GPL obligations:**
1. Provide complete source code
2. License derivative work under GPL
3. Include license and copyright
4. State changes made
5. (GPLv3) Provide installation information

## Compliance by Distribution Type

### SaaS/Cloud Service
| License | Triggered? | Notes |
|---------|------------|-------|
| MIT, BSD, Apache | No | Attribution in docs optional |
| LGPL | No | Not distributed |
| GPL | No | Not distributed |
| AGPL | **Yes** | Network interaction triggers |

### Binary Distribution
| License | Obligations |
|---------|-------------|
| MIT, BSD | Include license/copyright |
| Apache 2.0 | Include license, NOTICE, state changes |
| LGPL | Provide library source, allow replacement |
| GPL | Provide all source, license under GPL |

### Source Distribution
| License | Obligations |
|---------|-------------|
| Permissive | Include license files |
| LGPL/GPL | Ensure license propagation |

## Compliance Process

### 1. Inventory Dependencies

```bash
# Node.js
npm ls --all

# Python
pip freeze

# Go
go list -m all
```

### 2. Identify Licenses

Use tools like:
- FOSSA
- Snyk
- license-checker (npm)
- pip-licenses (Python)
- go-licenses (Go)

### 3. Assess Compatibility

Check that:
- Licenses are compatible with each other
- Licenses are compatible with your distribution model
- No license conflicts exist

### 4. Meet Obligations

**For permissive licenses:**
- Include LICENSES directory or file
- List attributions in documentation

**For copyleft:**
- Consult legal counsel
- Consider architectural isolation
- Implement compliance procedures

### 5. Document and Monitor

- Maintain license inventory
- Review new dependencies
- Update compliance artifacts

## Common Scenarios

### Scenario 1: MIT dependency in proprietary SaaS

**Risk:** Low

**Action:**
- Include license in source comments or LICENSES file
- Include in third-party notices (optional but good practice)

### Scenario 2: GPL dependency in proprietary product

**Risk:** High

**Options:**
1. Release product under GPL (usually not viable)
2. Replace with permissively licensed alternative
3. Isolate as separate process (may satisfy linking concerns)
4. Obtain commercial license (dual-licensed projects)

### Scenario 3: AGPL dependency in SaaS

**Risk:** Critical

**Options:**
1. Release entire service source under AGPL
2. Replace with different library
3. Obtain commercial license
4. Run as separate service with API barrier (consult counsel)

### Scenario 4: LGPL in mobile app

**Risk:** Medium

**Requirements:**
- Use dynamic linking (shared library)
- Allow user to replace the library
- Provide library source code

**Challenge:** iOS doesn't allow dynamic linking

**Options:**
1. Replace with permissive alternative
2. Obtain commercial license
3. Consult counsel on static linking implications

## License Compatibility

### Compatible Combinations

```
MIT + Apache 2.0 → Apache 2.0 (more restrictive)
MIT + GPL-3.0 → GPL-3.0 (copyleft dominates)
Apache 2.0 + GPL-3.0 → GPL-3.0 (one-way compatible)
```

### Incompatible Combinations

```
GPL-2.0 + Apache 2.0 → Conflict (Apache patent terms)
GPL-2.0 + GPL-3.0 → Conflict (unless "or later")
AGPL + proprietary → Conflict (no proprietary allowed)
```

## Best Practices

### Policy

1. **Establish approved license list**
   - Green: MIT, BSD, Apache 2.0, ISC
   - Yellow: LGPL, MPL (review required)
   - Red: GPL, AGPL (legal review required)

2. **Review process**
   - New dependencies require license check
   - Copyleft licenses require legal approval
   - Document decisions

### Technical

1. **Automate scanning**
   - CI/CD license checking
   - Block unapproved licenses

2. **Maintain compliance artifacts**
   - LICENSES directory
   - Third-party notices
   - SBOM (Software Bill of Materials)

### Documentation

1. **Attribution file**
   - List all third-party software
   - Include license text
   - Credit authors

2. **Internal records**
   - License decisions
   - Approval documentation
   - Compliance procedures

## Disclaimer

This guide provides general information about software licensing. It is not legal advice. License interpretation and compliance requirements can be complex and fact-specific. Consult a qualified attorney for legal advice regarding your specific situation.
