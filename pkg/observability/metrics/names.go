package metrics

// Stable Prometheus series names (namespace qjp + standard process metrics).
const (
	MetricOnlinePlayers       = "qjp_online_players"
	MetricBattleLatencyP99Ms  = "qjp_battle_latency_p99_ms"
	MetricQueueLen            = "qjp_queue_len"
	MetricTimeFlowRate        = "qjp_time_flow_rate"
	MetricActiveRooms         = "qjp_active_rooms"
	MetricProcessCPUSeconds   = "process_cpu_seconds_total"
)

// CoreSeries returns qjp_* gauges required for production alert/dashboard checks.
func CoreSeries() []string {
	return []string{
		MetricOnlinePlayers,
		MetricBattleLatencyP99Ms,
		MetricQueueLen,
		MetricTimeFlowRate,
		MetricActiveRooms,
	}
}
