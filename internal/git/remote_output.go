package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

var remoteURLPattern = regexp.MustCompile(`https?://\S+`)

// OpenLastRemoteURL opens the last URL from git output in the default browser.
func OpenLastRemoteURL(output string) error {
	url := lastRemoteURL(output)
	if url == "" {
		return nil
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("open browser failed: %w\n%s", err, string(output))
	}
	return nil
}

func lastRemoteURL(output string) string {
	var lastMatch string
	for line := range strings.SplitSeq(output, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if !strings.HasPrefix(trimmed, "remote:") {
			continue
		}
		matches := remoteURLPattern.FindAllString(trimmed, -1)
		if len(matches) == 0 {
			continue
		}
		lastMatch = matches[len(matches)-1]
	}
	return lastMatch
}
