package deferok

import "sync"

// Registry uses defer for both the read-lock and write-lock paths.
type Registry struct {
	mu   sync.RWMutex
	data map[string]int
}

func (r *Registry) Get(key string) (int, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	v, ok := r.data[key]
	if !ok {
		return 0, false
	}
	return v, true
}

func (r *Registry) Set(key string, val int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[key] = val
}
