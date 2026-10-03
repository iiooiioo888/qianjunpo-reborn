package janus

import (
	"context"
	"fmt"
	"sync"
	"time"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	gatewayv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/gateway/v1"
	romav1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/roma/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RomaClient forwards gateway battle commands to RomaZone gRPC.
type RomaClient struct {
	mu      sync.Mutex
	clients map[string]romav1.RomaZoneClient
	dial    func(target string) (romav1.RomaZoneClient, error)
}

// SetDialer overrides gRPC client creation (tests / custom transports).
func (c *RomaClient) SetDialer(fn func(target string) (romav1.RomaZoneClient, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dial = fn
}

// NewRomaClient creates a client pool keyed by Roma gRPC target.
func NewRomaClient() *RomaClient {
	return &RomaClient{
		clients: make(map[string]romav1.RomaZoneClient),
		dial: func(target string) (romav1.RomaZoneClient, error) {
			conn, err := grpc.NewClient(target,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return nil, err
			}
			return romav1.NewRomaZoneClient(conn), nil
		},
	}
}

func (c *RomaClient) client(target string) (romav1.RomaZoneClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if cli, ok := c.clients[target]; ok {
		return cli, nil
	}
	cli, err := c.dial(target)
	if err != nil {
		return nil, err
	}
	c.clients[target] = cli
	return cli, nil
}

// EnterBattle joins a Roma zone shard and returns battle metadata.
func (c *RomaClient) EnterBattle(ctx context.Context, target, accessToken string, zone *commonv1.ZoneRef) (*gatewayv1.EnterBattleResponse, error) {
	cli, err := c.client(target)
	if err != nil {
		return nil, err
	}
	resp, err := cli.JoinZone(ctx, &romav1.JoinZoneRequest{
		AccessToken: accessToken,
		Zone:        zone,
	})
	if err != nil {
		return nil, err
	}
	state, err := cli.GetBattleState(ctx, &romav1.GetBattleStateRequest{BattleId: resp.GetBattleId()})
	if err != nil {
		return nil, err
	}
	snapJSON, _, _, err := c.tacticalViewSnapshot(ctx, cli, resp.GetBattleId())
	if err != nil {
		return nil, err
	}
	return &gatewayv1.EnterBattleResponse{
		BattleId:          resp.GetBattleId(),
		SimTime:           resp.GetSimTime(),
		InitialStateHash:  state.GetStateHash(),
		ViewSnapshotJson:  snapJSON,
	}, nil
}

// SubmitTacticalCommand forwards a tactical input to Roma.
func (c *RomaClient) SubmitTacticalCommand(ctx context.Context, target string, req *gatewayv1.SubmitTacticalCommandRequest) (*gatewayv1.SubmitTacticalCommandResponse, error) {
	cli, err := c.client(target)
	if err != nil {
		return nil, err
	}
	resp, err := cli.SubmitTacticalCommand(ctx, &romav1.SubmitTacticalCommandRequest{
		BattleId: req.GetBattleId(),
		Command: &romav1.TacticalCommand{
			PlayerId: req.GetPlayerId(),
			Kind:     req.GetKind(),
			UnitId:   req.GetUnitId(),
			ToX:      req.GetToX(),
			ToY:      req.GetToY(),
			SkillId:  req.GetSkillId(),
		},
	})
	if err != nil {
		return nil, err
	}
	return &gatewayv1.SubmitTacticalCommandResponse{
		Accepted:       resp.GetAccepted(),
		RejectReason:   resp.GetRejectReason(),
		StateHash:      resp.GetStateHash(),
		LockstepFrame:  resp.GetLockstepFrame(),
	}, nil
}

// StepTacticalLockstep advances Roma lockstep frames.
func (c *RomaClient) StepTacticalLockstep(ctx context.Context, target string, req *gatewayv1.StepTacticalLockstepRequest) (*gatewayv1.StepTacticalLockstepResponse, error) {
	cli, err := c.client(target)
	if err != nil {
		return nil, err
	}
	resp, err := cli.StepLockstep(ctx, &romav1.StepLockstepRequest{
		BattleId:         req.GetBattleId(),
		Steps:            req.GetSteps(),
		AutoCommand:      req.GetAutoCommand(),
		AutoCommandMode:  req.GetAutoCommandMode(),
	})
	if err != nil {
		return nil, err
	}
	snapJSON, _, _, err := c.tacticalViewSnapshot(ctx, cli, req.GetBattleId())
	if err != nil {
		return nil, err
	}
	return &gatewayv1.StepTacticalLockstepResponse{
		LockstepFrame:    resp.GetLockstepFrame(),
		StateHash:        resp.GetStateHash(),
		Finished:         resp.GetFinished(),
		Winner:           resp.GetWinner(),
		ViewSnapshotJson: snapJSON,
		AutoCommandMode:  resp.GetAutoCommandMode(),
	}, nil
}

// GetBattleSnapshot pulls Roma tactical view JSON for the Cocos display layer.
func (c *RomaClient) GetBattleSnapshot(ctx context.Context, target string, req *gatewayv1.GetBattleSnapshotRequest) (*gatewayv1.GetBattleSnapshotResponse, error) {
	cli, err := c.client(target)
	if err != nil {
		return nil, err
	}
	jsonBytes, hash, frame, err := c.tacticalViewSnapshot(ctx, cli, req.GetBattleId())
	if err != nil {
		return nil, err
	}
	return &gatewayv1.GetBattleSnapshotResponse{
		ViewSnapshotJson: jsonBytes,
		StateHash:        hash,
		LockstepFrame:    frame,
	}, nil
}

func (c *RomaClient) tacticalViewSnapshot(ctx context.Context, cli romav1.RomaZoneClient, battleID string) ([]byte, uint64, uint64, error) {
	resp, err := cli.GetTacticalViewSnapshot(ctx, &romav1.GetTacticalViewSnapshotRequest{BattleId: battleID})
	if err != nil {
		return nil, 0, 0, err
	}
	return resp.GetViewSnapshotJson(), resp.GetStateHash(), resp.GetLockstepFrame(), nil
}

// DialRoma is a test helper with a short dial timeout.
func DialRoma(target string) (romav1.RomaZoneClient, *grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, target, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, nil, fmt.Errorf("janus: dial roma %s: %w", target, err)
	}
	return romav1.NewRomaZoneClient(conn), conn, nil
}
