package cmd

import (
	"github.com/spf13/cobra"
)

var (
	version = "dev" // This can be set during build time
)

var rootCmd = &cobra.Command{
	Use:   "cascade",
	Short: "cascade - apply changes across multiple git repositories",
	Long: `Cascade is a CLI tool that helps you apply changes to multiple git repositories.
It can fetch latest changes, create branches, apply patches, run scripts, or execute commands,
and create pull requests automatically.`,
	Version: version,
}

func Execute() error {
	return rootCmd.Execute()
}
