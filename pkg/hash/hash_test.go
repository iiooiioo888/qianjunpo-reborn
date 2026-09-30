package hash

import "testing"

func TestFNVGolden(t *testing.T) {
	// Known FNV-1a 64 for empty string
	h := New()
	if h.Sum64() != offset64 {
		t.Fatal("empty hash")
	}
	h.Write([]byte("hello"))
	want := uint64(0xa430d84680aabd0b)
	if h.Sum64() != want {
		t.Fatalf("hello fnv got %x want %x", h.Sum64(), want)
	}
}

func TestUint64Deterministic(t *testing.T) {
	a := New()
	a.WriteUint64(42)
	b := New()
	b.WriteUint64(42)
	if a.Sum64() != b.Sum64() {
		t.Fatal("uint64 mix not deterministic")
	}
}
