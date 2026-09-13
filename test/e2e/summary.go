//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"sync"
)

var summaryMu sync.Mutex

// WriteStepSummary appends a markdown section to GITHUB_STEP_SUMMARY file if defined, or prints to stdout.
func WriteStepSummary(summaryFile, title, content string) {
	if summaryFile == "" {
		return
	}

	summaryMu.Lock()
	defer summaryMu.Unlock()

	f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()

	text := fmt.Sprintf("## %s\n```\n%s\n```\n\n", title, content)
	_, _ = f.WriteString(text)
}
