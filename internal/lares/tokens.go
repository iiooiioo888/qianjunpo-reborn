package lares

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 7 * 24 * time.Hour
)

// TokenPair is the dual-token handshake result.
type TokenPair struct {
	AccessToken        string
	RefreshToken       string
	AccessExpiresUnix  int64
	RefreshExpiresUnix int64
	PlayerID           uint64
	AccountID          string
}

// TokenIssuer signs stub access/refresh tokens (not production JWT).
type TokenIssuer struct {
	Secret []byte
	Now    func() time.Time
}

func (ti *TokenIssuer) now() time.Time {
	if ti.Now != nil {
		return ti.Now()
	}
	return time.Now()
}

// Login validates credentials and mints token pair.
func (ti *TokenIssuer) Login(username, password string) (TokenPair, error) {
	if username == "" || password == "" {
		return TokenPair{}, errors.New("lares: missing credentials")
	}
	if ti.Secret == nil {
		return TokenPair{}, errors.New("lares: issuer not configured")
	}
	now := ti.now()
	playerID := hashAccount(username)
	return ti.mint(playerID, username, now)
}

// Refresh rotates access token when refresh token valid.
func (ti *TokenIssuer) Refresh(refreshToken string) (TokenPair, error) {
	playerID, account, exp, err := ti.parse(refreshToken, "refresh")
	if err != nil {
		return TokenPair{}, err
	}
	if ti.now().Unix() > exp {
		return TokenPair{}, errors.New("lares: refresh expired")
	}
	return ti.mint(playerID, account, ti.now())
}

// ValidateAccess checks access token signature and expiry.
func (ti *TokenIssuer) ValidateAccess(accessToken string) (uint64, string, error) {
	playerID, account, exp, err := ti.parse(accessToken, "access")
	if err != nil {
		return 0, "", err
	}
	if ti.now().Unix() > exp {
		return 0, "", errors.New("lares: access expired")
	}
	return playerID, account, nil
}

func (ti *TokenIssuer) mint(playerID uint64, account string, now time.Time) (TokenPair, error) {
	accessExp := now.Add(accessTTL)
	refreshExp := now.Add(refreshTTL)
	access, err := ti.sign("access", playerID, account, accessExp.Unix())
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := ti.sign("refresh", playerID, account, refreshExp.Unix())
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:        access,
		RefreshToken:       refresh,
		AccessExpiresUnix:  accessExp.Unix(),
		RefreshExpiresUnix: refreshExp.Unix(),
		PlayerID:           playerID,
		AccountID:          account,
	}, nil
}

func (ti *TokenIssuer) sign(kind string, playerID uint64, account string, expUnix int64) (string, error) {
	payload := fmt.Sprintf("%s|%d|%s|%d", kind, playerID, account, expUnix)
	mac := hmac.New(sha256.New, ti.Secret)
	_, _ = mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + sig, nil
}

func (ti *TokenIssuer) parse(token, wantKind string) (uint64, string, int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, "", 0, errors.New("lares: malformed token")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, "", 0, errors.New("lares: bad token encoding")
	}
	mac := hmac.New(sha256.New, ti.Secret)
	_, _ = mac.Write(raw)
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return 0, "", 0, errors.New("lares: invalid signature")
	}
	fields := strings.Split(string(raw), "|")
	if len(fields) != 4 || fields[0] != wantKind {
		return 0, "", 0, errors.New("lares: wrong token kind")
	}
	pid, _ := strconv.ParseUint(fields[1], 10, 64)
	exp, _ := strconv.ParseInt(fields[3], 10, 64)
	return pid, fields[2], exp, nil
}

func hashAccount(username string) uint64 {
	h := sha256.Sum256([]byte(username))
	return uint64(h[0])<<56 | uint64(h[1])<<48 | uint64(h[2])<<40 | uint64(h[3])<<32 |
		uint64(h[4])<<24 | uint64(h[5])<<16 | uint64(h[6])<<8 | uint64(h[7])
}
