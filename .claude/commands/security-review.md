# Security Code Review

Perform a comprehensive security code review on the specified target. This command uses AI-powered static analysis to identify high-confidence security vulnerabilities.

## Usage

```
/security-review [target]
```

Where `[target]` can be:
- A GitHub repository URL: `https://github.com/owner/repo`
- A local directory path: `./src` or `/path/to/code`
- A specific file: `./app.py`

## What This Reviews

The security review checks for these vulnerability categories:

### Injection Vulnerabilities
- SQL Injection (CWE-89)
- Command Injection (CWE-78)
- NoSQL Injection (CWE-943)
- LDAP/XPath Injection (CWE-90, CWE-91)
- Template/Expression Language Injection (CWE-1336, CWE-917)

### Authentication & Authorization
- Broken Access Control (CWE-284)
- Missing Authentication (CWE-306)
- IDOR - Insecure Direct Object References (CWE-639)
- Session Fixation (CWE-384)
- JWT Vulnerabilities (CWE-347)

### Cryptographic Issues
- Weak Algorithms (MD5, SHA1, DES, RC4) (CWE-327)
- Hardcoded Keys/Secrets (CWE-321, CWE-798)
- Insecure Random Number Generation (CWE-330)

### Code Execution
- Insecure Deserialization (CWE-502)
- eval/exec with User Input (CWE-95)
- Prototype Pollution (CWE-1321)

### Input Validation
- Cross-Site Scripting (XSS) (CWE-79)
- Path Traversal (CWE-22)
- Server-Side Request Forgery (SSRF) (CWE-918)
- XML External Entity (XXE) (CWE-611)
- Open Redirects (CWE-601)

### Secrets & Credentials
- Hardcoded API Keys (AWS, GCP, Azure, etc.)
- Hardcoded Passwords and Tokens
- Private Keys in Source Code

### Configuration
- Debug Mode Enabled (CWE-489)
- CORS Misconfiguration (CWE-942)
- Insecure Cookie Settings (CWE-614)
- Missing Security Headers

## Analysis Methodology

1. **Context Understanding** - Identify code purpose, data flows, and trust boundaries
2. **Data Flow Tracing** - Follow user input from entry points to sensitive sinks
3. **Pattern Detection** - Match against known vulnerability patterns with CWE classification
4. **Confidence Filtering** - Only report findings with ≥80% confidence

## Output

The review produces:
- **Terminal Output**: Real-time findings as files are analyzed
- **Markdown Report**: Detailed report with executive summary, severity breakdown, and remediation guidance
- **JSON/SARIF**: Machine-readable formats for CI/CD integration

## Example

```bash
# Review a GitHub repository
/security-review https://github.com/example/webapp

# Review local code
/security-review ./src

# Review a specific file
/security-review ./api/auth.py
```

## Configuration Options

Run the underlying tool directly for advanced options:

```bash
./utils/code-security/code-security-analyser.sh \
  --repo https://github.com/owner/repo \
  --output ./reports \
  --format markdown \
  --min-severity medium \
  --fail-on high \
  --supply-chain
```

---

When the user runs this command, execute the security analysis:

1. If a GitHub URL is provided, clone and analyze the repository
2. If a local path is provided, analyze that directory/file
3. Display findings in the terminal as they are discovered
4. Generate a comprehensive security report

Run: `./utils/code-security/code-security-analyser.sh --repo $ARGUMENTS` for repos
Or: `./utils/code-security/code-security-analyser.sh --local $ARGUMENTS` for local paths
