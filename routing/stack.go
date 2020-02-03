package routing

import (
	"errors"
	"github.com/mindstand/go-bolt/log"
	"sync"
)

type connectionStack struct {
	lock                sync.Mutex // you don't have to do this if you don't want thread safety
	s                   []*connectionPoolWrapper
	connRemovedDelegate chan *connectionPoolWrapper
}

func newStack() *connectionStack {
	return &connectionStack{sync.Mutex{}, make([]*connectionPoolWrapper, 0), make(chan *connectionPoolWrapper)}
}

func (s *connectionStack) delegateFunc(c *connectionPoolWrapper) {
	if s.connRemovedDelegate != nil && c != nil {
		s.connRemovedDelegate <- c
	}
}

func (s *connectionStack) Push(v *connectionPoolWrapper) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if v == nil {
		return errors.New("tried to push nil obj")
	}

	if v.markForDeletion {
		log.Trace("removing marked object")
		s.delegateFunc(v)
		return nil
	}

	v.borrowed = false

	s.s = append(s.s, v)
	return nil
}

func (s *connectionStack) Pop() (*connectionPoolWrapper, error) {
	s.lock.Lock()

	l := len(s.s)
	if l == 0 {
		s.lock.Unlock()
		return nil, errors.New("empty")
	}

	res := s.s[l-1]
	s.s = s.s[:l-1]

	// check if we should get rid of it
	if res == nil || res.markForDeletion {
		s.delegateFunc(res)
		s.lock.Unlock()
		return s.Pop()
	}

	// check if its working
	if !res.Connection.ValidateOpen() {
		// connection is screwed up, just dump it
		res.markForDeletion = true
		s.delegateFunc(res)
		s.lock.Unlock()
		return s.Pop()
	}

	res.numBorrows++
	res.borrowed = true

	s.lock.Unlock()
	return res, nil
}

func (s *connectionStack) PruneMarkedConnections() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for i := 0; i < len(s.s); i++ {
		s.delegateFunc(s.s[i])
		copy(s.s[i:], s.s[i+1:])
		s.s[len(s.s)-1] = nil
		s.s = s.s[:len(s.s)-1]
		i--
	}
}
