package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudcwfranck/kubac/internal/config"
	"github.com/spf13/cobra"
)

// NewInitCommand creates the init command
func NewInitCommand() *cobra.Command {
	var clusterProfile string
	var mode string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a kubac project (creates kubac.yaml config)",
		Long: `Initialize a new kubac project by creating a kubac.yaml configuration file
with sensible defaults based on the cluster profile.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if kubac.yaml already exists
			if _, err := os.Stat("kubac.yaml"); err == nil {
				return fmt.Errorf("kubac.yaml already exists, refusing to overwrite")
			}

			// Create default config
			cfg := config.DefaultConfig(clusterProfile, mode)

			// Write config file
			if err := config.WriteConfig("kubac.yaml", cfg); err != nil {
				return fmt.Errorf("failed to write kubac.yaml: %w", err)
			}

			fmt.Printf("✓ Created kubac.yaml with profile '%s' and mode '%s'\n", clusterProfile, mode)

			// If gitops mode, create directory structure
			if mode == "gitops" {
				gitopsPath := cfg.GitOps.Path
				if gitopsPath == "" {
					gitopsPath = "clusters/my-cluster"
				}

				dirs := []string{
					gitopsPath,
					filepath.Join(gitopsPath, "flux-system"),
					filepath.Join(gitopsPath, "platform"),
					filepath.Join(gitopsPath, "policies"),
					filepath.Join(gitopsPath, "apps"),
				}

				for _, dir := range dirs {
					if err := os.MkdirAll(dir, 0755); err != nil {
						return fmt.Errorf("failed to create directory %s: %w", dir, err)
					}
				}

				fmt.Printf("✓ Created GitOps directory structure at %s\n", gitopsPath)
			}

			fmt.Println("\nNext steps:")
			fmt.Println("  1. Review and customize kubac.yaml")
			fmt.Println("  2. Run 'kubac doctor' to verify cluster access")
			fmt.Println("  3. Run 'kubac install' to deploy the platform")

			return nil
		},
	}

	cmd.Flags().StringVar(&clusterProfile, "cluster-profile", "local", "Cluster profile: local, managed, onprem")
	cmd.Flags().StringVar(&mode, "mode", "direct", "Installation mode: gitops or direct")

	return cmd
}
