package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/iiooiioo888/qianjunpo-reborn/internal/lares"
)

type laresLoginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type laresLoginJSON struct {
	AccessToken        string `json:"access_token"`
	RefreshToken       string `json:"refresh_token,omitempty"`
	AccessExpiresUnix  int64  `json:"access_expires_unix,omitempty"`
	RefreshExpiresUnix int64  `json:"refresh_expires_unix,omitempty"`
	PlayerID           uint64 `json:"player_id,omitempty"`
	AccountID          string `json:"account_id,omitempty"`
}

type laresLoginHTTPConfig struct {
	Enabled func() bool
	Issuer  *lares.TokenIssuer
}

// httpLaresLoginEnabled gates POST /v1/lares/login on Janus (dev / Compose Live preview).
func httpLaresLoginEnabled() bool {
	return envBool("JANUS_HTTP_DEV_MINT", false)
}

func registerLaresLoginHTTPRoute(mux *http.ServeMux, cfg laresLoginHTTPConfig) {
	mux.HandleFunc("/v1/lares/login", func(w http.ResponseWriter, r *http.Request) {
		handleLaresLoginHTTP(w, r, cfg)
	})
}

func handleLaresLoginHTTP(w http.ResponseWriter, r *http.Request, cfg laresLoginHTTPConfig) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if cfg.Enabled == nil || !cfg.Enabled() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if cfg.Issuer == nil {
		http.Error(w, "lares login unavailable", http.StatusServiceUnavailable)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxGatewayMirrorBodyBytes)
	var body laresLoginBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err != io.EOF {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	pair, err := cfg.Issuer.Login(body.Username, body.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	writeGatewayMirrorJSON(w, http.StatusOK, laresLoginJSON{
		AccessToken:        pair.AccessToken,
		RefreshToken:       pair.RefreshToken,
		AccessExpiresUnix:  pair.AccessExpiresUnix,
		RefreshExpiresUnix: pair.RefreshExpiresUnix,
		PlayerID:           pair.PlayerID,
		AccountID:          pair.AccountID,
	})
}
