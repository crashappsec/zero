# Kubernetes

**Category**: developer-tools/containers
**Description**: Kubernetes - container orchestration platform for automating deployment, scaling, and management
**Homepage**: https://kubernetes.io

## Package Detection

### NPM
- `@kubernetes/client-node`

### PYPI
- `kubernetes`
- `kopf`

### GO
- `k8s.io/client-go`
- `k8s.io/api`
- `sigs.k8s.io/controller-runtime`

## Configuration Files

- `*.yaml` (with apiVersion/kind)
- `*.yml` (with apiVersion/kind)
- `kustomization.yaml`
- `kustomization.yml`
- `Chart.yaml` (Helm)
- `values.yaml` (Helm)
- `helmfile.yaml`
- `skaffold.yaml`
- `.kube/config`
- `kubeconfig`

## File Patterns

- `deployment.yaml`
- `service.yaml`
- `configmap.yaml`
- `secret.yaml`
- `ingress.yaml`
- `namespace.yaml`
- `pod.yaml`
- `statefulset.yaml`
- `daemonset.yaml`
- `cronjob.yaml`

## Environment Variables

- `KUBECONFIG`
- `KUBERNETES_SERVICE_HOST`
- `KUBERNETES_SERVICE_PORT`

## Detection Notes

- Look for YAML files with apiVersion and kind fields
- Check for kustomization.yaml (Kustomize)
- Chart.yaml indicates Helm chart
- k8s/ or kubernetes/ directories common
- Check for kubectl commands in scripts

## Detection Confidence

- **Configuration File Detection**: 95% (HIGH)
- **Helm Chart Detection**: 95% (HIGH)
- **Package Detection**: 90% (HIGH)
