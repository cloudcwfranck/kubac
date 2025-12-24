package render

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudcwfranck/kubac/internal/config"
)

// Renderer handles manifest rendering
type Renderer struct {
	config *config.Config
}

// NewRenderer creates a new renderer
func NewRenderer(cfg *config.Config) *Renderer {
	return &Renderer{
		config: cfg,
	}
}

// RenderGitOpsManifests renders GitOps manifests to the specified path
func (r *Renderer) RenderGitOpsManifests(destPath string) error {
	// Create Flux bootstrap manifests
	if err := r.renderFluxBootstrap(destPath); err != nil {
		return err
	}

	// Create kustomization files for each component
	if err := r.renderKustomizations(destPath); err != nil {
		return err
	}

	return nil
}

func (r *Renderer) renderFluxBootstrap(destPath string) error {
	fluxSystemPath := filepath.Join(destPath, "flux-system")

	// Create gotk-components.yaml (Flux components)
	fluxComponents := `---
apiVersion: v1
kind: Namespace
metadata:
  name: flux-system
---
# Flux toolkit components
# Install via: flux install --export > gotk-components.yaml
# This is a placeholder - use 'flux bootstrap' for production
`

	if err := os.WriteFile(filepath.Join(fluxSystemPath, "gotk-components.yaml"), []byte(fluxComponents), 0644); err != nil {
		return err
	}

	// Create gotk-sync.yaml
	syncManifest := fmt.Sprintf(`---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: flux-system
  namespace: flux-system
spec:
  interval: 1m0s
  ref:
    branch: %s
  url: %s
---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: flux-system
  namespace: flux-system
spec:
  interval: 10m0s
  path: %s
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
`, r.config.GitOps.Branch, r.config.GitOps.RepoURL, r.config.GitOps.Path)

	if err := os.WriteFile(filepath.Join(fluxSystemPath, "gotk-sync.yaml"), []byte(syncManifest), 0644); err != nil {
		return err
	}

	return nil
}

func (r *Renderer) renderKustomizations(destPath string) error {
	// Platform kustomization
	platformKust := `---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: platform
  namespace: flux-system
spec:
  interval: 10m0s
  path: ./platform
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
`
	if err := os.WriteFile(filepath.Join(destPath, "platform-kustomization.yaml"), []byte(platformKust), 0644); err != nil {
		return err
	}

	// Policies kustomization
	policiesKust := `---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: policies
  namespace: flux-system
spec:
  interval: 10m0s
  path: ./policies
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
  dependsOn:
    - name: platform
`
	if err := os.WriteFile(filepath.Join(destPath, "policies-kustomization.yaml"), []byte(policiesKust), 0644); err != nil {
		return err
	}

	return nil
}
