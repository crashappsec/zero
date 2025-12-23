# Dockerfile Security Patterns

**Category**: devops-security/dockerfile
**Description**: Security and best practice patterns for Dockerfiles
**CWE**: CWE-250 (Execution with Unnecessary Privileges), CWE-798 (Hard-coded Credentials)

---

## Security Patterns

### Using :latest Tag
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)^FROM\s+[^:]+:latest\s*$`
- Using :latest tag makes builds non-reproducible and may pull insecure versions
- Example: `FROM node:latest`
- Remediation: Use specific version tags like `FROM node:18.17.0-alpine`

### Running as Root
**Type**: regex
**Severity**: high
**Pattern**: `(?i)^USER\s+root\s*$`
- Explicitly running container as root user is a security risk
- Example: `USER root`
- Remediation: Use a non-root user like `USER nonroot` or `USER 1000`

### Hardcoded Secret in Dockerfile
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)(?:PASSWORD|SECRET|API_KEY|TOKEN|PRIVATE_KEY)\s*=\s*['\"]\S+['\"]`
- Secrets should never be hardcoded in Dockerfiles
- Example: `ENV DB_PASSWORD="secret123"`
- Remediation: Use build-time secrets or runtime environment variables

### Secret in ENV Instruction
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)^ENV\s+(?:[^\s=]+\s+)?(?:PASSWORD|SECRET|API_KEY|TOKEN|PRIVATE_KEY|AWS_SECRET|GITHUB_TOKEN)`
- ENV instructions persist secrets in image layers
- Example: `ENV AWS_SECRET_ACCESS_KEY=abc123`
- Remediation: Use --mount=type=secret for sensitive data

### Using ADD Instead of COPY
**Type**: regex
**Severity**: low
**Pattern**: `(?i)^ADD\s+(?!https?://)`
- ADD has implicit tar extraction and URL download that may be unexpected
- Example: `ADD package.json /app/`
- Remediation: Use COPY for local files: `COPY package.json /app/`

### apt-get Without Cleanup
**Type**: regex
**Severity**: low
**Pattern**: `(?i)apt-get\s+install[^&]*(?:$|\\$)`
- Not cleaning apt cache increases image size
- Example: `RUN apt-get update && apt-get install -y curl`
- Remediation: Add `&& rm -rf /var/lib/apt/lists/*` after install

### EXPOSE Wildcard Port
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)^EXPOSE\s+\*`
- Wildcard EXPOSE may expose unintended ports
- Example: `EXPOSE *`
- Remediation: Explicitly list required ports: `EXPOSE 80 443`

### Piping to Shell
**Type**: regex
**Severity**: high
**Pattern**: `(?i)(?:curl|wget)[^|]*\|\s*(?:bash|sh)`
- Piping untrusted content directly to shell is dangerous
- Example: `RUN curl https://example.com/script.sh | bash`
- Remediation: Download, verify, then execute: `RUN curl -o script.sh https://example.com/script.sh && chmod +x script.sh && ./script.sh`

### COPY with Wildcards
**Type**: regex
**Severity**: low
**Pattern**: `(?i)^COPY\s+\*`
- Wildcard COPY may include unintended files like .git or secrets
- Example: `COPY * /app/`
- Remediation: Use specific paths or .dockerignore file

---

## Best Practice Patterns

### Missing HEALTHCHECK
**Type**: regex
**Severity**: info
**Pattern**: `^(?!.*HEALTHCHECK).*$`
- HEALTHCHECK enables container orchestration health monitoring
- Missing HEALTHCHECK makes it harder to detect unhealthy containers
- Remediation: Add `HEALTHCHECK CMD curl -f http://localhost/ || exit 1`

### Missing Non-Root USER
**Type**: regex
**Severity**: high
**Pattern**: `^(?!.*USER\s+(?!root)).*CMD`
- Container runs as root if no USER directive before CMD
- Running as root increases attack surface if container is compromised
- Remediation: Add `USER nonroot` or `USER 1000:1000` before CMD

### Untagged Base Image
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)^FROM\s+[^:]+\s*$`
- Using untagged base image defaults to :latest
- Example: `FROM node`
- Remediation: Always specify a version tag: `FROM node:18`

---

## Detection Confidence

**Regex Detection**: 90%
**Best Practice Detection**: 85%

---

## References

- CIS Docker Benchmark
- Snyk Container Best Practices
- Docker Security Best Practices
