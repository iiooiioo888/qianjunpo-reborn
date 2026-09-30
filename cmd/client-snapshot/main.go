// Command client-snapshot writes tactical view JSON for the Cocos display layer.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

func main() {
	out := flag.String("out", "client/assets/resources/data/tactical/demo_initial.json", "output path")
	seed := flag.Uint64("seed", 0xcafe, "match seed (same as make match-play)")
	flag.Parse()

	v := tactical.InitialViewSnapshot(*seed)
	data, err := tactical.MarshalViewSnapshotJSON(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (units=%d frame=%d)\n", *out, len(v.Units), v.LockstepFrame)
}
