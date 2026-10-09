package missingunlock

import "sync"

// Cache has a bug: GetOrError locks, then returns early on the error
// path WITHOUT calling Unlock. Only the happy path releases the lock.
type Cache struct {
	mu   sync.Mutex
	data map[string]string
}

func (c *Cache) GetOrError(key string) (string, error) {
	c.mu.Lock()
	val, ok := c.data[key]
	if !ok {
		// BUG: returns without unlocking
		return "", errNotFound
	}
	c.mu.Unlock()
	return val, nil
}

var errNotFound = &notFoundErr{}

type notFoundErr struct{}

func (e *notFoundErr) Error() string { return "key not found" }
