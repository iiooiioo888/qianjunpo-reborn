package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	gatewayv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/gateway/v1"
	romav1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/roma/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/janus"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/roma"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/combat"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/lockstep"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timesync"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// janusRomaHTTPFixture wires in-process Janus gRPC + mock Roma for HTTP mirror tests.
type janusRomaHTTPFixture struct {
	GW               *janusGateway
	Store            *roma.Store
	BattleID         string
	SessionID        string
	InitialStateHash uint64
}

func TestHTTPTacticalConnectMirror(t *testing.T) {
	fix := startJanusGatewayRomaOnly(t)
	srv := newTacticalHTTPServer(t, fix.GW, defaultTacticalHTTPOptions())

	raw := []byte(`{"access_token":"test-token","target_zone":{"zone_id":"default","shard":0}}`)
	res, err := http.Post(srv.URL+"/v1/tactical/connect", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d: %s", res.StatusCode, b)
	}
	var out connectJSON
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.SessionID == "" {
		t.Fatal("expected non-empty session_id")
	}
	if out.ServerTime == nil || out.ServerTime.WallUnixMs == 0 {
		t.Fatal("expected server_time in connect response")
	}
}

func TestHTTPTacticalEnterBattleMirror(t *testing.T) {
	fix := startJanusGatewayRomaOnly(t)
	srv := newTacticalHTTPServer(t, fix.GW, defaultTacticalHTTPOptions())

	sessionRaw := []byte(`{"access_token":"test-token","target_zone":{"zone_id":"default","shard":0}}`)
	sessionRes, err := http.Post(srv.URL+"/v1/tactical/connect", "application/json", bytes.NewReader(sessionRaw))
	if err != nil {
		t.Fatal(err)
	}
	defer sessionRes.Body.Close()
	var connectOut connectJSON
	if err := json.NewDecoder(sessionRes.Body).Decode(&connectOut); err != nil {
		t.Fatal(err)
	}

	enterBody := map[string]interface{}{
		"session_id":   connectOut.SessionID,
		"access_token": "test-token",
		"target_zone":  map[string]interface{}{"zone_id": "default", "shard": 0},
	}
	enterRaw, _ := json.Marshal(enterBody)
	enterRes, err := http.Post(srv.URL+"/v1/tactical/enter-battle", "application/json", bytes.NewReader(enterRaw))
	if err != nil {
		t.Fatal(err)
	}
	defer enterRes.Body.Close()
	if enterRes.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(enterRes.Body)
		t.Fatalf("status %d: %s", enterRes.StatusCode, b)
	}
	var enterOut enterBattleJSON
	if err := json.NewDecoder(enterRes.Body).Decode(&enterOut); err != nil {
		t.Fatal(err)
	}
	if enterOut.BattleID == "" {
		t.Fatal("expected battle_id from HTTP EnterBattle")
	}
	if enterOut.InitialStateHash == 0 {
		t.Fatal("expected initial_state_hash")
	}
	if len(enterOut.ViewSnapshotJSON) == 0 {
		t.Fatal("expected view_snapshot_json")
	}
}

func TestHTTPTacticalEnterBattleThenCommandAccepted(t *testing.T) {
	fix := startJanusGatewayRomaOnly(t)
	srv := newTacticalHTTPServer(t, fix.GW, defaultTacticalHTTPOptions())

	sessionRes, err := http.Post(
		srv.URL+"/v1/tactical/connect",
		"application/json",
		bytes.NewReader([]byte(`{"access_token":"test-token","target_zone":{"zone_id":"default","shard":0}}`)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer sessionRes.Body.Close()
	var connectOut connectJSON
	if err := json.NewDecoder(sessionRes.Body).Decode(&connectOut); err != nil {
		t.Fatal(err)
	}

	enterRaw, _ := json.Marshal(map[string]interface{}{
		"session_id":   connectOut.SessionID,
		"access_token": "test-token",
		"target_zone":  map[string]interface{}{"zone_id": "default", "shard": 0},
	})
	enterRes, err := http.Post(srv.URL+"/v1/tactical/enter-battle", "application/json", bytes.NewReader(enterRaw))
	if err != nil {
		t.Fatal(err)
	}
	defer enterRes.Body.Close()
	var enterOut enterBattleJSON
	if err := json.NewDecoder(enterRes.Body).Decode(&enterOut); err != nil {
		t.Fatal(err)
	}

	cmdRaw, _ := json.Marshal(map[string]interface{}{
		"session_id": connectOut.SessionID,
		"battle_id":  enterOut.BattleID,
		"player_id":  0,
		"kind":       uint32(tactical.KindMove),
		"unit_id":    tactical.UnitIDPlayer0,
		"to_x":       5,
		"to_y":       8,
	})
	cmdRes, err := http.Post(srv.URL+"/v1/tactical/command", "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		t.Fatal(err)
	}
	defer cmdRes.Body.Close()
	if cmdRes.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(cmdRes.Body)
		t.Fatalf("status %d: %s", cmdRes.StatusCode, b)
	}
	var cmdOut tacticalCommandJSON
	if err := json.NewDecoder(cmdRes.Body).Decode(&cmdOut); err != nil {
		t.Fatal(err)
	}
	if !cmdOut.Accepted {
		t.Fatalf("expected accepted command after HTTP EnterBattle, reason=%q", cmdOut.RejectReason)
	}
}

func TestHTTPTacticalStepLockstepAdvancesFrame(t *testing.T) {
	fix := startJanusGatewayWithEnterBattle(t)
	srv := newTacticalHTTPServer(t, fix.GW, defaultTacticalHTTPOptions())

	cmdRaw, _ := json.Marshal(map[string]interface{}{
		"session_id": fix.SessionID,
		"battle_id":  fix.BattleID,
		"player_id":  0,
		"kind":       uint32(tactical.KindMove),
		"unit_id":    tactical.UnitIDPlayer0,
		"to_x":       5,
		"to_y":       8,
	})
	cmdRes, err := http.Post(srv.URL+"/v1/tactical/command", "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		t.Fatal(err)
	}
	defer cmdRes.Body.Close()
	var cmdOut tacticalCommandJSON
	if err := json.NewDecoder(cmdRes.Body).Decode(&cmdOut); err != nil {
		t.Fatal(err)
	}
	if !cmdOut.Accepted {
		t.Fatalf("command rejected: %s", cmdOut.RejectReason)
	}

	snapRes, err := http.Get(srv.URL + "/v1/tactical/snapshot?battle_id=" + fix.BattleID)
	if err != nil {
		t.Fatal(err)
	}
	defer snapRes.Body.Close()
	var before tactical.ViewSnapshot
	if err := json.NewDecoder(snapRes.Body).Decode(&before); err != nil {
		t.Fatal(err)
	}
	frameBefore := before.LockstepFrame

	stepRaw, _ := json.Marshal(map[string]interface{}{
		"session_id": fix.SessionID,
		"battle_id":  fix.BattleID,
		"steps":      1,
	})
	stepRes, err := http.Post(srv.URL+"/v1/tactical/step-lockstep", "application/json", bytes.NewReader(stepRaw))
	if err != nil {
		t.Fatal(err)
	}
	defer stepRes.Body.Close()
	if stepRes.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(stepRes.Body)
		t.Fatalf("step-lockstep status %d: %s", stepRes.StatusCode, b)
	}
	var stepOut stepLockstepJSON
	if err := json.NewDecoder(stepRes.Body).Decode(&stepOut); err != nil {
		t.Fatal(err)
	}
	if stepOut.LockstepFrame <= frameBefore {
		t.Fatalf("expected frame advance from %d, got %d", frameBefore, stepOut.LockstepFrame)
	}
	if stepOut.StateHash == 0 || len(stepOut.ViewSnapshotJSON) == 0 {
		t.Fatalf("expected state_hash and view_snapshot_json, got %+v", stepOut)
	}
}

func TestHTTPTacticalCommandMirror(t *testing.T) {
	fix := startJanusGatewayWithEnterBattle(t)
	srv := newTacticalHTTPServer(t, fix.GW, tacticalHTTPOptions{
		CommandMirrorEnabled: func() bool { return true },
	})

	body := map[string]interface{}{
		"session_id": fix.SessionID,
		"battle_id":  fix.BattleID,
		"player_id":  0,
		"kind":       uint32(tactical.KindMove),
		"unit_id":    tactical.UnitIDPlayer0,
		"to_x":       5,
		"to_y":       8,
	}
	raw, _ := json.Marshal(body)
	res, err := http.Post(srv.URL+"/v1/tactical/command", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d: %s", res.StatusCode, b)
	}
	var out tacticalCommandJSON
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if !out.Accepted {
		t.Fatalf("expected accepted command after EnterBattle, reason=%q", out.RejectReason)
	}
	if out.StateHash == 0 {
		t.Fatal("expected non-zero state_hash on accepted command")
	}
	if out.RejectReason != "" {
		t.Fatalf("unexpected reject_reason on success: %q", out.RejectReason)
	}
}

func TestHTTPTacticalCommandMirrorWithoutEnterBattle(t *testing.T) {
	fix := startJanusGatewayRomaOnly(t)
	srv := newTacticalHTTPServer(t, fix.GW, defaultTacticalHTTPOptions())

	// Docs/dev smoke often use default/0; without EnterBattle (JoinZone) Roma has no battle.
	raw := []byte(`{"battle_id":"default/0","player_id":0,"kind":1,"unit_id":1,"to_x":5,"to_y":8}`)
	res, err := http.Post(srv.URL+"/v1/tactical/command", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d: %s", res.StatusCode, b)
	}
	var out tacticalCommandJSON
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Accepted {
		t.Fatal("expected rejected command when battle was never created")
	}
	if !strings.Contains(out.RejectReason, "battle not found") {
		t.Fatalf("expected roma battle not found, got reason=%q", out.RejectReason)
	}
}

func TestHTTPTacticalSnapshotAfterEnterBattle(t *testing.T) {
	fix := startJanusGatewayWithEnterBattle(t)
	srv := newTacticalHTTPServer(t, fix.GW, defaultTacticalHTTPOptions())

	res, err := http.Get(srv.URL + "/v1/tactical/snapshot?battle_id=" + fix.BattleID)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d: %s", res.StatusCode, b)
	}
	if res.Header.Get("X-State-Hash") == "" {
		t.Fatal("missing X-State-Hash header")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) == 0 {
		t.Fatal("empty snapshot body")
	}
	var view tactical.ViewSnapshot
	if err := json.Unmarshal(body, &view); err != nil {
		t.Fatalf("snapshot json: %v", err)
	}
	if view.LockstepFrame != 0 {
		t.Fatalf("expected initial lockstep frame 0, got %d", view.LockstepFrame)
	}
}

func TestHTTPTacticalCommandMirrorDisabled(t *testing.T) {
	fix := startJanusGatewayWithEnterBattle(t)
	srv := newTacticalHTTPServer(t, fix.GW, tacticalHTTPOptions{
		CommandMirrorEnabled: func() bool { return false },
	})

	raw := []byte(`{"battle_id":"` + fix.BattleID + `","kind":1}`)
	res, err := http.Post(srv.URL+"/v1/tactical/command", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", res.StatusCode)
	}
	var out tacticalCommandJSON
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Accepted || out.RejectReason == "" {
		t.Fatalf("expected disabled reject payload, got %+v", out)
	}
}

func TestHTTPTacticalCommandKindSkillRequiresSkillID(t *testing.T) {
	fix := startJanusGatewayWithEnterBattle(t)
	srv := newTacticalHTTPServer(t, fix.GW, defaultTacticalHTTPOptions())

	raw, _ := json.Marshal(map[string]interface{}{
		"session_id": fix.SessionID,
		"battle_id":  fix.BattleID,
		"player_id":  0,
		"kind":       uint32(tactical.KindSkill),
		"unit_id":    tactical.UnitIDPlayer0,
		"to_x":       9,
		"to_y":       9,
	})
	res, err := http.Post(srv.URL+"/v1/tactical/command", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var out tacticalCommandJSON
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Accepted || !strings.Contains(out.RejectReason, "skill_id") {
		t.Fatalf("expected skill_id required, got %+v", out)
	}
}

func TestHTTPTacticalSnapshotPreservesLastSkillCastAndHP(t *testing.T) {
	fix := startJanusGatewayWithEnterBattle(t)
	srv := newTacticalHTTPServer(t, fix.GW, defaultTacticalHTTPOptions())

	b, err := fix.Store.Get(roma.BattleID(fix.BattleID))
	if err != nil {
		t.Fatal(err)
	}
	attacker := b.Match.Units[tactical.UnitIDPlayer0]
	neighbor := board.Coord{X: attacker.Pos.X + 1, Y: attacker.Pos.Y}
	b.Match.Board.ClearUnit(b.Match.Units[tactical.UnitIDPlayer1].Pos)
	defender := b.Match.Units[tactical.UnitIDPlayer1]
	defender.Pos = neighbor
	b.Match.Units[tactical.UnitIDPlayer1] = defender
	if !b.Match.Board.SetUnit(neighbor, tactical.UnitIDPlayer1) {
		t.Fatal("place defender adjacent")
	}
	hpBefore := defender.Stats.HP.Raw()

	skillBody, _ := json.Marshal(map[string]interface{}{
		"session_id": fix.SessionID,
		"battle_id":  fix.BattleID,
		"player_id":  0,
		"kind":       uint32(tactical.KindSkill),
		"unit_id":    tactical.UnitIDPlayer0,
		"to_x":       neighbor.X,
		"to_y":       neighbor.Y,
		"skill_id":   uint32(combat.SkillStubStrike),
	})
	cmdRes, err := http.Post(srv.URL+"/v1/tactical/command", "application/json", bytes.NewReader(skillBody))
	if err != nil {
		t.Fatal(err)
	}
	defer cmdRes.Body.Close()
	var cmdOut tacticalCommandJSON
	if err := json.NewDecoder(cmdRes.Body).Decode(&cmdOut); err != nil {
		t.Fatal(err)
	}
	if !cmdOut.Accepted {
		t.Fatalf("skill command rejected: %s", cmdOut.RejectReason)
	}

	steps := int(lockstep.CommandDelayFrames) + 1
	stepRaw, _ := json.Marshal(map[string]interface{}{
		"session_id": fix.SessionID,
		"battle_id":  fix.BattleID,
		"steps":      steps,
	})
	stepRes, err := http.Post(srv.URL+"/v1/tactical/step-lockstep", "application/json", bytes.NewReader(stepRaw))
	if err != nil {
		t.Fatal(err)
	}
	defer stepRes.Body.Close()
	var stepOut stepLockstepJSON
	if err := json.NewDecoder(stepRes.Body).Decode(&stepOut); err != nil {
		t.Fatal(err)
	}
	var stepSnap tactical.ViewSnapshot
	if err := json.Unmarshal(stepOut.ViewSnapshotJSON, &stepSnap); err != nil {
		t.Fatal(err)
	}
	if stepSnap.LastSkillCast == nil {
		t.Fatal("step-lockstep view_snapshot_json missing lastSkillCast")
	}
	if stepSnap.LastSkillCast.SkillID != uint16(combat.SkillStubStrike) {
		t.Fatalf("skillId=%d", stepSnap.LastSkillCast.SkillID)
	}
	var victimHP int64
	for _, u := range stepSnap.Units {
		if u.ID == tactical.UnitIDPlayer1 {
			victimHP = u.HP
			break
		}
	}
	if victimHP >= hpBefore {
		t.Fatalf("expected defender hp drop from %d, got %d", hpBefore, victimHP)
	}

	snapRes, err := http.Get(srv.URL + "/v1/tactical/snapshot?battle_id=" + fix.BattleID)
	if err != nil {
		t.Fatal(err)
	}
	defer snapRes.Body.Close()
	var snap tactical.ViewSnapshot
	if err := json.NewDecoder(snapRes.Body).Decode(&snap); err != nil {
		t.Fatal(err)
	}
	if snap.LastSkillCast == nil || snap.LastSkillCast.SkillID != uint16(combat.SkillStubStrike) {
		t.Fatalf("GET snapshot lastSkillCast: %+v", snap.LastSkillCast)
	}
	for _, u := range snap.Units {
		if u.ID == tactical.UnitIDPlayer1 && u.HP >= hpBefore {
			t.Fatalf("GET snapshot hp not updated: %d", u.HP)
		}
	}
}

func TestHTTPTacticalSnapshotVictoryOutcomeFields(t *testing.T) {
	fix := startJanusGatewayWithEnterBattle(t)
	srv := newTacticalHTTPServer(t, fix.GW, defaultTacticalHTTPOptions())

	b, err := fix.Store.Get(roma.BattleID(fix.BattleID))
	if err != nil {
		t.Fatal(err)
	}
	def := b.Match.Units[tactical.UnitIDPlayer1]
	def.Stats.HP = def.Stats.HP.Sub(def.Stats.HP)
	b.Match.Board.ClearUnit(def.Pos)

	stepRaw, _ := json.Marshal(map[string]interface{}{
		"session_id": fix.SessionID,
		"battle_id":  fix.BattleID,
		"steps":      1,
	})
	stepRes, err := http.Post(srv.URL+"/v1/tactical/step-lockstep", "application/json", bytes.NewReader(stepRaw))
	if err != nil {
		t.Fatal(err)
	}
	defer stepRes.Body.Close()
	var stepOut stepLockstepJSON
	if err := json.NewDecoder(stepRes.Body).Decode(&stepOut); err != nil {
		t.Fatal(err)
	}
	if !stepOut.Finished {
		t.Fatal("expected finished step-lockstep response")
	}
	var stepSnap tactical.ViewSnapshot
	if err := json.Unmarshal(stepOut.ViewSnapshotJSON, &stepSnap); err != nil {
		t.Fatal(err)
	}
	if !stepSnap.Finished || stepSnap.Winner == nil || *stepSnap.Winner != 0 || stepSnap.EndReason != tactical.EndAnnihilation {
		t.Fatalf("step snapshot outcome: finished=%v winner=%v reason=%s", stepSnap.Finished, stepSnap.Winner, stepSnap.EndReason)
	}

	snapRes, err := http.Get(srv.URL + "/v1/tactical/snapshot?battle_id=" + fix.BattleID)
	if err != nil {
		t.Fatal(err)
	}
	defer snapRes.Body.Close()
	var snap tactical.ViewSnapshot
	if err := json.NewDecoder(snapRes.Body).Decode(&snap); err != nil {
		t.Fatal(err)
	}
	if !snap.Finished || snap.Winner == nil || *snap.Winner != 0 || snap.EndReason != tactical.EndAnnihilation {
		t.Fatalf("GET snapshot outcome: finished=%v winner=%v reason=%s", snap.Finished, snap.Winner, snap.EndReason)
	}
}

func TestHTTPTacticalCommandMethodNotAllowed(t *testing.T) {
	fix := startJanusGatewayWithEnterBattle(t)
	srv := newTacticalHTTPServer(t, fix.GW, defaultTacticalHTTPOptions())

	res, err := http.Get(srv.URL + "/v1/tactical/command")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", res.StatusCode)
	}
}

func newTacticalHTTPServer(t *testing.T, gw *janusGateway, opts tacticalHTTPOptions) *httptest.Server {
	mux := http.NewServeMux()
	registerTacticalHTTPRoutes(mux, gw, opts)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// startJanusGatewayRomaOnly stands up Janus + mock Roma without EnterBattle (no live battle).
func startJanusGatewayRomaOnly(t *testing.T) janusRomaHTTPFixture {
	gw, store, _ := startJanusGRPCWithRoma(t)
	return janusRomaHTTPFixture{GW: gw, Store: store, BattleID: "default/0"}
}

// startJanusGatewayWithEnterBattle runs Connect → EnterBattle so Roma holds an in-memory battle.
func startJanusGatewayWithEnterBattle(t *testing.T) janusRomaHTTPFixture {
	gw, store, jc := startJanusGRPCWithRoma(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	connect, err := jc.Connect(ctx, &gatewayv1.ConnectRequest{
		AccessToken: "test-token",
		TargetZone:  &commonv1.ZoneRef{ZoneId: "default", Shard: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	enter, err := jc.EnterBattle(ctx, &gatewayv1.EnterBattleRequest{
		SessionId:   connect.GetSessionId(),
		AccessToken: "test-token",
		TargetZone:  &commonv1.ZoneRef{ZoneId: "default", Shard: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	battleID := enter.GetBattleId()
	if battleID == "" || enter.GetInitialStateHash() == 0 {
		t.Fatal("missing battle id or initial state hash from EnterBattle")
	}
	return janusRomaHTTPFixture{
		GW:               gw,
		Store:            store,
		BattleID:         battleID,
		SessionID:        connect.GetSessionId(),
		InitialStateHash: enter.GetInitialStateHash(),
	}
}

func startJanusGRPCWithRoma(t *testing.T) (*janusGateway, *roma.Store, gatewayv1.JanusGatewayClient) {
	romaLis := bufconn.Listen(bufSize)
	romaSrv := grpc.NewServer()
	romaStore := roma.NewStore(nil)
	romav1.RegisterRomaZoneServer(romaSrv, newE2ERomaServer(romaStore))
	go func() { _ = romaSrv.Serve(romaLis) }()
	t.Cleanup(romaSrv.Stop)

	romaDialer := func(context.Context, string) (net.Conn, error) {
		return romaLis.Dial()
	}
	romaConn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(romaDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = romaConn.Close() })
	romaTarget := "passthrough:///bufnet"

	gw := &janusGateway{
		clock: timesync.NewClock(nil),
		limit: janus.NewRateLimiter(10000),
		disco: &janus.Discovery{Static: map[string]string{"default": romaTarget}, Default: romaTarget},
		auth:  janus.StaticAuth{},
		roma:  janus.NewRomaClient(),
	}
	gw.roma.SetDialer(func(target string) (romav1.RomaZoneClient, error) {
		return romav1.NewRomaZoneClient(romaConn), nil
	})

	janusLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	janusSrv := grpc.NewServer()
	gatewayv1.RegisterJanusGatewayServer(janusSrv, gw)
	go func() { _ = janusSrv.Serve(janusLis) }()
	t.Cleanup(janusSrv.Stop)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	janusConn, err := grpc.DialContext(ctx, janusLis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = janusConn.Close() })
	return gw, romaStore, gatewayv1.NewJanusGatewayClient(janusConn)
}
