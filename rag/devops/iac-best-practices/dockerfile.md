# Dockerfile Best Practices Patterns

**Category**: devops/iac-best-practices
**Description**: Dockerfile organizational and operational best practices
**Type**: best-practice

---

## Build Optimization

### Missing Multi-Stage Build
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `^FROM\s+[^\n]+\n(?:(?!FROM\s).)*$`
- Use multi-stage builds to reduce image size
- Example: Single FROM with build tools in final image
- Remediation: Use `FROM builder AS build` then `FROM runtime` pattern

### Inefficient Layer Caching
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `COPY\s+\.\s+\.\s*\n.*RUN\s+(?:npm|pip|go)\s+install`
- Copy dependency files before source for better caching
- Example: `COPY . . && RUN npm install` (bad)
- Remediation: `COPY package*.json ./ && RUN npm install && COPY . .`

### Missing .dockerignore Reference
**Type**: structural
**Severity**: low
**Category**: best-practice
- Ensure .dockerignore exists to exclude unnecessary files
- Remediation: Create .dockerignore with node_modules, .git, etc.

---

## Image Size Optimization

### Using Full Base Image
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `FROM\s+(?:node|python|golang|ruby|java):\d+(?!-(?:alpine|slim|distroless))`
- Consider using slim or alpine variants for smaller images
- Example: `FROM node:18` (larger)
- Remediation: Use `FROM node:18-alpine` or `FROM node:18-slim`

### Not Cleaning Package Manager Cache
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `RUN\s+apt-get\s+install[^&]+(?<!rm\s+-rf\s+/var/lib/apt/lists/\*)`
- Clean package manager cache in same layer
- Remediation: End with `&& rm -rf /var/lib/apt/lists/*`

---

## Metadata Best Practices

### Missing LABEL Instructions
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `^FROM[^\n]+\n(?:(?!LABEL).)*CMD`
- Images should have metadata labels
- Remediation: Add `LABEL maintainer="email" version="1.0"`

### Missing WORKDIR
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `^FROM[^\n]+\n(?:(?!WORKDIR).)*(?:COPY|RUN)`
- Set explicit WORKDIR instead of relying on defaults
- Remediation: Add `WORKDIR /app` before COPY/RUN

---

## Instruction Best Practices

### Using ADD Instead of COPY
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `ADD\s+(?!https?://)[^\s]+\s+`
- Use COPY for local files; ADD only for URLs/archives
- Example: `ADD src/ /app/` (bad)
- Remediation: Use `COPY src/ /app/` (good)

### Combining RUN Instructions
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `RUN\s+[^\n]+\nRUN\s+[^\n]+\nRUN\s+`
- Combine related RUN instructions to reduce layers
- Example: Multiple sequential RUN commands
- Remediation: Combine with `&&` in single RUN

### Missing CMD or ENTRYPOINT
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `^(?:(?!CMD|ENTRYPOINT).)*$`
- Dockerfiles should define how to run the container
- Remediation: Add `CMD ["node", "server.js"]` or `ENTRYPOINT`

---

## Environment Best Practices

### Hardcoded Environment Values
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `ENV\s+(?:PASSWORD|SECRET|KEY|TOKEN)\s*=\s*[^\$][^\s]+`
- Don't hardcode secrets in Dockerfile
- Remediation: Use build args or runtime environment variables

### Missing Environment Documentation
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `^(?:(?!ENV).)*CMD`
- Document expected environment variables
- Remediation: Add ENV with placeholder values and comments
