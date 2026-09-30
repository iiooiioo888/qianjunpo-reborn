package roma

import (
	"sync"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadsample"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/observability/metrics"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timedilation"
)

// regionDilation runs the Phase 3 control loop for a Roma partition (scheduling only).
type regionDilation struct {
	mu    sync.Mutex
	loop  *timedilation.ControlLoop
	clock *timedilation.ManualClock
	rate  timedilation.Rate
}

func newRegionDilation() *regionDilation {
	clock := timedilation.NewManualClock(time.Unix(0, 0))
	rd := &regionDilation{
		clock: clock,
		loop:  timedilation.NewControlLoop(clock),
		rate:  timedilation.MaxRate,
	}
	return rd
}

func (rd *regionDilation) tick(sample loadsample.Sample) timedilation.Rate {
	rd.mu.Lock()
	defer rd.mu.Unlock()
	rec := rd.loop.Tick(sample)
	rd.rate = rec.Rate
	rd.clock.Advance(timedilation.DefaultControllerConfig().TickInterval)
	metrics.SetTimeFlowRate(int64(rd.rate))
	return rd.rate
}

func (rd *regionDilation) currentRate() timedilation.Rate {
	rd.mu.Lock()
	defer rd.mu.Unlock()
	return rd.rate
}

// TickRegionDilation drives the control loop from an explicit sample (tests / drill hooks).
func (s *Store) TickRegionDilation(sample loadsample.Sample) timedilation.Rate {
	if s.dilation == nil {
		return timedilation.MaxRate
	}
	return s.dilation.tick(sample)
}

func (s *Store) observeAndTickDilation(b *BattleState) {
	if s.dilation == nil || b == nil || b.Match == nil {
		return
	}
	queue := b.Match.PendingCommandCount()
	s.dilation.tick(loadsample.Sample{QueueLength: queue})
}

func (s *Store) timeFlowRateParts() uint32 {
	if s.dilation == nil {
		return uint32(timedilation.MaxRate)
	}
	return uint32(s.dilation.currentRate())
}
