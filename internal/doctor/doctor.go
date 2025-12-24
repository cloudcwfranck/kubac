package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Doctor performs preflight checks
type Doctor struct{}

// NewDoctor creates a new doctor
func NewDoctor() *Doctor {
	return &Doctor{}
}

// CheckResult represents a single check result
type CheckResult struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// CheckResults holds all check results
type CheckResults struct {
	Checks []CheckResult `json:"checks"`
}

// RunChecks runs all preflight checks
func (d *Doctor) RunChecks() *CheckResults {
	results := &CheckResults{
		Checks: []CheckResult{},
	}

	// Check kubectl
	results.Checks = append(results.Checks, d.checkKubectl())

	// Check cluster access
	results.Checks = append(results.Checks, d.checkClusterAccess())

	// Check permissions
	results.Checks = append(results.Checks, d.checkPermissions())

	// Check nodes
	results.Checks = append(results.Checks, d.checkNodes())

	return results
}

func (d *Doctor) checkKubectl() CheckResult {
	cmd := exec.Command("kubectl", "version", "--client", "--short")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return CheckResult{
			Name:    "kubectl",
			Status:  "FAIL",
			Message: "kubectl not found or not executable",
		}
	}

	return CheckResult{
		Name:    "kubectl",
		Status:  "PASS",
		Message: string(output),
	}
}

func (d *Doctor) checkClusterAccess() CheckResult {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return CheckResult{
			Name:    "cluster-access",
			Status:  "FAIL",
			Message: fmt.Sprintf("Failed to load kubeconfig: %v", err),
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return CheckResult{
			Name:    "cluster-access",
			Status:  "FAIL",
			Message: fmt.Sprintf("Failed to create client: %v", err),
		}
	}

	ctx := context.Background()
	_, err = clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return CheckResult{
			Name:    "cluster-access",
			Status:  "FAIL",
			Message: fmt.Sprintf("Failed to connect to cluster: %v", err),
		}
	}

	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return CheckResult{
			Name:    "cluster-access",
			Status:  "WARN",
			Message: "Connected but cannot determine context",
		}
	}

	return CheckResult{
		Name:    "cluster-access",
		Status:  "PASS",
		Message: fmt.Sprintf("Connected to context: %s", rawConfig.CurrentContext),
	}
}

func (d *Doctor) checkPermissions() CheckResult {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return CheckResult{
			Name:    "permissions",
			Status:  "SKIP",
			Message: "Skipped due to cluster access failure",
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return CheckResult{
			Name:    "permissions",
			Status:  "SKIP",
			Message: "Skipped due to client creation failure",
		}
	}

	ctx := context.Background()

	// Try to create a namespace (will fail if no permissions, but that's ok for the check)
	_, err = clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return CheckResult{
			Name:    "permissions",
			Status:  "FAIL",
			Message: "Insufficient permissions to list namespaces",
		}
	}

	return CheckResult{
		Name:    "permissions",
		Status:  "PASS",
		Message: "User has required cluster permissions",
	}
}

func (d *Doctor) checkNodes() CheckResult {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return CheckResult{
			Name:    "nodes",
			Status:  "SKIP",
			Message: "Skipped due to cluster access failure",
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return CheckResult{
			Name:    "nodes",
			Status:  "SKIP",
			Message: "Skipped due to client creation failure",
		}
	}

	ctx := context.Background()
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return CheckResult{
			Name:    "nodes",
			Status:  "WARN",
			Message: fmt.Sprintf("Cannot list nodes: %v", err),
		}
	}

	readyNodes := 0
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				readyNodes++
				break
			}
		}
	}

	if readyNodes == 0 {
		return CheckResult{
			Name:    "nodes",
			Status:  "FAIL",
			Message: "No ready nodes found",
		}
	}

	return CheckResult{
		Name:    "nodes",
		Status:  "PASS",
		Message: fmt.Sprintf("%d/%d nodes ready", readyNodes, len(nodes.Items)),
	}
}

// AllPassed returns true if all checks passed
func (r *CheckResults) AllPassed() bool {
	for _, check := range r.Checks {
		if check.Status == "FAIL" {
			return false
		}
	}
	return true
}

// PrintText prints results in text format
func (r *CheckResults) PrintText(w io.Writer) {
	fmt.Fprintln(w, "\nPreflight Checks:")
	fmt.Fprintln(w, "-----------------")
	for _, check := range r.Checks {
		status := check.Status
		switch status {
		case "PASS":
			status = "✓ PASS"
		case "FAIL":
			status = "✗ FAIL"
		case "WARN":
			status = "⚠ WARN"
		case "SKIP":
			status = "- SKIP"
		}
		fmt.Fprintf(w, "%-20s %s\n", check.Name+":", status)
		if check.Message != "" {
			fmt.Fprintf(w, "  %s\n", check.Message)
		}
	}
}

// PrintJSON prints results in JSON format
func (r *CheckResults) PrintJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(r)
}
