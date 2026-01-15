package log

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestNewApplyLogger(t *testing.T) {
	logger, err := NewApplyLogger()
	if err != nil {
		t.Fatalf("NewApplyLogger() error: %v", err)
	}

	path := logger.Path()
	if path == "" {
		t.Fatal("Path() returned empty string")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat log file: %v", err)
	}
	if info.IsDir() {
		t.Fatalf("expected log file, got directory: %s", path)
	}

	if err := logger.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
	_ = os.Remove(path)
}

func TestApplyLoggerLogRepoError(t *testing.T) {
	logger, err := NewApplyLogger()
	if err != nil {
		t.Fatalf("NewApplyLogger() error: %v", err)
	}

	path := logger.Path()
	repoErr := fmt.Errorf("outer: %w", errors.New("inner"))
	logger.LogRepoError("repo1", repoErr)

	if err := logger.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	text := string(data)

	for _, want := range []string{
		"repo: repo1",
		"- outer: inner",
		"  - inner",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected log output to contain %q, got:\n%s", want, text)
		}
	}

	_ = os.Remove(path)
}
