package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	patchFile  string
	scriptFile string
	repos      string
	branch     string
	message    string
)

var applyCmd = &cobra.Command{
	Use:     "apply",
	Short:   "Apply changes across multiple repositories",
	Long:    "Apply changes across multiple git repositories using either a patch file or a script.",
	Example: "cascade apply --script ./update.sh --repos ./repo1,./repo2 --branch refactor-components --message \"Refactor components\"",
	RunE:    runApply,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if patchFile == "" && scriptFile == "" {
			return fmt.Errorf("either --patch or --script must be specified")
		}
		if patchFile != "" && scriptFile != "" {
			return fmt.Errorf("--patch and --script cannot be used together")
		}
		if repos == "" {
			return fmt.Errorf("--repos is required")
		}
		if branch == "" {
			return fmt.Errorf("--branch is required")
		}
		if message == "" {
			return fmt.Errorf("--message is required")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Required flags
	applyCmd.Flags().StringVar(&patchFile, "patch", "", "Path to patch file")
	applyCmd.Flags().StringVar(&scriptFile, "script", "", "Path to executable script")
	applyCmd.Flags().StringVar(&repos, "repos", "", "Comma-separated list of repository paths to modify")
	applyCmd.Flags().StringVar(&branch, "branch", "", "Name for the new branch that will be created")
	applyCmd.Flags().StringVar(&message, "message", "", "Commit message used for the changes")
}

func runApply(cmd *cobra.Command, args []string) error {
	// For now, just print the parameters
	fmt.Printf("Applying changes with:\n")
	if patchFile != "" {
		fmt.Printf("Patch file: %s\n", patchFile)
	} else {
		fmt.Printf("Script file: %s\n", scriptFile)
	}
	fmt.Printf("Repositories: %s\n", repos)
	fmt.Printf("Branch: %s\n", branch)
	fmt.Printf("Message: %s\n", message)
	return nil
}
