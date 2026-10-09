// Source: grpc-go issue/PR #3017, via GoBench/GoKer, vendored by the
// Go project at src/runtime/testdata/testgoroutineleakprofile/goker/grpc3017.go
// MIT-licensed upstream. See cockroach584.go in this directory for
// full provenance notes.
//
// GoKer's README: "Line 65 is an execution path with a missing unlock."
// Notable for racedetect specifically: the Lock/Unlock pair here is
// NOT in a top-level method -- it's inside a closure passed to
// time.AfterFunc(). This is a real-world test of Stage 2's rule that
// every FuncLit (closure) gets analyzed as its own independent CFG
// unit, separate from its enclosing method. Reconstructed from the
// fetched fragment; surrounding cache-setup code omitted, the buggy
// closure body kept verbatim.
//
// Original: https://go.googlesource.com/go/+/HEAD/src/runtime/testdata/testgoroutineleakprofile/goker/grpc3017.go

package realworld

import (
	"runtime"
	"sync"
	"time"
)

type AddressGrpc3017 int
type SubConnGrpc3017 int

type subConnCacheEntryGrpc3017 struct {
	sc            SubConnGrpc3017
	cancel        func()
	abortDeleting bool
}

type lbCacheClientConnGrpc3017 struct {
	mu            sync.Mutex // L1
	timeout       time.Duration
	subConnCache  map[AddressGrpc3017]*subConnCacheEntryGrpc3017
	subConnToAddr map[SubConnGrpc3017]AddressGrpc3017
}

// RemoveSubConn: the real lock bug lives in the closure scheduled via
// time.AfterFunc below -- on the entry.abortDeleting branch, it
// returns without calling ccc.mu.Unlock().
func (ccc *lbCacheClientConnGrpc3017) RemoveSubConn(sc SubConnGrpc3017, addr AddressGrpc3017) {
	entry := &subConnCacheEntryGrpc3017{sc: sc}
	ccc.subConnCache[addr] = entry

	timer := time.AfterFunc(ccc.timeout, func() { // G3 -- separate CFG unit
		runtime.Gosched()
		ccc.mu.Lock() // L1
		if entry.abortDeleting {
			return // Missing unlock
		}
		delete(ccc.subConnToAddr, sc)
		delete(ccc.subConnCache, addr)
		ccc.mu.Unlock()
	})

	entry.cancel = func() {
		if !timer.Stop() {
			entry.abortDeleting = true
		}
	}
}
