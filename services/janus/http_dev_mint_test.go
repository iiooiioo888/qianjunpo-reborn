package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/internal/lares"
)

func TestHTTPDevMintIssuesLaresToken(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	issuer := &lares.TokenIssuer{Secret: []byte("phase4-dev"), Now: func() time.Time { return now }}
	mux := http.NewServeMux()
	registerDevMintHTTPRoute(mux, devMintConfig{
		Enabled: func() bool { return true },
		Issuer:  issuer,
		Account: func() string { return lares.DevStableAccount },
	})
	srv := newTacticalHTTPServerFromMux(t, mux)

	res, err := http.Post(srv.URL+"/v1/auth/dev-mint", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d: %s", res.StatusCode, b)
	}
	var out devMintJSON
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.AccessToken == "" || out.AccessExpiresUnix == 0 {
		t.Fatalf("expected access token payload, got %+v", out)
	}
	if out.AccountID != lares.DevStableAccount {
		t.Fatalf("account_id=%q", out.AccountID)
	}
	pid, acct, err := issuer.ValidateAccess(out.AccessToken)
	if err != nil || acct != lares.DevStableAccount || pid != out.PlayerID {
		t.Fatalf("token not valid via issuer: pid=%d acct=%s err=%v", pid, acct, err)
	}
}

func TestHTTPDevMintDisabledNotFound(t *testing.T) {
	issuer := &lares.TokenIssuer{Secret: []byte("k")}
	mux := http.NewServeMux()
	registerDevMintHTTPRoute(mux, devMintConfig{
		Enabled: func() bool { return false },
		Issuer:  issuer,
	})
	srv := newTacticalHTTPServerFromMux(t, mux)

	res, err := http.Post(srv.URL+"/v1/auth/dev-mint", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
}

func newTacticalHTTPServerFromMux(t *testing.T, mux *http.ServeMux) *httptest.Server {
	// reuse httptest helper pattern from http_tactical_test.go
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}
