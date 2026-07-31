package update

import (
	"fmt"
	"os/exec"
	"strings"
)

const installScriptURL = "https://raw.githubusercontent.com/hxsggsz/kanba/main/install.sh"

// Run downloads and executes install.sh to update kanba to the latest
// release. On failure it returns an error built from the script's output:
// a short human-readable reason (extracted from an "Error:"-prefixed line
// when the script produced one) plus the raw combined output, so failures
// that don't match install.sh's error-line convention are still
// diagnosable instead of being discarded.
func Run() error {
	cmd := exec.Command("bash", "-c", "curl -fsSL "+installScriptURL+" | bash")
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	reason, ok := extractErrorReason(string(out))
	if !ok {
		reason = "install script exited with an error"
	}
	return fmt.Errorf("install failed: %s\n\n%s", reason, strings.TrimSpace(string(out)))
}

// extractErrorReason returns the last "Error:"-prefixed line found in
// output, if any, along with whether one was found.
func extractErrorReason(output string) (string, bool) {
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if reason, ok := strings.CutPrefix(line, "Error:"); ok {
			reason = strings.TrimSpace(reason)
			if reason != "" {
				return reason, true
			}
		}
	}
	return "", false
}
