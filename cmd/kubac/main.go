package main

import (
	"fmt"
	"os"

	"github.com/cloudcwfranck/kubac/internal/cli"
)

const version = "0.1.0"

func main() {
	rootCmd := cli.NewRootCommand(version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
