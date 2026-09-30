// Package fixed provides FP64 (32.32) fixed-point arithmetic for deterministic simulation.
package fixed

import (
	"math"
	"math/bits"
)

const fracBits = 32

// Fixed is a signed 32.32 fixed-point value stored as raw int64 units.
type Fixed struct {
	raw int64
}

// Zero, One, and MaxValue are common fixed-point constants.
var (
	Zero     = Fixed{raw: 0}
	One      = Fixed{raw: 1 << fracBits}
	MaxValue = Fixed{raw: math.MaxInt64}
)

// FromInt converts an integer to fixed-point.
func FromInt(v int64) Fixed {
	return Fixed{raw: v << fracBits}
}

// FromFloat converts a float64 to fixed-point. Use only during initialization/config.
func FromFloat(v float64) Fixed {
	return Fixed{raw: int64(v * float64(uint64(1)<<fracBits))}
}

// ToInt truncates toward zero.
func ToInt(f Fixed) int64 {
	return f.raw >> fracBits
}

// Raw returns the underlying raw representation (for hashing and tests).
func (f Fixed) Raw() int64 {
	return f.raw
}

// FromRaw constructs a Fixed from raw storage (e.g. after snapshot restore).
func FromRaw(raw int64) Fixed {
	return Fixed{raw: raw}
}

func (a Fixed) Add(b Fixed) Fixed {
	return Fixed{raw: a.raw + b.raw}
}

func (a Fixed) Sub(b Fixed) Fixed {
	return Fixed{raw: a.raw - b.raw}
}

func (a Fixed) Mul(b Fixed) Fixed {
	neg := (a.raw < 0) != (b.raw < 0)
	ua := abs64(a.raw)
	ub := abs64(b.raw)
	hi, lo := mul64(ua, ub)
	shifted := shiftRight128(hi, lo, fracBits)
	if neg {
		if shifted > uint64(math.MaxInt64) {
			return MaxValue
		}
		return Fixed{raw: -int64(shifted)}
	}
	if shifted > uint64(math.MaxInt64) {
		return MaxValue
	}
	return Fixed{raw: int64(shifted)}
}

func (a Fixed) Div(b Fixed) Fixed {
	if b.raw == 0 {
		return MaxValue
	}
	neg := (a.raw < 0) != (b.raw < 0)
	ua := abs64(a.raw)
	ub := abs64(b.raw)
	numHi, numLo := shiftLeft128(ua, fracBits)
	quoHi, quoLo := div128(numHi, numLo, ub)
	// Quotient should fit in int64 for sane inputs.
	_ = quoHi
	if quoLo > uint64(math.MaxInt64) {
		return MaxValue
	}
	if neg {
		return Fixed{raw: -int64(quoLo)}
	}
	return Fixed{raw: int64(quoLo)}
}

func abs64(v int64) uint64 {
	if v < 0 {
		return uint64(-v)
	}
	return uint64(v)
}

func mul64(a, b uint64) (hi, lo uint64) {
	const mask32 = uint64(0xffffffff)
	al := a & mask32
	ah := a >> 32
	bl := b & mask32
	bh := b >> 32

	ll := al * bl
	lh := al * bh
	hl := ah * bl
	hh := ah * bh

	mid := (ll >> 32) + (lh & mask32) + (hl & mask32)
	hi = hh + (lh >> 32) + (hl >> 32) + (mid >> 32)
	lo = (mid << 32) | (ll & mask32)
	return hi, lo
}

func shiftRight128(hi, lo uint64, n uint) uint64 {
	if n == 0 {
		return lo
	}
	if n >= 64 {
		return hi >> (n - 64)
	}
	return hi<<(64-n) | lo>>n
}

func shiftLeft128(v uint64, n uint) (hi, lo uint64) {
	if n >= 64 {
		return v << (n - 64), 0
	}
	hi = v >> (64 - n)
	lo = v << n
	return hi, lo
}

func div128(numHi, numLo, den uint64) (quoHi, quoLo uint64) {
	if den == 0 {
		return 0, 0
	}
	q, _ := bits.Div64(numHi, numLo, den)
	return 0, q
}

// Sqrt returns the fixed-point square root using deterministic integer iteration.
func Sqrt(v Fixed) Fixed {
	if v.raw <= 0 {
		return Zero
	}
	lo := int64(0)
	hi := v.raw
	if hi < 0 {
		return Zero
	}
	// Scale upper bound: sqrt(max) ~ 2^31 in fixed units.
	hiBound := int64(1) << 48
	if hi > hiBound {
		hi = hiBound
	}
	for lo < hi-1 {
		mid := lo + (hi-lo)/2
		midSq := Fixed{raw: mid}.Mul(Fixed{raw: mid}).raw
		if midSq <= v.raw {
			lo = mid
		} else {
			hi = mid
		}
	}
	return Fixed{raw: lo}
}
