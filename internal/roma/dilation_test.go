package roma

import (
	"encoding/json"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadsample"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timedilation"
)

func TestViewSnapshotReportsLiveTimeFlowRate(t *testing.T) {
	store := NewStore(nil)
	b, err := store.Join("default", 0)
	if err != nil {
		t.Fatal(err)
	}
	store.TickRegionDilation(loadsample.Sample{QueueLength: timedilation.HighWatermark})
	raw, _, _, err := store.TacticalViewSnapshotJSON(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	var view tactical.ViewSnapshot
	if err := json.Unmarshal(raw, &view); err != nil {
		t.Fatal(err)
	}
	if view.TimeFlowRateParts >= uint32(timedilation.MaxRate) {
		t.Fatalf("expected slowed rate, got %d", view.TimeFlowRateParts)
	}
}
