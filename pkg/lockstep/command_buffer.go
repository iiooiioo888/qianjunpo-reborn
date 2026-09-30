package lockstep

// CommandPacket is a per-player input for one lockstep frame (empty = optimistic timeout).
type CommandPacket struct {
	PlayerID uint8
	MoveX    int8
	MoveY    int8
}

// EmptyCommand returns a zeroed packet for optimistic frame fill.
func EmptyCommand(playerID uint8) CommandPacket {
	return CommandPacket{PlayerID: playerID}
}

// CommandBuffer queues submissions and releases them after CommandDelayFrames.
type CommandBuffer struct {
	delay       int
	byExecFrame map[uint64][]CommandPacket
}

// NewCommandBuffer creates a buffer with the standard command delay.
func NewCommandBuffer() *CommandBuffer {
	return &CommandBuffer{
		delay:       CommandDelayFrames,
		byExecFrame: make(map[uint64][]CommandPacket),
	}
}

// QueueSubmission records a command at lockstep frame submitFrame; it executes at submitFrame+delay.
func (b *CommandBuffer) QueueSubmission(submitFrame uint64, cmd CommandPacket) {
	execFrame := submitFrame + uint64(b.delay)
	b.byExecFrame[execFrame] = append(b.byExecFrame[execFrame], cmd)
}

// PopExecutable returns and removes commands due at execFrame.
func (b *CommandBuffer) PopExecutable(execFrame uint64) []CommandPacket {
	cmds := b.byExecFrame[execFrame]
	delete(b.byExecFrame, execFrame)
	return cmds
}

// FillOptimistic ensures each player has a command, using empty packets on timeout.
func FillOptimistic(commands []CommandPacket, playerCount int) []CommandPacket {
	seen := make(map[uint8]bool, playerCount)
	out := make([]CommandPacket, 0, playerCount)
	for _, c := range commands {
		if int(c.PlayerID) < playerCount && !seen[c.PlayerID] {
			seen[c.PlayerID] = true
			out = append(out, c)
		}
	}
	for pid := 0; pid < playerCount; pid++ {
		if !seen[uint8(pid)] {
			out = append(out, EmptyCommand(uint8(pid)))
		}
	}
	return out
}
