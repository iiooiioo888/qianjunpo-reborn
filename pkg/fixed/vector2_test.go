package fixed

import "testing"

func TestVector2LengthSq(t *testing.T) {
	v := NewVector2(FromInt(3), FromInt(4))
	ls := v.LengthSq()
	if !ls.Equal(FromInt(25)) {
		t.Fatalf("length sq expected 25, got %d", ToInt(ls))
	}
}

func TestVector2Length(t *testing.T) {
	v := NewVector2(FromInt(3), FromInt(4))
	l := v.Length()
	if ToInt(l) != 5 {
		t.Fatalf("length expected 5, got %d", ToInt(l))
	}
}

func TestVector2Ops(t *testing.T) {
	a := NewVector2(FromInt(1), FromInt(2))
	b := NewVector2(FromInt(3), FromInt(4))
	sum := a.Add(b)
	if ToInt(sum.X) != 4 || ToInt(sum.Y) != 6 {
		t.Fatal("add")
	}
	scaled := a.Scale(FromInt(2))
	if ToInt(scaled.X) != 2 || ToInt(scaled.Y) != 4 {
		t.Fatal("scale")
	}
}
