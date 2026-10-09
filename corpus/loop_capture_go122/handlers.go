package main

import "sync"

// BuildHandlers: syntactically identical to the go1.21 buggy version,
// but safe here because `name` is scoped per-iteration under go1.22+.
func BuildHandlers(names []string) map[string]func() string {
	handlers := make(map[string]func() string)
	var mu sync.Mutex

	for _, name := range names {
		go func() {
			mu.Lock()
			handlers[name] = func() string { return name } // safe under go1.22+
			mu.Unlock()
		}()
	}
	return handlers
}
