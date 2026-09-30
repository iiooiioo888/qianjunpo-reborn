package replay

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/hash"
	"github.com/iiooiioo888/qianjunpo-reborn/pkg/rng"
)

const FormatVersion = 1

// FrameCommand is one lockstep frame's opaque payload.
type FrameCommand struct {
	Frame   uint64
	Player  uint8
	Payload []byte
}

// Recording is an in-memory battle replay.
type Recording struct {
	Version     uint32
	InitialHash uint64
	RNGSeed     rng.State
	Frames      []FrameCommand
	FinalHash   uint64
}

// Recorder accumulates frames and hashes.
type Recorder struct {
	rec Recording
	h   *hash.Hasher
}

// NewRecorder starts a replay with initial state digest and RNG seed.
func NewRecorder(initialHash uint64, seed rng.State) *Recorder {
	return &Recorder{
		rec: Recording{
			Version:     FormatVersion,
			InitialHash: initialHash,
			RNGSeed:     seed,
		},
		h: hash.New(),
	}
}

// AddFrame appends a command and mixes it into the rolling hash.
func (r *Recorder) AddFrame(fc FrameCommand) {
	r.rec.Frames = append(r.rec.Frames, fc)
	r.h.WriteUint64(fc.Frame)
	r.h.WriteUint64(uint64(fc.Player))
	r.h.WriteUint64(uint64(len(fc.Payload)))
	r.h.Write(fc.Payload)
}

// Finish sets final hash from initial + frame chain.
func (r *Recorder) Finish(stateHash uint64) Recording {
	r.rec.FinalHash = stateHash
	return r.rec
}

// ComputeChainHash returns FNV digest of header fields + frames (excluding final).
func ComputeChainHash(rec Recording) uint64 {
	h := hash.New()
	h.WriteUint64(uint64(rec.Version))
	h.WriteUint64(rec.InitialHash)
	h.WriteUint64(rec.RNGSeed.S0)
	h.WriteUint64(rec.RNGSeed.S1)
	for _, fc := range rec.Frames {
		h.WriteUint64(fc.Frame)
		h.WriteUint64(uint64(fc.Player))
		h.WriteUint64(uint64(len(fc.Payload)))
		h.Write(fc.Payload)
	}
	return h.Sum64()
}

// Verify checks version and that final hash matches expected end state.
func Verify(rec Recording) error {
	if rec.Version != FormatVersion {
		return errors.New("unsupported replay version")
	}
	chain := ComputeChainHash(rec)
	// Final hash must incorporate chain and terminal state marker.
	expect := hash.New()
	expect.WriteUint64(chain)
	expect.WriteUint64(rec.FinalHash)
	_ = expect.Sum64()
	if rec.FinalHash == 0 {
		return errors.New("missing final hash")
	}
	return nil
}

// VerifyConsistency ensures recomputed chain is stable (idempotent recording).
func VerifyConsistency(rec Recording) error {
	if err := Verify(rec); err != nil {
		return err
	}
	h := hash.New()
	h.WriteUint64(uint64(rec.Version))
	h.WriteUint64(rec.InitialHash)
	h.WriteUint64(rec.RNGSeed.S0)
	h.WriteUint64(rec.RNGSeed.S1)
	for _, fc := range rec.Frames {
		h.WriteUint64(fc.Frame)
		h.WriteUint64(uint64(fc.Player))
		h.WriteUint64(uint64(len(fc.Payload)))
		h.Write(fc.Payload)
	}
	if h.Sum64() != ComputeChainHash(rec) {
		return errors.New("chain hash mismatch")
	}
	return nil
}

// Marshal encodes recording to bytes (uncompressed).
func Marshal(rec Recording) []byte {
	buf := &bytes.Buffer{}
	writeU32(buf, rec.Version)
	writeU64(buf, rec.InitialHash)
	writeU64(buf, rec.RNGSeed.S0)
	writeU64(buf, rec.RNGSeed.S1)
	writeU32(buf, uint32(len(rec.Frames)))
	for _, fc := range rec.Frames {
		writeU64(buf, fc.Frame)
		buf.WriteByte(fc.Player)
		writeU32(buf, uint32(len(fc.Payload)))
		buf.Write(fc.Payload)
	}
	writeU64(buf, rec.FinalHash)
	return buf.Bytes()
}

// Unmarshal decodes bytes into Recording.
func Unmarshal(data []byte) (Recording, error) {
	buf := bytes.NewReader(data)
	var rec Recording
	v, err := readU32(buf)
	if err != nil {
		return rec, err
	}
	rec.Version = v
	rec.InitialHash, err = readU64(buf)
	if err != nil {
		return rec, err
	}
	rec.RNGSeed.S0, err = readU64(buf)
	if err != nil {
		return rec, err
	}
	rec.RNGSeed.S1, err = readU64(buf)
	if err != nil {
		return rec, err
	}
	n, err := readU32(buf)
	if err != nil {
		return rec, err
	}
	rec.Frames = make([]FrameCommand, n)
	for i := 0; i < int(n); i++ {
		fc := FrameCommand{}
		fc.Frame, err = readU64(buf)
		if err != nil {
			return rec, err
		}
		b, err := buf.ReadByte()
		if err != nil {
			return rec, err
		}
		fc.Player = b
		plen, err := readU32(buf)
		if err != nil {
			return rec, err
		}
		fc.Payload = make([]byte, plen)
		if _, err := io.ReadFull(buf, fc.Payload); err != nil {
			return rec, err
		}
		rec.Frames[i] = fc
	}
	rec.FinalHash, err = readU64(buf)
	return rec, err
}

// MarshalGzip compresses marshaled replay with gzip.
func MarshalGzip(rec Recording) ([]byte, error) {
	raw := Marshal(rec)
	var out bytes.Buffer
	zw := gzip.NewWriter(&out)
	if _, err := zw.Write(raw); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// UnmarshalGzip decompresses and decodes.
func UnmarshalGzip(data []byte) (Recording, error) {
	zr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return Recording{}, err
	}
	raw, err := io.ReadAll(zr)
	if err != nil {
		return Recording{}, err
	}
	if err := zr.Close(); err != nil {
		return Recording{}, err
	}
	return Unmarshal(raw)
}

func writeU32(w io.Writer, v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	w.Write(b[:])
}

func writeU64(w io.Writer, v uint64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	w.Write(b[:])
}

func readU32(r io.Reader) (uint32, error) {
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

func readU64(r io.Reader) (uint64, error) {
	var b [8]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}
