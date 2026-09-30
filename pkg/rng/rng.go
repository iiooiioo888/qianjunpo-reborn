// Package rng provides deterministic Xorshift128+ pseudo-random numbers.
package rng

import (
	"encoding/binary"
)

// State holds the two 64-bit words of Xorshift128+.
type State struct {
	S0 uint64
	S1 uint64
}

// RNG is a Xorshift128+ generator (Vigna).
type RNG struct {
	s0, s1 uint64
}

// New creates an RNG from a single seed by splitting into two state words.
func New(seed uint64) *RNG {
	r := &RNG{}
	r.s0 = splitmix64(seed)
	r.s1 = splitmix64(r.s0)
	if r.s0 == 0 && r.s1 == 0 {
		r.s1 = 1
	}
	return r
}

// NewFromState creates an RNG from explicit state.
func NewFromState(s0, s1 uint64) *RNG {
	if s0 == 0 && s1 == 0 {
		s1 = 1
	}
	return &RNG{s0: s0, s1: s1}
}

func splitmix64(x uint64) uint64 {
	z := x + 0x9e3779b97f4a7c15
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// NextUint64 advances the generator and returns the next value.
func (r *RNG) NextUint64() uint64 {
	s1 := r.s0
	s0 := r.s1
	s1 ^= s1 << 23
	s1 ^= s1 >> 17
	s1 ^= s0
	s1 ^= s0 >> 26
	r.s0 = s0
	r.s1 = s1
	return s0 + s1
}

// NextUint64Bounded returns a value in [0, bound).
func (r *RNG) NextUint64Bounded(bound uint64) uint64 {
	if bound <= 1 {
		return 0
	}
	return r.NextUint64() % bound
}

// NextIntBounded returns a uniform int in [0, bound).
func (r *RNG) NextIntBounded(bound int) int {
	if bound <= 0 {
		return 0
	}
	return int(r.NextUint64Bounded(uint64(bound)))
}

// GetState exports the current state.
func (r *RNG) GetState() State {
	return State{S0: r.s0, S1: r.s1}
}

// SetState restores state.
func (r *RNG) SetState(s State) {
	if s.S0 == 0 && s.S1 == 0 {
		s.S1 = 1
	}
	r.s0, r.s1 = s.S0, s.S1
}

// Clone returns an independent copy with the same state.
func (r *RNG) Clone() *RNG {
	return &RNG{s0: r.s0, s1: r.s1}
}

// MarshalState encodes state for hashing/snapshots.
func (s State) Marshal() []byte {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[0:], s.S0)
	binary.LittleEndian.PutUint64(buf[8:], s.S1)
	return buf
}
