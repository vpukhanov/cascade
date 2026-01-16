package cmd

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/vpukhanov/cascade/internal/git"
	applog "github.com/vpukhanov/cascade/internal/log"
	"github.com/vpukhanov/cascade/internal/validation"

	"github.com/spf13/cobra"
)

var (
	patchFile  string
	scriptFile string
	branch     string
	message    string
	baseBranch string
	pullLatest bool
	push       bool
	noVerify   bool
	stash      bool

	gitCheckoutBranch         = git.CheckoutBranch
	gitCheckoutExistingBranch = git.CheckoutExistingBranch
	gitApplyPatch             = git.ApplyPatch
	gitCommitChanges          = git.CommitChanges
	gitExecuteScript          = git.ExecuteScript
	gitPullLatest             = git.PullLatest
	gitPushChanges            = git.PushChanges
	gitStashChanges           = git.StashChanges
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
		if runtime.GOOS == "windows" && scriptFile != "" {
			return fmt.Errorf("--script option is not supported on Windows")
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
	applyCmd.Flags().BoolVar(&push, "push", false, "Push new branch to origin after applying the changes")
	applyCmd.Flags().BoolVar(&noVerify, "no-verify", false, "Skip git commit and push hooks")
	applyCmd.Flags().BoolVar(&stash, "stash", false, "Stash tracked and untracked changes before applying changes")
}

// ResetFlags resets all global flag variables to their zero values
func ResetFlags() {
	patchFile = ""
	scriptFile = ""
	branch = ""
	message = ""
	baseBranch = ""
	pullLatest = false
	push = false
	noVerify = false
	stash = false
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

	type repoResult struct {
		repo string
		err  error
	}

	results := make([]repoResult, 0, len(args))
	var logger *applog.ApplyLogger

	for _, repoPath := range args {
		var repoErr error

		if stash {
			if err := gitStashChanges(repoPath); err != nil {
				repoErr = fmt.Errorf("stash failed: %w", err)
			}
		}

		// If base branch is specified, check it out
		if repoErr == nil && baseBranch != "" {
			if err := gitCheckoutExistingBranch(repoPath, baseBranch); err != nil {
				repoErr = fmt.Errorf("base branch checkout failed: %w", err)
			}
		}

		// Pull latest changes if requested
		if repoErr == nil && pullLatest {
			if err := gitPullLatest(repoPath); err != nil {
				repoErr = fmt.Errorf("pull latest failed: %w", err)
			}
		}

		// Create and checkout the new branch
		if repoErr == nil {
			if err := gitCheckoutBranch(repoPath, branch); err != nil {
				repoErr = fmt.Errorf("branch checkout failed: %w", err)
			}
		}

		if repoErr == nil {
			if scriptFile != "" {
				if err := gitExecuteScript(repoPath, absPath); err != nil {
					repoErr = fmt.Errorf("script execution failed: %w", err)
				}
			} else {
				if err := gitApplyPatch(repoPath, absPath); err != nil {
					repoErr = fmt.Errorf("patch application failed: %w", err)
				}
			}
		}

		if repoErr == nil {
			if err := gitCommitChanges(repoPath, message, noVerify); err != nil {
				repoErr = fmt.Errorf("commit failed: %w", err)
			}
		}

		if repoErr == nil && push {
			if err := gitPushChanges(repoPath, branch, noVerify); err != nil {
				repoErr = fmt.Errorf("push failed: %w", err)
			}
		}

		if repoErr != nil {
			if logger == nil {
				var err error
				logger, err = applog.NewApplyLogger()
				if err != nil {
					return fmt.Errorf("failed to create error log: %w", err)
				}
			}
			logger.LogRepoError(repoPath, repoErr)
		}

		results = append(results, repoResult{repo: repoPath, err: repoErr})
	}

	// Print results
	fmt.Println()
	for _, result := range results {
		status := "ok"
		if result.err != nil {
			status = "fail"
		}
		fmt.Printf("%-4s %s\n", status, result.repo)
	}

	if logger != nil {
		fmt.Printf("\nError details: %s\n", logger.Path())
		_ = logger.Close()
	}

	return nil
}
