// Package roma holds authoritative in-memory battle partitions.
// Redis is allowed only for routing/cache metadata — never HP, position, or buff state.
package roma

import (
	"errors"
	"sync"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/board"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/observability/metrics"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timesync"
)

// BattleID identifies an in-memory battle instance.
type BattleID string

// Live zone ids select distinct tactical constructors in Join (see pkg/tactical/live.go).
const (
	ZoneLiveOccupy  = "live-occupy"
	ZoneLiveTimeout = "live-timeout"
)

// UnitState is legacy stub metadata (tests / docs); tactical truth lives in Match.
type UnitState struct {
	ID uint32
	X  int32
	Y  int32
	HP int32
}

// BattleState is the Roma partition authoritative snapshot.
type BattleState struct {
	ID      BattleID
	ZoneID  string
	Shard   uint32
	SimTime timesync.DualTime
	Match   *tactical.Match
}

// Store keeps battles keyed by id per zone shard.
type Store struct {
	mu       sync.RWMutex
	battles  map[BattleID]*BattleState
	clock    *timesync.Clock
	dilation *regionDilation
}

// NewStore creates an empty partition store.
func NewStore(clock *timesync.Clock) *Store {
	if clock == nil {
		clock = timesync.NewClock(nil)
	}
	return &Store{battles: make(map[BattleID]*BattleState), clock: clock, dilation: newRegionDilation()}
}

func newLiveMatchForZone(seed uint64, zoneID string) *tactical.Match {
	switch zoneID {
	case ZoneLiveOccupy:
		return tactical.NewLiveMatchOccupy(seed)
	case ZoneLiveTimeout:
		return tactical.NewLiveMatchTimeout(seed)
	default:
		return tactical.NewMatch(seed)
	}
}

// Join creates or returns a tactical duel for zone shard (in-memory only).
func (s *Store) Join(zoneID string, shard uint32) (*BattleState, error) {
	if zoneID == "" {
		return nil, errors.New("roma: empty zone")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := BattleID(zoneID + "/" + itoa(shard))
	if b, ok := s.battles[id]; ok {
		return b, nil
	}
	seed := tacticalSeed(zoneID, shard)
	b := &BattleState{
		ID:      id,
		ZoneID:  zoneID,
		Shard:   shard,
		SimTime: s.clock.Now(),
		Match:   newLiveMatchForZone(seed, zoneID),
	}
	s.battles[id] = b
	metrics.SetActiveRooms(len(s.battles))
	return b, nil
}

// Get returns battle by id.
func (s *Store) Get(id BattleID) (*BattleState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.battles[id]
	if !ok {
		return nil, errors.New("roma: battle not found")
	}
	return b, nil
}

// SubmitCommand applies a legacy delta move for the player's default duel unit.
func (s *Store) SubmitCommand(id BattleID, playerID uint32, moveX, moveY int32) (uint64, error) {
	var unitID uint32 = tactical.UnitIDPlayer0
	if playerID == 1 {
		unitID = tactical.UnitIDPlayer1
	}
	s.mu.RLock()
	b, ok := s.battles[id]
	s.mu.RUnlock()
	if !ok {
		return 0, errors.New("roma: battle not found")
	}
	if b.Match == nil {
		return 0, errors.New("roma: battle has no tactical match")
	}
	u := b.Match.Units[unitID]
	if u == nil {
		return 0, errors.New("roma: unit not found")
	}
	to := board.Coord{X: u.Pos.X + int(moveX), Y: u.Pos.Y + int(moveY)}
	_, hash, err := s.SubmitTacticalCommand(id, playerID, uint32(tactical.KindMove), unitID, int32(to.X), int32(to.Y), 0)
	return hash, err
}

func (b *BattleState) StateHash() uint64 {
	if b.Match != nil {
		return b.Match.StateHash()
	}
	return 0
}

func (b *BattleState) UnitCount() int32 {
	if b.Match == nil {
		return 0
	}
	return int32(len(b.Match.Units))
}

func itoa(v uint32) string {
	if v == 0 {
		return "0"
	}
	var buf [16]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
