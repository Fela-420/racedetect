// Source: CockroachDB issue/PR #584, as extracted and simplified by
// GoBench ("GoBench: A Benchmark Suite of Real-World Go Concurrency
// Bugs", Yuan et al., CGO 2021, doi:10.1109/CGO51591.2021.9370317),
// and vendored verbatim (with only cosmetic package-name tweaks) by
// the Go project itself in src/runtime/testdata/testgoroutineleakprofile/goker/cockroach584.go
// to validate the Go runtime's own experimental goroutine leak profiler.
//
// Upstream file is MIT-licensed per its own header. Pulled into this
// corpus unmodified (bootstrap/manage/struct) as corpus/real-world
// validation for racedetect Stage 2, per docs/SOURCES.md.
//
// Bug, per the GoKer README: "Missing call to mu.Unlock() before the
// break in the loop." Both bootstrap() and manage() have the identical
// defect -- a classic lock-without-unlock on an early-exit path,
// exactly the class of bug Stage 2's CFG reachability check targets.
//
// Original: https://go.googlesource.com/go/+/HEAD/src/runtime/testdata/testgoroutineleakprofile/goker/cockroach584.go

package realworld

import "sync"

type gossipCockroach584 struct {
	mu     sync.Mutex // L1
	closed bool
}

func (g *gossipCockroach584) bootstrap() {
	for {
		g.mu.Lock()
		if g.closed {
			// Missing g.mu.Unlock
			break
		}
		g.mu.Unlock()
	}
}

func (g *gossipCockroach584) manage() {
	for {
		g.mu.Lock()
		if g.closed {
			// Missing g.mu.Unlock
			break
		}
		g.mu.Unlock()
	}
}
