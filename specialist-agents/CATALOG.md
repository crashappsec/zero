# Specialist Agent Catalog

A complete catalog of all available specialist agents, their purposes, and invocation patterns.

## Quick Reference

| Agent | Category | Purpose | Guardrail Level |
|-------|----------|---------|-----------------|
| vulnerability-analyst | Security | CVE analysis, exploit research | Level 2 (Web) |
| threat-modeler | Security | Attack surface, STRIDE analysis | Level 2 (Web) |
| code-auditor | Security | SAST-style code review | Level 1 (Read-only) |
| secrets-scanner | Security | Credential detection | Level 1 (Read-only) |
| container-security | Security | Docker/K8s security | Level 2 (Web) |
| dependency-investigator | Supply Chain | Package health, alternatives | Level 3 (Commands) |
| license-auditor | Supply Chain | License compliance | Level 2 (Web) |
| code-reviewer | Engineering | PR review, best practices | Level 1 (Read-only) |
| refactoring-advisor | Engineering | Code improvements, tech debt | Level 1 (Read-only) |
| test-strategist | Engineering | Test coverage, strategies | Level 1 (Read-only) |
| performance-analyst | Engineering | Bottlenecks, optimization | Level 2 (Web) |
| infrastructure-auditor | DevOps | IaC security, cost | Level 2 (Web) |
| ci-cd-optimizer | DevOps | Pipeline optimization | Level 2 (Web) |
| remediation-planner | Planning | Fix prioritization, plans | Level 3 (Commands) |

---

## Security Agents

### vulnerability-analyst
**Purpose**: Deep CVE analysis with exploit context and reachability assessment.

**Best For**:
- Analyzing vulnerability scan results
- Researching specific CVEs
- Assessing exploit risk
- Prioritizing security fixes

**Input**: CVE data, scan results, codebase access
**Output**: Risk assessment, exploit analysis, prioritized recommendations

**Invocation**:
```
Task: security/vulnerability-analyst
Prompt: "Analyze the following vulnerability scan results and assess exploitability for each CVE found..."
```

---

### threat-modeler
**Purpose**: Systematic threat identification using STRIDE methodology.

**Best For**:
- Security architecture reviews
- Identifying attack surfaces
- Building threat models
- Risk assessment for new features

**Input**: Architecture docs, code, system description
**Output**: Threat catalog, attack trees, mitigation recommendations

**Invocation**:
```
Task: security/threat-modeler
Prompt: "Create a threat model for the authentication system, focusing on..."
```

---

### code-auditor
**Purpose**: SAST-style security code review with CWE mapping.

**Best For**:
- Security-focused code review
- Finding injection vulnerabilities
- Identifying auth/authz issues
- Pre-release security checks

**Input**: Source code
**Output**: Security findings with remediation guidance

**Invocation**:
```
Task: security/code-auditor
Prompt: "Perform a security audit of the API handlers in src/api/, focusing on injection vulnerabilities..."
```

---

### secrets-scanner
**Purpose**: Detect exposed credentials and API keys in source code.

**Best For**:
- Pre-commit secret scanning
- Repository audits
- Incident response (leaked secrets)
- Compliance verification

**Input**: Codebase
**Output**: Masked secret findings, rotation guidance

**Invocation**:
```
Task: security/secrets-scanner
Prompt: "Scan the repository for exposed secrets, API keys, and credentials..."
```

---

### container-security
**Purpose**: Analyze Docker and Kubernetes configurations for security issues.

**Best For**:
- Dockerfile review
- K8s manifest auditing
- Image security assessment
- CIS benchmark compliance

**Input**: Dockerfiles, compose files, K8s manifests
**Output**: Security findings, hardening recommendations

**Invocation**:
```
Task: security/container-security
Prompt: "Audit the Kubernetes deployment manifests for security misconfigurations..."
```

---

## Supply Chain Agents

### dependency-investigator
**Purpose**: Package health analysis and alternative recommendations.

**Best For**:
- Dependency health checks
- Finding abandoned packages
- Typosquatting detection
- Alternative research

**Input**: Package manifests, lock files
**Output**: Health assessments, alternative recommendations

**Invocation**:
```
Task: supply-chain/dependency-investigator
Prompt: "Analyze the health of all dependencies in package.json, identifying any abandoned or deprecated packages..."
```

---

### license-auditor
**Purpose**: License compliance and compatibility analysis.

**Best For**:
- Open source compliance
- License compatibility checks
- SBOM validation
- Pre-release compliance

**Input**: Dependencies, SBOM
**Output**: License inventory, compatibility issues, disclosure requirements

**Invocation**:
```
Task: supply-chain/license-auditor
Prompt: "Audit all dependencies for license compliance, identifying any conflicts with our MIT license..."
```

---

## Engineering Agents

### code-reviewer
**Purpose**: Comprehensive, constructive code review.

**Best For**:
- PR reviews
- Code quality assessment
- Best practice verification
- Knowledge sharing

**Input**: Code changes, context
**Output**: Review comments, suggestions, patterns observed

**Invocation**:
```
Task: engineering/code-reviewer
Prompt: "Review the changes in the following files, focusing on correctness, maintainability, and best practices..."
```

---

### refactoring-advisor
**Purpose**: Identify code improvement and technical debt opportunities.

**Best For**:
- Tech debt assessment
- Code smell detection
- Refactoring planning
- Architecture improvements

**Input**: Codebase
**Output**: Refactoring opportunities, prioritized improvements

**Invocation**:
```
Task: engineering/refactoring-advisor
Prompt: "Analyze the services directory for refactoring opportunities, focusing on code smells and complexity..."
```

---

### test-strategist
**Purpose**: Test coverage analysis and testing strategy recommendations.

**Best For**:
- Coverage gap analysis
- Test strategy planning
- Test quality assessment
- Testing roadmaps

**Input**: Source and test files
**Output**: Coverage gaps, test recommendations, quality issues

**Invocation**:
```
Task: engineering/test-strategist
Prompt: "Analyze test coverage for the order processing module and recommend additional tests..."
```

---

### performance-analyst
**Purpose**: Identify performance bottlenecks and optimization opportunities.

**Best For**:
- Performance audits
- Algorithm analysis
- N+1 query detection
- Memory leak identification

**Input**: Source code
**Output**: Performance issues, optimization recommendations

**Invocation**:
```
Task: engineering/performance-analyst
Prompt: "Analyze the search functionality for performance issues, focusing on database queries and algorithmic complexity..."
```

---

## DevOps Agents

### infrastructure-auditor
**Purpose**: Review IaC for security, cost, and reliability.

**Best For**:
- Terraform/CloudFormation review
- Security misconfiguration detection
- Cost optimization
- Compliance checking

**Input**: IaC files
**Output**: Security findings, cost recommendations, compliance status

**Invocation**:
```
Task: devops/infrastructure-auditor
Prompt: "Audit the Terraform configuration for security misconfigurations and cost optimization opportunities..."
```

---

### ci-cd-optimizer
**Purpose**: Analyze and improve CI/CD pipelines.

**Best For**:
- Pipeline optimization
- Build time reduction
- CI/CD security review
- Reliability improvements

**Input**: Pipeline configuration files
**Output**: Performance improvements, security findings, recommendations

**Invocation**:
```
Task: devops/ci-cd-optimizer
Prompt: "Analyze the GitHub Actions workflows for optimization opportunities and security issues..."
```

---

## Planning Agents

### remediation-planner
**Purpose**: Create prioritized fix plans from security and engineering findings.

**Best For**:
- Prioritizing security fixes
- Sprint planning for tech debt
- Creating fix roadmaps
- PR planning

**Input**: Findings from other agents
**Output**: Prioritized fix plans, PR suggestions, effort estimates

**Invocation**:
```
Task: planning/remediation-planner
Prompt: "Create a remediation plan for the following vulnerability findings, prioritized by risk and effort..."
```

---

## Chaining Agents

### Security Assessment Chain
```
1. security/secrets-scanner → Find exposed secrets
2. security/code-auditor → Find code vulnerabilities
3. security/container-security → Check container configs
4. planning/remediation-planner → Create fix plan
```

### Full Code Review Chain
```
1. engineering/code-reviewer → Quality review
2. security/code-auditor → Security review
3. engineering/test-strategist → Test coverage
4. engineering/performance-analyst → Performance check
```

### Supply Chain Assessment
```
1. supply-chain/dependency-investigator → Health check
2. security/vulnerability-analyst → CVE analysis
3. supply-chain/license-auditor → License compliance
4. planning/remediation-planner → Update plan
```

### Infrastructure Review
```
1. devops/infrastructure-auditor → IaC review
2. devops/ci-cd-optimizer → Pipeline review
3. security/secrets-scanner → Secret detection
4. planning/remediation-planner → Fix plan
```

---

## Guardrail Levels

### Level 1: Read-Only
- Tools: Read, Grep, Glob
- Use: Security audits where no external access needed
- Agents: code-auditor, secrets-scanner, code-reviewer, refactoring-advisor, test-strategist

### Level 2: Web Access
- Tools: Level 1 + WebFetch, WebSearch
- Use: Research, threat intelligence, best practices lookup
- Agents: vulnerability-analyst, threat-modeler, container-security, license-auditor, performance-analyst, infrastructure-auditor, ci-cd-optimizer

### Level 3: Limited Commands
- Tools: Level 2 + Bash (allowlisted commands only)
- Use: Package queries, version checks
- Agents: dependency-investigator, remediation-planner
