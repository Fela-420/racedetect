package main

import "sync"

// BuildHandlers: under go1.21 semantics, every closure captures the same
// `name` variable, so all registered handlers end up referencing the
// last value in `names` instead of their own.
func BuildHandlers(names []string) map[string]func() string {
	handlers := make(map[string]func() string)
	var mu sync.Mutex

	for _, name := range names {
		go func() {
			mu.Lock()
			handlers[name] = func() string { return name } // BUG under go1.21
			mu.Unlock()
		}()
	}
	return handlers
}
