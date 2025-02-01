package validation

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/vpukhanov/cascade/internal/git"
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

	// Delegate to git package for repository validation
	if err := git.IsGitRepository(path); err != nil {
		return fmt.Errorf("not a git repository: %s (%v)", path, err)
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
