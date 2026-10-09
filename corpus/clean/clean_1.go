package clean

import "sync"

// Counter demonstrates a correctly locked struct.
// Every path that acquires the lock also releases it.
type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) Increment() {
	c.mu.Lock()
	c.value++
	c.mu.Unlock()
}

func (c *Counter) IncrementConditional(skip bool) int {
	c.mu.Lock()
	if skip {
		c.mu.Unlock()
		return c.value
	}
	c.value++
	result := c.value
	c.mu.Unlock()
	return result
}
