package git

import (
	"fmt"
	"os/exec"
)

func CheckoutBranch(repoPath string, branch string) error {
	cmd := exec.Command("git", "checkout", "-B", branch)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error checking out branch: %w\n%s", err, string(output))
	}
	return nil
}

func ApplyPatch(repoPath string, patchFile string) error {
	cmd := exec.Command("git", "apply", patchFile)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("patch apply failed: %w\n%s", err, string(output))
	}
	return nil
}

func CommitChanges(repoPath string, message string) error {
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = repoPath
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\n%s", err, string(output))
	}

	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = repoPath
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git commit failed: %w\n%s", err, string(output))
	}
	return nil
}

func IsGitRepository(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--git-dir")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w\n%s", err, string(output))
	}
	return nil
}
