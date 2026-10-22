package cache

import "sync"

type cache struct{ mu sync.Mutex }

// Deadlocks: the second Lock is taken while the first is still held.
// The defer does not help, because it only runs at return.
func (c *cache) Bad() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mu.Lock()
	c.mu.Unlock()
}
