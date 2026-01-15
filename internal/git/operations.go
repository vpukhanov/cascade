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

func CommitChanges(repoPath string, message string, noVerify bool) error {
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = repoPath
	if output, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add failed: %w\n%s", err, string(output))
	}

	commitArgs := []string{"commit"}
	if noVerify {
		commitArgs = append(commitArgs, "--no-verify")
	}
	commitArgs = append(commitArgs, "-m", message)
	commitCmd := exec.Command("git", commitArgs...)
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

func ExecuteScript(repoPath string, scriptPath string) error {
	cmd := exec.Command(scriptPath)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("script execution failed: %w\n%s", err, string(output))
	}
	return nil
}

func CheckoutExistingBranch(repoPath string, branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error checking out branch: %w\n%s", err, string(output))
	}
	return nil
}

func PullLatest(repoPath string) error {
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error pulling latest changes: %w\n%s", err, string(output))
	}
	return nil
}

func PushChanges(repoPath string, branch string, noVerify bool) error {
	pushArgs := []string{"push"}
	if noVerify {
		pushArgs = append(pushArgs, "--no-verify")
	}
	pushArgs = append(pushArgs, "-u", "origin", branch)
	cmd := exec.Command("git", pushArgs...)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error pushing changes: %w\n%s", err, string(output))
	}
	return nil
}
