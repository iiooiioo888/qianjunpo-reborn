package lockstep

import "testing"

func TestCommandDelay(t *testing.T) {
	b := NewCommandBuffer()
	b.QueueSubmission(0, CommandPacket{PlayerID: 0, MoveX: 1})
	if b.PopExecutable(0) != nil {
		t.Fatal("should not execute at frame 0")
	}
	if b.PopExecutable(2) != nil {
		t.Fatal("should not execute at frame 2")
	}
	cmds := b.PopExecutable(3)
	if len(cmds) != 1 || cmds[0].MoveX != 1 {
		t.Fatalf("expected delayed cmd at frame 3, got %v", cmds)
	}
}

func TestFillOptimistic(t *testing.T) {
	out := FillOptimistic(nil, 2)
	if len(out) != 2 {
		t.Fatal("expected 2 empty cmds")
	}
	if out[0].MoveX != 0 || out[1].MoveX != 0 {
		t.Fatal("empty moves")
	}
}
