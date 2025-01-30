package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0" // This can be set during build time
)

var rootCmd = &cobra.Command{
	Use:   "cascade",
	Short: "cascade - apply changes across multiple git repositories",
	Long: `Cascade is a CLI tool that helps you apply changes to multiple git repositories.
It can fetch latest changes, create branches, apply patches or run scripts,
and create pull requests automatically.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
