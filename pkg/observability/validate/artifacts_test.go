package validate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/observability/metrics"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func TestAlertRulesReferenceCoreMetrics(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "deploy/observability/prometheus/alerts/qjp-production.rules.yml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	requiredAlerts := []string{
		"QjpHighCPU",
		"QjpQueueDepthHigh",
		"QjpTimeFlowDegraded",
		"QjpActiveRoomsHigh",
	}
	for _, alert := range requiredAlerts {
		if !strings.Contains(text, alert) {
			t.Errorf("rules file missing alert %q", alert)
		}
	}
	if !strings.Contains(text, metrics.MetricQueueLen) ||
		!strings.Contains(text, metrics.MetricTimeFlowRate) ||
		!strings.Contains(text, metrics.MetricActiveRooms) {
		t.Fatal("rules must reference queue, time_flow, and active_rooms metrics")
	}
	if !strings.Contains(text, "process_cpu_seconds_total") {
		t.Fatal("rules must reference process_cpu_seconds_total for CPU alert")
	}
}

func TestGrafanaDashboardReferencesCoreMetrics(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "deploy/observability/grafana-phase5-dashboard.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Panels []struct {
			Title   string `json:"title"`
			Targets []struct {
				Expr string `json:"expr"`
			} `json:"targets"`
		} `json:"panels"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	joined := string(raw)
	for _, metric := range metrics.CoreSeries() {
		if !strings.Contains(joined, metric) {
			t.Errorf("dashboard missing metric %q", metric)
		}
	}
	if !strings.Contains(joined, metrics.MetricProcessCPUSeconds) {
		t.Errorf("dashboard missing %q", metrics.MetricProcessCPUSeconds)
	}
}
