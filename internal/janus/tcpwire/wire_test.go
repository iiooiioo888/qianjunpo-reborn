package tcpwire

import (
	"bytes"
	"testing"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

func TestPacketRoundTrip(t *testing.T) {
	body := EncodeConnect("tok", "default", 0)
	var buf bytes.Buffer
	if err := WritePacket(&buf, OpConnect, body); err != nil {
		t.Fatal(err)
	}
	pkt, err := ReadPacket(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if pkt.Header.Opcode != OpConnect {
		t.Fatal(pkt.Header.Opcode)
	}
	tok, zone, shard, err := DecodeConnect(pkt.Body)
	if err != nil || tok != "tok" || zone != "default" || shard != 0 {
		t.Fatalf("%q %q %d err=%v", tok, zone, shard, err)
	}
}

func TestTacticalPayloadRoundTrip(t *testing.T) {
	cmd := tactical.Command{
		PlayerID: 0,
		Kind:     tactical.KindMove,
		UnitID:   tactical.UnitIDPlayer0,
		To:       board.Coord{X: 5, Y: 8},
	}
	body := EncodeSubmitTactical("sess", "default/0", cmd)
	_, _, got, err := DecodeSubmitTactical(body)
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != cmd.Kind || got.To != cmd.To {
		t.Fatalf("cmd mismatch %+v vs %+v", got, cmd)
	}
	ack := EncodeSubmitTacticalAck(true, "", 0xabc, 3)
	ok, _, hash, frame, err := DecodeSubmitTacticalAck(ack)
	if err != nil || !ok || hash != 0xabc || frame != 3 {
		t.Fatalf("ack %v %016x %d", ok, hash, frame)
	}
}
