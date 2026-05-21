package session

import (
	"testing"
)

type mockSender struct {
	id     string
	active bool
}

func (m *mockSender) ID() string       { return m.id }
func (m *mockSender) Activate()        { m.active = true }
func (m *mockSender) Deactivate()      { m.active = false }
func (m *mockSender) IsActive() bool   { return m.active }

func TestEnqueueFirstBecomesActive(t *testing.T) {
	q := NewQueueManager()
	id := q.Enqueue(&mockSender{id: "A"})
	if id != "A" {
		t.Fatalf("expected A, got %s", id)
	}
	if q.Active() == nil {
		t.Fatal("first sender should be auto-activated")
	}
	if q.Active().ID() != "A" {
		t.Fatalf("expected A active, got %s", q.Active().ID())
	}
}

func TestSecondEnqueued(t *testing.T) {
	q := NewQueueManager()
	q.Enqueue(&mockSender{id: "A"})
	id := q.Enqueue(&mockSender{id: "B"})
	if id != "B" {
		t.Fatalf("expected B, got %s", id)
	}
	if q.Len() != 2 {
		t.Fatalf("expected len 2, got %d", q.Len())
	}
}

func TestRemoveActiveActivatesNext(t *testing.T) {
	q := NewQueueManager()
	q.Enqueue(&mockSender{id: "A"})
	q.Enqueue(&mockSender{id: "B"})
	q.Remove("A")
	if q.Active() == nil {
		t.Fatal("next sender should be activated")
	}
	if q.Active().ID() != "B" {
		t.Fatalf("expected B active after A removed, got %s", q.Active().ID())
	}
	if q.Len() != 1 {
		t.Fatalf("expected len 1 after removal, got %d", q.Len())
	}
}

func TestRemoveQueuedDoesNotAffectActive(t *testing.T) {
	q := NewQueueManager()
	q.Enqueue(&mockSender{id: "A"})
	q.Enqueue(&mockSender{id: "B"})
	q.Remove("B")
	if q.Active().ID() != "A" {
		t.Fatal("active should not change when queued sender removed")
	}
	if q.Len() != 1 {
		t.Fatalf("expected len 1, got %d", q.Len())
	}
}

func TestRemoveOnlySender(t *testing.T) {
	q := NewQueueManager()
	q.Enqueue(&mockSender{id: "A"})
	q.Remove("A")
	if q.Active() != nil {
		t.Fatal("no active sender after removing the only one")
	}
	if q.Len() != 0 {
		t.Fatalf("expected len 0, got %d", q.Len())
	}
}

func TestList(t *testing.T) {
	q := NewQueueManager()
	q.Enqueue(&mockSender{id: "A"})
	q.Enqueue(&mockSender{id: "B"})
	list := q.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 in list, got %d", len(list))
	}
	if list[0] != "A" || list[1] != "B" {
		t.Fatalf("unexpected list order: %v", list)
	}
}
