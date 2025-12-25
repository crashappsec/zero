# Kubernetes Best Practices Patterns

**Category**: devops/iac-best-practices
**Description**: Kubernetes manifest organizational and operational best practices
**Type**: best-practice

---

## Label Patterns

### Missing Standard Labels
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `metadata:\s*\n\s+name:[^\n]+\n(?:(?!labels:).)*spec:`
- Resources should have standard labels for organization
- Required labels: app, version, component
- Remediation: Add `labels: { app.kubernetes.io/name: myapp, app.kubernetes.io/version: "1.0" }`

### Missing App Label
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `labels:\s*\n(?:(?!app\.kubernetes\.io/name|app:).)*\n\s+[a-z]`
- All resources should have an app/name label for identification
- Remediation: Add `app.kubernetes.io/name: <app-name>` label

---

## Resource Management

### Missing Resource Requests
**Type**: regex
**Severity**: high
**Category**: best-practice
**Pattern**: `containers:\s*\n\s*-\s*name:[^\n]+\n(?:(?!resources:).)*image:`
- Containers should define resource requests for scheduling
- Example: Container without `resources.requests` block
- Remediation: Add `resources: { requests: { cpu: "100m", memory: "128Mi" } }`

### Missing Resource Limits
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `resources:\s*\n(?:(?!limits:).)*requests:`
- Containers should define resource limits to prevent resource exhaustion
- Remediation: Add `limits: { cpu: "500m", memory: "512Mi" }`

---

## Health Checks

### Missing Liveness Probe
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `containers:\s*\n\s*-\s*name:[^\n]+\n(?:(?!livenessProbe:).)*ports:`
- Containers should have liveness probes for restart on failure
- Remediation: Add `livenessProbe: { httpGet: { path: /health, port: 8080 } }`

### Missing Readiness Probe
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `containers:\s*\n\s*-\s*name:[^\n]+\n(?:(?!readinessProbe:).)*ports:`
- Containers should have readiness probes for traffic management
- Remediation: Add `readinessProbe: { httpGet: { path: /ready, port: 8080 } }`

---

## Replica Management

### Single Replica Deployment
**Type**: regex
**Severity**: medium
**Category**: best-practice
**Pattern**: `replicas:\s*1\s*\n`
- Production deployments should have multiple replicas for HA
- Remediation: Set `replicas: 2` or higher for production

### Missing PodDisruptionBudget
**Type**: structural
**Severity**: low
**Category**: best-practice
- Critical deployments should have PodDisruptionBudget
- Remediation: Create PDB with `minAvailable` or `maxUnavailable`

---

## Namespace Best Practices

### Using Default Namespace
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `namespace:\s*["']?default["']?\s*\n`
- Avoid using the default namespace in production
- Remediation: Create and use application-specific namespaces

### Missing Namespace
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `metadata:\s*\n\s+name:[^\n]+\n(?:(?!namespace:).)*spec:`
- Resources should explicitly specify namespace
- Remediation: Add `namespace: <namespace>` to metadata

---

## Image Best Practices

### Missing Image Pull Policy
**Type**: regex
**Severity**: low
**Category**: best-practice
**Pattern**: `image:\s*[^\n]+\n(?:(?!imagePullPolicy:).)*ports:`
- Containers should specify imagePullPolicy explicitly
- Remediation: Add `imagePullPolicy: IfNotPresent` or `Always`
