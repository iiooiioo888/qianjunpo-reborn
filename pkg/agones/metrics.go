package agones

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricsOnce sync.Once

	lifecycleOps = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "qjp",
		Name:      "agones_lifecycle_operations_total",
		Help:      "Agones room lifecycle operations (allocate, ready, shutdown, health).",
	}, []string{"operation", "result"})
)

func registerMetrics(reg prometheus.Registerer) {
	metricsOnce.Do(func() {
		if reg == nil {
			reg = prometheus.DefaultRegisterer
		}
		reg.MustRegister(lifecycleOps)
	})
}

func observeOp(operation, result string) {
	registerMetrics(nil)
	lifecycleOps.WithLabelValues(operation, result).Inc()
}
