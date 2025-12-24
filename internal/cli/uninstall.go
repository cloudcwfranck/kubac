package cli

import (
	"fmt"

	"github.com/cloudcwfranck/kubac/internal/config"
	"github.com/cloudcwfranck/kubac/internal/install"
	"github.com/spf13/cobra"
)

// NewUninstallCommand creates the uninstall command
func NewUninstallCommand() *cobra.Command {
	var configFile string
	var force bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall kubac platform components",
		Long:  `Safely remove kubac platform components from the cluster.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if !force {
				fmt.Println("WARNING: This will remove all kubac components from the cluster.")
				fmt.Print("Are you sure? (yes/no): ")
				var response string
				fmt.Scanln(&response)
				if response != "yes" {
					fmt.Println("Uninstall cancelled.")
					return nil
				}
			}

			fmt.Println("Uninstalling kubac platform...")

			// Create installer
			installer, err := install.NewInstaller(cfg)
			if err != nil {
				return fmt.Errorf("failed to create installer: %w", err)
			}

			// Run uninstall
			if err := installer.Uninstall(); err != nil {
				return fmt.Errorf("uninstall failed: %w", err)
			}

			fmt.Println("\n✓ Uninstall completed successfully!")

			return nil
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "kubac.yaml", "Path to kubac config file")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}
