package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the kubac configuration
type Config struct {
	ClusterProfile string          `yaml:"clusterProfile"`
	Mode           string          `yaml:"mode"`
	GitOps         GitOpsConfig    `yaml:"gitops"`
	Platform       PlatformConfig  `yaml:"platform"`
	Policy         PolicyConfig    `yaml:"policy"`
	NetworkPolicy  NetworkPolicy   `yaml:"networkPolicy"`
	Autoscaling    AutoscalingConfig `yaml:"autoscaling"`
	Demo           DemoConfig      `yaml:"demo"`
	Verify         VerifyConfig    `yaml:"verify"`
}

// GitOpsConfig holds GitOps settings
type GitOpsConfig struct {
	Provider string `yaml:"provider"`
	RepoURL  string `yaml:"repoURL"`
	Branch   string `yaml:"branch"`
	Path     string `yaml:"path"`
}

// PlatformConfig holds platform component settings
type PlatformConfig struct {
	MetricsServer     ComponentConfig `yaml:"metricsServer"`
	KubeStateMetrics  ComponentConfig `yaml:"kubeStateMetrics"`
	PrometheusStack   ComponentConfig `yaml:"prometheusStack"`
	Ingress           IngressConfig   `yaml:"ingress"`
	CertManager       ComponentConfig `yaml:"certManager"`
}

// ComponentConfig holds component settings
type ComponentConfig struct {
	Enabled bool   `yaml:"enabled"`
	Version string `yaml:"version"`
}

// IngressConfig holds ingress settings
type IngressConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Provider string `yaml:"provider"`
	Version  string `yaml:"version"`
}

// PolicyConfig holds policy engine settings
type PolicyConfig struct {
	Enabled              bool     `yaml:"enabled"`
	Provider             string   `yaml:"provider"`
	Version              string   `yaml:"version"`
	PodSecurityStandard  string   `yaml:"podSecurityStandard"`
	CustomPolicies       []string `yaml:"customPolicies"`
}

// NetworkPolicy holds network policy settings
type NetworkPolicy struct {
	Enabled          bool     `yaml:"enabled"`
	DefaultDeny      bool     `yaml:"defaultDeny"`
	SystemNamespaces []string `yaml:"systemNamespaces"`
}

// AutoscalingConfig holds autoscaling settings
type AutoscalingConfig struct {
	HPA            HPAConfig            `yaml:"hpa"`
	NodeAutoscaler NodeAutoscalerConfig `yaml:"nodeAutoscaler"`
}

// HPAConfig holds HPA settings
type HPAConfig struct {
	Enabled bool `yaml:"enabled"`
}

// NodeAutoscalerConfig holds node autoscaler settings
type NodeAutoscalerConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Provider string `yaml:"provider"`
	Version  string `yaml:"version"`
}

// DemoConfig holds demo app settings
type DemoConfig struct {
	Namespace string        `yaml:"namespace"`
	Replicas  int           `yaml:"replicas"`
	Resources ResourcesSpec `yaml:"resources"`
	HPA       HPASpec       `yaml:"hpa"`
	PDB       PDBSpec       `yaml:"pdb"`
}

// ResourcesSpec holds resource requirements
type ResourcesSpec struct {
	Requests ResourceList `yaml:"requests"`
	Limits   ResourceList `yaml:"limits"`
}

// ResourceList holds CPU and memory
type ResourceList struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

// HPASpec holds HPA specification
type HPASpec struct {
	MinReplicas           int `yaml:"minReplicas"`
	MaxReplicas           int `yaml:"maxReplicas"`
	TargetCPUUtilization  int `yaml:"targetCPUUtilization"`
}

// PDBSpec holds PDB specification
type PDBSpec struct {
	MinAvailable int `yaml:"minAvailable"`
}

// VerifyConfig holds verification settings
type VerifyConfig struct {
	Timeout  string   `yaml:"timeout"`
	Parallel bool     `yaml:"parallel"`
	Tests    []string `yaml:"tests"`
}

// DefaultConfig returns a default configuration
func DefaultConfig(profile, mode string) *Config {
	return &Config{
		ClusterProfile: profile,
		Mode:           mode,
		GitOps: GitOpsConfig{
			Provider: "flux",
			RepoURL:  "",
			Branch:   "main",
			Path:     "clusters/my-cluster",
		},
		Platform: PlatformConfig{
			MetricsServer: ComponentConfig{
				Enabled: true,
				Version: "v0.7.0",
			},
			KubeStateMetrics: ComponentConfig{
				Enabled: true,
				Version: "v2.10.1",
			},
			PrometheusStack: ComponentConfig{
				Enabled: false,
				Version: "v55.5.0",
			},
			Ingress: IngressConfig{
				Enabled:  false,
				Provider: "nginx",
				Version:  "v1.9.5",
			},
			CertManager: ComponentConfig{
				Enabled: false,
				Version: "v1.13.3",
			},
		},
		Policy: PolicyConfig{
			Enabled:             true,
			Provider:            "kyverno",
			Version:             "v1.11.4",
			PodSecurityStandard: "restricted",
			CustomPolicies: []string{
				"require-non-root-user",
				"require-ro-rootfs",
				"disallow-privileged",
			},
		},
		NetworkPolicy: NetworkPolicy{
			Enabled:     true,
			DefaultDeny: true,
			SystemNamespaces: []string{
				"kube-system",
				"kube-public",
				"kubac-system",
			},
		},
		Autoscaling: AutoscalingConfig{
			HPA: HPAConfig{
				Enabled: true,
			},
			NodeAutoscaler: NodeAutoscalerConfig{
				Enabled:  false,
				Provider: "cluster-autoscaler",
				Version:  "v1.29.0",
			},
		},
		Demo: DemoConfig{
			Namespace: "kubac-demo",
			Replicas:  2,
			Resources: ResourcesSpec{
				Requests: ResourceList{
					CPU:    "100m",
					Memory: "128Mi",
				},
				Limits: ResourceList{
					CPU:    "200m",
					Memory: "256Mi",
				},
			},
			HPA: HPASpec{
				MinReplicas:          2,
				MaxReplicas:          10,
				TargetCPUUtilization: 50,
			},
			PDB: PDBSpec{
				MinAvailable: 1,
			},
		},
		Verify: VerifyConfig{
			Timeout:  "300s",
			Parallel: true,
			Tests: []string{
				"pod-selfheal",
				"hpa-scale",
				"policy-deny",
				"network-deny",
			},
		},
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// WriteConfig writes configuration to a file
func WriteConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
