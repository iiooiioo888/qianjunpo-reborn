package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	gatewayv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/gateway/v1"
)

const maxTacticalCommandBodyBytes = 16 << 10
const maxGatewayMirrorBodyBytes = 16 << 10

type tacticalCommandBody struct {
	SessionID string `json:"session_id"`
	BattleID  string `json:"battle_id"`
	PlayerID  uint32 `json:"player_id"`
	Kind      uint32 `json:"kind"`
	UnitID    uint32 `json:"unit_id"`
	ToX       int32  `json:"to_x"`
	ToY       int32  `json:"to_y"`
	SkillID   uint32 `json:"skill_id"`
}

type tacticalCommandJSON struct {
	Accepted      bool   `json:"accepted"`
	RejectReason  string `json:"reject_reason,omitempty"`
	LockstepFrame uint64 `json:"lockstep_frame,omitempty"`
	StateHash     uint64 `json:"state_hash,omitempty"`
}

type stepLockstepBody struct {
	SessionID string `json:"session_id"`
	BattleID  string `json:"battle_id"`
	Steps     uint32 `json:"steps"`
}

type stepLockstepJSON struct {
	LockstepFrame    uint64          `json:"lockstep_frame"`
	StateHash        uint64          `json:"state_hash"`
	Finished         bool            `json:"finished"`
	Winner           uint32          `json:"winner,omitempty"`
	ViewSnapshotJSON json.RawMessage `json:"view_snapshot_json,omitempty"`
}

type zoneRefJSON struct {
	ZoneID string `json:"zone_id"`
	Shard  uint32 `json:"shard"`
}

type dualTimeJSON struct {
	WallUnixMs int64 `json:"wall_unix_ms"`
	SimTick    int64 `json:"sim_tick"`
}

type connectBody struct {
	ClientVersion string       `json:"client_version"`
	AccessToken   string       `json:"access_token"`
	TargetZone    *zoneRefJSON `json:"target_zone"`
}

type connectJSON struct {
	SessionID    string        `json:"session_id"`
	ServerTime   *dualTimeJSON `json:"server_time,omitempty"`
	RomaEndpoint string        `json:"roma_endpoint,omitempty"`
}

type enterBattleBody struct {
	SessionID   string       `json:"session_id"`
	AccessToken string       `json:"access_token"`
	TargetZone  *zoneRefJSON `json:"target_zone"`
}

type enterBattleJSON struct {
	BattleID         string          `json:"battle_id"`
	InitialStateHash uint64          `json:"initial_state_hash"`
	SimTime          *dualTimeJSON   `json:"sim_time,omitempty"`
	ViewSnapshotJSON json.RawMessage `json:"view_snapshot_json,omitempty"`
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
	mux.HandleFunc("/v1/tactical/connect", func(w http.ResponseWriter, r *http.Request) {
		handleTacticalConnect(w, r, gw)
	})
	mux.HandleFunc("/v1/tactical/enter-battle", func(w http.ResponseWriter, r *http.Request) {
		handleTacticalEnterBattle(w, r, gw)
	})
	mux.HandleFunc("/v1/tactical/step-lockstep", func(w http.ResponseWriter, r *http.Request) {
		handleTacticalStepLockstep(w, r, gw)
	})
}

func zoneRefFromJSON(z *zoneRefJSON) *commonv1.ZoneRef {
	if z == nil {
		return &commonv1.ZoneRef{ZoneId: "default"}
	}
	zoneID := strings.TrimSpace(z.ZoneID)
	if zoneID == "" {
		zoneID = "default"
	}
	return &commonv1.ZoneRef{ZoneId: zoneID, Shard: z.Shard}
}

func dualTimeJSONFromProto(ts *commonv1.DualTimestamp) *dualTimeJSON {
	if ts == nil {
		return nil
	}
	return &dualTimeJSON{
		WallUnixMs: ts.GetWallUnixMs(),
		SimTick:    ts.GetSimTick(),
	}
}

func writeGatewayMirrorJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func decodeGatewayMirrorBody(w http.ResponseWriter, r *http.Request, dest interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxGatewayMirrorBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		if err == io.EOF {
			http.Error(w, "empty body", http.StatusBadRequest)
			return false
		}
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return false
	}
	return true
}

func handleTacticalConnect(w http.ResponseWriter, r *http.Request, gw *janusGateway) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body connectBody
	if !decodeGatewayMirrorBody(w, r, &body) {
		return
	}
	resp, err := gw.Connect(r.Context(), &gatewayv1.ConnectRequest{
		ClientVersion: body.ClientVersion,
		AccessToken:   body.AccessToken,
		TargetZone:    zoneRefFromJSON(body.TargetZone),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeGatewayMirrorJSON(w, http.StatusOK, connectJSON{
		SessionID:    resp.GetSessionId(),
		ServerTime:   dualTimeJSONFromProto(resp.GetServerTime()),
		RomaEndpoint: resp.GetRomaEndpoint(),
	})
}

func handleTacticalEnterBattle(w http.ResponseWriter, r *http.Request, gw *janusGateway) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body enterBattleBody
	if !decodeGatewayMirrorBody(w, r, &body) {
		return
	}
	resp, err := gw.EnterBattle(r.Context(), &gatewayv1.EnterBattleRequest{
		SessionId:   body.SessionID,
		AccessToken: body.AccessToken,
		TargetZone:  zoneRefFromJSON(body.TargetZone),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	var viewSnap json.RawMessage
	if raw := resp.GetViewSnapshotJson(); len(raw) > 0 {
		viewSnap = json.RawMessage(raw)
	}
	writeGatewayMirrorJSON(w, http.StatusOK, enterBattleJSON{
		BattleID:         resp.GetBattleId(),
		InitialStateHash: resp.GetInitialStateHash(),
		SimTime:          dualTimeJSONFromProto(resp.GetSimTime()),
		ViewSnapshotJSON: viewSnap,
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
		SkillId:   body.SkillID,
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

func handleTacticalStepLockstep(w http.ResponseWriter, r *http.Request, gw *janusGateway) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body stepLockstepBody
	if !decodeGatewayMirrorBody(w, r, &body) {
		return
	}
	if body.BattleID == "" {
		http.Error(w, "missing battle_id", http.StatusBadRequest)
		return
	}
	steps := body.Steps
	if steps == 0 {
		steps = 1
	}
	resp, err := gw.StepTacticalLockstep(r.Context(), &gatewayv1.StepTacticalLockstepRequest{
		SessionId: body.SessionID,
		BattleId:  body.BattleID,
		Steps:     steps,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	var viewSnap json.RawMessage
	if raw := resp.GetViewSnapshotJson(); len(raw) > 0 {
		viewSnap = json.RawMessage(raw)
	}
	writeGatewayMirrorJSON(w, http.StatusOK, stepLockstepJSON{
		LockstepFrame:    resp.GetLockstepFrame(),
		StateHash:        resp.GetStateHash(),
		Finished:         resp.GetFinished(),
		Winner:           resp.GetWinner(),
		ViewSnapshotJSON: viewSnap,
	})
}
