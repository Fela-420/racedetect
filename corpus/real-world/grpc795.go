// Source: grpc-go issue/PR #795, via GoBench/GoKer, vendored by the
// Go project in src/runtime/testdata/testgoroutineleakprofile/goker/grpc795.go
// MIT-licensed upstream. See cockroach584.go in this directory for
// full provenance notes.
//
// GoKer's README describes this as "Line 20 is an execution path with
// a missing unlock" -- but read GracefulStop() closely: the real root
// cause is a DOUBLE LOCK. When s.drain is already true, the function
// calls s.mu.Lock() a SECOND time while still holding it from line 24.
// Since sync.Mutex is not reentrant, that second Lock() blocks
// forever -- it's a self-deadlock, not a "forgot to write Unlock" bug.
//
// racedetect's Stage 2 only checks lock/unlock path balance, with no
// concept of double-locking. It will likely still flag this function,
// but for the technically-narrow reason that the `return` on that
// branch never reaches the real Unlock() at line 20 of the original --
// not because it understands the deadlock. Correct catch, wrong full
// diagnosis. This is exactly the gap goconcurrencylint's GCL1011
// double-lock check exists for; see docs/SOURCES.md. Kept in the
// corpus specifically to document this limitation, not just to pad
// the pass count.
//
// Original: https://go.googlesource.com/go/+/HEAD/src/runtime/testdata/testgoroutineleakprofile/goker/grpc795.go

package realworld

import "sync"

type ServerGrpc795 struct {
	mu    sync.Mutex
	drain bool
}

func (s *ServerGrpc795) GracefulStop() {
	s.mu.Lock()
	if s.drain {
		s.mu.Lock() // BUG: double lock -> self-deadlock, not a missing-unlock in the usual sense
		return
	}
	s.drain = true
	s.mu.Unlock()
}

func (s *ServerGrpc795) Serve() {
	s.mu.Lock()
	s.mu.Unlock()
}
