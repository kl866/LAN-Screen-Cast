package protocol

import (
	"encoding/binary"
	"errors"
)

func EncodeVideoBlock(x, y, w, h int, pngData []byte) []byte {
	buf := make([]byte, 10+len(pngData))
	binary.BigEndian.PutUint16(buf[0:2], uint16(x))
	binary.BigEndian.PutUint16(buf[2:4], uint16(y))
	binary.BigEndian.PutUint16(buf[4:6], uint16(w))
	binary.BigEndian.PutUint16(buf[6:8], uint16(h))
	binary.BigEndian.PutUint16(buf[8:10], uint16(len(pngData)))
	copy(buf[10:], pngData)
	return buf
}

var ErrVideoBlockTooShort = errors.New("video block too short")
var ErrVideoBlockTruncated = errors.New("video block data truncated")

func DecodeVideoBlock(data []byte) (x, y, w, h int, pngData []byte, err error) {
	if len(data) < 10 {
		return 0, 0, 0, 0, nil, ErrVideoBlockTooShort
	}
	x = int(binary.BigEndian.Uint16(data[0:2]))
	y = int(binary.BigEndian.Uint16(data[2:4]))
	w = int(binary.BigEndian.Uint16(data[4:6]))
	h = int(binary.BigEndian.Uint16(data[6:8]))
	size := int(binary.BigEndian.Uint16(data[8:10]))
	if len(data) < 10+size {
		return 0, 0, 0, 0, nil, ErrVideoBlockTruncated
	}
	pngData = make([]byte, size)
	copy(pngData, data[10:10+size])
	return x, y, w, h, pngData, nil
}
