# Flux GitOps

This directory contains Flux configuration for GitOps mode.

## Installation

For GitOps mode, kubac generates Flux manifests in your specified repository path.

### Bootstrap Flux

```bash
# GitHub
flux bootstrap github \
  --owner=<organization> \
  --repository=<repository> \
  --branch=main \
  --path=clusters/my-cluster \
  --personal

# GitLab
flux bootstrap gitlab \
  --owner=<group> \
  --repository=<repository> \
  --branch=main \
  --path=clusters/my-cluster

# Generic Git
flux bootstrap git \
  --url=ssh://git@<host>/<org>/<repository> \
  --branch=main \
  --path=clusters/my-cluster
```

## Directory Structure

After running `kubac init --mode=gitops`, you'll have:

```
clusters/my-cluster/
├── flux-system/
│   ├── gotk-components.yaml
│   └── gotk-sync.yaml
├── platform/
│   └── (platform components)
├── policies/
│   └── (kyverno policies)
├── netpol/
│   └── (network policies)
└── apps/
    └── (your applications)
```

## Next Steps

1. Commit the generated manifests to your Git repository
2. Run the bootstrap command above
3. Flux will automatically sync and apply all manifests
4. Run `kubac verify` to validate the installation
