package kitevent

import "sync"

// Queue .
type Queue struct {
	queue []*KitEvent
	pos   int
	lock  sync.RWMutex
}

// NewQueue .
func NewQueue(cap int) *Queue {
	return &Queue{
		queue: make([]*KitEvent, cap),
	}
}

// Push .
func (q *Queue) Push(e *KitEvent) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.queue[q.pos] = e
	q.pos++
	if q.pos >= len(q.queue) {
		q.pos = 0
	}
}

// Dump .
func (q *Queue) Dump() []*KitEvent {
	results := make([]*KitEvent, 0, len(q.queue))
	q.lock.RLock()
	defer q.lock.RUnlock()
	pos := q.pos
	for i := 0; i < len(q.queue); i++ {
		pos--
		if pos < 0 {
			pos = len(q.queue) - 1
		}

		e := q.queue[pos]
		if e == nil {
			return results
		}

		results = append(results, e)
	}

	return results
}
