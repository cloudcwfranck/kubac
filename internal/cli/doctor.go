package cli

import (
	"fmt"
	"os"

	"github.com/cloudcwfranck/kubac/internal/doctor"
	"github.com/spf13/cobra"
)

// NewDoctorCommand creates the doctor command
func NewDoctorCommand() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run preflight checks",
		Long: `Run preflight checks to verify:
  - kubectl is installed and accessible
  - Current context is set
  - Cluster is reachable
  - User has required permissions
  - Cluster meets minimum requirements`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Running kubac preflight checks...")

			// Create doctor
			doc := doctor.NewDoctor()

			// Run checks
			results := doc.RunChecks()

			// Print results
			if output == "json" {
				if err := results.PrintJSON(os.Stdout); err != nil {
					return fmt.Errorf("failed to print JSON: %w", err)
				}
			} else {
				results.PrintText(os.Stdout)
			}

			// Exit with error if any critical checks failed
			if !results.AllPassed() {
				return fmt.Errorf("some preflight checks failed")
			}

			fmt.Println("\n✓ All preflight checks passed!")

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "text", "Output format: text or json")

	return cmd
}
