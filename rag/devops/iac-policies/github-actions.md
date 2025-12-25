# GitHub Actions Security Patterns

**Category**: devops/iac-policies
**Description**: GitHub Actions workflow security and organizational policy patterns
**CWE**: CWE-78 (OS Command Injection), CWE-732 (Incorrect Permission Assignment)

---

## Action Security Patterns

### Unpinned Action Reference
**Type**: regex
**Severity**: high
**Pattern**: `(?i)uses:\s*[^@]+@(?:main|master|latest|v[0-9]+)\s*$`
- Actions should be pinned to full commit SHA
- Example: `uses: actions/checkout@v4`
- Remediation: Pin to commit SHA: `uses: actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11`

### Third-Party Action Without SHA
**Type**: regex
**Severity**: high
**Pattern**: `(?i)uses:\s*(?!actions/|github/)[^@\s]+@(?!sha:)[^@\s]+\s*$`
- Third-party actions should be pinned to SHA
- Example: `uses: some-org/action@v1`
- Remediation: Pin to specific commit SHA for security

### Action from Untrusted Source
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)uses:\s*docker://[^/]+/`
- Docker actions from public registries may be untrusted
- Example: `uses: docker://someimage:latest`
- Remediation: Use verified actions or self-hosted images

---

## Permission Patterns

### Write-All Permissions
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)permissions:\s*write-all`
- Workflows should not have write-all permissions
- Example: `permissions: write-all`
- Remediation: Specify minimum required permissions

### Contents Write Permission
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)contents:\s*write`
- Contents write allows modifying repository files
- Example: `contents: write`
- Remediation: Only use when necessary (releases, commits)

### Packages Write Permission
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)packages:\s*write`
- Packages write allows publishing to GHCR
- Example: `packages: write`
- Remediation: Only use in publishing workflows

### ID Token Write Permission
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)id-token:\s*write`
- ID token write enables OIDC authentication
- Example: `id-token: write`
- Remediation: Only use with trusted cloud providers

### Missing Permission Block
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)^on:(?:(?!permissions:).)*jobs:`
- Workflows should explicitly define permissions
- Example: Workflow without permissions block
- Remediation: Add permissions block with minimum required

---

## Injection Patterns

### Script Injection via Title
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)\$\{\{\s*github\.event\.(?:issue|pull_request)\.title\s*\}\}`
- Issue/PR titles can contain malicious content
- Example: `${{ github.event.issue.title }}`
- Remediation: Use environment variable with proper quoting

### Script Injection via Body
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)\$\{\{\s*github\.event\.(?:issue|pull_request|comment)\.body\s*\}\}`
- Issue/PR/comment bodies can contain malicious content
- Example: `${{ github.event.pull_request.body }}`
- Remediation: Use environment variable with proper quoting

### Script Injection via Commit Message
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)\$\{\{\s*github\.event\.(?:head_commit|commits\[\d+\])\.message\s*\}\}`
- Commit messages can contain malicious content
- Example: `${{ github.event.head_commit.message }}`
- Remediation: Use environment variable with proper quoting

### Script Injection via Branch Name
**Type**: regex
**Severity**: high
**Pattern**: `(?i)\$\{\{\s*github\.(?:head_ref|ref_name)\s*\}\}(?:\s|"|'|`)`
- Branch names can contain shell metacharacters
- Example: `${{ github.head_ref }}`
- Remediation: Sanitize or use environment variable

### Unsafe Interpolation in Run
**Type**: regex
**Severity**: high
**Pattern**: `(?i)run:\s*.*\$\{\{\s*github\.event\.`
- Direct interpolation in run commands is unsafe
- Example: `run: echo "${{ github.event.issue.title }}"`
- Remediation: Pass via environment variable

---

## Secret Handling Patterns

### Secret in Run Command
**Type**: regex
**Severity**: high
**Pattern**: `(?i)run:.*\$\{\{\s*secrets\.[^}]+\s*\}\}`
- Secrets should be passed via environment variables
- Example: `run: curl -H "Authorization: ${{ secrets.TOKEN }}"`
- Remediation: Use env block to pass secrets

### Debug Logging with Secrets
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)ACTIONS_STEP_DEBUG:\s*true`
- Debug logging can expose secrets
- Example: `ACTIONS_STEP_DEBUG: true`
- Remediation: Disable debug logging in production

### Secret in Output
**Type**: regex
**Severity**: high
**Pattern**: `(?i)echo\s+["']?.*\$\{\{\s*secrets\.[^}]+\s*\}\}`
- Echoing secrets can expose them in logs
- Example: `echo "Token: ${{ secrets.TOKEN }}"`
- Remediation: Never echo secrets, use masking

---

## Workflow Triggers

### Pull Request Target Trigger
**Type**: regex
**Severity**: high
**Pattern**: `(?i)on:\s*\[?\s*pull_request_target`
- pull_request_target runs in base repo context with secrets
- Example: `on: pull_request_target`
- Remediation: Use pull_request unless secrets needed for forks

### Workflow Dispatch Without Inputs Validation
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)workflow_dispatch:(?:(?!inputs:).)*$`
- Manual triggers should validate inputs
- Example: workflow_dispatch without inputs
- Remediation: Add input validation with required fields

### Schedule Without Protection
**Type**: regex
**Severity**: low
**Pattern**: `(?i)schedule:\s*-\s*cron:`
- Scheduled workflows run on default branch
- Example: `schedule: - cron: "0 0 * * *"`
- Remediation: Ensure scheduled jobs have appropriate permissions

---

## Runner Security

### Self-Hosted Runner Without Labels
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)runs-on:\s*self-hosted\s*$`
- Self-hosted runners should use specific labels
- Example: `runs-on: self-hosted`
- Remediation: Use `runs-on: [self-hosted, linux, x64]`

### Public Repo Self-Hosted Runner
**Type**: regex
**Severity**: high
**Pattern**: `(?i)runs-on:\s*\[?\s*self-hosted`
- Self-hosted runners in public repos are risky
- Example: Self-hosted runner in public repository
- Remediation: Use GitHub-hosted runners for public repos

---

## Organizational Policies

### Missing Timeout
**Type**: regex
**Severity**: low
**Pattern**: `(?i)jobs:[\s\S]*?runs-on:(?:(?!timeout-minutes:).)*steps:`
- Jobs should have timeout to prevent stuck workflows
- Example: Job without timeout-minutes
- Remediation: Add `timeout-minutes: 30` (or appropriate)

### Missing Concurrency
**Type**: regex
**Severity**: low
**Pattern**: `(?i)^on:(?:(?!concurrency:).)*jobs:`
- Workflows should use concurrency to prevent duplicates
- Example: Workflow without concurrency block
- Remediation: Add concurrency group with cancel-in-progress

### Using Deprecated Node Version
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)node-version:\s*['"]?(?:12|14|16)['"]?`
- Node.js 12/14/16 are end-of-life
- Example: `node-version: "16"`
- Remediation: Update to Node.js 18 or 20

---

## Detection Confidence

**Regex Detection**: 90%
**Security Pattern Detection**: 85%

---

## References

- GitHub Actions Security Hardening
- OWASP CI/CD Security Guide
- StepSecurity hardening recommendations
