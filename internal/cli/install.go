package cli

import (
	"fmt"

	"github.com/cloudcwfranck/kubac/internal/config"
	"github.com/cloudcwfranck/kubac/internal/install"
	"github.com/spf13/cobra"
)

// NewInstallCommand creates the install command
func NewInstallCommand() *cobra.Command {
	var configFile string
	var mode string
	var clusterProfile string

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install kubac platform components",
		Long: `Install the kubac baseline platform stack including:
  - Metrics server (required)
  - Policy engine (Kyverno)
  - Network policies
  - Optional: Prometheus, Ingress, Cert Manager`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Override with CLI flags if provided
			if mode != "" {
				cfg.Mode = mode
			}
			if clusterProfile != "" {
				cfg.ClusterProfile = clusterProfile
			}

			fmt.Printf("Installing kubac platform (mode=%s, profile=%s)...\n", cfg.Mode, cfg.ClusterProfile)

			// Create installer
			installer, err := install.NewInstaller(cfg)
			if err != nil {
				return fmt.Errorf("failed to create installer: %w", err)
			}

			// Run installation
			if err := installer.Install(); err != nil {
				return fmt.Errorf("installation failed: %w", err)
			}

			fmt.Println("\n✓ Installation completed successfully!")
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Run 'kubac verify' to validate the installation")
			fmt.Println("  2. Run 'kubac demo deploy' to deploy a sample application")

			return nil
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "kubac.yaml", "Path to kubac config file")
	cmd.Flags().StringVar(&mode, "mode", "", "Installation mode: gitops or direct (overrides config)")
	cmd.Flags().StringVar(&clusterProfile, "cluster-profile", "", "Cluster profile (overrides config)")

	return cmd
}
