package balance

import "fmt"

// WinRateReport aggregates species win rates from batch simulation.
type WinRateReport struct {
	Matches   int
	WinCounts map[Species]int
	WinRates  map[Species]float64
}

func emptyCounts() map[Species]int {
	return map[Species]int{
		0: 0, 1: 0, 2: 0,
	}
}

// RunBatch simulates n matches with rotating seeds (CI-friendly small n).
func RunBatch(n int, mk func(i int) MatchConfig) WinRateReport {
	if n <= 0 {
		n = 1
	}
	counts := emptyCounts()
	for i := 0; i < n; i++ {
		res := SimulateMatch(mk(i))
		counts[res.Winner]++
	}
	rates := map[Species]float64{}
	for sp, c := range counts {
		rates[sp] = float64(c) / float64(n)
	}
	return WinRateReport{Matches: n, WinCounts: counts, WinRates: rates}
}

// DefaultBatch runs infantry vs archer alternation for quick CI stats.
func DefaultBatch(n int) WinRateReport {
	return RunBatch(n, func(i int) MatchConfig {
		a := StubAgent{Label: "A", Type: 0} // infantry
		b := StubAgent{Label: "B", Type: 1} // archer
		if i%2 == 1 {
			a, b = b, a
		}
		return MatchConfig{Seed: uint64(1000 + i), AgentA: a, AgentB: b}
	})
}

// MaxWinRateSwing returns the largest absolute delta per species between two reports.
func MaxWinRateSwing(baseline, current WinRateReport) float64 {
	var max float64
	for sp, base := range baseline.WinRates {
		cur := current.WinRates[sp]
		d := cur - base
		if d < 0 {
			d = -d
		}
		if d > max {
			max = d
		}
	}
	return max
}

// GateError blocks deploy/balance merges when swing exceeds threshold (e.g. 0.05 = 5%).
type GateError struct {
	Threshold float64
	Swing     float64
}

func (e GateError) Error() string {
	return fmt.Sprintf("balance gate: win-rate swing %.2f%% exceeds threshold %.2f%%", e.Swing*100, e.Threshold*100)
}

// CheckWinRateGate returns GateError when any species swing > threshold.
func CheckWinRateGate(baseline, current WinRateReport, threshold float64) error {
	swing := MaxWinRateSwing(baseline, current)
	if swing > threshold {
		return GateError{Threshold: threshold, Swing: swing}
	}
	return nil
}
