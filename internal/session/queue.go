package session

import "sync"

type QueueManager struct {
	mu      sync.Mutex
	senders []Sender
}

func NewQueueManager() *QueueManager {
	return &QueueManager{}
}

func (q *QueueManager) Enqueue(s Sender) string {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.senders = append(q.senders, s)
	if len(q.senders) == 1 {
		s.Activate()
	}
	return s.ID()
}

func (q *QueueManager) Remove(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	wasActive := false
	for i, s := range q.senders {
		if s.ID() == id {
			wasActive = s.IsActive()
			s.Deactivate()
			q.senders = append(q.senders[:i], q.senders[i+1:]...)
			break
		}
	}
	if wasActive && len(q.senders) > 0 {
		q.senders[0].Activate()
	}
}

func (q *QueueManager) Active() Sender {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, s := range q.senders {
		if s.IsActive() {
			return s
		}
	}
	return nil
}

func (q *QueueManager) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.senders)
}

func (q *QueueManager) List() []string {
	q.mu.Lock()
	defer q.mu.Unlock()
	ids := make([]string, len(q.senders))
	for i, s := range q.senders {
		ids[i] = s.ID()
	}
	return ids
}

func (q *QueueManager) ActivateNext() Sender {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, s := range q.senders {
		if s.IsActive() {
			s.Deactivate()
			break
		}
	}
	if len(q.senders) > 0 {
		q.senders[0].Activate()
		return q.senders[0]
	}
	return nil
}
