// Package cmdqueue implements hot / warm / cold command tiers with P0–P2 FIFO lanes.
//
// Defaults align with Phase 2 brief: hot 1000, warm 5000, cold TTL 5s.
// See docs/cmdqueue.md for overflow and stale-packet policy.
package cmdqueue

import "time"

// Priority levels (lower number = higher priority).
type Priority uint8

const (
	PriorityP0 Priority = 0
	PriorityP1 Priority = 1
	PriorityP2 Priority = 2
)

// Temperature selects hot / warm / cold queue bucket.
type Temperature uint8

const (
	TempHot  Temperature = 0
	TempWarm Temperature = 1
	TempCold Temperature = 2
)

// Production tier budgets (total items per temperature, all priorities).
const (
	DefaultHotCap  = 1000
	DefaultWarmCap = 5000
)

// DefaultColdTTL is the cold-tier retention window.
const DefaultColdTTL = 5 * time.Second

// Command is an opaque queued item.
type Command struct {
	ID       uint64
	Priority Priority
	Payload  []byte
}

// Config defines per-temperature capacity and cold retention.
type Config struct {
	HotCap  int
	WarmCap int
	// ColdTTL drops cold-tier entries after this duration (wall clock via Queue clock).
	ColdTTL time.Duration
}

// DefaultConfig matches Phase 2 production defaults (hot / warm / cold TTL).
func DefaultConfig() Config {
	return Config{HotCap: DefaultHotCap, WarmCap: DefaultWarmCap, ColdTTL: DefaultColdTTL}
}

type fifoQueue struct {
	items []Command
}

func (q *fifoQueue) enqueue(c Command) {
	q.items = append(q.items, c)
}

func (q *fifoQueue) dequeue() (Command, bool) {
	if len(q.items) == 0 {
		return Command{}, false
	}
	c := q.items[0]
	q.items = q.items[1:]
	return c, true
}

func (q *fifoQueue) len() int { return len(q.items) }

type coldEntry struct {
	cmd       Command
	expiresAt time.Time
}

type coldLane struct {
	entries []coldEntry
	ttl     time.Duration
}

func (l *coldLane) enqueue(c Command, now time.Time, ttl time.Duration) {
	l.ttl = ttl
	l.entries = append(l.entries, coldEntry{cmd: c, expiresAt: now.Add(ttl)})
}

func (l *coldLane) purgeExpired(now time.Time, onP0 func(Command)) {
	i := 0
	for _, e := range l.entries {
		if !now.Before(e.expiresAt) {
			if e.cmd.Priority == PriorityP0 && onP0 != nil {
				onP0(e.cmd)
			}
			continue
		}
		l.entries[i] = e
		i++
	}
	l.entries = l.entries[:i]
}

func (l *coldLane) len(now time.Time, onP0 func(Command)) int {
	l.purgeExpired(now, onP0)
	return len(l.entries)
}

func (l *coldLane) dequeue(now time.Time, onP0 func(Command)) (Command, bool) {
	l.purgeExpired(now, onP0)
	if len(l.entries) == 0 {
		return Command{}, false
	}
	c := l.entries[0].cmd
	l.entries = l.entries[1:]
	return c, true
}

// EnqueueOutcome reports spill, drop, or storage tier.
type EnqueueOutcome struct {
	OK       bool
	Dropped  bool
	StoredAt Temperature
}

// Queue multiplexes hot/warm/cold with P0/P1/P2 ordering inside each temperature.
type Queue struct {
	cfg  Config
	hot  [3]fifoQueue
	warm [3]fifoQueue
	cold [3]coldLane
	seq  uint64
	now  func() time.Time
	// p0StaleHook runs when a P0 cold-tier entry expires before dequeue (rollback path signal).
	p0StaleHook func(Command)
}

// New creates queues with per-priority FIFO lanes.
func New(cfg Config) *Queue {
	if cfg.ColdTTL <= 0 {
		cfg.ColdTTL = DefaultColdTTL
	}
	return &Queue{cfg: cfg, now: time.Now}
}

// SetClock replaces wall-clock reads (tests).
func (q *Queue) SetClock(now func() time.Time) {
	if now == nil {
		q.now = time.Now
		return
	}
	q.now = now
}

// SetP0StaleHook registers a callback when cold-tier P0 entries expire (rollback signal).
func (q *Queue) SetP0StaleHook(h func(Command)) {
	q.p0StaleHook = h
}

func (q *Queue) coldHook() func(Command) { return q.p0StaleHook }

func (q *Queue) laneHotWarm(temp Temperature, pri Priority) *fifoQueue {
	switch temp {
	case TempHot:
		return &q.hot[pri]
	case TempWarm:
		return &q.warm[pri]
	default:
		return nil
	}
}

func (q *Queue) tierLen(temp Temperature) int {
	now := q.now()
	n := 0
	for pri := PriorityP0; pri <= PriorityP2; pri++ {
		switch temp {
		case TempCold:
			n += q.cold[pri].len(now, q.coldHook())
		default:
			n += q.laneHotWarm(temp, pri).len()
		}
	}
	return n
}

func (q *Queue) tierCap(temp Temperature) int {
	switch temp {
	case TempHot:
		return q.cfg.HotCap
	case TempWarm:
		return q.cfg.WarmCap
	default:
		return 0
	}
}

func (q *Queue) canAcceptAt(temp Temperature) bool {
	if temp == TempCold {
		return true
	}
	cap := q.tierCap(temp)
	if cap <= 0 {
		return false
	}
	return q.tierLen(temp) < cap
}

func (q *Queue) pushAt(temp Temperature, cmd Command) {
	now := q.now()
	if temp == TempCold {
		q.cold[cmd.Priority].enqueue(cmd, now, q.cfg.ColdTTL)
		return
	}
	q.laneHotWarm(temp, cmd.Priority).enqueue(cmd)
}

// Enqueue adds a command to the given temperature lane (no spill).
func (q *Queue) Enqueue(temp Temperature, cmd Command) bool {
	if cmd.Priority > PriorityP2 {
		cmd.Priority = PriorityP2
	}
	if temp == TempCold {
		q.pushAt(TempCold, cmd)
		return true
	}
	if !q.canAcceptAt(temp) {
		return false
	}
	q.pushAt(temp, cmd)
	return true
}

// TryEnqueue stores cmd by spilling hot → warm → cold on tier pressure.
// When hot and warm are both at cap, P2 is droppable (not spilled to cold).
// P0/P1 still use cold; P0 never returns Dropped (cold accepts without count cap).
func (q *Queue) TryEnqueue(cmd Command) EnqueueOutcome {
	if cmd.Priority > PriorityP2 {
		cmd.Priority = PriorityP2
	}
	if cmd.Priority == PriorityP2 && !q.canAcceptAt(TempHot) && !q.canAcceptAt(TempWarm) {
		return EnqueueOutcome{Dropped: true}
	}
	for _, temp := range []Temperature{TempHot, TempWarm, TempCold} {
		if temp != TempCold && !q.canAcceptAt(temp) {
			continue
		}
		q.pushAt(temp, cmd)
		return EnqueueOutcome{OK: true, StoredAt: temp}
	}
	if cmd.Priority == PriorityP0 {
		q.pushAt(TempCold, cmd)
		return EnqueueOutcome{OK: true, StoredAt: TempCold}
	}
	return EnqueueOutcome{Dropped: true}
}

// Dequeue pops the highest-priority oldest command across temperatures (hot before warm before cold).
func (q *Queue) Dequeue() (Command, bool) {
	now := q.now()
	for _, temp := range []Temperature{TempHot, TempWarm, TempCold} {
		for pri := PriorityP0; pri <= PriorityP2; pri++ {
			if temp == TempCold {
				if c, ok := q.cold[pri].dequeue(now, q.coldHook()); ok {
					return c, true
				}
				continue
			}
			lane := q.laneHotWarm(temp, pri)
			if lane.len() > 0 {
				return lane.dequeue()
			}
		}
	}
	return Command{}, false
}

// Len returns total queued commands (after cold TTL purge).
func (q *Queue) Len() int {
	return q.tierLen(TempHot) + q.tierLen(TempWarm) + q.tierLen(TempCold)
}

// HotLen is the hot-tier depth (timedilation high watermark uses 1000).
func (q *Queue) HotLen() int { return q.tierLen(TempHot) }

// WarmLen returns warm-tier depth.
func (q *Queue) WarmLen() int { return q.tierLen(TempWarm) }

// ColdLen returns cold-tier depth after TTL purge.
func (q *Queue) ColdLen() int { return q.tierLen(TempCold) }

// PurgeColdExpired removes stale cold entries (also runs on dequeue/len).
func (q *Queue) PurgeColdExpired() {
	now := q.now()
	h := q.coldHook()
	for pri := PriorityP0; pri <= PriorityP2; pri++ {
		q.cold[pri].purgeExpired(now, h)
	}
}

// NextID assigns monotonic command ids.
func (q *Queue) NextID() uint64 {
	q.seq++
	return q.seq
}
