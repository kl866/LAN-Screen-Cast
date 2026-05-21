package protocol

import (
	"encoding/json"
	"testing"
)

func TestJoinMsgRoundTrip(t *testing.T) {
	msg := ControlMessage{Type: "join", ID: "PC-001"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ControlMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Type != "join" || decoded.ID != "PC-001" {
		t.Fatalf("round-trip failed: %+v", decoded)
	}
}

func TestQueuePosMsg(t *testing.T) {
	msg := ControlMessage{Type: "queue_pos", Pos: 3}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ControlMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Pos != 3 {
		t.Fatalf("expected pos=3, got %d", decoded.Pos)
	}
}

func TestActivateNextMsg(t *testing.T) {
	msg := ControlMessage{Type: "activate", Next: "PC-002"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ControlMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Next != "PC-002" {
		t.Fatalf("expected next=PC-002, got %s", decoded.Next)
	}
}
