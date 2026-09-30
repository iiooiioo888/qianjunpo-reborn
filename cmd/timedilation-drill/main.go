// Command timedilation-drill runs the deterministic overload scenario (slowdown → recovery).
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/timedilation"
)

func main() {
	clock := timedilation.NewManualClock(time.Unix(0, 0))
	report := timedilation.RunOverloadDrill(clock, timedilation.DefaultOverloadSegments())
	fmt.Println("qianjunpo timedilation overload drill")
	fmt.Println(report.Summary())
	if err := report.Pass(timedilation.DefaultDrillEnvelope()); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		fmt.Fprint(os.Stderr, timedilation.FormatDrillLog(report, 20))
		os.Exit(1)
	}
	fmt.Println("PASS")
	fmt.Print(timedilation.FormatDrillLog(report, 25))
}
