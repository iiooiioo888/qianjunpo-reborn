package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "qjp"

var (
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
)

// Register installs collectors on reg (defaults to prometheus default registry).
func Register(reg prometheus.Registerer) {
	registerOnce.Do(func() {
		if reg == nil {
			reg = prometheus.DefaultRegisterer
		}
		reg.MustRegister(OnlinePlayers, BattleLatencyP99, QueueLen, TimeFlowRate)
	})
}

// Handler exposes Prometheus text format on /metrics when mounted.
func Handler() http.Handler {
	Register(nil)
	return promhttp.Handler()
}

// SetTimeFlowRate stores rate using timedilation convention (10000 = 1.0x).
func SetTimeFlowRate(rate int64) {
	TimeFlowRate.Set(float64(rate) / 10000.0)
}
