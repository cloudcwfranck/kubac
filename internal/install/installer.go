package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudcwfranck/kubac/internal/config"
	"github.com/cloudcwfranck/kubac/internal/render"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Installer manages kubac installation
type Installer struct {
	config    *config.Config
	clientset *kubernetes.Clientset
	dynamic   dynamic.Interface
	renderer  *render.Renderer
}

// NewInstaller creates a new installer
func NewInstaller(cfg *config.Config) (*Installer, error) {
	// Load kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Create renderer
	renderer := render.NewRenderer(cfg)

	return &Installer{
		config:    cfg,
		clientset: clientset,
		dynamic:   dynamicClient,
		renderer:  renderer,
	}, nil
}

// Install installs the kubac platform
func (i *Installer) Install() error {
	ctx := context.Background()

	fmt.Println("\n→ Creating kubac-system namespace...")
	if err := i.createNamespace(ctx, "kubac-system"); err != nil {
		return err
	}

	// Install based on mode
	if i.config.Mode == "gitops" {
		return i.installGitOps(ctx)
	}
	return i.installDirect(ctx)
}

// installDirect installs components directly via kubectl apply
func (i *Installer) installDirect(ctx context.Context) error {
	// Install metrics server
	if i.config.Platform.MetricsServer.Enabled {
		fmt.Println("\n→ Installing metrics-server...")
		if err := i.applyManifests(ctx, "bundles/platform/metrics-server"); err != nil {
			return fmt.Errorf("failed to install metrics-server: %w", err)
		}
		fmt.Println("  ✓ metrics-server installed")
	}

	// Install kube-state-metrics
	if i.config.Platform.KubeStateMetrics.Enabled {
		fmt.Println("\n→ Installing kube-state-metrics...")
		if err := i.applyManifests(ctx, "bundles/platform/kube-state-metrics"); err != nil {
			return fmt.Errorf("failed to install kube-state-metrics: %w", err)
		}
		fmt.Println("  ✓ kube-state-metrics installed")
	}

	// Install policy engine
	if i.config.Policy.Enabled {
		fmt.Println("\n→ Installing policy engine (Kyverno)...")
		if err := i.applyManifests(ctx, "bundles/policies/kyverno"); err != nil {
			return fmt.Errorf("failed to install kyverno: %w", err)
		}
		fmt.Println("  ✓ Kyverno installed")

		// Wait for Kyverno to be ready
		fmt.Println("  Waiting for Kyverno to be ready...")
		if err := i.waitForDeployment(ctx, "kyverno", "kyverno", 120*time.Second); err != nil {
			return fmt.Errorf("kyverno not ready: %w", err)
		}

		// Apply policies
		fmt.Println("\n→ Applying baseline policies...")
		if err := i.applyManifests(ctx, "bundles/policies/baseline"); err != nil {
			return fmt.Errorf("failed to apply policies: %w", err)
		}
		fmt.Println("  ✓ Baseline policies applied")
	}

	// Install network policies
	if i.config.NetworkPolicy.Enabled {
		fmt.Println("\n→ Installing network policies...")
		if err := i.applyManifests(ctx, "bundles/netpol"); err != nil {
			return fmt.Errorf("failed to install network policies: %w", err)
		}
		fmt.Println("  ✓ Network policies installed")
	}

	return nil
}

// installGitOps generates Flux manifests and directory structure
func (i *Installer) installGitOps(ctx context.Context) error {
	fmt.Println("\n→ Generating GitOps manifests...")

	gitopsPath := i.config.GitOps.Path
	if gitopsPath == "" {
		gitopsPath = "clusters/my-cluster"
	}

	// Create directory structure
	dirs := []string{
		filepath.Join(gitopsPath, "flux-system"),
		filepath.Join(gitopsPath, "platform"),
		filepath.Join(gitopsPath, "policies"),
		filepath.Join(gitopsPath, "netpol"),
		filepath.Join(gitopsPath, "apps"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Copy and render manifests
	if err := i.renderer.RenderGitOpsManifests(gitopsPath); err != nil {
		return fmt.Errorf("failed to render GitOps manifests: %w", err)
	}

	fmt.Printf("  ✓ GitOps manifests generated at %s\n", gitopsPath)
	fmt.Println("\nNext steps for GitOps mode:")
	fmt.Println("  1. Commit the generated manifests to your Git repository")
	fmt.Println("  2. Bootstrap Flux: flux bootstrap github --owner=<org> --repository=<repo> --path=" + gitopsPath)
	fmt.Println("  3. Flux will automatically sync and apply the manifests")

	return nil
}

// applyManifests applies manifests from the filesystem
func (i *Installer) applyManifests(ctx context.Context, path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read manifests directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".yaml" && filepath.Ext(entry.Name()) != ".yml" {
			continue
		}

		manifestPath := filepath.Join(path, entry.Name())
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			return fmt.Errorf("failed to read manifest %s: %w", manifestPath, err)
		}

		// Apply manifest
		if err := i.applyManifest(ctx, data); err != nil {
			return fmt.Errorf("failed to apply manifest %s: %w", manifestPath, err)
		}
	}

	return nil
}

// applyManifest applies a single manifest
func (i *Installer) applyManifest(ctx context.Context, data []byte) error {
	// Parse and apply using kubectl apply logic
	// For simplicity, we'll use server-side apply
	obj := &unstructured.Unstructured{}
	if err := obj.UnmarshalJSON(data); err != nil {
		return err
	}

	gvr := schema.GroupVersionResource{
		Group:    obj.GroupVersionKind().Group,
		Version:  obj.GroupVersionKind().Version,
		Resource: obj.GetKind() + "s", // Simple pluralization
	}

	_, err := i.dynamic.Resource(gvr).Namespace(obj.GetNamespace()).Apply(
		ctx,
		obj.GetName(),
		obj,
		metav1.ApplyOptions{
			FieldManager: "kubac",
		},
	)

	return err
}

// createNamespace creates a namespace if it doesn't exist
func (i *Installer) createNamespace(ctx context.Context, name string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	_, err := i.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil && !isAlreadyExists(err) {
		return err
	}

	return nil
}

// waitForDeployment waits for a deployment to be ready
func (i *Installer) waitForDeployment(ctx context.Context, namespace, name string, timeout time.Duration) error {
	return wait.PollImmediate(5*time.Second, timeout, func() (bool, error) {
		deployment, err := i.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		return deployment.Status.ReadyReplicas == *deployment.Spec.Replicas, nil
	})
}

// Uninstall removes kubac components
func (i *Installer) Uninstall() error {
	ctx := context.Background()

	fmt.Println("\n→ Removing kubac components...")

	// Delete namespaces (this will cascade delete all resources)
	namespaces := []string{"kubac-system", "kyverno"}
	if i.config.Demo.Namespace != "" {
		namespaces = append(namespaces, i.config.Demo.Namespace)
	}

	for _, ns := range namespaces {
		fmt.Printf("  Deleting namespace %s...\n", ns)
		err := i.clientset.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{})
		if err != nil && !isNotFound(err) {
			fmt.Printf("  Warning: failed to delete namespace %s: %v\n", ns, err)
		}
	}

	return nil
}

// DeployDemo deploys the demo application
func (i *Installer) DeployDemo() error {
	ctx := context.Background()

	// Create demo namespace
	fmt.Printf("\n→ Creating namespace %s...\n", i.config.Demo.Namespace)
	if err := i.createNamespace(ctx, i.config.Demo.Namespace); err != nil {
		return err
	}

	// Apply demo manifests
	fmt.Println("\n→ Deploying demo application...")
	if err := i.applyManifests(ctx, "bundles/demos/demo-app"); err != nil {
		return fmt.Errorf("failed to deploy demo app: %w", err)
	}

	fmt.Println("  ✓ Demo application deployed")

	return nil
}

// RunLoadTest runs a load test
func (i *Installer) RunLoadTest(duration string, requests int) error {
	ctx := context.Background()

	// Apply load test job
	fmt.Println("\n→ Starting load test job...")
	if err := i.applyManifests(ctx, "bundles/demos/load-test"); err != nil {
		return fmt.Errorf("failed to start load test: %w", err)
	}

	fmt.Println("  ✓ Load test job started")
	fmt.Println("  Monitor HPA with: kubectl get hpa -n " + i.config.Demo.Namespace + " -w")

	return nil
}

// RunChaosTest runs chaos tests
func (i *Installer) RunChaosTest() error {
	ctx := context.Background()

	// Delete a pod to test self-healing
	fmt.Println("\n→ Deleting a demo pod to test self-healing...")

	pods, err := i.clientset.CoreV1().Pods(i.config.Demo.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=demo",
	})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("no demo pods found")
	}

	podName := pods.Items[0].Name
	fmt.Printf("  Deleting pod %s...\n", podName)

	err = i.clientset.CoreV1().Pods(i.config.Demo.Namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete pod: %w", err)
	}

	fmt.Println("  ✓ Pod deleted, Kubernetes should recreate it automatically")
	fmt.Println("  Monitor pods with: kubectl get pods -n " + i.config.Demo.Namespace + " -w")

	return nil
}

// CleanupDemo removes the demo application
func (i *Installer) CleanupDemo() error {
	ctx := context.Background()

	fmt.Printf("\n→ Deleting namespace %s...\n", i.config.Demo.Namespace)
	err := i.clientset.CoreV1().Namespaces().Delete(ctx, i.config.Demo.Namespace, metav1.DeleteOptions{})
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("failed to delete demo namespace: %w", err)
	}

	return nil
}

func isAlreadyExists(err error) bool {
	return err != nil && err.Error() == "already exists"
}

func isNotFound(err error) bool {
	return err != nil && err.Error() == "not found"
}
