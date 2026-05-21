package codec

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	// 32x32 RGBA test image — red-green gradient
	w, h := 32, 32
	pixels := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			off := (y*w + x) * 4
			pixels[off] = byte(x * 8) // R
			pixels[off+1] = byte(y * 8) // G
			pixels[off+2] = 0 // B
			pixels[off+3] = 255 // A
		}
	}

	encoded, err := EncodeRGBA(pixels, w, h)
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) == 0 {
		t.Fatal("empty encoded data")
	}

	decoded, err := DecodeToRGBA(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.W != w || decoded.H != h {
		t.Fatalf("size mismatch: got %dx%d, want %dx%d", decoded.W, decoded.H, w, h)
	}
	if len(decoded.Pixels) != w*h*4 {
		t.Fatalf("pixel buffer size mismatch: got %d", len(decoded.Pixels))
	}
	if !bytes.Equal(pixels, decoded.Pixels) {
		t.Fatal("pixel data mismatch after round-trip")
	}
}

func TestDecodeInvalidData(t *testing.T) {
	_, err := DecodeToRGBA([]byte{0x00, 0x01, 0x02})
	if err == nil {
		t.Fatal("expected error for invalid PNG data")
	}
}
