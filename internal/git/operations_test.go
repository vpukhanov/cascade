package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCheckoutBranch(t *testing.T) {
	t.Run("create new branch", func(t *testing.T) {
		repoPath := createTestRepo(t)
		branch := "new-feature"

		err := CheckoutBranch(repoPath, branch)
		if err != nil {
			t.Fatalf("CheckoutBranch failed: %v", err)
		}

		current := currentBranch(t, repoPath)
		if current != branch {
			t.Errorf("Expected branch %q, got %q", branch, current)
		}
	})

	t.Run("switch existing branch", func(t *testing.T) {
		repoPath := createTestRepo(t)
		initialBranch := "main"
		newBranch := "develop"

		// Create and checkout new branch
		runGit(t, repoPath, "checkout", "-b", newBranch)

		// Switch back to initial branch using our function
		err := CheckoutBranch(repoPath, initialBranch)
		if err != nil {
			t.Fatalf("CheckoutBranch failed: %v", err)
		}

		current := currentBranch(t, repoPath)
		if current != initialBranch {
			t.Errorf("Expected branch %q, got %q", initialBranch, current)
		}
	})
}

func TestApplyPatch(t *testing.T) {
	repoPath := createTestRepo(t)
	patchFile := createTestPatch(t)

	err := ApplyPatch(repoPath, patchFile)
	if err != nil {
		t.Fatalf("ApplyPatch failed: %v", err)
	}

	// Verify file was created by patch
	patchedFile := filepath.Join(repoPath, "testfile.txt")
	if _, err := os.Stat(patchedFile); os.IsNotExist(err) {
		t.Error("Patch did not create expected file")
	}
}

func TestCommitChanges(t *testing.T) {
	repoPath := createTestRepo(t)
	commitMessage := "Add new feature\n"

	// Create test file
	err := os.WriteFile(filepath.Join(repoPath, "test.txt"), []byte("content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = CommitChanges(repoPath, commitMessage, false)
	if err != nil {
		t.Fatalf("CommitChanges failed: %v", err)
	}

	// Verify commit exists
	log := runGit(t, repoPath, "log", "-1", "--pretty=%s")
	if log != commitMessage {
		t.Errorf("Expected commit message %q, got %q", commitMessage, log)
	}
}

func TestCommitChangesNoVerifySkipsHook(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("git hook shell scripts are not supported on Windows in this test")
	}

	repoPath := createTestRepo(t)
	commitMessage := "Skip hooks\n"

	// Create a pre-commit hook that would fail if executed.
	hookPath := filepath.Join(repoPath, ".git", "hooks", "pre-commit")
	markerPath := filepath.Join(repoPath, "hook_ran")
	hookContents := "#!/bin/sh\n" +
		"echo ran > \"" + markerPath + "\"\n" +
		"exit 1\n"
	if err := os.WriteFile(hookPath, []byte(hookContents), 0755); err != nil {
		t.Fatalf("failed to write pre-commit hook: %v", err)
	}

	// Create test file.
	err := os.WriteFile(filepath.Join(repoPath, "test.txt"), []byte("content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = CommitChanges(repoPath, commitMessage, true)
	if err != nil {
		t.Fatalf("CommitChanges with --no-verify failed: %v", err)
	}

	// Verify commit exists.
	log := runGit(t, repoPath, "log", "-1", "--pretty=%s")
	if log != commitMessage {
		t.Errorf("Expected commit message %q, got %q", commitMessage, log)
	}

	if _, err := os.Stat(markerPath); err == nil {
		t.Errorf("Expected pre-commit hook to be skipped, but it ran")
	} else if !os.IsNotExist(err) {
		t.Fatalf("failed to check hook marker: %v", err)
	}
}

func TestIsGitRepository(t *testing.T) {
	t.Run("valid repository", func(t *testing.T) {
		repoPath := createTestRepo(t)
		err := IsGitRepository(repoPath)
		if err != nil {
			t.Errorf("IsGitRepository failed for valid repo: %v", err)
		}
	})

	t.Run("non-repository directory", func(t *testing.T) {
		dir := t.TempDir()
		err := IsGitRepository(dir)
		if err == nil {
			t.Error("IsGitRepository should fail for non-repo directory")
		}
	})

	t.Run("non-existent path", func(t *testing.T) {
		err := IsGitRepository("/path/does/not/exist")
		if err == nil {
			t.Error("IsGitRepository should fail for non-existent path")
		}
	})
}

// Helper functions
func createTestRepo(t *testing.T) string {
	repoPath := t.TempDir()
	runGit(t, repoPath, "init")
	runGit(t, repoPath, "config", "user.email", "test@example.com")
	runGit(t, repoPath, "config", "user.name", "Test User")

	// Create initial commit
	file := filepath.Join(repoPath, "README.md")
	os.WriteFile(file, []byte("# Test Repo"), 0644)
	runGit(t, repoPath, "add", ".")
	runGit(t, repoPath, "commit", "-m", "Initial commit")

	return repoPath
}

func createTestPatch(t *testing.T) string {
	// Create source repo
	srcRepo := createTestRepo(t)

	// Create test file
	filePath := filepath.Join(srcRepo, "testfile.txt")
	err := os.WriteFile(filePath, []byte("patch content"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Generate patch
	patchFile := filepath.Join(t.TempDir(), "test.patch")
	runGit(t, srcRepo, "add", ".")
	runGit(t, srcRepo, "diff", "--cached", "--output", patchFile)

	return patchFile
}

func runGit(t *testing.T, repoPath string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, output)
	}
	return string(output)
}

func currentBranch(t *testing.T, repoPath string) string {
	output := runGit(t, repoPath, "branch", "--show-current")
	return strings.TrimSpace(output)
}
