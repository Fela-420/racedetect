package deferok

import "sync"

// Ledger uses the standard defer-unlock-immediately-after-lock idiom.
// Every exit path (including panics) releases the lock via defer.
type Ledger struct {
	mu      sync.Mutex
	entries []int
}

func (l *Ledger) Add(n int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if n < 0 {
		return // covered by defer, must NOT flag
	}
	l.entries = append(l.entries, n)
}

func (l *Ledger) Sum() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	total := 0
	for _, e := range l.entries {
		if e == 0 {
			continue
		}
		total += e
	}
	return total
}
