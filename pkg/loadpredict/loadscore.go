package loadpredict

import "github.com/iiooiioo888/qianjunpo-reborn/pkg/loadsample"

// CompositeLoad maps a sample to 0..1 using CPU, queue watermark, and P99 latency.
func CompositeLoad(s loadsample.Sample) float64 {
	cpu := s.CPUPercent / 100.0
	if cpu < 0 {
		cpu = 0
	}
	if cpu > 1 {
		cpu = 1
	}
	queue := float64(s.QueueLength) / 1000.0
	if queue > 1 {
		queue = 1
	}
	lat := s.P99LatencyMs / 500.0
	if lat > 1 {
		lat = 1
	}
	if lat < 0 {
		lat = 0
	}
	// Weight queue highest — aligns with timedilation watermarks.
	return 0.25*cpu + 0.55*queue + 0.20*lat
}
