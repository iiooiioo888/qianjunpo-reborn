package rng

import (
	"testing"
)

// Golden values: first 10 outputs of Xorshift128+ with explicit state (1, 2).
var goldenFirst10 = []uint64{
	0x800045,
	0x2000104,
	0x4000020010c3,
	0xc00002103045,
	0x1000801c450c4,
	0x148200440334b,
	0x40118200d8e16da,
	0xc02e8191918b86e,
	0x1008a12455ca592a,
	0x1089632414ba78d4,
}

func TestGoldenSequence(t *testing.T) {
	r := NewFromState(1, 2)
	for i := 0; i < len(goldenFirst10); i++ {
		got := r.NextUint64()
		if got != goldenFirst10[i] {
			t.Fatalf("step %d: got %x want %x", i, got, goldenFirst10[i])
		}
	}
}

func TestMillionStepsChecksum(t *testing.T) {
	r := NewFromState(1, 2)
	var sum uint64
	const n = 1_000_000
	for i := 0; i < n; i++ {
		sum += r.NextUint64()
	}
	// Checksum captured from reference run; detects algorithm drift.
	const wantSum = 0x34c65acf3c9ecf5f
	if sum != wantSum {
		// Recompute on first failure to update - use known reference
		t.Fatalf("1M sum mismatch: got %016x", sum)
	}
}

func TestStateExportClone(t *testing.T) {
	a := New(12345)
	_ = a.NextUint64()
	_ = a.NextUint64()
	st := a.GetState()
	b := a.Clone()
	c := NewFromState(st.S0, st.S1)
	for i := 0; i < 1000; i++ {
		va := a.NextUint64()
		vb := b.NextUint64()
		vc := c.NextUint64()
		if va != vb || va != vc {
			t.Fatalf("diverged at %d", i)
		}
	}
}

func TestSetState(t *testing.T) {
	r1 := New(99)
	for i := 0; i < 500; i++ {
		r1.NextUint64()
	}
	st := r1.GetState()
	r2 := New(0)
	r2.SetState(st)
	for i := 0; i < 100; i++ {
		if r1.NextUint64() != r2.NextUint64() {
			t.Fatal("set state mismatch")
		}
	}
}

func TestBounded(t *testing.T) {
	r := New(7)
	for i := 0; i < 10000; i++ {
		v := r.NextIntBounded(10)
		if v < 0 || v >= 10 {
			t.Fatalf("out of range: %d", v)
		}
	}
}
