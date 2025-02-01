package cmd

import (
	"fmt"
	"path/filepath"

	"cascade/internal/git"
	"cascade/internal/validation"

	"github.com/spf13/cobra"
)

var (
	patchFile  string
	scriptFile string
	branch     string
	message    string
	baseBranch string
	pullLatest bool

	gitCheckoutBranch         = git.CheckoutBranch
	gitCheckoutExistingBranch = git.CheckoutExistingBranch
	gitApplyPatch             = git.ApplyPatch
	gitCommitChanges          = git.CommitChanges
	gitExecuteScript          = git.ExecuteScript
	gitPullLatest             = git.PullLatest
)

var applyCmd = &cobra.Command{
	Use:     "apply [repositories...]",
	Short:   "Apply changes across multiple repositories",
	Long:    "Apply changes across multiple git repositories using either a patch file or a script.",
	Example: `cascade apply --patch ./changes.patch --branch update-logging --message "Update logging" ./repo1 ./repo2`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    runApply,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if patchFile == "" && scriptFile == "" {
			return fmt.Errorf("either --patch or --script must be specified")
		}
		if patchFile != "" && scriptFile != "" {
			return fmt.Errorf("--patch and --script cannot be used together")
		}
		if branch == "" {
			return fmt.Errorf("--branch is required")
		}
		if message == "" {
			return fmt.Errorf("--message is required")
		}

		if patchFile != "" {
			if err := validation.ValidateFile(patchFile, "patch"); err != nil {
				return err
			}
		}
		if scriptFile != "" {
			if err := validation.ValidateFile(scriptFile, "script"); err != nil {
				return err
			}
		}

		for _, repo := range args {
			if err := validation.ValidateGitRepo(repo); err != nil {
				return err
			}
		}

		if err := validation.ValidateBranchName(branch); err != nil {
			return fmt.Errorf("invalid target branch name: %w", err)
		}

		if baseBranch != "" {
			if err := validation.ValidateBranchName(baseBranch); err != nil {
				return fmt.Errorf("invalid base branch name: %w", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Required flags
	applyCmd.Flags().StringVar(&patchFile, "patch", "", "Path to patch file")
	applyCmd.Flags().StringVar(&scriptFile, "script", "", "Path to executable script")
	applyCmd.Flags().StringVar(&branch, "branch", "", "Name for the new branch that will be created")
	applyCmd.Flags().StringVar(&message, "message", "", "Commit message used for the changes")

	// Optional flags
	applyCmd.Flags().StringVar(&baseBranch, "base-branch", "", "Branch to check out and apply changes to")
	applyCmd.Flags().BoolVar(&pullLatest, "pull", false, "Pull latest changes from remote before applying changes")
}

// ResetFlags resets all global flag variables to their zero values
func ResetFlags() {
	patchFile = ""
	scriptFile = ""
	branch = ""
	message = ""
	baseBranch = ""
	pullLatest = false
}

func runApply(cmd *cobra.Command, args []string) error {
	var absPath string
	var err error

	if scriptFile != "" {
		absPath, err = filepath.Abs(scriptFile)
	} else {
		absPath, err = filepath.Abs(patchFile)
	}
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	results := struct {
		success []string
		errors  map[string]error
	}{
		errors: make(map[string]error),
	}

	for _, repoPath := range args {
		// If base branch is specified, check it out first
		if baseBranch != "" {
			if err := gitCheckoutExistingBranch(repoPath, baseBranch); err != nil {
				results.errors[repoPath] = fmt.Errorf("base branch checkout failed: %w", err)
				continue
			}
		}

		// Pull latest changes if requested
		if pullLatest {
			if err := gitPullLatest(repoPath); err != nil {
				results.errors[repoPath] = fmt.Errorf("pull latest failed: %w", err)
				continue
			}
		}

		// Create and checkout the new branch
		if err := gitCheckoutBranch(repoPath, branch); err != nil {
			results.errors[repoPath] = fmt.Errorf("branch checkout failed: %w", err)
			continue
		}

		if scriptFile != "" {
			if err := gitExecuteScript(repoPath, absPath); err != nil {
				results.errors[repoPath] = fmt.Errorf("script execution failed: %w", err)
				continue
			}
		} else {
			if err := gitApplyPatch(repoPath, absPath); err != nil {
				results.errors[repoPath] = fmt.Errorf("patch application failed: %w", err)
				continue
			}
		}

		if err := gitCommitChanges(repoPath, message); err != nil {
			results.errors[repoPath] = fmt.Errorf("commit failed: %w", err)
			continue
		}

		results.success = append(results.success, repoPath)
	}

	// Print results
	fmt.Printf("\nSuccessfully processed (%d):\n", len(results.success))
	for _, repo := range results.success {
		fmt.Printf("  ✓ %s\n", repo)
	}

	fmt.Printf("\nFailed (%d):\n", len(results.errors))
	for repo, err := range results.errors {
		fmt.Printf("  ✗ %s: %v\n", repo, err)
	}

	return nil
}
