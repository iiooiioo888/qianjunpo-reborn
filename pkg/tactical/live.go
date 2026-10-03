package tactical

import (
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
)

// MaxTurnFrames is the lockstep tick threshold for timeout victory (Live / Roma step-lockstep).
const MaxTurnFrames = 64

// Live board anchors on the standard 19×19 duel (river y=9, pass at center).
var (
	// LiveSpawnP0 is the default P0 infantry spawn (board setup in NewMatch).
	LiveSpawnP0 = board.Coord{X: 2, Y: 8}
	// LiveSpawnP1 is the default P1 cavalry spawn.
	LiveSpawnP1 = board.Coord{X: 16, Y: 10}
	// LiveCenterPass is the bridged pass tile (9,9).
	LiveCenterPass = board.Coord{X: 9, Y: 9}
)

// LiveOccupyHoldFrames is the sole-occupancy hold length for Live occupy presets (lockstep ticks).
const LiveOccupyHoldFrames = 3

// LiveControlPointsOccupy returns the canonical Live occupy objective: P0 spawn tile.
// P0 starts on this cell; full-auto with LiveAutoOccupy holds the point before wipeout.
func LiveControlPointsOccupy() []ControlPoint {
	return []ControlPoint{
		{Pos: LiveSpawnP0, HoldFrames: LiveOccupyHoldFrames},
	}
}

// LiveControlPointsOccupyCenter is an alternate occupy objective on the center pass (9,9).
func LiveControlPointsOccupyCenter() []ControlPoint {
	return []ControlPoint{
		{Pos: LiveCenterPass, HoldFrames: LiveOccupyHoldFrames},
	}
}

// LiveAutoProfile steers simple AI under AutoCommandBoth / AutoCommandMissing (auto_command_mode wire).
type LiveAutoProfile uint8

const (
	// LiveAutoDefault pursues and attacks (legacy spectator wipeout on the default duel).
	LiveAutoDefault LiveAutoProfile = 0
	// LiveAutoOccupy prioritizes holding or reaching control points (use with LiveControlPointsOccupy).
	LiveAutoOccupy LiveAutoProfile = 1
	// LiveAutoTimeout avoids combat and stalls until MaxTurnFrames (timeout).
	LiveAutoTimeout LiveAutoProfile = 2
)

// SetLiveAutoProfile selects occupy / timeout / default simple AI behavior.
func (m *Match) SetLiveAutoProfile(profile LiveAutoProfile) {
	if m == nil {
		return
	}
	m.liveAutoProfile = profile
}

// LiveAutoProfile returns the current Live simple AI steering mode.
func (m *Match) LiveAutoProfile() LiveAutoProfile {
	if m == nil {
		return LiveAutoDefault
	}
	return m.liveAutoProfile
}

// NewLiveMatchOccupy builds the standard duel with Live occupy control points and LiveAutoOccupy AI profile.
// Core: pair with auto_command_mode=2 (both) on step-lockstep after Join placement is done.
func NewLiveMatchOccupy(seed uint64) *Match {
	m := NewMatchWithControlPoints(seed, LiveControlPointsOccupy())
	m.SetLiveAutoProfile(LiveAutoOccupy)
	return m
}

// NewLiveMatchOccupyCenter is like NewLiveMatchOccupy but the objective is LiveCenterPass (9,9).
func NewLiveMatchOccupyCenter(seed uint64) *Match {
	m := NewMatchWithControlPoints(seed, LiveControlPointsOccupyCenter())
	m.SetLiveAutoProfile(LiveAutoOccupy)
	return m
}

// NewLiveMatchTimeout builds the standard duel without control points and LiveAutoTimeout AI profile.
// Full auto passes each tick so the match reaches MaxTurnFrames without annihilation.
func NewLiveMatchTimeout(seed uint64) *Match {
	m := NewMatch(seed)
	m.SetLiveAutoProfile(LiveAutoTimeout)
	return m
}

// NewLiveMatchWithControlPoints is for Core wiring: duel + explicit points + occupy AI profile.
func NewLiveMatchWithControlPoints(seed uint64, points []ControlPoint) *Match {
	m := NewMatchWithControlPoints(seed, points)
	m.SetLiveAutoProfile(LiveAutoOccupy)
	return m
}
