<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# Example: License Audit

Example of using the legal review skill to audit open source licenses in a project.

## Scenario

A development team wants to audit their project's dependencies before a major release to ensure license compliance.

## Conversation

**User**:
```
Please audit the licenses in our project. We're using MIT license and will be distributing the software to customers. Here's our package.json dependencies:

{
  "dependencies": {
    "express": "^4.18.0",
    "axios": "^1.6.0",
    "lodash": "^4.17.21",
    "readline": "^1.3.0",
    "jsonwebtoken": "^9.0.2"
  },
  "devDependencies": {
    "jest": "^29.7.0",
    "eslint": "^8.55.0"
  }
}

Please check:
1. Are all licenses compatible with MIT?
2. Any licenses that require special attention?
3. What attribution is required?
4. Are there any licensing risks?
```

**Assistant (Legal Review Skill)**:
```markdown
# License Audit Report

## Summary
✅ 4 packages approved
❌ 1 package with licensing concern
ℹ️  2 dev dependencies (not distributed)

## Detailed Analysis

### Production Dependencies

#### ✅ express - MIT License
- **Status**: Approved
- **Compatibility**: Fully compatible with MIT
- **Attribution**: Include copyright notice and license text
- **Risk**: Low
- **Action**: None required

#### ✅ axios - MIT License
- **Status**: Approved
- **Compatibility**: Fully compatible with MIT
- **Attribution**: Include copyright notice and license text
- **Risk**: Low
- **Action**: None required

#### ✅ lodash - MIT License
- **Status**: Approved
- **Compatibility**: Fully compatible with MIT
- **Attribution**: Include copyright notice and license text
- **Risk**: Low
- **Action**: None required

#### ❌ readline - GPL-3.0 License
- **Status**: LICENSE VIOLATION
- **Compatibility**: ⚠️ Incompatible with proprietary distribution
- **Issue**: GPL-3.0 requires derivative works to be GPL-3.0
- **Impact**: Cannot distribute proprietary software using GPL code
- **Risk**: HIGH - Legal liability
- **Options**:
  1. **Remove dependency** (Recommended)
  2. Find MIT/Apache-2.0 alternative
  3. Obtain commercial license from author
  4. Release your software as GPL-3.0 (not viable for proprietary)

**Recommended Alternatives**:
- `inquirer` (MIT) - Interactive command-line prompts
- `prompts` (MIT) - Lightweight prompting library
- Node.js built-in `readline` module (no external dependency)

#### ✅ jsonwebtoken - MIT License
- **Status**: Approved
- **Compatibility**: Fully compatible with MIT
- **Attribution**: Include copyright notice and license text
- **Risk**: Low
- **Action**: None required

### Development Dependencies (Not Distributed)

#### ✅ jest - MIT License
- **Status**: Approved (dev dependency)
- **Note**: Dev dependencies not included in distribution
- **Action**: None required for distribution

#### ✅ eslint - MIT License
- **Status**: Approved (dev dependency)
- **Note**: Dev dependencies not included in distribution
- **Action**: None required for distribution

## Attribution Requirements

Create a `NOTICE` file or `ATTRIBUTION.md` with:

```
This software uses the following open source packages:

express (MIT)
Copyright (c) 2009-2014 TJ Holowaychuk
Copyright (c) 2013-2014 Roman Shtylman
Copyright (c) 2014-2015 Douglas Christopher Wilson
https://github.com/expressjs/express

axios (MIT)
Copyright (c) 2014-present Matt Zabriskie
https://github.com/axios/axios

lodash (MIT)
Copyright JS Foundation and other contributors
https://github.com/lodash/lodash

jsonwebtoken (MIT)
Copyright (c) 2015 Auth0, Inc.
https://github.com/auth0/node-jsonwebtoken

For full license texts, see LICENSE file.
```

## Critical Action Required

**IMMEDIATE**: Remove or replace `readline` package (GPL-3.0)

### Recommended Resolution
```bash
# Remove GPL dependency
npm uninstall readline

# Install MIT alternative
npm install inquirer

# Update code to use inquirer instead
```

### Code Migration Example
```javascript
// Before (readline - GPL-3.0)
const readline = require('readline');
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout
});

rl.question('What is your name? ', (answer) => {
  console.log(`Hello, ${answer}!`);
  rl.close();
});

// After (inquirer - MIT)
const inquirer = require('inquirer');

inquirer.prompt([
  {
    type: 'input',
    name: 'name',
    message: 'What is your name?'
  }
]).then((answers) => {
  console.log(`Hello, ${answers.name}!`);
});
```

## Compliance Checklist

- [ ] Remove `readline` dependency
- [ ] Install MIT-licensed alternative
- [ ] Update code to use new library
- [ ] Test functionality
- [ ] Create NOTICE/ATTRIBUTION file
- [ ] Include LICENSE file with MIT license text
- [ ] Document dependency licenses in README
- [ ] Re-scan after changes

## Next Steps

1. **Immediate** (Today):
   - Remove GPL dependency
   - Install alternative

2. **Short-term** (This Week):
   - Update code
   - Create attribution file
   - Verify all tests pass

3. **Before Release**:
   - Final license scan
   - Legal review sign-off
   - Update documentation

## Risk Assessment

**Current Risk**: HIGH - GPL-3.0 violation
**Post-Remediation**: LOW - All MIT compatible

**Business Impact**: Cannot release with current dependencies
**Timeline**: 1-2 days to remediate

## Conclusion

Your project currently has one critical licensing issue (GPL-3.0 dependency) that prevents proprietary distribution. This is easily remediated by switching to a MIT-licensed alternative. All other dependencies are approved and compatible with your MIT license.

After removing the GPL dependency, your project will be in full license compliance and ready for distribution.
```

## Outcome

The team:
1. Removed `readline` package
2. Installed `inquirer` (MIT)
3. Updated code
4. Created NOTICE file
5. Passed final license audit
6. Released software with confidence

## Key Takeaways

- Always audit licenses before adding dependencies
- GPL licenses incompatible with proprietary distribution
- Dev dependencies don't affect distribution licensing
- Attribution requirements are simple to fulfill
- Quick identification saves time and legal risk
