package protocol

import (
	"bytes"
	"errors"
	"testing"
)

func TestEncodeDecodeVideoBlock(t *testing.T) {
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x01, 0x02, 0x03} // fake PNG header
	encoded := EncodeVideoBlock(100, 200, 64, 32, pngData)

	x, y, w, h, decoded, err := DecodeVideoBlock(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if x != 100 || y != 200 || w != 64 || h != 32 {
		t.Fatalf("rect mismatch: got (%d,%d,%d,%d)", x, y, w, h)
	}
	if !bytes.Equal(decoded, pngData) {
		t.Fatal("PNG data mismatch")
	}
}

func TestDecodeVideoBlockTooShort(t *testing.T) {
	_, _, _, _, _, err := DecodeVideoBlock([]byte{0, 0, 0, 0})
	if !errors.Is(err, ErrVideoBlockTooShort) {
		t.Fatal("expected ErrVideoBlockTooShort")
	}
}
