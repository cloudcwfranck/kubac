# kubac — Kubernetes Accelerator (Cloud-Agnostic)
Version: 0.1 (Foundation)
Status: Build from scratch

## 0) One-line mission
kubac is a cloud-agnostic Kubernetes accelerator that installs a secure, autoscaling, self-healing “baseline platform” on any cluster using GitOps, with opinionated defaults, policy enforcement, and verifiable operational tests.

## 1) Primary outcomes (must be demonstrably true)
kubac MUST provide:
1) Cluster baseline install (GitOps) that works on:
   - local: kind AND k3d (required)
   - at least one managed: EKS OR AKS OR GKE (docs + manifests; optional CI)
   - generic on-prem: “vanilla” Kubernetes with CNI (docs + manifests)
2) Self-healing workloads:
   - prove via chaos test that pods are restarted/rescheduled
   - prove PDB + readiness/liveness work
3) Autoscaling:
   - HPA for CPU + custom metrics example
   - Cluster Autoscaler OR Karpenter (choose one default) with abstraction layer
   - prove via load test that replicas scale and nodes scale (where supported)
4) Security and governance baseline:
   - Pod Security Standards enforced (restricted baseline)
   - NetworkPolicy default-deny with explicit allow
   - Supply chain: pinned images, SBOM generation, signature verification hooks
   - Policy engine: Kyverno OR OPA Gatekeeper (pick one default, support both later)
5) “One command” user experience:
   - `kubac init`, `kubac install`, `kubac verify`, `kubac doctor`, `kubac uninstall`
   - outputs are deterministic, scriptable, CI-friendly
6) A real reference app:
   - `kubac demo deploy` installs a sample microservice with HPA + PDB + policies
   - `kubac demo load` triggers scaling
   - `kubac demo chaos` triggers recovery proof

## 2) Non-goals (explicitly NOT in v0.1)
- No custom Kubernetes distribution
- No proprietary cloud dependencies
- No “magic” controllers that bypass upstream primitives
- No UI required in v0.1 (CLI + docs only)

## 3) Design principles (non-negotiable)
- Cloud-agnostic: never require AWS/Azure/GCP APIs for core install.
- Reproducible: every install step is idempotent.
- Secure-by-default: deny-by-default networking + policy enforcement.
- Minimal moving parts: prefer upstream, widely adopted components.
- Operable: everything has `verify` tests and `doctor` checks.

## 4) Architecture (v0.1)
kubac = CLI + GitOps bundles + verification tests.

A) CLI (Go) responsibilities:
- render templates (Helm/Kustomize) with user config
- create GitOps repo structure OR apply directly (two modes)
- install baseline components into namespaces
- run verification suite (conformance-style checks)
- produce a machine-readable report (JSON) and human summary

B) Bundles (delivered as versioned artifacts inside repo):
- gitops/
  - flux/ or argocd/ (choose Flux for v0.1)
- platform/
  - ingress (nginx or traefik) [optional in v0.1, behind flag]
  - cert-manager (optional)
  - metrics-server (required)
  - kube-state-metrics (required)
  - prometheus stack OR lightweight metrics path (choose kube-prometheus-stack for now)
  - policy engine (Kyverno recommended)
  - network policies baseline
  - autoscaling:
    - HPA examples (required)
    - node autoscaler: Karpenter OR Cluster Autoscaler (default to Cluster Autoscaler abstraction; implement generic interface first; document cloud-specific adapters)

C) Verification suite:
- Must run on kind/k3d without cloud creds.
- Must include:
  - pod self-heal test (kill pod -> replaced)
  - HPA test (load -> replicas scale up/down)
  - policy test (attempt forbidden pod -> denied)
  - network test (default deny blocks traffic)
  - reporting (JSON + markdown)

## 5) UX Contract (commands)
CLI name: `kubac`

Required commands:
- `kubac version`
- `kubac init`        # scaffolds a kubac project (config + gitops layout)
- `kubac install`     # installs baseline stack (direct apply OR GitOps bootstrap)
- `kubac verify`      # runs verification suite and prints report path
- `kubac doctor`      # preflight checks (kubectl access, context, privileges, node OS)
- `kubac uninstall`   # removes installed components safely
- `kubac demo deploy|load|chaos|cleanup`

Flags (minimum):
- `--mode=gitops|direct` (default gitops)
- `--gitops=flux` (v0.1 fixed to flux)
- `--cluster-profile=local|managed|onprem` (drives defaults)
- `--output=json|text` for verify/doctor

Config file:
- `kubac.yaml` at repo root created by `kubac init`

## 6) Repo structure (MUST match exactly)
/
  AGENTS.md
  README.md
  kubac.yaml.example
  cmd/kubac/main.go
  internal/
    config/
    cli/
    render/
    install/
    verify/
    doctor/
    report/
  bundles/
    flux/
    platform/
    policies/
    netpol/
    demos/
  test/
    e2e/
    kind/
    k3d/
  scripts/
  .github/workflows/
  LICENSE

## 7) Quality gates (must be enforced)
- `go test ./...` must pass
- `golangci-lint` configured and passing
- e2e tests runnable locally:
  - kind and k3d scripts in `test/`
- GitHub Actions:
  - unit + lint on PR
  - nightly e2e on kind (required)
- Security:
  - SBOM generation for CLI release artifact (Syft)
  - vulnerability scan step (Grype) as “advisory” in CI

## 8) “Real environment” usage requirements
kubac MUST support:
- Local developer cluster validation (kind/k3d) with zero cloud deps.
- GitOps mode:
  - creates Flux bootstrap manifests that point to a repo path
  - supports “bring your own repo” by generating a folder tree and instructions
- Direct mode:
  - applies the same rendered manifests without GitOps

kubac MUST document:
- how to deploy to an existing org repo
- how to promote environment overlays (dev/stage/prod) via kustomize
- how to run `kubac verify` in CI against a cluster

## 9) Documentation requirements
README must include:
- what kubac is (no fluff)
- quickstart on kind
- install/verify/uninstall examples
- architecture diagram (ASCII ok)
- component list and why chosen
- security model (policies + netpol + PSS)
- limitations and roadmap

## 10) Implementation constraints for the AI
- Do not invent cloud integrations in v0.1.
- Prefer battle-tested OSS components and pin versions.
- Every bundle must be reproducible (version pinned).
- Keep defaults minimal; everything optional behind flags.
- Provide clear TODO markers for v0.2+ expansion.
