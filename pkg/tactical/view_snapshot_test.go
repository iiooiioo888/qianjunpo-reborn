package tactical

import (
	"encoding/json"
	"testing"
)

func TestInitialViewSnapshotShape(t *testing.T) {
	v := InitialViewSnapshot(0xcafe)
	if v.SchemaVersion != viewSnapshotSchemaVersion || v.BoardSize != 19 {
		t.Fatalf("schema=%d size=%d", v.SchemaVersion, v.BoardSize)
	}
	if len(v.Cells) != 19 || len(v.Cells[0]) != 19 {
		t.Fatalf("cells grid len=%d", len(v.Cells))
	}
	if len(v.Units) != 2 {
		t.Fatalf("units=%d", len(v.Units))
	}
	river := v.Cells[9][0]
	if river.Terrain != 3 || river.Passable {
		t.Fatalf("y=9 x=0 river=%+v", river)
	}
	pass := v.Cells[9][9]
	if pass.Terrain != 5 || !pass.Passable {
		t.Fatalf("pass tile=%+v", pass)
	}
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var back ViewSnapshot
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.Seed != 0xcafe {
		t.Fatalf("seed=%d", back.Seed)
	}
	if back.EndReason != "none" || back.Winner != nil {
		t.Fatalf("in-progress outcome winner=%v reason=%q", back.Winner, back.EndReason)
	}
}
