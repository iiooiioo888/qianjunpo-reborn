// Demo: two lockstep clients with identical seeds and inputs produce matching hashes.
package main

import (
	"fmt"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/lockstep"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/sim"
)

func main() {
	const frames = 15
	sched := make([]sim.ScheduledCommand, frames)
	for f := 0; f < frames; f++ {
		sched[f] = sim.ScheduledCommand{
			SubmitFrame: uint64(f),
			Cmd: lockstep.CommandPacket{
				PlayerID: uint8(f % 2),
				MoveX:    1,
				MoveY:    -1,
			},
		}
	}
	a := sim.NewEngine(0xcafe)
	b := sim.NewEngine(0xcafe)
	ha := a.RunWithSchedule(sched, uint64(frames))
	hb := b.RunWithSchedule(sched, uint64(frames))
	fmt.Printf("lockstep frames=%d delay=%d subframes=%d\n", frames, lockstep.CommandDelayFrames, lockstep.GameTurnFrames)
	fmt.Printf("client A hash: %016x\n", ha)
	fmt.Printf("client B hash: %016x\n", hb)
	if ha == hb {
		fmt.Println("deterministic: OK")
	} else {
		fmt.Println("deterministic: DESYNC")
	}
}
