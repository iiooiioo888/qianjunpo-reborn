package tactical

import (
	"sort"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
)

// ControlPoint is a capture objective: sole occupancy for HoldFrames lockstep ticks wins.
type ControlPoint struct {
	Pos        board.Coord
	HoldFrames uint8
}

// NewMatchWithControlPoints builds a standard duel with capture objectives (does not affect NewMatch replays).
func NewMatchWithControlPoints(seed uint64, points []ControlPoint) *Match {
	m := NewMatch(seed)
	m.controlPoints = append([]ControlPoint(nil), points...)
	m.captureHold = make([]uint8, len(points)*PlayerCount)
	return m
}

func (m *Match) tickCapture() {
	if m.Finished || len(m.controlPoints) == 0 {
		return
	}
	for i, cp := range m.controlPoints {
		owners := uniqueOwnersOnCell(m, cp.Pos)
		base := i * PlayerCount
		if len(owners) != 1 {
			for p := 0; p < PlayerCount; p++ {
				m.captureHold[base+p] = 0
			}
			continue
		}
		owner := owners[0]
		for p := 0; p < PlayerCount; p++ {
			if uint8(p) == owner {
				m.captureHold[base+p]++
				if m.captureHold[base+p] >= cp.HoldFrames {
					m.Finished = true
					m.Winner = owner
					m.EndReason = EndCapture
					return
				}
			} else {
				m.captureHold[base+p] = 0
			}
		}
	}
}

func uniqueOwnersOnCell(m *Match, c board.Coord) []uint8 {
	seen := make(map[uint8]struct{})
	for _, u := range m.Units {
		if u == nil || u.Stats.HP.Raw() <= 0 {
			continue
		}
		if u.Pos != c {
			continue
		}
		seen[u.Owner] = struct{}{}
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([]uint8, 0, len(seen))
	for o := range seen {
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
