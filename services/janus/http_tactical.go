package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	gatewayv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/gateway/v1"
)

const maxTacticalCommandBodyBytes = 16 << 10

type tacticalCommandBody struct {
	SessionID string `json:"session_id"`
	BattleID  string `json:"battle_id"`
	PlayerID  uint32 `json:"player_id"`
	Kind      uint32 `json:"kind"`
	UnitID    uint32 `json:"unit_id"`
	ToX       int32  `json:"to_x"`
	ToY       int32  `json:"to_y"`
}

type tacticalCommandJSON struct {
	Accepted      bool   `json:"accepted"`
	RejectReason  string `json:"reject_reason,omitempty"`
	LockstepFrame uint64 `json:"lockstep_frame,omitempty"`
	StateHash     uint64 `json:"state_hash,omitempty"`
}

type tacticalHTTPOptions struct {
	CommandMirrorEnabled func() bool
}

func defaultTacticalHTTPOptions() tacticalHTTPOptions {
	return tacticalHTTPOptions{
		CommandMirrorEnabled: httpTacticalCommandMirrorEnabled,
	}
}

// httpTacticalCommandMirrorEnabled gates POST /v1/tactical/command (default on for dev preview).
func httpTacticalCommandMirrorEnabled() bool {
	return envBool("JANUS_HTTP_TACTICAL_COMMAND", true)
}

func envBool(k string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on", "enabled":
		return true
	case "0", "false", "no", "off", "disabled":
		return false
	default:
		return def
	}
}

func registerTacticalHTTPRoutes(mux *http.ServeMux, gw *janusGateway, opts tacticalHTTPOptions) {
	mux.HandleFunc("/v1/tactical/snapshot", func(w http.ResponseWriter, r *http.Request) {
		handleTacticalSnapshot(w, r, gw)
	})
	mux.HandleFunc("/v1/tactical/command", func(w http.ResponseWriter, r *http.Request) {
		handleTacticalCommand(w, r, gw, opts.CommandMirrorEnabled)
	})
}

func handleTacticalSnapshot(w http.ResponseWriter, r *http.Request, gw *janusGateway) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	battleID := r.URL.Query().Get("battle_id")
	if battleID == "" {
		http.Error(w, "missing battle_id", http.StatusBadRequest)
		return
	}
	resp, err := gw.GetBattleSnapshot(r.Context(), &gatewayv1.GetBattleSnapshotRequest{
		SessionId: r.URL.Query().Get("session_id"),
		BattleId:  battleID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Lockstep-Frame", fmt.Sprintf("%d", resp.GetLockstepFrame()))
	w.Header().Set("X-State-Hash", fmt.Sprintf("%016x", resp.GetStateHash()))
	_, _ = w.Write(resp.GetViewSnapshotJson())
}

func handleTacticalCommand(w http.ResponseWriter, r *http.Request, gw *janusGateway, enabled func() bool) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !enabled() {
		writeTacticalCommandJSON(w, http.StatusServiceUnavailable, tacticalCommandJSON{
			Accepted:     false,
			RejectReason: "HTTP /v1/tactical/command disabled (set JANUS_HTTP_TACTICAL_COMMAND=1)",
		})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxTacticalCommandBodyBytes)
	var body tacticalCommandBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		if err == io.EOF {
			http.Error(w, "empty body", http.StatusBadRequest)
			return
		}
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if body.BattleID == "" {
		http.Error(w, "missing battle_id", http.StatusBadRequest)
		return
	}
	resp, err := gw.SubmitTacticalCommand(r.Context(), &gatewayv1.SubmitTacticalCommandRequest{
		SessionId: body.SessionID,
		BattleId:  body.BattleID,
		PlayerId:  body.PlayerID,
		Kind:      body.Kind,
		UnitId:    body.UnitID,
		ToX:       body.ToX,
		ToY:       body.ToY,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeTacticalCommandJSON(w, http.StatusOK, tacticalCommandJSON{
		Accepted:      resp.GetAccepted(),
		RejectReason:  resp.GetRejectReason(),
		LockstepFrame: resp.GetLockstepFrame(),
		StateHash:     resp.GetStateHash(),
	})
}

func writeTacticalCommandJSON(w http.ResponseWriter, status int, payload tacticalCommandJSON) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
