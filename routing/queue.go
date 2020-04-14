package routing

import "sync"

type Queue struct {
	items []*connectionPoolWrapper
	lock  sync.RWMutex
}

// New creates a new OldQueue
func NewQueue() *Queue {
	return &Queue{
		items: []*connectionPoolWrapper{},
		lock: sync.RWMutex{},
	}
}

func NewQueueFromSlice(q ...*connectionPoolWrapper) *Queue{
	return &Queue{
		items: q,
		lock: sync.RWMutex{},
	}
}

// Enqueue adds an Item to the end of the queue
func (s *Queue) Enqueue(t *connectionPoolWrapper) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.items = append(s.items, t)
}

// Dequeue removes an Item from the start of the queue
func (s *Queue) Dequeue() *connectionPoolWrapper {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.items) == 0 {
		return nil
	}

	item := s.items[0]
	s.items = s.items[1:len(s.items)]
	return item
}

// Front returns the item next in the queue, without removing it
func (s *Queue) Front() *connectionPoolWrapper {
	s.lock.RLock()
	item := s.items[0]
	s.lock.RUnlock()
	return item
}

// IsEmpty returns true if the queue is empty
func (s *Queue) IsEmpty() bool {
	return len(s.items) == 0
}

// Size returns the number of Items in the queue
func (s *Queue) Size() int {
	return len(s.items)
}
