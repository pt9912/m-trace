# Kubernetes example manifests

These manifests are an optional `0.13.0` seed for `MVP-42` / `NF-18`.
They are deliberately small and are not a production-ready Kubernetes
distribution.

## Scope

- The manifests run the API, analyzer service and dashboard with the
  same lab-oriented assumptions as `docker-compose.yml`.
- SQLite is stored in a `PersistentVolumeClaim` mounted by the single API
  replica.
- Secrets are placeholders and must be supplied by an operator-specific
  Secret or external secret system before any non-local use.
- No ingress, TLS, HPA, network policy, pod security profile, backup
  policy or managed database setup is included.

## Apply

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/analyzer-service.yaml
kubectl apply -f deploy/k8s/api.yaml
kubectl apply -f deploy/k8s/dashboard.yaml
```

For local inspection with a private image registry, override image names
before applying or use `kubectl set image` after apply.

## Validate

The example set has a cluster-free validation target:

```bash
make k8s-validate
```

The check parses every manifest, limits the allowed resource kinds to
`Namespace`, `PersistentVolumeClaim`, `Deployment` and `Service`, keeps
all workloads single-replica, verifies the shared `m-trace` labels and
requires image tags to match the root `package.json` version. It also
guards the R-9 boundary by rejecting example labels named `pod`,
`namespace` or `container`; those labels belong to a future
Kubernetes-specific smoke profile, not to the Compose default allowlist.

## R-9 observability note

These examples do not add a Kubernetes observability smoke gate. A future
K8s smoke must use a Kubernetes-specific infrastructure-label allowlist
instead of widening the Compose default silently.
