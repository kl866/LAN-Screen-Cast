package protocol

import (
	"bytes"
	"errors"
	"testing"
)

func TestEncodeDecodeControlMessage(t *testing.T) {
	payload := []byte(`{"type":"join","id":"PC-001"}`)
	encoded := EncodeMessage(TypeControl, payload)

	magic := string(encoded[:4])
	if magic != "PROJ" {
		t.Fatalf("bad magic: %s", magic)
	}

	typ, decoded, err := DecodeMessage(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if typ != TypeControl {
		t.Fatalf("expected type 0x01, got 0x%02x", typ)
	}
	if !bytes.Equal(decoded, payload) {
		t.Fatalf("payload mismatch: got %s", decoded)
	}
}

func TestEncodeDecodeVideoMessage(t *testing.T) {
	payload := make([]byte, 1024)
	for i := range payload {
		payload[i] = byte(i % 256)
	}
	encoded := EncodeMessage(TypeVideo, payload)

	typ, decoded, err := DecodeMessage(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if typ != TypeVideo {
		t.Fatalf("expected type 0x02, got 0x%02x", typ)
	}
	if !bytes.Equal(decoded, payload) {
		t.Fatal("payload mismatch")
	}
}

func TestDecodeIncompleteHeader(t *testing.T) {
	_, _, err := DecodeMessage([]byte{0x50, 0x52})
	if !errors.Is(err, ErrIncompleteHeader) {
		t.Fatalf("expected ErrIncompleteHeader, got %v", err)
	}
}

func TestDecodeBadMagic(t *testing.T) {
	data := EncodeMessage(TypeControl, []byte("test"))
	data[0] = 0x00
	_, _, err := DecodeMessage(data)
	if !errors.Is(err, ErrBadMagic) {
		t.Fatalf("expected ErrBadMagic, got %v", err)
	}
}

func TestDecodeIncompletePayload(t *testing.T) {
	data := EncodeMessage(TypeControl, []byte("hello"))
	// Truncate the payload so only part of it remains.
	data = data[:HeaderSize+2] // only 2 bytes of "hello" present
	_, _, err := DecodeMessage(data)
	if !errors.Is(err, ErrIncompletePayload) {
		t.Fatalf("expected ErrIncompletePayload, got %v", err)
	}
}

func TestEmptyPayloadRoundTrip(t *testing.T) {
	encoded := EncodeMessage(TypeControl, nil)
	typ, decoded, err := DecodeMessage(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if typ != TypeControl {
		t.Fatalf("expected type 0x01, got 0x%02x", typ)
	}
	if len(decoded) != 0 {
		t.Fatalf("expected empty payload, got %d bytes", len(decoded))
	}
}
