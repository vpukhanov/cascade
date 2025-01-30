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
	gitApplyPatch = func(repoPath, patchPath string) error { return nil }
	gitCommitChanges = func(repoPath, message string) error { return nil }
}

func TestRunApply(t *testing.T) {
	tests := []struct {
		name        string
		repos       []string
		mockSetup   func()
		wantSuccess int
		wantErrors  int
	}{
		{
			name:  "all_success",
			repos: []string{"repo1", "repo2"},
			mockSetup: func() {
				resetMocks()
			},
			wantSuccess: 2,
			wantErrors:  0,
		},
		{
			name:  "mixed_results",
			repos: []string{"repo1", "repo2", "repo3"},
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
				gitCommitChanges = func(repoPath, _ string) error {
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
			name:  "all_fail",
			repos: []string{"repo1"},
			mockSetup: func() {
				resetMocks()
				gitCheckoutBranch = func(_, _ string) error {
					return fmt.Errorf("checkout failed")
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

			// Verify results
			if err != nil {
				t.Errorf("runApply() unexpected error: %v", err)
			}

			output := string(out)

			// Check success header
			expectedSuccessHeader := fmt.Sprintf("Successfully processed (%d):", tt.wantSuccess)
			if tt.wantSuccess > 0 && !strings.Contains(output, expectedSuccessHeader) {
				t.Errorf("Missing success header: %q\n\nGot:\n%s", expectedSuccessHeader, output)
			}

			// Check failure header
			expectedErrorHeader := fmt.Sprintf("Failed (%d):", tt.wantErrors)
			if tt.wantErrors > 0 && !strings.Contains(output, expectedErrorHeader) {
				t.Errorf("Missing error header: %q\n\nGot:\n%s", expectedErrorHeader, output)
			}
		})
	}
}
