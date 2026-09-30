package lares

import (
	"testing"
	"time"
)

func TestDualTokenHandshake(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	issuer := &TokenIssuer{Secret: []byte("phase4-dev"), Now: func() time.Time { return now }}

	pair, err := issuer.Login("hero", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected dual tokens")
	}
	pid, acct, err := issuer.ValidateAccess(pair.AccessToken)
	if err != nil || acct != "hero" || pid != pair.PlayerID {
		t.Fatalf("validate access: pid=%d acct=%s err=%v", pid, acct, err)
	}

	rotated, err := issuer.Refresh(pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	issuer.Now = func() time.Time { return now.Add(time.Second) }
	rotated2, err := issuer.Refresh(rotated.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if rotated2.AccessToken == pair.AccessToken {
		t.Fatal("expected new access token after refresh with advanced clock")
	}
	if _, _, err := issuer.ValidateAccess(rotated2.AccessToken); err != nil {
		t.Fatal(err)
	}
}

func TestAccessExpires(t *testing.T) {
	start := time.Unix(1_700_000_000, 0)
	issuer := &TokenIssuer{Secret: []byte("k"), Now: func() time.Time { return start }}
	pair, _ := issuer.Login("u", "p")
	issuer.Now = func() time.Time { return start.Add(accessTTL + time.Second) }
	if _, _, err := issuer.ValidateAccess(pair.AccessToken); err == nil {
		t.Fatal("expected expired access")
	}
}
