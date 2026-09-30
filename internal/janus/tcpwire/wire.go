// Package tcpwire defines the Janus client TCP framing for tactical gateway ops.
package tcpwire

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/iiooiioo888/qianjunpo-reborn/pkg/tactical"
)

const (
	Magic uint32 = 0x544a5051 // "QJPT" little-endian
	Vers  uint16 = 1
)

// Opcode identifies a tactical bridge message.
type Opcode uint8

const (
	OpConnect       Opcode = 0x01
	OpConnectAck    Opcode = 0x02
	OpEnterBattle   Opcode = 0x10
	OpEnterBattleAck Opcode = 0x11
	OpSubmitTactical Opcode = 0x20
	OpSubmitTacticalAck Opcode = 0x21
	OpStepLockstep  Opcode = 0x30
	OpStepLockstepAck Opcode = 0x31
	OpError         Opcode = 0x7f
)

// Header is fixed prefix before body.
type Header struct {
	Magic  uint32
	Vers   uint16
	Opcode Opcode
}

// Packet is one framed message.
type Packet struct {
	Header Header
	Body   []byte
}

var (
	ErrBadMagic   = errors.New("tcpwire: bad magic")
	ErrBadVersion = errors.New("tcpwire: bad version")
	ErrShortBody  = errors.New("tcpwire: short body")
)

// ReadPacket reads one length-prefixed frame.
func ReadPacket(r io.Reader) (Packet, error) {
	var hdr [11]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return Packet{}, err
	}
	p := Packet{}
	p.Header.Magic = binary.LittleEndian.Uint32(hdr[0:4])
	p.Header.Vers = binary.LittleEndian.Uint16(hdr[4:6])
	p.Header.Opcode = Opcode(hdr[6])
	bodyLen := binary.LittleEndian.Uint32(hdr[7:11])
	if p.Header.Magic != Magic {
		return Packet{}, ErrBadMagic
	}
	if p.Header.Vers != Vers {
		return Packet{}, ErrBadVersion
	}
	if bodyLen > 1<<20 {
		return Packet{}, errors.New("tcpwire: body too large")
	}
	p.Body = make([]byte, bodyLen)
	if _, err := io.ReadFull(r, p.Body); err != nil {
		return Packet{}, err
	}
	return p, nil
}

// WritePacket writes header + body.
func WritePacket(w io.Writer, opcode Opcode, body []byte) error {
	var hdr [11]byte
	binary.LittleEndian.PutUint32(hdr[0:4], Magic)
	binary.LittleEndian.PutUint16(hdr[4:6], Vers)
	hdr[6] = byte(opcode)
	binary.LittleEndian.PutUint32(hdr[7:11], uint32(len(body)))
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	_, err := w.Write(body)
	return err
}

func writeString(buf *[]byte, s string) {
	b := []byte(s)
	var lenb [2]byte
	binary.LittleEndian.PutUint16(lenb[:], uint16(len(b)))
	*buf = append(*buf, lenb[0], lenb[1])
	*buf = append(*buf, b...)
}

func readString(body []byte, off int) (string, int, error) {
	if off+2 > len(body) {
		return "", off, ErrShortBody
	}
	n := int(binary.LittleEndian.Uint16(body[off : off+2]))
	off += 2
	if off+n > len(body) {
		return "", off, ErrShortBody
	}
	return string(body[off : off+n]), off + n, nil
}

// EncodeConnect builds OpConnect body.
func EncodeConnect(accessToken, zoneID string, shard uint32) []byte {
	var b []byte
	writeString(&b, accessToken)
	writeString(&b, zoneID)
	var sh [4]byte
	binary.LittleEndian.PutUint32(sh[:], shard)
	b = append(b, sh[0], sh[1], sh[2], sh[3])
	return b
}

// DecodeConnect parses OpConnect body.
func DecodeConnect(body []byte) (accessToken, zoneID string, shard uint32, err error) {
	off := 0
	accessToken, off, err = readString(body, off)
	if err != nil {
		return "", "", 0, err
	}
	zoneID, off, err = readString(body, off)
	if err != nil {
		return "", "", 0, err
	}
	if off+4 > len(body) {
		return "", "", 0, ErrShortBody
	}
	shard = binary.LittleEndian.Uint32(body[off : off+4])
	return accessToken, zoneID, shard, nil
}

// EncodeConnectAck builds OpConnectAck body.
func EncodeConnectAck(sessionID, romaEndpoint string, wallMs uint64, simTick uint64) []byte {
	var b []byte
	writeString(&b, sessionID)
	writeString(&b, romaEndpoint)
	var ts [16]byte
	binary.LittleEndian.PutUint64(ts[0:8], wallMs)
	binary.LittleEndian.PutUint64(ts[8:16], simTick)
	b = append(b, ts[0], ts[1], ts[2], ts[3], ts[4], ts[5], ts[6], ts[7])
	b = append(b, ts[8], ts[9], ts[10], ts[11], ts[12], ts[13], ts[14], ts[15])
	return b
}

// DecodeConnectAck parses OpConnectAck body.
func DecodeConnectAck(body []byte) (sessionID, romaEndpoint string, wallMs, simTick uint64, err error) {
	off := 0
	sessionID, off, err = readString(body, off)
	if err != nil {
		return "", "", 0, 0, err
	}
	romaEndpoint, off, err = readString(body, off)
	if err != nil {
		return "", "", 0, 0, err
	}
	if off+16 > len(body) {
		return "", "", 0, 0, ErrShortBody
	}
	wallMs = binary.LittleEndian.Uint64(body[off : off+8])
	simTick = binary.LittleEndian.Uint64(body[off+8 : off+16])
	return sessionID, romaEndpoint, wallMs, simTick, nil
}

// EncodeEnterBattle builds OpEnterBattle body.
func EncodeEnterBattle(sessionID, accessToken, zoneID string, shard uint32) []byte {
	var b []byte
	writeString(&b, sessionID)
	writeString(&b, accessToken)
	writeString(&b, zoneID)
	var sh [4]byte
	binary.LittleEndian.PutUint32(sh[:], shard)
	b = append(b, sh[0], sh[1], sh[2], sh[3])
	return b
}

// DecodeEnterBattle parses OpEnterBattle body.
func DecodeEnterBattle(body []byte) (sessionID, accessToken, zoneID string, shard uint32, err error) {
	off := 0
	sessionID, off, err = readString(body, off)
	if err != nil {
		return "", "", "", 0, err
	}
	accessToken, off, err = readString(body, off)
	if err != nil {
		return "", "", "", 0, err
	}
	zoneID, off, err = readString(body, off)
	if err != nil {
		return "", "", "", 0, err
	}
	if off+4 > len(body) {
		return "", "", "", 0, ErrShortBody
	}
	shard = binary.LittleEndian.Uint32(body[off : off+4])
	return sessionID, accessToken, zoneID, shard, nil
}

// EncodeEnterBattleAck builds OpEnterBattleAck body.
func EncodeEnterBattleAck(battleID string, initialHash uint64) []byte {
	var b []byte
	writeString(&b, battleID)
	var h [8]byte
	binary.LittleEndian.PutUint64(h[:], initialHash)
	b = append(b, h[0], h[1], h[2], h[3], h[4], h[5], h[6], h[7])
	return b
}

// DecodeEnterBattleAck parses OpEnterBattleAck body.
func DecodeEnterBattleAck(body []byte) (battleID string, initialHash uint64, err error) {
	off := 0
	battleID, off, err = readString(body, off)
	if err != nil {
		return "", 0, err
	}
	if off+8 > len(body) {
		return "", 0, ErrShortBody
	}
	initialHash = binary.LittleEndian.Uint64(body[off : off+8])
	return battleID, initialHash, nil
}

// EncodeSubmitTactical builds OpSubmitTactical body (8-byte tactical command payload).
func EncodeSubmitTactical(sessionID, battleID string, cmd tactical.Command) []byte {
	var b []byte
	writeString(&b, sessionID)
	writeString(&b, battleID)
	payload := tactical.Encode(cmd)
	b = append(b, payload...)
	return b
}

// DecodeSubmitTactical parses OpSubmitTactical body.
func DecodeSubmitTactical(body []byte) (sessionID, battleID string, cmd tactical.Command, err error) {
	off := 0
	sessionID, off, err = readString(body, off)
	if err != nil {
		return "", "", cmd, err
	}
	battleID, off, err = readString(body, off)
	if err != nil {
		return "", "", cmd, err
	}
	if off+8 > len(body) {
		return "", "", cmd, ErrShortBody
	}
	cmd, err = tactical.Decode(body[off : off+8])
	return sessionID, battleID, cmd, err
}

// EncodeSubmitTacticalAck builds OpSubmitTacticalAck body.
func EncodeSubmitTacticalAck(accepted bool, rejectReason string, stateHash, lockstepFrame uint64) []byte {
	var b []byte
	if accepted {
		b = append(b, 1)
	} else {
		b = append(b, 0)
	}
	writeString(&b, rejectReason)
	var nums [16]byte
	binary.LittleEndian.PutUint64(nums[0:8], stateHash)
	binary.LittleEndian.PutUint64(nums[8:16], lockstepFrame)
	b = append(b, nums[0], nums[1], nums[2], nums[3], nums[4], nums[5], nums[6], nums[7])
	b = append(b, nums[8], nums[9], nums[10], nums[11], nums[12], nums[13], nums[14], nums[15])
	return b
}

// DecodeSubmitTacticalAck parses OpSubmitTacticalAck body.
func DecodeSubmitTacticalAck(body []byte) (accepted bool, rejectReason string, stateHash, lockstepFrame uint64, err error) {
	if len(body) < 1 {
		return false, "", 0, 0, ErrShortBody
	}
	accepted = body[0] != 0
	off := 1
	rejectReason, off, err = readString(body, off)
	if err != nil {
		return false, "", 0, 0, err
	}
	if off+16 > len(body) {
		return false, "", 0, 0, ErrShortBody
	}
	stateHash = binary.LittleEndian.Uint64(body[off : off+8])
	lockstepFrame = binary.LittleEndian.Uint64(body[off+8 : off+16])
	return accepted, rejectReason, stateHash, lockstepFrame, nil
}

// EncodeStepLockstep builds OpStepLockstep body.
func EncodeStepLockstep(sessionID, battleID string, steps uint32) []byte {
	var b []byte
	writeString(&b, sessionID)
	writeString(&b, battleID)
	var st [4]byte
	binary.LittleEndian.PutUint32(st[:], steps)
	b = append(b, st[0], st[1], st[2], st[3])
	return b
}

// DecodeStepLockstep parses OpStepLockstep body.
func DecodeStepLockstep(body []byte) (sessionID, battleID string, steps uint32, err error) {
	off := 0
	sessionID, off, err = readString(body, off)
	if err != nil {
		return "", "", 0, err
	}
	battleID, off, err = readString(body, off)
	if err != nil {
		return "", "", 0, err
	}
	if off+4 > len(body) {
		return "", "", 0, ErrShortBody
	}
	steps = binary.LittleEndian.Uint32(body[off : off+4])
	return sessionID, battleID, steps, nil
}

// EncodeStepLockstepAck builds OpStepLockstepAck body.
func EncodeStepLockstepAck(lockstepFrame, stateHash uint64, finished bool, winner uint32) []byte {
	var b []byte
	var nums [16]byte
	binary.LittleEndian.PutUint64(nums[0:8], lockstepFrame)
	binary.LittleEndian.PutUint64(nums[8:16], stateHash)
	b = append(b, nums[0], nums[1], nums[2], nums[3], nums[4], nums[5], nums[6], nums[7])
	b = append(b, nums[8], nums[9], nums[10], nums[11], nums[12], nums[13], nums[14], nums[15])
	if finished {
		b = append(b, 1)
	} else {
		b = append(b, 0)
	}
	var w [4]byte
	binary.LittleEndian.PutUint32(w[:], winner)
	b = append(b, w[0], w[1], w[2], w[3])
	return b
}

// DecodeStepLockstepAck parses OpStepLockstepAck body.
func DecodeStepLockstepAck(body []byte) (lockstepFrame, stateHash uint64, finished bool, winner uint32, err error) {
	if len(body) < 21 {
		return 0, 0, false, 0, ErrShortBody
	}
	lockstepFrame = binary.LittleEndian.Uint64(body[0:8])
	stateHash = binary.LittleEndian.Uint64(body[8:16])
	finished = body[16] != 0
	winner = binary.LittleEndian.Uint32(body[17:21])
	return lockstepFrame, stateHash, finished, winner, nil
}

// EncodeError builds OpError body.
func EncodeError(msg string) []byte {
	var b []byte
	writeString(&b, msg)
	return b
}

// DecodeError parses OpError body.
func DecodeError(body []byte) (string, error) {
	msg, _, err := readString(body, 0)
	return msg, err
}
