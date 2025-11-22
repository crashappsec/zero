<!--
Copyright (c) 2024 Gibson Powers Contributors

SPDX-License-Identifier: GPL-3.0
-->

# Prompt: SBOM License Compliance Analysis

## Purpose
Analyze software licenses in an SBOM to identify compliance issues, conflicts, and obligations.

## When to Use
- Pre-deployment compliance checks
- Open source audit requirements
- Commercial software release preparation
- M&A due diligence
- Policy compliance verification

## Prompt

```
Please perform a comprehensive license compliance analysis of this SBOM:

1. Extract all license declarations from components
2. Categorize licenses by type:
   - Permissive (MIT, Apache, BSD, etc.)
   - Weak copyleft (LGPL, MPL, etc.)
   - Strong copyleft (GPL, AGPL, etc.)
   - Proprietary/Commercial
   - Unknown/Missing
3. Identify potential license conflicts
4. Flag any GPL/AGPL components that require source disclosure
5. Validate SPDX license expressions
6. Provide a compliance matrix
7. Recommend actions for any compliance issues

[Paste SBOM content here]
```

## Expected Output
- License inventory table
- License type breakdown (counts and percentages)
- Conflict identification
- Copyleft obligation summary
- Missing license warnings
- Compliance recommendations

## Variations

### GPL/Copyleft Focus
```
Review this SBOM specifically for GPL and copyleft licenses.
Identify any components that would require source code disclosure.
Assess compatibility with our proprietary application.

[Paste SBOM]
```

### Commercial Use Clearance
```
We plan to use this software in a commercial SaaS product.
Identify any licenses that would prohibit or restrict commercial use.

[Paste SBOM]
```

### License Conflict Detection
```
Check this SBOM for license conflicts between components.
Identify incompatible license combinations (e.g., GPL + proprietary).

[Paste SBOM]
```

## Examples

### Example Usage
```
Please perform a comprehensive license compliance analysis...

[SBOM with various licenses]
```

### Example Output Structure
```markdown
# License Compliance Report

## License Summary
- Total Components: 45
- Licensed: 43
- Unknown/Missing: 2

## License Breakdown
| License Type | Count | Percentage |
|--------------|-------|------------|
| MIT | 25 | 56% |
| Apache-2.0 | 10 | 22% |
| BSD-3-Clause | 5 | 11% |
| GPL-3.0 | 2 | 4% |
| Proprietary | 1 | 2% |
| Unknown | 2 | 4% |

## Issues Detected

### ⚠️ Copyleft Licenses Requiring Source Disclosure
1. component-x v1.2.3 (GPL-3.0)
2. component-y v2.0.1 (AGPL-3.0)

### ❌ License Conflicts
- GPL-3.0 component linked with proprietary component

### ⚠️ Missing Licenses
1. internal-util v1.0.0
2. legacy-module v0.9.2

## Recommendations
[Specific actions...]
```

## Tips
- Specify your use case (internal, commercial, SaaS, etc.)
- Mention if you have specific license policies
- Include information about static vs. dynamic linking if relevant
- Request comparison with approved license lists if you have one
