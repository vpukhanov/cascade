package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"cascade/cmd"
)

func TestIntegrationApply(t *testing.T) {
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
		{"git", "init"},
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
