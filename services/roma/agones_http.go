package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/agones"
)

func mountAgonesRoutes(mux *http.ServeMux, coord *agones.Coordinator) {
	mux.HandleFunc("/v1/agones/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, coord.Status(r.Context()))
	})
	mux.HandleFunc("/v1/agones/allocate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body struct {
			ZoneID string `json:"zone_id"`
			Shard  uint32 `json:"shard"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if body.ZoneID == "" {
			body.ZoneID = "default"
		}
		ident, err := coord.AllocateRoom(r.Context(), agones.AllocateRequest{
			ZoneID: body.ZoneID,
			Shard:  body.Shard,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		writeJSON(w, ident)
	})
	mux.HandleFunc("/v1/agones/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := coord.ShutdownGameServer(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		writeJSON(w, map[string]string{"status": "shutdown"})
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func agonesHealthSnippet(ctx context.Context, coord *agones.Coordinator) string {
	st := coord.Status(ctx)
	return `"agones_phase":"` + string(st.Phase) + `","agones_ready":` + strconv.FormatBool(st.Ready)
}
