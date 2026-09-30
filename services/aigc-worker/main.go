package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	addr := env("AIGC_HTTP_ADDR", ":8096")
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]string{"status": "ok", "service": "aigc-worker", "gpu": "disabled"})
	})
	mux.HandleFunc("/v1/comfy/workflow", handleComfyPlaceholder)
	mux.HandleFunc("/v1/ipadapter/style", handleIPAdapterPlaceholder)

	log.Printf("aigc-worker listening on %s (ComfyUI/IP-Adapter placeholders)", addr)
	s := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(s.ListenAndServe())
}

func handleComfyPlaceholder(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"status":  "stub",
		"engine":  "comfyui",
		"message": "Submit workflow JSON in prod; CI does not download checkpoints",
		"asset_import": "see docs/aigc/PIPELINE.md",
	})
}

func handleIPAdapterPlaceholder(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"status": "stub",
		"engine": "ip-adapter",
		"note":   "Reference image id only; models live outside repo",
	})
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
