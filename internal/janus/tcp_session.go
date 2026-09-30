package janus

import (
	"context"
	"net"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	gatewayv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/gateway/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/janus/tcpwire"
)

// TacticalGateway is the subset of JanusGateway used by the TCP bridge.
type TacticalGateway interface {
	Connect(ctx context.Context, req *gatewayv1.ConnectRequest) (*gatewayv1.ConnectResponse, error)
	EnterBattle(ctx context.Context, req *gatewayv1.EnterBattleRequest) (*gatewayv1.EnterBattleResponse, error)
	SubmitTacticalCommand(ctx context.Context, req *gatewayv1.SubmitTacticalCommandRequest) (*gatewayv1.SubmitTacticalCommandResponse, error)
	StepTacticalLockstep(ctx context.Context, req *gatewayv1.StepTacticalLockstepRequest) (*gatewayv1.StepTacticalLockstepResponse, error)
}

// ServeTCPBridge handles framed tactical packets on one connection.
func ServeTCPBridge(conn net.Conn, gw TacticalGateway) {
	defer conn.Close()
	ctx := context.Background()
	var sessionID string
	for {
		pkt, err := tcpwire.ReadPacket(conn)
		if err != nil {
			return
		}
		switch pkt.Header.Opcode {
		case tcpwire.OpConnect:
			if err := handleTCPConnect(ctx, conn, gw, pkt.Body, &sessionID); err != nil {
				_ = tcpwire.WritePacket(conn, tcpwire.OpError, tcpwire.EncodeError(err.Error()))
				return
			}
		case tcpwire.OpEnterBattle:
			if err := handleTCPEnterBattle(ctx, conn, gw, pkt.Body, sessionID); err != nil {
				_ = tcpwire.WritePacket(conn, tcpwire.OpError, tcpwire.EncodeError(err.Error()))
				return
			}
		case tcpwire.OpSubmitTactical:
			if err := handleTCPSubmit(ctx, conn, gw, pkt.Body); err != nil {
				_ = tcpwire.WritePacket(conn, tcpwire.OpError, tcpwire.EncodeError(err.Error()))
				return
			}
		case tcpwire.OpStepLockstep:
			if err := handleTCPStep(ctx, conn, gw, pkt.Body); err != nil {
				_ = tcpwire.WritePacket(conn, tcpwire.OpError, tcpwire.EncodeError(err.Error()))
				return
			}
		default:
			_ = tcpwire.WritePacket(conn, tcpwire.OpError, tcpwire.EncodeError("unknown opcode"))
			return
		}
	}
}

func handleTCPConnect(ctx context.Context, conn net.Conn, gw TacticalGateway, body []byte, sessionID *string) error {
	token, zoneID, shard, err := tcpwire.DecodeConnect(body)
	if err != nil {
		return err
	}
	resp, err := gw.Connect(ctx, &gatewayv1.ConnectRequest{
		AccessToken: token,
		TargetZone:  &commonv1.ZoneRef{ZoneId: zoneID, Shard: shard},
	})
	if err != nil {
		return err
	}
	*sessionID = resp.GetSessionId()
	ack := tcpwire.EncodeConnectAck(
		resp.GetSessionId(),
		resp.GetRomaEndpoint(),
		uint64(resp.GetServerTime().GetWallUnixMs()),
		uint64(resp.GetServerTime().GetSimTick()),
	)
	return tcpwire.WritePacket(conn, tcpwire.OpConnectAck, ack)
}

func handleTCPEnterBattle(ctx context.Context, conn net.Conn, gw TacticalGateway, body []byte, fallbackSession string) error {
	sess, token, zoneID, shard, err := tcpwire.DecodeEnterBattle(body)
	if err != nil {
		return err
	}
	if sess == "" {
		sess = fallbackSession
	}
	resp, err := gw.EnterBattle(ctx, &gatewayv1.EnterBattleRequest{
		SessionId:   sess,
		AccessToken: token,
		TargetZone:  &commonv1.ZoneRef{ZoneId: zoneID, Shard: shard},
	})
	if err != nil {
		return err
	}
	ack := tcpwire.EncodeEnterBattleAck(resp.GetBattleId(), resp.GetInitialStateHash())
	return tcpwire.WritePacket(conn, tcpwire.OpEnterBattleAck, ack)
}

func handleTCPSubmit(ctx context.Context, conn net.Conn, gw TacticalGateway, body []byte) error {
	sess, battleID, cmd, err := tcpwire.DecodeSubmitTactical(body)
	if err != nil {
		return err
	}
	resp, err := gw.SubmitTacticalCommand(ctx, &gatewayv1.SubmitTacticalCommandRequest{
		SessionId: sess,
		BattleId:  battleID,
		PlayerId:  uint32(cmd.PlayerID),
		Kind:      uint32(cmd.Kind),
		UnitId:    cmd.UnitID,
		ToX:       int32(cmd.To.X),
		ToY:       int32(cmd.To.Y),
	})
	if err != nil {
		return err
	}
	ack := tcpwire.EncodeSubmitTacticalAck(
		resp.GetAccepted(),
		resp.GetRejectReason(),
		resp.GetStateHash(),
		resp.GetLockstepFrame(),
	)
	return tcpwire.WritePacket(conn, tcpwire.OpSubmitTacticalAck, ack)
}

func handleTCPStep(ctx context.Context, conn net.Conn, gw TacticalGateway, body []byte) error {
	sess, battleID, steps, err := tcpwire.DecodeStepLockstep(body)
	if err != nil {
		return err
	}
	resp, err := gw.StepTacticalLockstep(ctx, &gatewayv1.StepTacticalLockstepRequest{
		SessionId: sess,
		BattleId:  battleID,
		Steps:     steps,
	})
	if err != nil {
		return err
	}
	ack := tcpwire.EncodeStepLockstepAck(
		resp.GetLockstepFrame(),
		resp.GetStateHash(),
		resp.GetFinished(),
		resp.GetWinner(),
	)
	return tcpwire.WritePacket(conn, tcpwire.OpStepLockstepAck, ack)
}
