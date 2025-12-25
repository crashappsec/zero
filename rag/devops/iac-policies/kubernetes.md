# Kubernetes Security Patterns

**Category**: devops/iac-policies
**Description**: Kubernetes manifest security and organizational policy patterns
**CWE**: CWE-250 (Execution with Unnecessary Privileges), CWE-732 (Incorrect Permission Assignment)

---

## Container Security Patterns

### Container Running as Root
**Type**: regex
**Severity**: high
**Pattern**: `(?i)runAsUser:\s*0\b`
- Containers should not run as root user
- Example: `runAsUser: 0`
- Remediation: Set `runAsUser` to a non-root UID (e.g., 1000)

### Privileged Container
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)privileged:\s*true`
- Containers should not run in privileged mode
- Example: `privileged: true`
- Remediation: Set `privileged: false` and use specific capabilities instead

### Allow Privilege Escalation
**Type**: regex
**Severity**: high
**Pattern**: `(?i)allowPrivilegeEscalation:\s*true`
- Containers should not allow privilege escalation
- Example: `allowPrivilegeEscalation: true`
- Remediation: Set `allowPrivilegeEscalation: false`

### Host Network Namespace
**Type**: regex
**Severity**: high
**Pattern**: `(?i)hostNetwork:\s*true`
- Pods should not use the host network namespace
- Example: `hostNetwork: true`
- Remediation: Set `hostNetwork: false` and use NetworkPolicies

### Host PID Namespace
**Type**: regex
**Severity**: high
**Pattern**: `(?i)hostPID:\s*true`
- Pods should not use the host PID namespace
- Example: `hostPID: true`
- Remediation: Set `hostPID: false`

### Host IPC Namespace
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)hostIPC:\s*true`
- Pods should not use the host IPC namespace
- Example: `hostIPC: true`
- Remediation: Set `hostIPC: false`

### All Capabilities Added
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)capabilities:[\s\S]*?add:\s*\[\s*["']?ALL["']?\s*\]`
- Containers should not have all Linux capabilities
- Example: `add: ["ALL"]`
- Remediation: Add only specific required capabilities

### SYS_ADMIN Capability
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)capabilities:[\s\S]*?add:[\s\S]*?["']?SYS_ADMIN["']?`
- SYS_ADMIN capability is dangerous and rarely needed
- Example: `add: ["SYS_ADMIN"]`
- Remediation: Remove SYS_ADMIN and use more specific capabilities

### Writable Root Filesystem
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)readOnlyRootFilesystem:\s*false`
- Container root filesystem should be read-only
- Example: `readOnlyRootFilesystem: false`
- Remediation: Set `readOnlyRootFilesystem: true` and use volumes for writable paths

---

## Resource Management Patterns

### Missing Resource Limits
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)containers:[\s\S]*?name:\s*[^\n]+(?:(?!resources:).)*$`
- Containers should have resource limits defined
- Example: Container without `resources` block
- Remediation: Add `resources.limits` for CPU and memory

### Missing Memory Limit
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)resources:[\s\S]*?limits:(?:(?!memory:).)*$`
- Containers should have memory limits to prevent OOM issues
- Example: Missing `memory` in limits
- Remediation: Add `limits.memory` (e.g., "512Mi")

### Missing CPU Limit
**Type**: regex
**Severity**: low
**Pattern**: `(?i)resources:[\s\S]*?limits:(?:(?!cpu:).)*$`
- Containers should have CPU limits for fair scheduling
- Example: Missing `cpu` in limits
- Remediation: Add `limits.cpu` (e.g., "500m")

---

## Network Security Patterns

### Service Exposed via NodePort
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)type:\s*NodePort`
- NodePort exposes service on all cluster nodes
- Example: `type: NodePort`
- Remediation: Use LoadBalancer or ClusterIP with Ingress

### Service Exposed via LoadBalancer Without Annotation
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)type:\s*LoadBalancer(?:(?!annotations:).)*$`
- LoadBalancer services should have annotations for security
- Example: LoadBalancer without internal annotation
- Remediation: Add cloud provider annotations for internal load balancer if needed

### Missing NetworkPolicy
**Type**: structural
**Severity**: medium
**Pattern**: `NetworkPolicy`
- Namespaces should have NetworkPolicies for traffic control
- Example: Namespace without NetworkPolicy
- Remediation: Create NetworkPolicy to restrict pod communication

---

## RBAC Patterns

### ClusterRoleBinding to cluster-admin
**Type**: regex
**Severity**: critical
**Pattern**: `(?i)roleRef:[\s\S]*?name:\s*["']?cluster-admin["']?`
- Binding to cluster-admin grants full cluster access
- Example: `name: cluster-admin`
- Remediation: Create custom ClusterRole with minimal permissions

### Wildcard Verb in Role
**Type**: regex
**Severity**: high
**Pattern**: `(?i)verbs:\s*\[\s*["']?\*["']?\s*\]`
- Roles should not grant wildcard verb permissions
- Example: `verbs: ["*"]`
- Remediation: Specify explicit verbs (get, list, watch, create, etc.)

### Wildcard Resource in Role
**Type**: regex
**Severity**: high
**Pattern**: `(?i)resources:\s*\[\s*["']?\*["']?\s*\]`
- Roles should not grant access to all resources
- Example: `resources: ["*"]`
- Remediation: Specify explicit resource types

### Secrets Access in Role
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)resources:[\s\S]*?["']?secrets["']?[\s\S]*?verbs:\s*\[[\s\S]*?(?:get|list|\*)[\s\S]*?\]`
- Access to secrets should be carefully controlled
- Example: Role granting secrets access
- Remediation: Ensure secrets access is necessary and audited

---

## Pod Security Patterns

### Missing SecurityContext
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)spec:[\s\S]*?containers:(?:(?!securityContext:).)*$`
- Pods should have securityContext defined
- Example: Pod without securityContext
- Remediation: Add securityContext with runAsNonRoot, capabilities, etc.

### Missing RunAsNonRoot
**Type**: regex
**Severity**: high
**Pattern**: `(?i)securityContext:(?:(?!runAsNonRoot).)*$`
- Pods should enforce non-root execution
- Example: Missing runAsNonRoot in securityContext
- Remediation: Add `runAsNonRoot: true`

### Default Service Account
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)serviceAccountName:\s*["']?default["']?`
- Pods should not use the default service account
- Example: `serviceAccountName: default`
- Remediation: Create and use dedicated service account

### Missing ServiceAccount Token Mount
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)automountServiceAccountToken:\s*true`
- Service account tokens should not be auto-mounted unless needed
- Example: `automountServiceAccountToken: true`
- Remediation: Set `automountServiceAccountToken: false` unless required

---

## Image Security Patterns

### Image with Latest Tag
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)image:\s*[^\s:]+:latest`
- Images should not use the :latest tag
- Example: `image: nginx:latest`
- Remediation: Use specific version tags (e.g., nginx:1.25.0)

### Image without Tag
**Type**: regex
**Severity**: medium
**Pattern**: `(?i)image:\s*[^\s:]+\s*$`
- Images should have explicit tags
- Example: `image: nginx`
- Remediation: Specify version tag (e.g., nginx:1.25.0)

### Image Pull Policy Always
**Type**: regex
**Severity**: low
**Pattern**: `(?i)imagePullPolicy:\s*Always`
- Always pulling images may slow deployments
- Example: `imagePullPolicy: Always`
- Remediation: Consider IfNotPresent for production

---

## Organizational Policies

### Missing Labels
**Type**: regex
**Severity**: low
**Pattern**: `(?i)metadata:(?:(?!labels:).)*$`
- Resources should have labels for organization
- Example: Resource without labels
- Remediation: Add labels including app, environment, version, owner

### Missing Namespace
**Type**: regex
**Severity**: low
**Pattern**: `(?i)metadata:(?:(?!namespace:).)*kind:`
- Resources should specify namespace explicitly
- Example: Resource without namespace
- Remediation: Specify namespace or use kustomize/helm for namespace management

### Missing PodDisruptionBudget
**Type**: structural
**Severity**: low
**Pattern**: `PodDisruptionBudget`
- Deployments should have PodDisruptionBudget for availability
- Example: Deployment without PDB
- Remediation: Create PodDisruptionBudget with minAvailable

---

## Detection Confidence

**Regex Detection**: 85%
**Policy Compliance**: 90%

---

## References

- CIS Kubernetes Benchmark
- NSA/CISA Kubernetes Hardening Guide
- Pod Security Standards
- Kubernetes Security Best Practices
