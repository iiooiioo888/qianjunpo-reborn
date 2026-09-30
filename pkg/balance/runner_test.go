package balance

import "testing"

func TestRunBatchWinRates(t *testing.T) {
	report := DefaultBatch(20)
	if report.Matches != 20 {
		t.Fatalf("matches=%d", report.Matches)
	}
	sum := 0.0
	for _, r := range report.WinRates {
		sum += r
	}
	if sum < 0.99 || sum > 1.01 {
		t.Fatalf("win rates should sum ~1, got %v", sum)
	}
}

func TestCheckWinRateGatePass(t *testing.T) {
	base := DefaultBatch(40)
	cur := DefaultBatch(40)
	if err := CheckWinRateGate(base, cur, 0.05); err != nil {
		t.Fatalf("identical batches should pass: %v", err)
	}
}

func TestCheckWinRateGateDeliberateFail(t *testing.T) {
	base := WinRateReport{
		Matches: 100,
		WinRates: map[Species]float64{
			0: 0.50, 1: 0.30, 2: 0.20,
		},
	}
	skewed := WinRateReport{
		Matches: 100,
		WinRates: map[Species]float64{
			0: 0.10, 1: 0.70, 2: 0.20,
		},
	}
	err := CheckWinRateGate(base, skewed, 0.05)
	if err == nil {
		t.Fatal("expected gate failure for >5% species swing")
	}
	if _, ok := err.(GateError); !ok {
		t.Fatalf("expected GateError, got %T", err)
	}
}
