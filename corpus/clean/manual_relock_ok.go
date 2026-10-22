package cache

import (
	"sync"
	"time"
)

type cache struct{ mu sync.Mutex }

func (c *cache) wait(done chan struct{}, d time.Duration) {
	c.mu.Lock()
	if d > 0 {
		c.mu.Unlock()
		select {
		case <-time.After(d):
		case <-done:
			return
		}
		c.mu.Lock()
		select {
		case <-done:
			c.mu.Unlock()
			return
		default:
		}
	}
	c.mu.Unlock()
}
