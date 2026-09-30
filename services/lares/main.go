package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	commonv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/common/v1"
	laresv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/lares/v1"
	"github.com/iiooiioo888/qianjunpo-reborn/internal/lares"
	"google.golang.org/grpc"
)

type laresServer struct {
	laresv1.UnimplementedLaresAuthServer
	issuer *lares.TokenIssuer
}

func (s *laresServer) Login(_ context.Context, req *laresv1.LoginRequest) (*laresv1.LoginResponse, error) {
	pair, err := s.issuer.Login(req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, err
	}
	return &laresv1.LoginResponse{
		AccessToken:        pair.AccessToken,
		RefreshToken:       pair.RefreshToken,
		AccessExpiresUnix:  pair.AccessExpiresUnix,
		RefreshExpiresUnix: pair.RefreshExpiresUnix,
		Player:             &commonv1.PlayerRef{PlayerId: pair.PlayerID, AccountId: pair.AccountID},
	}, nil
}

func (s *laresServer) Refresh(_ context.Context, req *laresv1.RefreshRequest) (*laresv1.RefreshResponse, error) {
	pair, err := s.issuer.Refresh(req.GetRefreshToken())
	if err != nil {
		return nil, err
	}
	return &laresv1.RefreshResponse{
		AccessToken:        pair.AccessToken,
		RefreshToken:       pair.RefreshToken,
		AccessExpiresUnix:  pair.AccessExpiresUnix,
		RefreshExpiresUnix: pair.RefreshExpiresUnix,
	}, nil
}

func (s *laresServer) Validate(_ context.Context, req *laresv1.ValidateRequest) (*laresv1.ValidateResponse, error) {
	pid, acct, err := s.issuer.ValidateAccess(req.GetAccessToken())
	if err != nil {
		return &laresv1.ValidateResponse{Valid: false}, nil
	}
	return &laresv1.ValidateResponse{
		Valid:  true,
		Player: &commonv1.PlayerRef{PlayerId: pid, AccountId: acct},
	}, nil
}

func main() {
	grpcAddr := env("LARES_GRPC_ADDR", ":9091")
	httpAddr := env("LARES_HTTP_ADDR", ":8091")
	secret := env("LARES_TOKEN_SECRET", "phase4-dev-secret")

	flag.Parse()
	issuer := &lares.TokenIssuer{Secret: []byte(secret)}

	go serveHTTP(httpAddr)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	laresv1.RegisterLaresAuthServer(srv, &laresServer{issuer: issuer})
	log.Printf("lares grpc on %s http health on %s", grpcAddr, httpAddr)
	log.Fatal(srv.Serve(lis))
}

func serveHTTP(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"lares"}`))
	})
	s := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(s.ListenAndServe())
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
