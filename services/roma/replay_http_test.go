package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/internal/roma"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/replay"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

func TestReplayHTTPExport(t *testing.T) {
	store := roma.NewStore(nil)
	b, err := store.Join("default", 0)
	if err != nil {
		t.Fatal(err)
	}
	_ = b.Match.RunSchedule(tactical.DemoSchedule(), tactical.DemoTargetFrame())

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/battles/replay", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("battle_id")
		rec, err := store.ExportRecording(roma.BattleID(id))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		gz, err := replay.MarshalGzip(rec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/gzip")
		_, _ = w.Write(gz)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/battles/replay?battle_id="+string(b.ID), nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	data, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	back, err := replay.UnmarshalGzip(data)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tactical.ReplayFromRecording(back); err != nil {
		t.Fatal(err)
	}
}
