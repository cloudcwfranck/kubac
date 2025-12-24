# kubac — Kubernetes Accelerator (Cloud-Agnostic)

[![CI](https://github.com/cloudcwfranck/kubac/actions/workflows/pr.yaml/badge.svg)](https://github.com/cloudcwfranck/kubac/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudcwfranck/kubac)](https://goreportcard.com/report/github.com/cloudcwfranck/kubac)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

kubac is a cloud-agnostic Kubernetes accelerator that installs a secure, autoscaling, self-healing "baseline platform" on any cluster using GitOps, with opinionated defaults, policy enforcement, and verifiable operational tests.

## What is kubac?

kubac provides a production-ready Kubernetes platform foundation that works on **any** cluster:
- **Local development**: kind, k3d, minikube
- **Managed clouds**: EKS, AKS, GKE
- **On-premises**: vanilla Kubernetes with CNI

It installs and configures essential platform components with security best practices baked in, then verifies everything works through automated tests.

## Features

- **Cloud-agnostic**: No AWS/Azure/GCP APIs required for core functionality
- **Security-first**: Pod Security Standards, default-deny NetworkPolicies, policy enforcement
- **Autoscaling**: HPA for workloads, optional node autoscaling
- **Self-healing**: Automated pod recovery, PodDisruptionBudgets
- **GitOps ready**: Flux integration for declarative cluster management
- **Verifiable**: Built-in test suite validates platform capabilities
- **One-command UX**: Simple CLI for init, install, verify, and demo workflows

## Quickstart on kind

### Prerequisites

- Go 1.21+ (to build kubac)
- kubectl
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)

### Installation

```bash
# Clone the repository
git clone https://github.com/cloudcwfranck/kubac.git
cd kubac

# Build kubac CLI
go build -o kubac ./cmd/kubac

# Create a kind cluster
kind create cluster --name kubac-demo

# Run preflight checks
./kubac doctor

# Initialize kubac configuration
./kubac init --cluster-profile=local --mode=direct

# Install the platform stack
./kubac install

# Deploy demo application
./kubac demo deploy

# Verify the installation
./kubac verify
```

### What Gets Installed

The default installation includes:

**Core Components**:
- `metrics-server`: Pod and node metrics (required for HPA)
- `kube-state-metrics`: Cluster state metrics

**Policy & Security**:
- `Kyverno`: Policy engine for security enforcement
- Baseline policies: non-root users, read-only root filesystem, no privileged containers
- Pod Security Standards: Restricted baseline
- NetworkPolicies: Default-deny with explicit allow rules

**Demo Application**:
- Sample HTTP service with HPA and PDB
- Configured for autoscaling and self-healing

### Testing the Platform

```bash
# Test HPA scaling
./kubac demo load --duration=60s

# Watch replicas scale up
kubectl get hpa -n kubac-demo -w

# Test self-healing
./kubac demo chaos

# Watch pod replacement
kubectl get pods -n kubac-demo -w

# Run full verification suite
./kubac verify --output=text
```

### Cleanup

```bash
# Remove demo app
./kubac demo cleanup

# Uninstall kubac platform
./kubac uninstall

# Delete kind cluster
kind delete cluster --name kubac-demo
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         kubac CLI                                │
│  (Go binary: init, install, verify, doctor, demo, uninstall)    │
└────────────────────────┬────────────────────────────────────────┘
                         │
         ┌───────────────┴───────────────┐
         │                               │
    ┌────▼─────┐                    ┌───▼────┐
    │  Direct  │                    │ GitOps │
    │   Mode   │                    │  Mode  │
    └────┬─────┘                    └───┬────┘
         │                              │
         │ kubectl apply                │ flux bootstrap
         │                              │
         ▼                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                            │
├─────────────────────────────────────────────────────────────────┤
│  Platform Stack:                                                 │
│    • metrics-server (required)                                   │
│    • kube-state-metrics (required)                               │
│    • Kyverno (policy engine)                                     │
│    • NetworkPolicies (default deny)                              │
│                                                                   │
│  Optional Components:                                            │
│    • Prometheus stack                                            │
│    • Ingress controller                                          │
│    • cert-manager                                                │
│    • Node autoscaler                                             │
└─────────────────────────────────────────────────────────────────┘
```

### Installation Modes

**Direct Mode** (default for local):
- kubac applies manifests directly via kubectl
- Fast, suitable for development and testing
- No Git repository required

**GitOps Mode** (recommended for production):
- kubac generates Flux manifests and directory structure
- You commit manifests to Git
- Flux syncs and applies changes automatically
- Full audit trail and declarative operations

## Components and Rationale

| Component | Purpose | Why This One |
|-----------|---------|--------------|
| **Kyverno** | Policy engine | Native Kubernetes resources, simpler than OPA for common use cases, active community |
| **Flux** | GitOps operator | CNCF graduated, lightweight, good multi-tenancy support |
| **metrics-server** | Resource metrics | Official Kubernetes SIG project, required for HPA |
| **kube-state-metrics** | Cluster state metrics | De facto standard for Prometheus-style monitoring |

All components are:
- Widely adopted in production
- Actively maintained
- Version-pinned for reproducibility
- Cloud-agnostic

## Security Model

kubac enforces defense-in-depth security:

### 1. Pod Security Standards
- Enforced via Kyverno policies
- Default: `restricted` baseline
- Blocks: privileged containers, host namespace access, privilege escalation

### 2. Network Policies
- Default deny all ingress/egress
- Explicit allow rules for:
  - DNS resolution (kube-dns)
  - Application-specific traffic
  - System namespace communication

### 3. Policy Enforcement
- Required: `runAsNonRoot: true`
- Required: `readOnlyRootFilesystem: true`
- Denied: `privileged: true`
- All policies auditable and versioned

### 4. Future: Supply Chain Security (v0.2+)
- SBOM generation for releases
- Image signature verification
- Vulnerability scanning

## Commands Reference

### Core Commands

```bash
# Check cluster access and prerequisites
kubac doctor

# Initialize kubac configuration
kubac init [--cluster-profile=local|managed|onprem] [--mode=direct|gitops]

# Install platform components
kubac install [--config=kubac.yaml]

# Verify installation
kubac verify [--output=text|json] [--report=path]

# Uninstall components
kubac uninstall [--force]

# Show version
kubac version
```

### Demo Commands

```bash
# Deploy demo application
kubac demo deploy

# Generate load (test HPA)
kubac demo load [--duration=60s] [--requests=100]

# Trigger chaos (test self-healing)
kubac demo chaos

# Remove demo app
kubac demo cleanup
```

## Configuration

kubac uses a `kubac.yaml` configuration file:

```yaml
clusterProfile: local  # local, managed, onprem
mode: direct           # direct or gitops

platform:
  metricsServer:
    enabled: true
    version: v0.7.0
  kubeStateMetrics:
    enabled: true
    version: v2.10.1

policy:
  enabled: true
  provider: kyverno
  podSecurityStandard: restricted

networkPolicy:
  enabled: true
  defaultDeny: true

demo:
  namespace: kubac-demo
  replicas: 2
  hpa:
    minReplicas: 2
    maxReplicas: 10
    targetCPUUtilization: 50
```

See [kubac.yaml.example](kubac.yaml.example) for full configuration options.

## Real Environment Usage

### GitOps Mode with Existing Org Repo

```bash
# Initialize with GitOps mode
kubac init --mode=gitops --cluster-profile=managed

# Edit kubac.yaml to set your repo
vim kubac.yaml
# Set gitops.repoURL to your repo
# Set gitops.path to clusters/production

# Generate manifests
kubac install --mode=gitops

# Commit to your repo
git add clusters/production/
git commit -m "Add kubac platform baseline"
git push

# Bootstrap Flux
flux bootstrap github \
  --owner=your-org \
  --repository=your-repo \
  --branch=main \
  --path=clusters/production
```

### Environment Overlays (dev/stage/prod)

Use Kustomize overlays for environment-specific configs:

```
your-repo/
├── base/
│   └── kubac/           # Generated by kubac install
├── overlays/
│   ├── dev/
│   │   └── kustomization.yaml
│   ├── stage/
│   │   └── kustomization.yaml
│   └── prod/
│       └── kustomization.yaml
```

### CI Usage

Run `kubac verify` in CI against any cluster:

```yaml
- name: Verify platform
  run: |
    kubac verify --output=json --report=verify-report.json

- name: Upload report
  uses: actions/upload-artifact@v4
  with:
    name: verification-report
    path: verify-report.json
```

## Limitations

**Current (v0.1)**:
- No cloud-specific integrations (EBS volumes, ALB, etc.)
- Node autoscaling requires manual cloud provider configuration
- No UI/dashboard included
- Limited policy library (extensible via custom Kyverno policies)

**Not Goals**:
- Not a Kubernetes distribution
- Not a replacement for your CI/CD pipeline
- Not an application platform (use Knative, etc. on top)

## Roadmap

**v0.2** (planned):
- Cloud provider adapters (AWS, Azure, GCP)
- Enhanced policy library
- Automatic SBOM generation
- Image signature verification
- Prometheus + Grafana dashboards

**v0.3** (future):
- Multi-cluster support
- Advanced autoscaling (KEDA integration)
- Disaster recovery automation
- Cost optimization recommendations

## Development

### Building

```bash
go build -o kubac ./cmd/kubac
```

### Testing

```bash
# Unit tests
go test ./...

# Linting
golangci-lint run

# E2E on kind
./test/kind/test-kind.sh

# E2E on k3d
./test/k3d/test-k3d.sh
```

### Project Structure

```
kubac/
├── cmd/kubac/           # CLI entry point
├── internal/
│   ├── cli/             # Command implementations
│   ├── config/          # Configuration handling
│   ├── install/         # Installation logic
│   ├── verify/          # Verification tests
│   ├── doctor/          # Preflight checks
│   ├── render/          # Template rendering
│   └── report/          # Report generation
├── bundles/             # Platform component manifests
│   ├── platform/
│   ├── policies/
│   ├── netpol/
│   └── demos/
├── test/                # E2E tests
│   ├── kind/
│   └── k3d/
└── .github/workflows/   # CI/CD
```

## Contributing

Contributions welcome! Please:
1. Open an issue to discuss major changes
2. Follow existing code style
3. Add tests for new functionality
4. Update documentation

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

## Credits

kubac builds on these excellent open-source projects:
- [Kubernetes](https://kubernetes.io/)
- [Kyverno](https://kyverno.io/)
- [Flux](https://fluxcd.io/)
- [kind](https://kind.sigs.k8s.io/)
- [k3d](https://k3d.io/)

---

**Status**: Alpha (v0.1) - Not recommended for production use yet. Feedback and testing welcome!