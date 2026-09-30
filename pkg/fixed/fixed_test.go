package fixed

import (
	"math"
	"testing"
)

func TestFromIntToInt(t *testing.T) {
	v := FromInt(42)
	if ToInt(v) != 42 {
		t.Fatalf("expected 42, got %d", ToInt(v))
	}
}

func TestFromFloatInit(t *testing.T) {
	v := FromFloat(1.5)
	if ToInt(v) != 1 {
		t.Fatalf("truncated int expected 1")
	}
	oneHalf := FromInt(1).Add(FromInt(1).Div(FromInt(2)))
	if v.Sub(oneHalf).raw != 0 && v.Sub(oneHalf).raw != -1 && v.Sub(oneHalf).raw != 1 {
		t.Fatalf("FromFloat(1.5) raw=%d expected ~%d", v.raw, oneHalf.raw)
	}
}

func TestAddSub(t *testing.T) {
	a := FromInt(3)
	b := FromInt(7)
	if !a.Add(b).Equal(FromInt(10)) {
		t.Fatal("add")
	}
	if !b.Sub(a).Equal(FromInt(4)) {
		t.Fatal("sub")
	}
}

func TestMulBasic(t *testing.T) {
	two := FromInt(2)
	three := FromInt(3)
	if !two.Mul(three).Equal(FromInt(6)) {
		t.Fatalf("2*3 = %d", ToInt(two.Mul(three)))
	}
	frac := FromInt(1).Div(FromInt(2)) // 0.5
	if !frac.Mul(FromInt(4)).Equal(FromInt(2)) {
		t.Fatal("0.5 * 4")
	}
}

func TestDivBasic(t *testing.T) {
	if !FromInt(10).Div(FromInt(4)).Equal(FromFloat(2.5)) {
		t.Fatal("10/4")
	}
	if FromInt(1).Div(Zero) != MaxValue {
		t.Fatal("div by zero -> MaxValue")
	}
}

func TestMulOverflowClamp(t *testing.T) {
	big := Fixed{raw: math.MaxInt64 >> 1}
	out := big.Mul(big)
	if out.raw <= 0 {
		t.Fatal("expected positive clamped product")
	}
}

func TestSqrt(t *testing.T) {
	four := FromInt(4)
	two := FromInt(2)
	s := Sqrt(four)
	diff := s.Sub(two).raw
	if diff < -2 || diff > 2 {
		t.Fatalf("sqrt(4) expected ~2, got raw %d", s.raw)
	}
	if Sqrt(Zero) != Zero {
		t.Fatal("sqrt(0)")
	}
}

func (a Fixed) Equal(b Fixed) bool {
	return a.raw == b.raw
}
