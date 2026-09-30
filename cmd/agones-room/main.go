// Command agones-room exercises Roma pre-prod Agones room lifecycle (mock default).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/agones"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "allocate":
		runAllocate(os.Args[2:])
	case "ready":
		runReady()
	case "shutdown":
		runShutdown()
	case "status":
		runStatus()
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: agones-room allocate [-zone ID] [-shard N]\n")
	fmt.Fprintf(os.Stderr, "       agones-room ready\n")
	fmt.Fprintf(os.Stderr, "       agones-room shutdown\n")
	fmt.Fprintf(os.Stderr, "       agones-room status\n")
	fmt.Fprintf(os.Stderr, "env: ROMA_AGONES_BACKEND=mock|sidecar, ROMA_AGONES_ALLOCATOR=mock|http\n")
}

func runAllocate(args []string) {
	fs := flag.NewFlagSet("allocate", flag.ExitOnError)
	zone := fs.String("zone", "default", "zone id")
	shard := fs.Uint("shard", 0, "shard")
	_ = fs.Parse(args)

	coord, err := agones.NewFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ident, err := coord.AllocateRoom(ctx, agones.AllocateRequest{ZoneID: *zone, Shard: uint32(*shard)})
	if err != nil {
		fmt.Fprintf(os.Stderr, "allocate: %v\n", err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(ident)
}

func runReady() {
	coord, err := agones.NewFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	if err := coord.BootstrapGameServer(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "ready: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("ok: ready")
}

func runShutdown() {
	coord, err := agones.NewFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	if err := coord.ShutdownGameServer(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "shutdown: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("ok: shutdown")
}

func runStatus() {
	coord, err := agones.NewFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	st := coord.Status(context.Background())
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(st)
}
