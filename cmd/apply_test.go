package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// Override git functions with mocks
func resetMocks() {
	gitCheckoutBranch = func(repoPath, branch string) error { return nil }
	gitCheckoutExistingBranch = func(repoPath, branch string) error { return nil }
	gitApplyPatch = func(repoPath, patchPath string) error { return nil }
	gitCommitChanges = func(repoPath, message string, noVerify bool) error { return nil }
	gitExecuteScript = func(repoPath, scriptPath string) error { return nil }
	gitPullLatest = func(repoPath string) error { return nil }
	gitPushChanges = func(repoPath, branch string, noVerify bool) error { return nil }
	gitStashChanges = func(repoPath string) error { return nil }
}

func TestRunApply(t *testing.T) {
	tests := []struct {
		name        string
		repos       []string
		useScript   bool
		baseBranch  string
		pullLatest  bool
		push        bool
		stash       bool
		mockSetup   func()
		wantSuccess int
		wantErrors  int
	}{
		{
			name:        "all_success_patch",
			repos:       []string{"repo1", "repo2"},
			useScript:   false,
			mockSetup:   func() { resetMocks() },
			wantSuccess: 2,
			wantErrors:  0,
		},
		{
			name:        "all_success_script",
			repos:       []string{"repo1", "repo2"},
			useScript:   true,
			mockSetup:   func() { resetMocks() },
			wantSuccess: 2,
			wantErrors:  0,
		},
		{
			name:        "success_with_base_branch",
			repos:       []string{"repo1", "repo2"},
			useScript:   false,
			baseBranch:  "main",
			mockSetup:   func() { resetMocks() },
			wantSuccess: 2,
			wantErrors:  0,
		},
		{
			name:        "success_with_pull",
			repos:       []string{"repo1", "repo2"},
			useScript:   false,
			pullLatest:  true,
			mockSetup:   func() { resetMocks() },
			wantSuccess: 2,
			wantErrors:  0,
		},
		{
			name:        "success_with_push",
			repos:       []string{"repo1", "repo2"},
			useScript:   false,
			push:        true,
			mockSetup:   func() { resetMocks() },
			wantSuccess: 2,
			wantErrors:  0,
		},
		{
			name:        "success_with_stash",
			repos:       []string{"repo1", "repo2"},
			useScript:   false,
			stash:       true,
			mockSetup:   func() { resetMocks() },
			wantSuccess: 2,
			wantErrors:  0,
		},
		{
			name:       "fail_base_branch_checkout",
			repos:      []string{"repo1", "repo2"},
			useScript:  false,
			baseBranch: "main",
			mockSetup: func() {
				resetMocks()
				gitCheckoutExistingBranch = func(_, _ string) error {
					return fmt.Errorf("base branch checkout failed")
				}
			},
			wantSuccess: 0,
			wantErrors:  2,
		},
		{
			name:      "fail_stash",
			repos:     []string{"repo1", "repo2"},
			useScript: false,
			stash:     true,
			mockSetup: func() {
				resetMocks()
				gitStashChanges = func(_ string) error {
					return fmt.Errorf("stash failed")
				}
			},
			wantSuccess: 0,
			wantErrors:  2,
		},
		{
			name:       "fail_pull_latest",
			repos:      []string{"repo1", "repo2"},
			useScript:  false,
			pullLatest: true,
			mockSetup: func() {
				resetMocks()
				gitPullLatest = func(_ string) error {
					return fmt.Errorf("pull failed")
				}
			},
			wantSuccess: 0,
			wantErrors:  2,
		},
		{
			name:      "fail_push",
			repos:     []string{"repo1", "repo2"},
			useScript: false,
			push:      true,
			mockSetup: func() {
				resetMocks()
				gitPushChanges = func(_, _ string, _ bool) error {
					return fmt.Errorf("push failed")
				}
			},
			wantSuccess: 0,
			wantErrors:  2,
		},
		{
			name:      "mixed_results_patch",
			repos:     []string{"repo1", "repo2", "repo3"},
			useScript: false,
			mockSetup: func() {
				resetMocks()
				// Fail checkout for repo1
				gitCheckoutBranch = func(repoPath, branch string) error {
					if repoPath == "repo1" {
						return fmt.Errorf("checkout error")
					}
					return nil
				}
				// Fail patch apply for repo2
				gitApplyPatch = func(repoPath, _ string) error {
					if repoPath == "repo2" {
						return fmt.Errorf("apply error")
					}
					return nil
				}
				// Fail commit for repo3
				gitCommitChanges = func(repoPath, _ string, _ bool) error {
					if repoPath == "repo3" {
						return fmt.Errorf("commit error")
					}
					return nil
				}
			},
			wantSuccess: 0,
			wantErrors:  3,
		},
		{
			name:      "some_success_some_fail_patch",
			repos:     []string{"repo1", "repo2", "repo3"},
			useScript: false,
			mockSetup: func() {
				resetMocks()
				gitCommitChanges = func(repoPath, _ string, _ bool) error {
					if repoPath == "repo2" {
						return fmt.Errorf("commit error")
					}
					return nil
				}
			},
			wantSuccess: 2,
			wantErrors:  1,
		},
		{
			name:      "mixed_results_script",
			repos:     []string{"repo1", "repo2", "repo3"},
			useScript: true,
			mockSetup: func() {
				resetMocks()
				// Fail checkout for repo1
				gitCheckoutBranch = func(repoPath, _ string) error {
					if repoPath == "repo1" {
						return fmt.Errorf("checkout error")
					}
					return nil
				}
				// Fail script execution for repo2
				gitExecuteScript = func(repoPath, _ string) error {
					if repoPath == "repo2" {
						return fmt.Errorf("script error")
					}
					return nil
				}
				// Fail commit for repo3
				gitCommitChanges = func(repoPath, _ string, _ bool) error {
					if repoPath == "repo3" {
						return fmt.Errorf("commit error")
					}
					return nil
				}
			},
			wantSuccess: 0,
			wantErrors:  3,
		},
		{
			name:      "all_fail_script",
			repos:     []string{"repo1"},
			useScript: true,
			mockSetup: func() {
				resetMocks()
				gitExecuteScript = func(_, _ string) error {
					return fmt.Errorf("script failed")
				}
			},
			wantSuccess: 0,
			wantErrors:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simplified setup without temp files
			tt.mockSetup()

			// Set script or patch file
			if tt.useScript {
				scriptFile = "test.sh"
				patchFile = ""
			} else {
				scriptFile = ""
				patchFile = "test.patch"
			}

			// Set optional flags
			baseBranch = tt.baseBranch
			pullLatest = tt.pullLatest
			push = tt.push
			noVerify = false
			stash = tt.stash

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run test
			err := runApply(nil, tt.repos)

			// Restore stdout
			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = oldStdout

			// Reset global variables
			scriptFile = ""
			patchFile = ""
			baseBranch = ""
			pullLatest = false
			push = false
			noVerify = false
			stash = false

			// Verify results
			if err != nil {
				t.Errorf("runApply() unexpected error: %v", err)
			}

			output := string(out)

			okCount := 0
			failCount := 0
			var logPath string
			for _, line := range strings.Split(output, "\n") {
				switch {
				case strings.HasPrefix(line, "ok "):
					okCount++
				case strings.HasPrefix(line, "fail "):
					failCount++
				case strings.HasPrefix(line, "Error details: "):
					logPath = strings.TrimSpace(strings.TrimPrefix(line, "Error details: "))
				}
			}

			if okCount != tt.wantSuccess {
				t.Errorf("Expected %d ok lines, got %d\n\nOutput:\n%s", tt.wantSuccess, okCount, output)
			}
			if failCount != tt.wantErrors {
				t.Errorf("Expected %d fail lines, got %d\n\nOutput:\n%s", tt.wantErrors, failCount, output)
			}
			if okCount+failCount != len(tt.repos) {
				t.Errorf("Expected %d repo result lines, got %d\n\nOutput:\n%s", len(tt.repos), okCount+failCount, output)
			}

			for _, repo := range tt.repos {
				okLine := fmt.Sprintf("%-4s %s", "ok", repo)
				failLine := fmt.Sprintf("%-4s %s", "fail", repo)
				if !strings.Contains(output, okLine) && !strings.Contains(output, failLine) {
					t.Errorf("Missing result line for repo %q\n\nOutput:\n%s", repo, output)
				}
			}

			if tt.wantErrors > 0 {
				if logPath == "" {
					t.Errorf("Missing error log path\n\nOutput:\n%s", output)
				} else if info, err := os.Stat(logPath); err != nil {
					t.Errorf("Error log not found at %q: %v", logPath, err)
				} else if info.IsDir() {
					t.Errorf("Expected error log file, got directory: %q", logPath)
				}
			} else if logPath != "" {
				t.Errorf("Unexpected error log path in success output: %q\n\nOutput:\n%s", logPath, output)
			}

			if logPath != "" {
				_ = os.Remove(logPath)
			}
		})
	}
}
