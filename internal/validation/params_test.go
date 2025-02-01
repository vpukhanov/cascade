package validation

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestValidateFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "cascade-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	regularFile := filepath.Join(tmpDir, "regular.txt")
	if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	executableFile := filepath.Join(tmpDir, "script.sh")
	if err := os.WriteFile(executableFile, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		path     string
		fileType string
		wantErr  bool
	}{
		{"valid patch file", regularFile, "patch", false},
		{"valid script file", executableFile, "script", false},
		{"non-existent file", filepath.Join(tmpDir, "nonexistent"), "patch", true},
		{"directory as file", tmpDir, "patch", true},
		{"non-executable script", regularFile, "script", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFile(tt.path, tt.fileType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGitRepo(t *testing.T) {
	// Create a temporary directory for test repos
	tmpDir, err := os.MkdirTemp("", "cascade-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test git repository
	gitRepo := filepath.Join(tmpDir, "git-repo")
	if err := os.Mkdir(gitRepo, 0755); err != nil {
		t.Fatal(err)
	}
	if err := runGitInit(gitRepo); err != nil {
		t.Fatal(err)
	}

	// Create a regular directory
	regularDir := filepath.Join(tmpDir, "regular-dir")
	if err := os.Mkdir(regularDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid git repo", gitRepo, false},
		{"non-git directory", regularDir, true},
		{"non-existent directory", filepath.Join(tmpDir, "nonexistent"), true},
		{"file as directory", filepath.Join(tmpDir, "file.txt"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitRepo(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGitRepo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		wantErr bool
	}{
		{"valid branch name", "feature/new-feature", false},
		{"valid simple name", "main", false},
		{"valid with hyphens", "fix-bug-123", false},
		{"starts with dot", ".hidden", true},
		{"contains double dot", "feature..branch", true},
		{"ends with slash", "feature/", true},
		{"ends with .lock", "feature.lock", true},
		{"contains space", "feature branch", true},
		{"contains tilde", "feature~1", true},
		{"contains caret", "feature^1", true},
		{"contains colon", "feature:branch", true},
		{"contains question mark", "feature?", true},
		{"contains asterisk", "feature*", true},
		{"contains brackets", "feature[1]", true},
		{"contains backslash", "feature\\branch", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBranchName(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBranchName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to initialize a git repository
func runGitInit(dir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	return cmd.Run()
}
