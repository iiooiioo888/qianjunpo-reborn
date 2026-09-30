package cmdqueue

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

// Command is an opaque queued item.
type Command struct {
	ID       uint64
	Priority Priority
	Payload  []byte
}

// Config defines per-temperature capacity.
type Config struct {
	HotCap  int
	WarmCap int
	ColdCap int
}

// DefaultConfig matches brief-style dev defaults.
func DefaultConfig() Config {
	return Config{HotCap: 64, WarmCap: 256, ColdCap: 1024}
}

type fifoQueue struct {
	items []Command
	cap   int
}

func (q *fifoQueue) enqueue(c Command) bool {
	if len(q.items) >= q.cap {
		return false
	}
	q.items = append(q.items, c)
	return true
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

// Queue multiplexes hot/warm/cold with P0/P1/P2 ordering inside each temperature.
type Queue struct {
	cfg  Config
	hot  [3]fifoQueue
	warm [3]fifoQueue
	cold [3]fifoQueue
	seq  uint64
}

// New creates queues with per-priority FIFO lanes.
func New(cfg Config) *Queue {
	q := &Queue{cfg: cfg}
	for i := 0; i < 3; i++ {
		q.hot[i] = fifoQueue{cap: cfg.HotCap}
		q.warm[i] = fifoQueue{cap: cfg.WarmCap}
		q.cold[i] = fifoQueue{cap: cfg.ColdCap}
	}
	return q
}

func (q *Queue) lane(temp Temperature, pri Priority) *fifoQueue {
	switch temp {
	case TempHot:
		return &q.hot[pri]
	case TempWarm:
		return &q.warm[pri]
	default:
		return &q.cold[pri]
	}
}

// Enqueue adds a command to the given temperature lane.
func (q *Queue) Enqueue(temp Temperature, cmd Command) bool {
	if cmd.Priority > PriorityP2 {
		cmd.Priority = PriorityP2
	}
	lane := q.lane(temp, cmd.Priority)
	if !lane.enqueue(cmd) {
		return false
	}
	return true
}

// Dequeue pops the highest-priority oldest command across temperatures (hot before warm before cold).
func (q *Queue) Dequeue() (Command, bool) {
	for _, temp := range []Temperature{TempHot, TempWarm, TempCold} {
		for pri := PriorityP0; pri <= PriorityP2; pri++ {
			lane := q.lane(temp, pri)
			if lane.len() > 0 {
				return lane.dequeue()
			}
		}
	}
	return Command{}, false
}

// Len returns total queued commands.
func (q *Queue) Len() int {
	n := 0
	for i := 0; i < 3; i++ {
		n += q.hot[i].len() + q.warm[i].len() + q.cold[i].len()
	}
	return n
}

// NextID assigns monotonic command ids.
func (q *Queue) NextID() uint64 {
	q.seq++
	return q.seq
}
