package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	senatev1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/senate/v1"
	"google.golang.org/grpc"
)

type senateServer struct {
	senatev1.UnimplementedSenateOpsServer
	maintenance bool
}

func (s *senateServer) Ping(context.Context, *senatev1.PingRequest) (*senatev1.PingResponse, error) {
	st := "ok"
	if s.maintenance {
		st = "maintenance"
	}
	return &senatev1.PingResponse{Status: st}, nil
}

func (s *senateServer) SetMaintenance(_ context.Context, req *senatev1.SetMaintenanceRequest) (*senatev1.SetMaintenanceResponse, error) {
	s.maintenance = req.GetEnabled()
	return &senatev1.SetMaintenanceResponse{Applied: true}, nil
}

func main() {
	grpcAddr := env("SENATE_GRPC_ADDR", ":9093")
	httpAddr := env("SENATE_HTTP_ADDR", ":8093")
	srvImpl := &senateServer{}

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok","service":"senate"}`))
		})
		s := &http.Server{Addr: httpAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
		log.Fatal(s.ListenAndServe())
	}()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	senatev1.RegisterSenateOpsServer(srv, srvImpl)
	log.Printf("senate grpc on %s", grpcAddr)
	log.Fatal(srv.Serve(lis))
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
