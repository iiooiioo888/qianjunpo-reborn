package validate

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Optional: run when promtool is installed locally (e.g. brew install prometheus).
func TestPromtoolCheckRulesWhenAvailable(t *testing.T) {
	if _, err := exec.LookPath("promtool"); err != nil {
		t.Skip("promtool not installed")
	}
	root := repoRoot(t)
	path := filepath.Join(root, "deploy/observability/prometheus/alerts/qjp-production.rules.yml")
	cmd := exec.Command("promtool", "check", "rules", path)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("promtool check rules: %v\n%s", err, out)
	}
}
