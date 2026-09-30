// Package roma holds authoritative in-memory battle partitions.
// Redis is allowed only for routing/cache metadata — never HP, position, or buff state.
package roma

import (
	"errors"
	"sync"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/hash"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timesync"
)

// BattleID identifies an in-memory battle instance.
type BattleID string

// UnitState is authoritative combat state (memory-only).
type UnitState struct {
	ID uint32
	X  int32
	Y  int32
	HP int32
}

// BattleState is the Roma partition authoritative snapshot.
type BattleState struct {
	ID        BattleID
	ZoneID    string
	Shard     uint32
	SimTime   timesync.DualTime
	Units     map[uint32]*UnitState
	stateHash uint64
}

// Store keeps battles keyed by id per zone shard.
type Store struct {
	mu      sync.RWMutex
	battles map[BattleID]*BattleState
	clock   *timesync.Clock
}

// NewStore creates an empty partition store.
func NewStore(clock *timesync.Clock) *Store {
	if clock == nil {
		clock = timesync.NewClock(nil)
	}
	return &Store{battles: make(map[BattleID]*BattleState), clock: clock}
}

// Join creates or returns a battle for zone shard (in-memory only).
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
	b := &BattleState{
		ID:      id,
		ZoneID:  zoneID,
		Shard:   shard,
		SimTime: s.clock.Now(),
		Units:   map[uint32]*UnitState{1: {ID: 1, X: 0, Y: 0, HP: 100}},
	}
	b.rehash()
	s.battles[id] = b
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

// SubmitCommand applies a lockstep-style move stub and updates hash.
func (s *Store) SubmitCommand(id BattleID, playerID uint32, moveX, moveY int32) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.battles[id]
	if !ok {
		return 0, errors.New("roma: battle not found")
	}
	u, ok := b.Units[playerID]
	if !ok {
		u = &UnitState{ID: playerID, HP: 100}
		b.Units[playerID] = u
	}
	u.X += moveX
	u.Y += moveY
	s.clock.AdvanceSim(1)
	b.SimTime = s.clock.Now()
	b.rehash()
	return b.stateHash, nil
}

func (b *BattleState) rehash() {
	h := hash.New()
	for _, u := range b.Units {
		h.WriteUint64(uint64(u.ID))
		h.WriteInt64(int64(u.X))
		h.WriteInt64(int64(u.Y))
		h.WriteInt64(int64(u.HP))
	}
	b.stateHash = h.Sum64()
}

func (b *BattleState) StateHash() uint64 { return b.stateHash }

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
