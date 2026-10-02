package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/iiooiioo888/qianjunpo-reborn/internal/lares"
)

type devMintBody struct {
	Account string `json:"account"`
}

type devMintJSON struct {
	AccessToken       string `json:"access_token"`
	AccessExpiresUnix int64  `json:"access_expires_unix"`
	PlayerID          uint64 `json:"player_id"`
	AccountID         string `json:"account_id"`
}

type devMintConfig struct {
	Enabled func() bool
	Issuer  *lares.TokenIssuer
	Account func() string
}

func defaultDevMintAccount() string {
	if v := strings.TrimSpace(os.Getenv("JANUS_DEV_MINT_ACCOUNT")); v != "" {
		return v
	}
	return lares.DevStableAccount
}

func httpDevMintEnabled() bool {
	return envBool("JANUS_HTTP_DEV_MINT", false)
}

func registerDevMintHTTPRoute(mux *http.ServeMux, cfg devMintConfig) {
	mux.HandleFunc("/v1/auth/dev-mint", func(w http.ResponseWriter, r *http.Request) {
		handleDevMint(w, r, cfg)
	})
}

func handleDevMint(w http.ResponseWriter, r *http.Request, cfg devMintConfig) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if cfg.Enabled == nil || !cfg.Enabled() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if cfg.Issuer == nil {
		http.Error(w, "dev mint unavailable", http.StatusServiceUnavailable)
		return
	}
	account := defaultDevMintAccount()
	if cfg.Account != nil {
		account = cfg.Account()
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxGatewayMirrorBodyBytes)
	var body devMintBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err != io.EOF {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Account) != "" {
		account = strings.TrimSpace(body.Account)
	}
	pair, err := cfg.Issuer.MintDevStableAccess(account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeGatewayMirrorJSON(w, http.StatusOK, devMintJSON{
		AccessToken:       pair.AccessToken,
		AccessExpiresUnix: pair.AccessExpiresUnix,
		PlayerID:          pair.PlayerID,
		AccountID:         pair.AccountID,
	})
}
