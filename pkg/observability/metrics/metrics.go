package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "qjp"

var (
	svcRegistry  = prometheus.NewRegistry()
	registerOnce sync.Once

	OnlinePlayers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "online_players",
		Help:      "Current online players (stub gauge for Phase 5 observability).",
	})
	BattleLatencyP99 = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "battle_latency_p99_ms",
		Help:      "Battle path P99 latency in milliseconds (stub).",
	})
	QueueLen = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "queue_len",
		Help:      "Command queue depth driving time dilation.",
	})
	TimeFlowRate = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "time_flow_rate",
		Help:      "Region time_flow_rate as float (10000=1.0x in sim).",
	})
	ActiveRooms = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "active_rooms",
		Help:      "Active in-memory battles/rooms on this Roma partition.",
	})
)

// Register installs collectors on reg (defaults to prometheus default registry).
func Register(reg prometheus.Registerer) {
	registerOnce.Do(func() {
		if reg == nil {
			reg = svcRegistry
		}
		reg.MustRegister(
			OnlinePlayers,
			BattleLatencyP99,
			QueueLen,
			TimeFlowRate,
			ActiveRooms,
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
		)
	})
}

// Handler exposes Prometheus text format on /metrics when mounted.
func Handler() http.Handler {
	Register(nil)
	return promhttp.HandlerFor(svcRegistry, promhttp.HandlerOpts{})
}

// SetTimeFlowRate stores rate using timedilation convention (10000 = 1.0x).
func SetTimeFlowRate(rate int64) {
	TimeFlowRate.Set(float64(rate) / 10000.0)
}

// SetQueueLen publishes command queue depth for timedilation alerts.
func SetQueueLen(depth int) {
	QueueLen.Set(float64(depth))
}

// SetActiveRooms publishes the number of active battles/rooms.
func SetActiveRooms(n int) {
	ActiveRooms.Set(float64(n))
}
