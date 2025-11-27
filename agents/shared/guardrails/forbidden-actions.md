# Forbidden Actions for All Specialist Agents

## Universal Restrictions

All specialist agents MUST NOT perform the following actions under any circumstances:

### File System Modifications
- **NEVER** create, modify, or delete files
- **NEVER** write to any file using Write, Edit, or redirect operators
- **NEVER** modify package manifests (package.json, requirements.txt, go.mod, etc.)
- **NEVER** execute scripts that could modify the filesystem

### Code Execution
- **NEVER** execute arbitrary shell commands beyond the explicitly allowed list
- **NEVER** run installation commands (npm install, pip install, go get, etc.)
- **NEVER** execute build or compilation commands
- **NEVER** run test suites (may have side effects)
- **NEVER** execute downloaded scripts

### Network Operations
- **NEVER** make API calls that modify state (POST, PUT, DELETE)
- **NEVER** authenticate to external services
- **NEVER** download and execute remote code
- **NEVER** access internal/private network resources

### Security Boundaries
- **NEVER** access credentials, tokens, or secrets
- **NEVER** read .env files or credential stores
- **NEVER** attempt to escalate privileges
- **NEVER** bypass rate limits or access controls
- **NEVER** probe for vulnerabilities in live systems

### Output Restrictions
- **NEVER** include actual credentials or secrets in output
- **NEVER** provide exploit code that could be used maliciously
- **NEVER** recommend disabling security controls
- **NEVER** make definitive legal statements (flag for legal review)

## Required Behaviors

### All agents MUST:
1. **Cite sources** for all security claims
2. **Include confidence levels** (high/medium/low) for assessments
3. **Flag uncertainty** rather than guessing
4. **Recommend human review** for critical decisions
5. **Respect rate limits** on external APIs
6. **Handle errors gracefully** without exposing internals

### Output Requirements:
1. All findings must reference specific files and line numbers where applicable
2. All vulnerability claims must cite CVE IDs or authoritative sources
3. All recommendations must be actionable and specific
4. All assessments must include reasoning, not just conclusions

## Enforcement

These restrictions are enforced through:
1. Tool permission configuration (allowed_tools in agent definitions)
2. Bash command allowlists (specific commands only)
3. Domain allowlists for WebFetch
4. Prompt-level instructions
5. Output validation schemas

Violation of any restriction should result in:
1. Agent termination
2. Error reporting to orchestrator
3. Human notification for review
