// Command match runs or verifies a Phase 2 tactical duel replay.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/lockstep"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/replay"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "play":
		runPlay(os.Args[2:])
	case "verify":
		runVerify(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: match play [-seed N] [-out file.rgz]\n")
	fmt.Fprintf(os.Stderr, "       match verify -in file.rgz\n")
}

func runPlay(args []string) {
	fs := flag.NewFlagSet("play", flag.ExitOnError)
	seed := fs.Uint64("seed", 0xcafe, "match RNG seed")
	out := fs.String("out", "", "write gzip replay to path (optional)")
	_ = fs.Parse(args)

	sched := tactical.DemoSchedule()
	m := tactical.NewMatch(*seed)
	target := tactical.DemoTargetFrame()
	final := m.RunSchedule(sched, target)
	rec := m.Recording()

	fmt.Printf("tactical match seed=%d frames=%d delay=%d\n", *seed, target, lockstep.CommandDelayFrames)
	fmt.Printf("initial_hash=%016x\n", rec.InitialHash)
	fmt.Printf("chain_hash=%016x\n", replay.ComputeChainHash(rec))
	fmt.Printf("final_hash=%016x\n", final)
	if m.Winner != 255 {
		fmt.Printf("winner=player%d\n", m.Winner)
	} else {
		fmt.Println("winner=draw")
	}

	if err := replay.VerifyTerminal(rec, final); err != nil {
		fmt.Fprintf(os.Stderr, "replay verify: %v\n", err)
		os.Exit(1)
	}

	if *out != "" {
		gz, err := replay.MarshalGzip(rec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(*out, gz, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s\n", *out)
	}
}

func runVerify(args []string) {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	in := fs.String("in", "", "gzip replay path")
	_ = fs.Parse(args)
	if *in == "" {
		fmt.Fprintln(os.Stderr, "verify requires -in")
		os.Exit(2)
	}
	data, err := os.ReadFile(*in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read: %v\n", err)
		os.Exit(1)
	}
	rec, err := replay.UnmarshalGzip(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "decode: %v\n", err)
		os.Exit(1)
	}
	final, err := tactical.ReplayFromRecording(rec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "replay: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("OK final_hash=%016x chain_hash=%016x\n", final, replay.ComputeChainHash(rec))
}

func maxFrame(sched []tactical.ScheduledCommand) uint64 {
	var max uint64
	for _, s := range sched {
		if s.SubmitFrame > max {
			max = s.SubmitFrame
		}
	}
	return max
}
