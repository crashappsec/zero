<!--
Copyright (c) 2025 Crash Override Inc. - https://crashoverride.com

SPDX-License-Identifier: GPL-3.0
-->

# RAG Technology Catalog Implementation Plan

**Date**: 2025-11-24
**Status**: Implementation Phase
**Goal**: Expand from 5 → 30+ technologies with standardized RAG patterns

## Current State

**Existing RAG Patterns** (5 technologies):
- ✅ Stripe (business-tools/payment) - 4 files
- ✅ AWS SDK (cloud-providers/aws) - 4 files
- ✅ OpenSSL (cryptographic-libraries/tls) - 3 files
- ✅ Terraform (developer-tools/infrastructure) - 3 files
- ✅ Docker (developer-tools/containers) - 3 files

**Total**: 17 pattern files

## Pattern File Standard

Each technology should have these pattern files:

### Required Files
1. **package-patterns.json** - Package manager patterns (npm, PyPI, etc.)
2. **import-patterns.json** - Import/require statement patterns
3. **api-patterns.json** - API endpoint patterns
4. **env-patterns.json** - Environment variable patterns
5. **config-patterns.json** - Configuration file patterns
6. **versions.json** - Version history, EOL dates, breaking changes

### File Format: JSON for Easy Parsing

**Example: package-patterns.json**
```json
{
  "technology": "React",
  "category": "web-frameworks/frontend",
  "confidence": 95,
  "patterns": [
    {
      "ecosystem": "npm",
      "names": ["react", "react-dom"],
      "description": "React library packages"
    },
    {
      "ecosystem": "yarn",
      "names": ["react", "react-dom"],
      "description": "React library packages"
    }
  ]
}
```

**Example: import-patterns.json**
```json
{
  "technology": "React",
  "category": "web-frameworks/frontend",
  "confidence": 85,
  "patterns": [
    {
      "language": "javascript",
      "file_extensions": [".js", ".jsx"],
      "patterns": [
        "import\\s+.*\\s+from\\s+['\"]react['\"]",
        "import\\s+React\\s+from\\s+['\"]react['\"]",
        "require\\(['\"]react['\"]\\)"
      ]
    },
    {
      "language": "typescript",
      "file_extensions": [".ts", ".tsx"],
      "patterns": [
        "import\\s+.*\\s+from\\s+['\"]react['\"]",
        "import\\s+type\\s+.*\\s+from\\s+['\"]react['\"]"
      ]
    }
  ]
}
```

## Technology Priority Matrix

### Tier 1: Critical (Implement First) - 15 technologies
Most commonly used, high detection value

**Web Frameworks (5)**:
1. React - Frontend framework (most popular)
2. Vue.js - Frontend framework
3. Angular - Frontend framework
4. Express - Backend framework (Node.js)
5. Django - Backend framework (Python)

**Cloud Providers (3)**:
6. AWS SDK - Already have baseline, expand
7. Google Cloud SDK - GCP services
8. Azure SDK - Microsoft cloud

**Databases (4)**:
9. PostgreSQL - Relational database
10. MySQL - Relational database
11. MongoDB - NoSQL database
12. Redis - Key-value store (already detected, add patterns)

**Developer Tools (3)**:
13. Kubernetes - Container orchestration
14. GitHub Actions - CI/CD (already detected, add patterns)
15. GitLab CI - CI/CD

### Tier 2: High Value (Next Phase) - 10 technologies
Commonly used, important for coverage

**Business Tools (4)**:
16. Twilio - Communication API
17. SendGrid - Email service
18. Salesforce - CRM
19. Datadog - Monitoring

**Web Frameworks (3)**:
20. Flask - Python backend
21. FastAPI - Python backend
22. Next.js - React framework

**Developer Tools (3)**:
23. Ansible - Infrastructure automation
24. Jenkins - CI/CD
25. Prometheus - Monitoring

### Tier 3: Nice to Have (Future) - 10 technologies
Less common but still valuable

**Message Queues (2)**:
26. RabbitMQ - Message broker
27. Apache Kafka - Event streaming

**Testing (2)**:
28. Jest - JavaScript testing
29. pytest - Python testing

**Build Tools (2)**:
30. Webpack - JavaScript bundler
31. Maven - Java build tool

**Cryptographic (2)**:
32. libsodium - Modern crypto library
33. bcrypt - Password hashing

**Other (2)**:
34. Elasticsearch - Search engine
35. GraphQL - API query language

## Implementation Approach

### Phase 1: Tier 1 Technologies (15 techs, ~90 files)
**Estimated Time**: 3-4 days
**Files per Tech**: ~6 files (package, import, api, env, config, versions)

**Order of Implementation**:
1. **Day 1**: Web frameworks (React, Vue, Angular, Express, Django) - 5 × 6 = 30 files
2. **Day 2**: Databases (PostgreSQL, MySQL, MongoDB, Redis) - 4 × 6 = 24 files
3. **Day 3**: Cloud providers (AWS expand, GCP, Azure) - 3 × 6 = 18 files
4. **Day 4**: Developer tools (Kubernetes, GitHub Actions, GitLab CI) - 3 × 6 = 18 files

### Phase 2: Tier 2 Technologies (10 techs, ~60 files)
**Estimated Time**: 2-3 days

### Phase 3: Tier 3 Technologies (10 techs, ~60 files)
**Estimated Time**: 2-3 days

### Total
- **35 technologies**
- **210 pattern files**
- **7-10 days** for complete catalog

## RAG Pattern Generation Strategy

### Manual Creation (High Priority)
For Tier 1 technologies, manually create patterns to ensure accuracy:
- Research official documentation
- Test against real repositories
- Validate detection accuracy

### Semi-Automated (Future)
Use Claude AI to assist pattern extraction:
```bash
# Future tool
./rag-generator.sh extract \
  --technology react \
  --docs-url https://react.dev/reference \
  --output rag/technology-identification/web-frameworks/frontend/react/
```

### Automated Updates (Future)
Scheduled updates for version information:
```bash
# Future tool
./rag-generator.sh update-versions \
  --technology react \
  --source npm
```

## Directory Structure

```
rag/technology-identification/
├── web-frameworks/
│   ├── frontend/
│   │   ├── react/
│   │   │   ├── package-patterns.json
│   │   │   ├── import-patterns.json
│   │   │   ├── api-patterns.json
│   │   │   ├── env-patterns.json
│   │   │   ├── config-patterns.json
│   │   │   └── versions.json
│   │   ├── vue/
│   │   │   └── [same 6 files]
│   │   └── angular/
│   │       └── [same 6 files]
│   └── backend/
│       ├── express/
│       ├── django/
│       ├── flask/
│       └── fastapi/
├── databases/
│   ├── relational/
│   │   ├── postgresql/
│   │   ├── mysql/
│   │   └── sqlite/
│   ├── nosql/
│   │   ├── mongodb/
│   │   └── cassandra/
│   └── keyvalue/
│       ├── redis/
│       └── memcached/
├── cloud-providers/
│   ├── aws/
│   ├── gcp/
│   └── azure/
├── developer-tools/
│   ├── containers/
│   │   ├── docker/
│   │   └── kubernetes/
│   ├── cicd/
│   │   ├── github-actions/
│   │   ├── gitlab-ci/
│   │   └── jenkins/
│   └── infrastructure/
│       ├── terraform/
│       └── ansible/
└── business-tools/
    ├── payment/
    │   ├── stripe/
    │   ├── paypal/
    │   └── square/
    ├── communication/
    │   ├── twilio/
    │   └── sendgrid/
    └── crm/
        └── salesforce/
```

## Pattern Quality Guidelines

### Package Patterns
- Include all ecosystems (npm, PyPI, Maven, RubyGems, etc.)
- List official package names and common aliases
- Document scoped packages (e.g., @aws-sdk/*, @angular/*)

### Import Patterns
- Cover all programming languages
- Include both ES6 imports and CommonJS requires
- Account for namespace imports vs. named imports
- Regex patterns should be precise but not too restrictive

### API Patterns
- Official API endpoints only (not third-party wrappers)
- Include authentication patterns where relevant
- Document API versions if endpoint structure changes

### Environment Variables
- Official variable names from documentation
- Common variations and legacy names
- Configuration patterns (e.g., DATABASE_URL patterns)

### Config Patterns
- File name patterns (exact and wildcards)
- Directory conventions
- Content patterns (YAML/JSON structures)

### Versions
- Release history with dates
- EOL (end-of-life) dates
- Breaking changes between major versions
- Current stable/LTS versions
- Security advisories

## Testing Strategy

### Per-Technology Testing
After creating patterns for each technology:
1. Create test fixture with that technology
2. Run analyzer and verify detection
3. Check confidence scores are appropriate
4. Validate all detection methods work

### Integration Testing
After completing tier:
1. Create multi-technology test repository
2. Verify all technologies detected correctly
3. Check for false positives/negatives
4. Validate composite confidence scoring

### Regression Testing
Before considering phase complete:
1. Run full test suite
2. Test against real-world repositories
3. Compare against manual audits
4. Measure detection accuracy (target: 95%+)

## Success Criteria

### Phase 1 Complete When:
- ✅ 15 Tier 1 technologies have complete RAG patterns
- ✅ All 90 pattern files created and validated
- ✅ Dynamic pattern loader implemented
- ✅ Detection functions use RAG patterns (no hardcoded)
- ✅ All existing tests pass
- ✅ New tests for dynamic loading pass
- ✅ Detection accuracy >90% on test repositories

### Full Catalog Complete When:
- ✅ 35+ technologies with complete patterns
- ✅ 210+ pattern files
- ✅ Detection accuracy >95% on real repositories
- ✅ No hardcoded patterns remain in code
- ✅ Documentation complete

## Maintenance Plan

### Weekly Updates
- Check for new versions of popular technologies
- Update EOL dates
- Add new package patterns if ecosystem changes

### Monthly Reviews
- Analyze false positives/negatives from production usage
- Refine patterns based on real-world feedback
- Add new technologies based on usage trends

### Quarterly Audits
- Comprehensive pattern validation
- Technology popularity re-assessment
- Documentation updates

## Implementation Checklist

### Immediate (Today)
- [x] Create RAG catalog plan
- [ ] Design pattern loader library
- [ ] Create first 5 Tier 1 technology patterns (React, Vue, Angular, Express, Django)
- [ ] Implement basic pattern loader
- [ ] Test with new patterns

### This Week
- [ ] Complete all 15 Tier 1 technology patterns
- [ ] Full pattern loader implementation
- [ ] Update all detection functions to use patterns
- [ ] Comprehensive testing
- [ ] Documentation

### Next Week
- [ ] Tier 2 technologies (10 techs)
- [ ] Pattern validation and refinement
- [ ] Performance optimization

## Resources Needed

### Documentation Sources
- Official project documentation (react.dev, vuejs.org, etc.)
- Package registry APIs (npm, PyPI, Maven Central)
- GitHub repositories for real-world examples
- Stack Overflow for common usage patterns

### Tools
- Claude AI for pattern extraction assistance
- jq for JSON manipulation
- Regex testing tools
- Package registry CLIs

---

**Status**: Ready to implement Phase 1
**Next Action**: Create React, Vue, Angular, Express, Django patterns
