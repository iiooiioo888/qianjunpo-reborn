package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadpredict"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/loadsample"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/observability/metrics"
)

func main() {
	addr := env("LOADPREDICT_HTTP_ADDR", ":8095")
	metrics.Register(nil)
	metrics.OnlinePlayers.Set(0)

	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]string{"status": "ok", "service": "loadpredict"})
	})
	mux.HandleFunc("/v1/forecast", handleForecast)
	mux.HandleFunc("/v1/prescale", handlePrescale)

	log.Printf("loadpredict listening on %s (stub LSTM; no GPU)", addr)
	s := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(s.ListenAndServe())
}

func handleForecast(w http.ResponseWriter, r *http.Request) {
	sample := loadsample.Sample{QueueLength: 900, CPUPercent: 88, P99LatencyMs: 420}
	series := make([]loadsample.Sample, 400)
	for i := range series {
		series[i] = loadsample.Sample{QueueLength: i * 3, CPUPercent: float64(i) / 5, P99LatencyMs: float64(i * 2)}
	}
	pred := loadpredict.NewSeriesPredictor(series)
	load := pred.ForecastLoad(sample)
	writeJSON(w, map[string]any{
		"horizon_sec":     loadpredict.ForecastHorizon.Seconds(),
		"forecast_load":   load,
		"high_load_hook":  loadpredict.HookHighLoad(load),
		"predicted_queue": func() int { q, _ := pred.Predict(sample); return q }(),
	})
}

func handlePrescale(w http.ResponseWriter, _ *http.Request) {
	planner := loadpredict.DefaultPreScalePlanner()
	now := time.Now().UTC()
	sig, ok := planner.Plan(now, 0.91, 880)
	if !ok {
		writeJSON(w, map[string]string{"signal": "none"})
		return
	}
	writeJSON(w, sig)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
