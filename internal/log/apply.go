package log

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

// ApplyLogger writes detailed apply errors to a file for later inspection.
type ApplyLogger struct {
	mu     sync.Mutex
	file   *os.File
	logger *log.Logger
}

// NewApplyLogger creates a log file in the OS temp directory.
func NewApplyLogger() (*ApplyLogger, error) {
	file, err := os.CreateTemp("", "cascade-apply-*.log")
	if err != nil {
		return nil, fmt.Errorf("create apply log file: %w", err)
	}

	return &ApplyLogger{
		file:   file,
		logger: log.New(file, "", log.LstdFlags),
	}, nil
}

// Path returns the log file path.
func (l *ApplyLogger) Path() string {
	return l.file.Name()
}

// Close closes the underlying log file.
func (l *ApplyLogger) Close() error {
	return l.file.Close()
}

// LogRepoError records the full error chain for a repository.
func (l *ApplyLogger) LogRepoError(repo string, err error) {
	if err == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.logger.Printf("repo: %s", repo)
	for _, line := range formatErrorChain(err) {
		l.logger.Print(line)
	}
	l.logger.Print("")
}

func formatErrorChain(err error) []string {
	var lines []string
	var visit func(err error, depth int)

	visit = func(err error, depth int) {
		if err == nil {
			return
		}

		indent := strings.Repeat("  ", depth)
		lines = append(lines, fmt.Sprintf("%s- %s", indent, err.Error()))

		if unwrapped, ok := err.(interface{ Unwrap() []error }); ok {
			for _, nested := range unwrapped.Unwrap() {
				visit(nested, depth+1)
			}
			return
		}

		if next := errors.Unwrap(err); next != nil {
			visit(next, depth+1)
		}
	}

	visit(err, 0)
	return lines
}
