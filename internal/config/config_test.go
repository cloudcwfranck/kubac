package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		mode    string
	}{
		{"local-direct", "local", "direct"},
		{"managed-gitops", "managed", "gitops"},
		{"onprem-direct", "onprem", "direct"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig(tt.profile, tt.mode)

			if cfg.ClusterProfile != tt.profile {
				t.Errorf("ClusterProfile = %s, want %s", cfg.ClusterProfile, tt.profile)
			}

			if cfg.Mode != tt.mode {
				t.Errorf("Mode = %s, want %s", cfg.Mode, tt.mode)
			}

			// Check that required components are enabled
			if !cfg.Platform.MetricsServer.Enabled {
				t.Error("MetricsServer should be enabled by default")
			}

			if !cfg.Policy.Enabled {
				t.Error("Policy should be enabled by default")
			}

			if !cfg.NetworkPolicy.Enabled {
				t.Error("NetworkPolicy should be enabled by default")
			}
		})
	}
}

func TestWriteAndLoadConfig(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "kubac-test-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Create a config
	cfg := DefaultConfig("local", "direct")
	cfg.Demo.Namespace = "test-namespace"

	// Write it
	if err := WriteConfig(tmpfile.Name(), cfg); err != nil {
		t.Fatalf("WriteConfig failed: %v", err)
	}

	// Load it back
	loaded, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify
	if loaded.ClusterProfile != "local" {
		t.Errorf("ClusterProfile = %s, want local", loaded.ClusterProfile)
	}

	if loaded.Demo.Namespace != "test-namespace" {
		t.Errorf("Demo.Namespace = %s, want test-namespace", loaded.Demo.Namespace)
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	_, err := LoadConfig("/nonexistent/file.yaml")
	if err == nil {
		t.Error("LoadConfig should fail for nonexistent file")
	}
}
