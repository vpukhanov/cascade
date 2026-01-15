package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/vpukhanov/cascade/cmd"
)

func TestIntegrationApply(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test directory
	testDir := t.TempDir()

	// Create test repositories
	repo1Path := filepath.Join(testDir, "repo1")
	repo2Path := filepath.Join(testDir, "repo2")
	createTestRepo(t, repo1Path)
	createTestRepo(t, repo2Path)

	// Reset flags before each test
	resetFlags := func() {
		cmd.ResetFlags()
	}

	t.Run("apply patch to multiple repositories", func(t *testing.T) {
		resetFlags()
		// Create a patch file
		patchContent := `diff --git a/test.txt b/test.txt
new file mode 100644
index 0000000..9daeafb
--- /dev/null
+++ b/test.txt
@@ -0,0 +1 @@
+test
`
		patchFile := filepath.Join(testDir, "test.patch")
		if err := os.WriteFile(patchFile, []byte(patchContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Set up command line arguments
		os.Args = []string{
			"cascade",
			"apply",
			"--patch", patchFile,
			"--branch", "feature/test-patch",
			"--message", "Add test.txt",
			repo1Path,
			repo2Path,
		}

		// Run the command
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		// Verify changes in both repositories
		for _, repoPath := range []string{repo1Path, repo2Path} {
			// Check if branch exists
			if branch := getCurrentBranch(t, repoPath); branch != "feature/test-patch" {
				t.Errorf("Expected branch feature/test-patch, got %s in %s", branch, repoPath)
			}

			// Check if file exists
			testFile := filepath.Join(repoPath, "test.txt")
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				t.Errorf("test.txt not found in %s", repoPath)
			}

			// Check commit message
			if msg := getLastCommitMessage(t, repoPath); msg != "Add test.txt" {
				t.Errorf("Expected commit message 'Add test.txt', got '%s' in %s", msg, repoPath)
			}
		}
	})

	t.Run("apply script to multiple repositories", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Script execution is not supported on Windows [https://github.com/vpukhanov/cascade/issues/1]")
		}

		resetFlags()
		// Create a test script
		scriptContent := `#!/bin/sh
echo "modified content" > modified.txt
`
		scriptFile := filepath.Join(testDir, "test.sh")
		if err := os.WriteFile(scriptFile, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		// Set up command line arguments
		os.Args = []string{
			"cascade",
			"apply",
			"--script", scriptFile,
			"--branch", "feature/test-script",
			"--message", "Add modified.txt",
			repo1Path,
			repo2Path,
		}

		// Run the command
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		// Verify changes in both repositories
		for _, repoPath := range []string{repo1Path, repo2Path} {
			// Check if branch exists
			if branch := getCurrentBranch(t, repoPath); branch != "feature/test-script" {
				t.Errorf("Expected branch feature/test-script, got %s in %s", branch, repoPath)
			}

			// Check if file exists and has correct content
			modifiedFile := filepath.Join(repoPath, "modified.txt")
			content, err := os.ReadFile(modifiedFile)
			if err != nil {
				t.Errorf("modified.txt not found in %s", repoPath)
			} else if string(content) != "modified content\n" {
				t.Errorf("Expected content 'modified content', got '%s' in %s", string(content), repoPath)
			}

			// Check commit message
			if msg := getLastCommitMessage(t, repoPath); msg != "Add modified.txt" {
				t.Errorf("Expected commit message 'Add modified.txt', got '%s' in %s", msg, repoPath)
			}
		}
	})

	t.Run("apply changes to specific base branch", func(t *testing.T) {
		resetFlags()

		// Create a development branch with some changes
		for _, repoPath := range []string{repo1Path, repo2Path} {
			// Create and switch to dev branch
			runGitCmd(t, repoPath, "checkout", "-b", "development")
			// Add a file that would conflict with our patch
			err := os.WriteFile(filepath.Join(repoPath, "dev.txt"), []byte("dev branch"), 0644)
			if err != nil {
				t.Fatal(err)
			}
			runGitCmd(t, repoPath, "add", "dev.txt")
			runGitCmd(t, repoPath, "commit", "-m", "Development changes")
			// Switch back to main
			runGitCmd(t, repoPath, "checkout", "main")
		}

		// Create a patch file
		patchContent := `diff --git a/test.txt b/test.txt
new file mode 100644
index 0000000..9daeafb
--- /dev/null
+++ b/test.txt
@@ -0,0 +1 @@
+test
`
		patchFile := filepath.Join(testDir, "base-branch.patch")
		if err := os.WriteFile(patchFile, []byte(patchContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Set up command line arguments
		os.Args = []string{
			"cascade",
			"apply",
			"--patch", patchFile,
			"--branch", "feature/base-test",
			"--message", "Add test.txt",
			"--base-branch", "development",
			repo1Path,
			repo2Path,
		}

		// Run the command
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		// Verify changes in both repositories
		for _, repoPath := range []string{repo1Path, repo2Path} {
			// Verify we're on the right branch
			if branch := getCurrentBranch(t, repoPath); branch != "feature/base-test" {
				t.Errorf("Expected branch feature/base-test, got %s in %s", branch, repoPath)
			}

			// Verify both files exist (dev.txt from base branch and test.txt from patch)
			files := []string{"dev.txt", "test.txt"}
			for _, file := range files {
				path := filepath.Join(repoPath, file)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("%s not found in %s", file, repoPath)
				}
			}
		}
	})

	t.Run("pull latest changes before applying", func(t *testing.T) {
		resetFlags()

		// Create a remote repository and clone it
		remoteRepo := filepath.Join(testDir, "remote")
		createTestRepo(t, remoteRepo)
		clonedRepo := filepath.Join(testDir, "cloned")
		runGitCmd(t, testDir, "clone", remoteRepo, clonedRepo)

		// Add new changes to remote
		err := os.WriteFile(filepath.Join(remoteRepo, "remote.txt"), []byte("remote change"), 0644)
		if err != nil {
			t.Fatal(err)
		}
		runGitCmd(t, remoteRepo, "add", "remote.txt")
		runGitCmd(t, remoteRepo, "commit", "-m", "Remote changes")

		// Create a patch file
		patchContent := `diff --git a/local.txt b/local.txt
new file mode 100644
index 0000000..9daeafb
--- /dev/null
+++ b/local.txt
@@ -0,0 +1 @@
+local
`
		patchFile := filepath.Join(testDir, "pull.patch")
		if err := os.WriteFile(patchFile, []byte(patchContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Set up command line arguments
		os.Args = []string{
			"cascade",
			"apply",
			"--patch", patchFile,
			"--branch", "feature/pull-test",
			"--message", "Add local.txt",
			"--pull",
			clonedRepo,
		}

		// Run the command
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		// Verify changes
		// Check if both remote and local changes are present
		files := []string{"remote.txt", "local.txt"}
		for _, file := range files {
			path := filepath.Join(clonedRepo, file)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("%s not found in cloned repo", file)
			}
		}
	})

	t.Run("push changes to remote", func(t *testing.T) {
		resetFlags()

		// Create a bare remote repository
		remoteRepo := filepath.Join(testDir, "remote-push")
		if err := os.MkdirAll(remoteRepo, 0755); err != nil {
			t.Fatal(err)
		}
		runGitCmd(t, remoteRepo, "init", "--bare", "-b", "main")

		// Clone the remote repository
		clonedRepo := filepath.Join(testDir, "cloned-push")
		runGitCmd(t, testDir, "clone", remoteRepo, clonedRepo)

		// Set up git config in cloned repo
		runGitCmd(t, clonedRepo, "config", "user.email", "test@example.com")
		runGitCmd(t, clonedRepo, "config", "user.name", "Test User")
		runGitCmd(t, clonedRepo, "config", "commit.gpgsign", "false")

		// Create initial commit in cloned repo
		readmeFile := filepath.Join(clonedRepo, "README.md")
		if err := os.WriteFile(readmeFile, []byte("# Test Repository"), 0644); err != nil {
			t.Fatal(err)
		}
		runGitCmd(t, clonedRepo, "add", "README.md")
		runGitCmd(t, clonedRepo, "commit", "-m", "Initial commit")
		runGitCmd(t, clonedRepo, "push", "origin", "main")

		// Create a patch file
		patchContent := `diff --git a/pushed.txt b/pushed.txt
new file mode 100644
index 0000000..9daeafb
--- /dev/null
+++ b/pushed.txt
@@ -0,0 +1 @@
+pushed content
`
		patchFile := filepath.Join(testDir, "push.patch")
		if err := os.WriteFile(patchFile, []byte(patchContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Set up command line arguments
		os.Args = []string{
			"cascade",
			"apply",
			"--patch", patchFile,
			"--branch", "feature/push-test",
			"--message", "Add pushed.txt",
			"--push",
			clonedRepo,
		}

		// Run the command
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		// Verify changes were pushed to remote by cloning to a new directory
		verifyRepo := filepath.Join(testDir, "verify-push")
		runGitCmd(t, testDir, "clone", remoteRepo, verifyRepo)
		runGitCmd(t, verifyRepo, "checkout", "feature/push-test")

		// Verify file exists and has correct content
		pushedFile := filepath.Join(verifyRepo, "pushed.txt")
		content, err := os.ReadFile(pushedFile)
		if err != nil {
			t.Errorf("pushed.txt not found in remote repository")
		} else if strings.TrimSpace(string(content)) != "pushed content" {
			t.Errorf("Expected content 'pushed content', got '%s' in remote", string(content))
		}

		// Verify commit message
		if msg := getLastCommitMessage(t, verifyRepo); msg != "Add pushed.txt" {
			t.Errorf("Expected commit message 'Add pushed.txt', got '%s'", msg)
		}
	})

	t.Run("push with no-verify skips pre-push hook", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Git hook shell scripts are not supported on Windows in this test")
		}

		resetFlags()

		// Create a bare remote repository
		remoteRepo := filepath.Join(testDir, "remote-push-no-verify")
		if err := os.MkdirAll(remoteRepo, 0755); err != nil {
			t.Fatal(err)
		}
		runGitCmd(t, remoteRepo, "init", "--bare", "-b", "main")

		// Clone the remote repository
		clonedRepo := filepath.Join(testDir, "cloned-push-no-verify")
		runGitCmd(t, testDir, "clone", remoteRepo, clonedRepo)

		// Set up git config in cloned repo
		runGitCmd(t, clonedRepo, "config", "user.email", "test@example.com")
		runGitCmd(t, clonedRepo, "config", "user.name", "Test User")
		runGitCmd(t, clonedRepo, "config", "commit.gpgsign", "false")

		// Create initial commit in cloned repo
		readmeFile := filepath.Join(clonedRepo, "README.md")
		if err := os.WriteFile(readmeFile, []byte("# Test Repository"), 0644); err != nil {
			t.Fatal(err)
		}
		runGitCmd(t, clonedRepo, "add", "README.md")
		runGitCmd(t, clonedRepo, "commit", "-m", "Initial commit")
		runGitCmd(t, clonedRepo, "push", "origin", "main")

		// Add a pre-push hook that would fail if executed.
		hookPath := filepath.Join(clonedRepo, ".git", "hooks", "pre-push")
		markerPath := filepath.Join(clonedRepo, "pre_push_ran")
		hookContents := "#!/bin/sh\n" +
			"echo ran > \"" + markerPath + "\"\n" +
			"exit 1\n"
		if err := os.WriteFile(hookPath, []byte(hookContents), 0755); err != nil {
			t.Fatalf("failed to write pre-push hook: %v", err)
		}

		// Create a patch file
		patchContent := `diff --git a/pushed.txt b/pushed.txt
new file mode 100644
index 0000000..9daeafb
--- /dev/null
+++ b/pushed.txt
@@ -0,0 +1 @@
+pushed content
`
		patchFile := filepath.Join(testDir, "push-no-verify.patch")
		if err := os.WriteFile(patchFile, []byte(patchContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Set up command line arguments
		os.Args = []string{
			"cascade",
			"apply",
			"--patch", patchFile,
			"--branch", "feature/push-no-verify-test",
			"--message", "Add pushed.txt",
			"--push",
			"--no-verify",
			clonedRepo,
		}

		// Run the command
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		if _, err := os.Stat(markerPath); err == nil {
			t.Errorf("Expected pre-push hook to be skipped, but it ran")
		} else if !os.IsNotExist(err) {
			t.Fatalf("failed to check hook marker: %v", err)
		}

		// Verify changes were pushed to remote by cloning to a new directory
		verifyRepo := filepath.Join(testDir, "verify-push-no-verify")
		runGitCmd(t, testDir, "clone", remoteRepo, verifyRepo)
		runGitCmd(t, verifyRepo, "checkout", "feature/push-no-verify-test")

		// Verify file exists and has correct content
		pushedFile := filepath.Join(verifyRepo, "pushed.txt")
		content, err := os.ReadFile(pushedFile)
		if err != nil {
			t.Errorf("pushed.txt not found in remote repository")
		} else if strings.TrimSpace(string(content)) != "pushed content" {
			t.Errorf("Expected content 'pushed content', got '%s' in remote", string(content))
		}
	})

	t.Run("fail on invalid repository", func(t *testing.T) {
		resetFlags()
		invalidRepo := filepath.Join(testDir, "not-a-repo")
		if err := os.Mkdir(invalidRepo, 0755); err != nil {
			t.Fatal(err)
		}

		// Set up command line arguments
		os.Args = []string{
			"cascade",
			"apply",
			"--patch", filepath.Join(testDir, "test.patch"),
			"--branch", "feature/test",
			"--message", "Test",
			invalidRepo,
		}

		// Run the command and expect error
		if err := cmd.Execute(); err == nil {
			t.Error("Expected error when using invalid repository")
		}
	})

	t.Run("fail on non-executable script", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Script execution is not supported on Windows [https://github.com/vpukhanov/cascade/issues/1]")
		}

		resetFlags()
		nonExecutableScript := filepath.Join(testDir, "non-executable.sh")
		if err := os.WriteFile(nonExecutableScript, []byte("#!/bin/sh\necho test"), 0644); err != nil {
			t.Fatal(err)
		}

		// Set up command line arguments
		os.Args = []string{
			"cascade",
			"apply",
			"--script", nonExecutableScript,
			"--branch", "feature/test",
			"--message", "Test",
			repo1Path,
		}

		// Run the command and expect error
		if err := cmd.Execute(); err == nil {
			t.Error("Expected error when using non-executable script")
		}
	})
}

// Helper functions

func createTestRepo(t *testing.T, path string) {
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}

	cmds := [][]string{
		{"git", "init", "-b", "main"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "commit.gpgsign", "false"},
		{"git", "config", "init.defaultBranch", "main"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = path
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to run %v: %v\n%s", cmdArgs, err, output)
		}
	}

	// Create and commit initial file
	readmeFile := filepath.Join(path, "README.md")
	if err := os.WriteFile(readmeFile, []byte("# Test Repository"), 0644); err != nil {
		t.Fatal(err)
	}

	cmds = [][]string{
		{"git", "add", "README.md"},
		{"git", "commit", "-m", "Initial commit"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = path
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to run %v: %v\n%s", cmdArgs, err, output)
		}
	}
}

func getCurrentBranch(t *testing.T, repoPath string) string {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v\n%s", err, output)
	}
	return strings.TrimSpace(string(output))
}

func getLastCommitMessage(t *testing.T, repoPath string) string {
	cmd := exec.Command("git", "log", "-1", "--pretty=%s")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to get last commit message: %v\n%s", err, output)
	}
	return strings.TrimSpace(string(output))
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to run git %v: %v\n%s", args, err, output)
	}
}
