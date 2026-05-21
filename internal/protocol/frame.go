package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var Magic = [4]byte{'P', 'R', 'O', 'J'}

const HeaderSize = 9 // Magic(4) + Type(1) + PayloadLen(4)

const (
	TypeControl byte = 0x01
	TypeVideo   byte = 0x02
)

var (
	ErrIncompleteHeader  = errors.New("incomplete header")
	ErrBadMagic          = errors.New("bad magic")
	ErrPayloadTooLarge   = errors.New("payload too large")
	ErrIncompletePayload = errors.New("incomplete payload")
)

func EncodeMessage(typ byte, payload []byte) []byte {
	buf := make([]byte, HeaderSize+len(payload))
	copy(buf[:4], Magic[:])
	buf[4] = typ
	binary.BigEndian.PutUint32(buf[5:9], uint32(len(payload)))
	copy(buf[9:], payload)
	return buf
}

func DecodeMessage(data []byte) (typ byte, payload []byte, err error) {
	if len(data) < HeaderSize {
		return 0, nil, ErrIncompleteHeader
	}
	if string(data[:4]) != string(Magic[:]) {
		return 0, nil, ErrBadMagic
	}
	typ = data[4]
	payLen := binary.BigEndian.Uint32(data[5:9])
	if payLen > 100*1024*1024 { // 100MB sanity cap
		return 0, nil, ErrPayloadTooLarge
	}
	if len(data) < HeaderSize+int(payLen) {
		return 0, nil, fmt.Errorf("have %d, need %d: %w", len(data), HeaderSize+payLen, ErrIncompletePayload)
	}
	payload = make([]byte, payLen)
	copy(payload, data[9:9+payLen])
	return typ, payload, nil
}
