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

func TestHTTPLaresLoginIssuesAccessToken(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	issuer := &lares.TokenIssuer{Secret: []byte("phase4-dev"), Now: func() time.Time { return now }}
	mux := http.NewServeMux()
	registerLaresLoginHTTPRoute(mux, laresLoginHTTPConfig{
		Enabled: func() bool { return true },
		Issuer:  issuer,
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	raw := []byte(`{"username":"smoke","password":"smoke"}`)
	res, err := http.Post(srv.URL+"/v1/lares/login", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("status %d: %s", res.StatusCode, b)
	}
	var out laresLoginJSON
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.AccessToken == "" {
		t.Fatalf("expected access_token, got %+v", out)
	}
	pid, acct, err := issuer.ValidateAccess(out.AccessToken)
	if err != nil || acct != "smoke" || pid != out.PlayerID {
		t.Fatalf("token not valid: pid=%d acct=%s err=%v", pid, acct, err)
	}
}

func TestHTTPLaresLoginDisabledNotFound(t *testing.T) {
	issuer := &lares.TokenIssuer{Secret: []byte("k")}
	mux := http.NewServeMux()
	registerLaresLoginHTTPRoute(mux, laresLoginHTTPConfig{
		Enabled: func() bool { return false },
		Issuer:  issuer,
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	res, err := http.Post(srv.URL+"/v1/lares/login", "application/json",
		bytes.NewReader([]byte(`{"username":"smoke","password":"smoke"}`)))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
}
