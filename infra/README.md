# infra

Infrastructure-as-Code for the Kleff platform.

## Structure

```
infra/
├── kubernetes/     # Raw Kubernetes manifests (Deployments, Services, Ingresses, ConfigMaps, etc.)
├── helm/           # Helm charts for packaging and templating Kubernetes resources
└── argocd/         # ArgoCD Application definitions for GitOps-based continuous delivery
```

## Deployment Strategy

The platform uses a **GitOps** model:

1. Container images are built and pushed by CI (GitHub Actions)
2. ArgoCD watches this repository for changes to `infra/argocd/`
3. ArgoCD syncs desired state from `infra/helm/` or `infra/kubernetes/` into the cluster

## Directory Details

### `kubernetes/`

Raw manifests for environments where Helm is not used. Useful for simple, environment-specific overrides or for resources that don't belong in a chart (e.g., cluster-level RBAC, namespaces).

### `helm/`

Helm charts per service/app. Each chart packages the Kubernetes resources for a single deployable unit with configurable values (`values.yaml`, `values.prod.yaml`, etc.).

### `argocd/`

ArgoCD `Application` custom resources that point to charts or manifest directories in this repo. These define what gets deployed, from which source, and into which cluster/namespace.
