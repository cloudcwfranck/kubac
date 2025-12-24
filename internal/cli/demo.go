package cli

import (
	"fmt"

	"github.com/cloudcwfranck/kubac/internal/config"
	"github.com/cloudcwfranck/kubac/internal/install"
	"github.com/spf13/cobra"
)

// NewDemoCommand creates the demo command with subcommands
func NewDemoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Manage demo application",
		Long:  `Deploy, test, and manage the kubac demo application.`,
	}

	cmd.AddCommand(newDemoDeployCommand())
	cmd.AddCommand(newDemoLoadCommand())
	cmd.AddCommand(newDemoChaosCommand())
	cmd.AddCommand(newDemoCleanupCommand())

	return cmd
}

func newDemoDeployCommand() *cobra.Command {
	var configFile string

	return &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the demo application",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Println("Deploying kubac demo application...")

			installer, err := install.NewInstaller(cfg)
			if err != nil {
				return fmt.Errorf("failed to create installer: %w", err)
			}

			if err := installer.DeployDemo(); err != nil {
				return fmt.Errorf("demo deployment failed: %w", err)
			}

			fmt.Println("\n✓ Demo application deployed successfully!")
			fmt.Printf("  Namespace: %s\n", cfg.Demo.Namespace)
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Run 'kubac demo load' to test HPA scaling")
			fmt.Println("  2. Run 'kubac demo chaos' to test self-healing")

			return nil
		},
	}
}

func newDemoLoadCommand() *cobra.Command {
	var configFile string
	var duration string
	var requests int

	cmd := &cobra.Command{
		Use:   "load",
		Short: "Generate load to test HPA scaling",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Printf("Generating load (duration=%s, requests=%d)...\n", duration, requests)

			installer, err := install.NewInstaller(cfg)
			if err != nil {
				return fmt.Errorf("failed to create installer: %w", err)
			}

			if err := installer.RunLoadTest(duration, requests); err != nil {
				return fmt.Errorf("load test failed: %w", err)
			}

			fmt.Println("\n✓ Load test completed!")
			fmt.Println("  Check HPA scaling with: kubectl get hpa -n " + cfg.Demo.Namespace)

			return nil
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "kubac.yaml", "Path to kubac config file")
	cmd.Flags().StringVar(&duration, "duration", "60s", "Load test duration")
	cmd.Flags().IntVar(&requests, "requests", 100, "Requests per second")

	return cmd
}

func newDemoChaosCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "chaos",
		Short: "Run chaos tests to validate self-healing",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Println("Running chaos tests...")

			installer, err := install.NewInstaller(cfg)
			if err != nil {
				return fmt.Errorf("failed to create installer: %w", err)
			}

			if err := installer.RunChaosTest(); err != nil {
				return fmt.Errorf("chaos test failed: %w", err)
			}

			fmt.Println("\n✓ Chaos test completed!")

			return nil
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "kubac.yaml", "Path to kubac config file")

	return cmd
}

func newDemoCleanupCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Remove the demo application",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Println("Cleaning up demo application...")

			installer, err := install.NewInstaller(cfg)
			if err != nil {
				return fmt.Errorf("failed to create installer: %w", err)
			}

			if err := installer.CleanupDemo(); err != nil {
				return fmt.Errorf("demo cleanup failed: %w", err)
			}

			fmt.Println("\n✓ Demo application removed!")

			return nil
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "kubac.yaml", "Path to kubac config file")

	return cmd
}
