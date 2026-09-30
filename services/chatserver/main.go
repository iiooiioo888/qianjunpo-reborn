package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	chatv1 "github.com/iiooiioo888/qianjunpo-reborn/gen/go/chat/v1"
	"google.golang.org/grpc"
)

type chatServer struct {
	chatv1.UnimplementedChatServiceServer
	seq uint64
}

func (s *chatServer) Send(_ context.Context, req *chatv1.SendRequest) (*chatv1.SendResponse, error) {
	if req.GetSender().GetPlayerId() == 0 {
		return nil, fmt.Errorf("chat: missing sender")
	}
	s.seq++
	return &chatv1.SendResponse{
		MessageId:  formatID(s.seq),
		WallUnixMs: time.Now().UnixMilli(),
	}, nil
}

func main() {
	grpcAddr := env("CHAT_GRPC_ADDR", ":9094")
	httpAddr := env("CHAT_HTTP_ADDR", ":8094")

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok","service":"chat"}`))
		})
		s := &http.Server{Addr: httpAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
		log.Fatal(s.ListenAndServe())
	}()

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	chatv1.RegisterChatServiceServer(srv, &chatServer{})
	log.Printf("chat grpc on %s", grpcAddr)
	log.Fatal(srv.Serve(lis))
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func formatID(v uint64) string {
	return "msg-" + itoa(v)
}

func itoa(v uint64) string {
	if v == 0 {
		return "0"
	}
	var buf [32]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
