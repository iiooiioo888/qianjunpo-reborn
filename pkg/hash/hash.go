// Package hash provides deterministic FNV-1a 64-bit hashing for desync detection.
package hash

const (
	offset64 = 14695981039346656037
	prime64  = 1099511628211
)

// Hasher accumulates FNV-1a 64-bit hash bytes.
type Hasher struct {
	h uint64
}

// New creates a hasher with the standard offset basis.
func New() *Hasher {
	return &Hasher{h: offset64}
}

// Write mixes bytes into the hash.
func (hasher *Hasher) Write(p []byte) {
	for _, b := range p {
		hasher.h ^= uint64(b)
		hasher.h *= prime64
	}
}

// WriteUint64 mixes a uint64 in little-endian order.
func (hasher *Hasher) WriteUint64(v uint64) {
	hasher.h ^= v & 0xff
	hasher.h *= prime64
	hasher.h ^= (v >> 8) & 0xff
	hasher.h *= prime64
	hasher.h ^= (v >> 16) & 0xff
	hasher.h *= prime64
	hasher.h ^= (v >> 24) & 0xff
	hasher.h *= prime64
	hasher.h ^= (v >> 32) & 0xff
	hasher.h *= prime64
	hasher.h ^= (v >> 40) & 0xff
	hasher.h *= prime64
	hasher.h ^= (v >> 48) & 0xff
	hasher.h *= prime64
	hasher.h ^= (v >> 56) & 0xff
	hasher.h *= prime64
}

// WriteInt64 mixes a signed int64 deterministically.
func (hasher *Hasher) WriteInt64(v int64) {
	hasher.WriteUint64(uint64(v))
}

// Sum64 returns the current digest.
func (hasher *Hasher) Sum64() uint64 {
	return hasher.h
}

// HashBytes is a convenience one-shot hash.
func HashBytes(p []byte) uint64 {
	h := New()
	h.Write(p)
	return h.Sum64()
}
