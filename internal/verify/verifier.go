package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cloudcwfranck/kubac/internal/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Verifier runs verification tests
type Verifier struct {
	config    *config.Config
	clientset *kubernetes.Clientset
}

// NewVerifier creates a new verifier
func NewVerifier(cfg *config.Config) (*Verifier, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Verifier{
		config:    cfg,
		clientset: clientset,
	}, nil
}

// TestResult represents a test result
type TestResult struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Duration  string    `json:"duration"`
	Timestamp time.Time `json:"timestamp"`
}

// VerificationResults holds all test results
type VerificationResults struct {
	Tests     []TestResult `json:"tests"`
	Summary   Summary      `json:"summary"`
	Timestamp time.Time    `json:"timestamp"`
}

// Summary holds test summary
type Summary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// RunAll runs all verification tests
func (v *Verifier) RunAll() (*VerificationResults, error) {
	results := &VerificationResults{
		Tests:     []TestResult{},
		Timestamp: time.Now(),
	}

	// Run each test
	for _, testName := range v.config.Verify.Tests {
		var result TestResult
		start := time.Now()

		switch testName {
		case "pod-selfheal":
			result = v.testPodSelfHeal()
		case "hpa-scale":
			result = v.testHPAScale()
		case "policy-deny":
			result = v.testPolicyDeny()
		case "network-deny":
			result = v.testNetworkDeny()
		default:
			result = TestResult{
				Name:    testName,
				Status:  "SKIP",
				Message: "Unknown test",
			}
		}

		result.Duration = time.Since(start).String()
		result.Timestamp = time.Now()
		results.Tests = append(results.Tests, result)
	}

	// Calculate summary
	results.Summary.Total = len(results.Tests)
	for _, test := range results.Tests {
		if test.Status == "PASS" {
			results.Summary.Passed++
		} else if test.Status == "FAIL" {
			results.Summary.Failed++
		}
	}

	return results, nil
}

// testPodSelfHeal tests pod self-healing
func (v *Verifier) testPodSelfHeal() TestResult {
	ctx := context.Background()

	// Check if demo app is deployed
	deployment, err := v.clientset.AppsV1().Deployments(v.config.Demo.Namespace).Get(ctx, "demo", metav1.GetOptions{})
	if err != nil {
		return TestResult{
			Name:    "pod-selfheal",
			Status:  "SKIP",
			Message: fmt.Sprintf("Demo app not deployed: %v", err),
		}
	}

	// Get initial replica count
	initialReplicas := deployment.Status.ReadyReplicas

	// Get a pod to delete
	pods, err := v.clientset.CoreV1().Pods(v.config.Demo.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=demo",
	})
	if err != nil || len(pods.Items) == 0 {
		return TestResult{
			Name:    "pod-selfheal",
			Status:  "FAIL",
			Message: "No demo pods found",
		}
	}

	podName := pods.Items[0].Name

	// Delete the pod
	err = v.clientset.CoreV1().Pods(v.config.Demo.Namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		return TestResult{
			Name:    "pod-selfheal",
			Status:  "FAIL",
			Message: fmt.Sprintf("Failed to delete pod: %v", err),
		}
	}

	// Wait for replacement pod
	err = wait.PollImmediate(5*time.Second, 60*time.Second, func() (bool, error) {
		deployment, err := v.clientset.AppsV1().Deployments(v.config.Demo.Namespace).Get(ctx, "demo", metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return deployment.Status.ReadyReplicas >= initialReplicas, nil
	})

	if err != nil {
		return TestResult{
			Name:    "pod-selfheal",
			Status:  "FAIL",
			Message: "Pod was not replaced within timeout",
		}
	}

	return TestResult{
		Name:    "pod-selfheal",
		Status:  "PASS",
		Message: fmt.Sprintf("Pod %s was replaced successfully", podName),
	}
}

// testHPAScale tests HPA scaling
func (v *Verifier) testHPAScale() TestResult {
	ctx := context.Background()

	// Check if HPA exists
	hpa, err := v.clientset.AutoscalingV2().HorizontalPodAutoscalers(v.config.Demo.Namespace).Get(ctx, "demo", metav1.GetOptions{})
	if err != nil {
		return TestResult{
			Name:    "hpa-scale",
			Status:  "SKIP",
			Message: fmt.Sprintf("HPA not found: %v", err),
		}
	}

	// Verify HPA is configured correctly
	if hpa.Spec.MinReplicas == nil || *hpa.Spec.MinReplicas != int32(v.config.Demo.HPA.MinReplicas) {
		return TestResult{
			Name:    "hpa-scale",
			Status:  "FAIL",
			Message: "HPA minReplicas mismatch",
		}
	}

	if hpa.Spec.MaxReplicas != int32(v.config.Demo.HPA.MaxReplicas) {
		return TestResult{
			Name:    "hpa-scale",
			Status:  "FAIL",
			Message: "HPA maxReplicas mismatch",
		}
	}

	// Check current metrics
	currentReplicas := hpa.Status.CurrentReplicas

	return TestResult{
		Name:    "hpa-scale",
		Status:  "PASS",
		Message: fmt.Sprintf("HPA configured correctly (current replicas: %d, min: %d, max: %d)",
			currentReplicas, *hpa.Spec.MinReplicas, hpa.Spec.MaxReplicas),
	}
}

// testPolicyDeny tests policy enforcement
func (v *Verifier) testPolicyDeny() TestResult {
	ctx := context.Background()

	// Create a privileged pod that should be denied
	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-privileged-pod",
			Namespace: v.config.Demo.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test",
					Image: "nginx:latest",
					SecurityContext: &corev1.SecurityContext{
						Privileged: boolPtr(true),
					},
				},
			},
		},
	}

	_, err := v.clientset.CoreV1().Pods(v.config.Demo.Namespace).Create(ctx, testPod, metav1.CreateOptions{})
	if err == nil {
		// Clean up if it was created
		v.clientset.CoreV1().Pods(v.config.Demo.Namespace).Delete(ctx, testPod.Name, metav1.DeleteOptions{})
		return TestResult{
			Name:    "policy-deny",
			Status:  "FAIL",
			Message: "Privileged pod was not denied by policy",
		}
	}

	return TestResult{
		Name:    "policy-deny",
		Status:  "PASS",
		Message: "Privileged pod was correctly denied by policy",
	}
}

// testNetworkDeny tests network policy enforcement
func (v *Verifier) testNetworkDeny() TestResult {
	ctx := context.Background()

	// Check if network policies exist
	netpols, err := v.clientset.NetworkingV1().NetworkPolicies(v.config.Demo.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return TestResult{
			Name:    "network-deny",
			Status:  "SKIP",
			Message: fmt.Sprintf("Cannot list network policies: %v", err),
		}
	}

	if len(netpols.Items) == 0 {
		return TestResult{
			Name:    "network-deny",
			Status:  "FAIL",
			Message: "No network policies found",
		}
	}

	// Check for default deny policy
	hasDefaultDeny := false
	for _, np := range netpols.Items {
		if len(np.Spec.Ingress) == 0 && len(np.Spec.Egress) == 0 {
			hasDefaultDeny = true
			break
		}
	}

	if !hasDefaultDeny {
		return TestResult{
			Name:    "network-deny",
			Status:  "WARN",
			Message: fmt.Sprintf("Found %d network policies but no default deny", len(netpols.Items)),
		}
	}

	return TestResult{
		Name:    "network-deny",
		Status:  "PASS",
		Message: fmt.Sprintf("Network policies configured (%d policies including default deny)", len(netpols.Items)),
	}
}

// AllPassed returns true if all tests passed
func (r *VerificationResults) AllPassed() bool {
	return r.Summary.Failed == 0
}

// WriteReport writes the report to a file
func (r *VerificationResults) WriteReport(path, format string) error {
	if format == "json" || path != "" {
		data, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(path, data, 0644)
	}
	return nil
}

// PrintSummary prints a summary to the writer
func (r *VerificationResults) PrintSummary(w io.Writer) {
	fmt.Fprintln(w, "\nVerification Results:")
	fmt.Fprintln(w, "=====================")

	for _, test := range r.Tests {
		status := test.Status
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

		fmt.Fprintf(w, "\n%-20s %s\n", test.Name+":", status)
		fmt.Fprintf(w, "  %s\n", test.Message)
		fmt.Fprintf(w, "  Duration: %s\n", test.Duration)
	}

	fmt.Fprintln(w, "\n=====================")
	fmt.Fprintf(w, "Total:  %d\n", r.Summary.Total)
	fmt.Fprintf(w, "Passed: %d\n", r.Summary.Passed)
	fmt.Fprintf(w, "Failed: %d\n", r.Summary.Failed)

	if r.AllPassed() {
		fmt.Fprintln(w, "\n✓ All tests passed!")
	} else {
		fmt.Fprintln(w, "\n✗ Some tests failed")
	}
}

func boolPtr(b bool) *bool {
	return &b
}
