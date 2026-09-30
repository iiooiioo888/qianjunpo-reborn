package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerExposesCoreAndProcessMetrics(t *testing.T) {
	for _, name := range CoreSeries() {
		switch name {
		case MetricOnlinePlayers:
			OnlinePlayers.Set(1)
		case MetricBattleLatencyP99Ms:
			BattleLatencyP99.Set(3)
		case MetricQueueLen:
			SetQueueLen(7)
		case MetricTimeFlowRate:
			SetTimeFlowRate(10000)
		case MetricActiveRooms:
			SetActiveRooms(2)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body := rec.Body.String()
	for _, name := range CoreSeries() {
		if !strings.Contains(body, name) {
			t.Errorf("missing metric %q in exposition", name)
		}
	}
	if !strings.Contains(body, MetricProcessCPUSeconds) {
		t.Errorf("missing %q (process collector)", MetricProcessCPUSeconds)
	}
}
