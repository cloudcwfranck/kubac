package cli

import (
	"fmt"
	"os"

	"github.com/cloudcwfranck/kubac/internal/config"
	"github.com/cloudcwfranck/kubac/internal/verify"
	"github.com/spf13/cobra"
)

// NewVerifyCommand creates the verify command
func NewVerifyCommand() *cobra.Command {
	var configFile string
	var output string
	var reportPath string

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Run verification tests against the cluster",
		Long: `Run the kubac verification suite to validate:
  - Pod self-healing (delete pod, ensure replacement)
  - HPA scaling (load test, observe replicas)
  - Policy enforcement (forbidden pod rejection)
  - Network policies (default deny enforcement)

Results are written to a report file in JSON format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Println("Running kubac verification suite...")

			// Create verifier
			verifier, err := verify.NewVerifier(cfg)
			if err != nil {
				return fmt.Errorf("failed to create verifier: %w", err)
			}

			// Run verification
			results, err := verifier.RunAll()
			if err != nil {
				return fmt.Errorf("verification failed: %w", err)
			}

			// Determine report path
			if reportPath == "" {
				reportPath = "kubac-verify-report.json"
			}

			// Write report
			if err := results.WriteReport(reportPath, output); err != nil {
				return fmt.Errorf("failed to write report: %w", err)
			}

			// Print summary
			results.PrintSummary(os.Stdout)

			fmt.Printf("\n✓ Full report written to: %s\n", reportPath)

			// Exit with error if any tests failed
			if !results.AllPassed() {
				return fmt.Errorf("some verification tests failed")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "kubac.yaml", "Path to kubac config file")
	cmd.Flags().StringVar(&output, "output", "text", "Output format: text or json")
	cmd.Flags().StringVar(&reportPath, "report", "", "Path to write report (default: kubac-verify-report.json)")

	return cmd
}
