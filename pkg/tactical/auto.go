package tactical

import (
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/validate"
)

// AutoCommandMode controls simple AI command fill before each lockstep tick.
type AutoCommandMode uint8

const (
	AutoCommandOff AutoCommandMode = iota
	// AutoCommandBoth emits one simple AI command per alive side every tick (spectator / full auto).
	AutoCommandBoth
	// AutoCommandMissing fills only players who did not Submit since the last StepWithAuto.
	AutoCommandMissing
)

// SetAutoCommandMode enables or disables spectator-style auto command production.
func (m *Match) SetAutoCommandMode(mode AutoCommandMode) {
	if m == nil {
		return
	}
	m.autoMode = mode
}

// AutoCommandMode returns the current auto fill policy.
func (m *Match) AutoCommandMode() AutoCommandMode {
	if m == nil {
		return AutoCommandOff
	}
	return m.autoMode
}

// StepWithAuto runs optional simple AI submit for this frame, then StepLockstep.
func (m *Match) StepWithAuto() {
	if m == nil || m.Finished {
		if m != nil {
			m.StepLockstep()
		}
		return
	}
	if m.autoMode != AutoCommandOff {
		m.fillAutoCommands()
	}
	m.StepLockstep()
	for i := range m.submittedThisStep {
		m.submittedThisStep[i] = false
	}
}

// RunSpectatorAuto advances with AutoCommandBoth until finished or maxFrames.
func (m *Match) RunSpectatorAuto(maxFrames uint64) {
	prev := m.autoMode
	m.autoMode = AutoCommandBoth
	defer func() { m.autoMode = prev }()
	for !m.Finished && m.Frame < maxFrames {
		m.StepWithAuto()
	}
}

func (m *Match) fillAutoCommands() {
	for pid := uint8(0); pid < PlayerCount; pid++ {
		if m.autoMode == AutoCommandMissing && m.submittedThisStep[pid] {
			continue
		}
		if m.autoMode == AutoCommandBoth || !m.submittedThisStep[pid] {
			if cmd, ok := planSimpleAutoCommand(m, pid); ok {
				_ = m.Submit(cmd)
			}
		}
	}
}

func planSimpleAutoCommand(m *Match, playerID uint8) (Command, bool) {
	u := aliveUnitForOwner(m, playerID)
	if u == nil {
		return Command{}, false
	}
	enemy := aliveUnitForOwner(m, enemyOwner(playerID))
	if enemy == nil {
		return Command{PlayerID: playerID, Kind: KindPass, UnitID: u.ID, To: u.Pos}, true
	}
	atk := Command{PlayerID: playerID, Kind: KindAttack, UnitID: u.ID, To: enemy.Pos}
	if err := m.validateAttack(atk, u); err == nil {
		return atk, true
	}
	if dest, ok := bestMoveToward(m, u, enemy.Pos); ok {
		return Command{PlayerID: playerID, Kind: KindMove, UnitID: u.ID, To: dest}, true
	}
	return Command{PlayerID: playerID, Kind: KindPass, UnitID: u.ID, To: u.Pos}, true
}

func enemyOwner(playerID uint8) uint8 {
	if playerID == 0 {
		return 1
	}
	return 0
}

func aliveUnitForOwner(m *Match, owner uint8) *Unit {
	for _, u := range m.Units {
		if u != nil && u.Owner == owner && u.Stats.HP.Raw() > 0 {
			return u
		}
	}
	return nil
}

func coordLess(a, b board.Coord) bool {
	if a.Y != b.Y {
		return a.Y < b.Y
	}
	return a.X < b.X
}

func bestMoveToward(m *Match, u *Unit, target board.Coord) (board.Coord, bool) {
	bestDist := board.Chebyshev(u.Pos, target)
	var best board.Coord
	found := false
	for y := 0; y < board.Size; y++ {
		for x := 0; x < board.Size; x++ {
			to := board.Coord{X: x, Y: y}
			if to == u.Pos {
				continue
			}
			req := validate.MoveRequest{
				UnitID: u.ID, From: u.Pos, To: to, MovePoints: u.Stats.Move,
			}
			if _, err := m.validate.ValidateMove(req); err != nil {
				continue
			}
			d := board.Chebyshev(to, target)
			if d < bestDist || (d == bestDist && (!found || coordLess(to, best))) {
				bestDist = d
				best = to
				found = true
			}
		}
	}
	return best, found
}
