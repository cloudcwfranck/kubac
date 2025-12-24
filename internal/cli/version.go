package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewVersionCommand creates the version command
func NewVersionCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print kubac version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("kubac version %s\n", version)
		},
	}
}
