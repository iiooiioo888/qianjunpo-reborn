package janus

import (
	"context"
	"sync"
	"time"

	laresv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/lares/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/lares"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// LaresAuth validates access tokens through Lares gRPC Validate or a shared TokenIssuer.
type LaresAuth struct {
	mu          sync.Mutex
	grpcClient  laresv1.LaresAuthClient
	issuer      *lares.TokenIssuer
	dial        func(target string) (laresv1.LaresAuthClient, error)
	target      string
}

// NewLaresAuth dials Lares at grpcTarget. When secret is non-empty, local ValidateAccess
// is used as a fast path matching the Lares service signing key.
func NewLaresAuth(grpcTarget, secret string) (*LaresAuth, error) {
	a := &LaresAuth{
		target: grpcTarget,
		dial: func(target string) (laresv1.LaresAuthClient, error) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			conn, err := grpc.DialContext(ctx, target,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithBlock(),
			)
			if err != nil {
				return nil, err
			}
			return laresv1.NewLaresAuthClient(conn), nil
		},
	}
	if secret != "" {
		a.issuer = &lares.TokenIssuer{Secret: []byte(secret)}
	}
	return a, nil
}

// SetDialer overrides gRPC client creation (tests).
func (a *LaresAuth) SetDialer(fn func(target string) (laresv1.LaresAuthClient, error)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.dial = fn
}

func (a *LaresAuth) laresClient() (laresv1.LaresAuthClient, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.grpcClient != nil {
		return a.grpcClient, nil
	}
	if a.dial == nil {
		return nil, context.Canceled
	}
	cli, err := a.dial(a.target)
	if err != nil {
		return nil, err
	}
	a.grpcClient = cli
	return a.grpcClient, nil
}

// ValidateAccess implements AuthHook.
func (a *LaresAuth) ValidateAccess(ctx context.Context, token string) (uint64, bool) {
	if token == "" {
		return 0, false
	}
	if a.issuer != nil {
		pid, _, err := a.issuer.ValidateAccess(token)
		if err == nil {
			return pid, true
		}
	}
	cli, err := a.laresClient()
	if err != nil {
		return 0, false
	}
	resp, err := cli.Validate(ctx, &laresv1.ValidateRequest{AccessToken: token})
	if err != nil || resp == nil || !resp.GetValid() || resp.GetPlayer() == nil {
		return 0, false
	}
	return resp.GetPlayer().GetPlayerId(), true
}
