package missingunlock

import "sync"

// Store has a bug in its RWMutex usage: the write-lock branch releases
// correctly, but the read-lock branch has an early return that skips
// RUnlock when validate fails.
type Store struct {
	mu     sync.RWMutex
	values map[string]int
}

func (s *Store) Read(key string, validate func(int) bool) (int, bool) {
	s.mu.RLock()
	v, ok := s.values[key]
	if !ok {
		s.mu.RUnlock()
		return 0, false
	}
	if !validate(v) {
		// BUG: returns without RUnlock
		return 0, false
	}
	s.mu.RUnlock()
	return v, true
}

func (s *Store) Write(key string, val int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = val
}
