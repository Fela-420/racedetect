package lru

type cache struct{ mu sync.Mutex }

func (c *cache) Remove(key string) bool {
	var evicted []string
	c.mu.Lock()
	defer func() {
		c.mu.Unlock()
		_ = evicted
	}()
	if key == "" {
		return false
	}
	return true
}
