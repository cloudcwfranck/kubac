package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root kubac command
func NewRootCommand(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubac",
		Short: "Kubernetes Accelerator (Cloud-Agnostic)",
		Long: `kubac is a cloud-agnostic Kubernetes accelerator that installs a secure,
autoscaling, self-healing "baseline platform" on any cluster using GitOps,
with opinionated defaults, policy enforcement, and verifiable operational tests.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add subcommands
	cmd.AddCommand(NewVersionCommand(version))
	cmd.AddCommand(NewInitCommand())
	cmd.AddCommand(NewInstallCommand())
	cmd.AddCommand(NewVerifyCommand())
	cmd.AddCommand(NewDoctorCommand())
	cmd.AddCommand(NewUninstallCommand())
	cmd.AddCommand(NewDemoCommand())

	return cmd
}
