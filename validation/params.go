package validation

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ValidateFile checks if a file exists and meets the requirements for its type
func ValidateFile(path string, fileType string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s file does not exist: %s", fileType, path)
		}
		return fmt.Errorf("error accessing %s file: %v", fileType, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s file is a directory: %s", fileType, path)
	}

	if fileType == "script" && info.Mode()&0111 == 0 {
		return fmt.Errorf("script file is not executable: %s", path)
	}

	return nil
}

// ValidateGitRepo checks if the directory exists and is a git repository
func ValidateGitRepo(path string) error {
	// Check if directory exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", path)
		}
		return fmt.Errorf("error accessing directory: %v", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Check if it's a git repository by running 'git rev-parse --git-dir'
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not a git repository: %s", path)
	}
	return nil
}

// ValidateGitRepos validates a comma-separated list of git repository paths
func ValidateGitRepos(repos string) error {
	if repos == "" {
		return fmt.Errorf("repository list is empty")
	}

	for _, repo := range strings.Split(repos, ",") {
		repo = strings.TrimSpace(repo)
		if repo == "" {
			return fmt.Errorf("empty repository path in list")
		}
		if err := ValidateGitRepo(repo); err != nil {
			return err
		}
	}
	return nil
}

// ValidateBranchName checks if the branch name is valid according to git rules
func ValidateBranchName(name string) error {
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("branch name cannot start with '.'")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("branch name cannot contain '..'")
	}
	if strings.HasSuffix(name, "/") {
		return fmt.Errorf("branch name cannot end with '/'")
	}
	if strings.HasSuffix(name, ".lock") {
		return fmt.Errorf("branch name cannot end with '.lock'")
	}

	// Check for invalid characters
	invalidChars := regexp.MustCompile(`[\s~^:?*\[\\\x00-\x20]`)
	if invalidChars.MatchString(name) {
		return fmt.Errorf("branch name contains invalid characters")
	}

	return nil
}
