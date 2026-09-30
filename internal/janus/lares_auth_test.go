package janus

import (
	"context"
	"net"
	"testing"
	"time"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	laresv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/lares/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/lares"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type stubLares struct {
	laresv1.UnimplementedLaresAuthServer
	issuer *lares.TokenIssuer
}

func (s *stubLares) Validate(_ context.Context, req *laresv1.ValidateRequest) (*laresv1.ValidateResponse, error) {
	pid, acct, err := s.issuer.ValidateAccess(req.GetAccessToken())
	if err != nil {
		return &laresv1.ValidateResponse{Valid: false}, nil
	}
	return &laresv1.ValidateResponse{
		Valid:  true,
		Player: &commonv1.PlayerRef{PlayerId: pid, AccountId: acct},
	}, nil
}

func TestLaresAuthValidateGRPC(t *testing.T) {
	const secret = "phase4-dev"
	issuer := &lares.TokenIssuer{Secret: []byte(secret), Now: func() time.Time { return time.Unix(1_700_000_000, 0) }}
	pair, err := issuer.Login("hero", "pw")
	if err != nil {
		t.Fatal(err)
	}

	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	laresv1.RegisterLaresAuthServer(srv, &stubLares{issuer: issuer})
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	dialer := func(string) (laresv1.LaresAuthClient, error) {
		conn, err := grpc.DialContext(context.Background(), "bufnet",
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, err
		}
		return laresv1.NewLaresAuthClient(conn), nil
	}

	auth, err := NewLaresAuth("bufnet", "")
	if err != nil {
		t.Fatal(err)
	}
	auth.SetDialer(dialer)

	pid, ok := auth.ValidateAccess(context.Background(), pair.AccessToken)
	if !ok || pid != pair.PlayerID {
		t.Fatalf("validate ok=%v pid=%d", ok, pid)
	}
	_, ok = auth.ValidateAccess(context.Background(), "bad-token")
	if ok {
		t.Fatal("expected invalid token")
	}
}

func TestLaresAuthLocalSecret(t *testing.T) {
	now := time.Now()
	issuer := &lares.TokenIssuer{Secret: []byte("k"), Now: func() time.Time { return now }}
	pair, _ := issuer.Login("u", "p")
	auth, err := NewLaresAuth("unreachable:1", "k")
	if err != nil {
		t.Fatal(err)
	}
	pid, ok := auth.ValidateAccess(context.Background(), pair.AccessToken)
	if !ok || pid != pair.PlayerID {
		t.Fatal("local issuer path failed")
	}
}
