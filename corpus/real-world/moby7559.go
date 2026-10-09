// Source: moby/moby PR #7559, via GoBench/GoKer, vendored by the Go
// project at src/runtime/testdata/testgoroutineleakprofile/goker/moby7559.go
// MIT-licensed upstream, kept verbatim except the pprof-driven
// Moby7559() harness. See cockroach584.go in this directory for full
// provenance notes.
//
// Buggy version: 64579f51fcb439c36377c0068ccc9a007b368b5a
// Fix commit:    6cbb8e070d6c3a66bf48fbe5cbf689557eee23db
//
// GoKer's README: "Line 25 is missing a call to .Unlock." Worth
// looking closely at the shape of this one: Lock() is called on
// EVERY loop iteration, but Unlock() is called only ONCE, after the
// loop exits. If DialUDP succeeds on iteration 0 and i != 0 never
// triggers break, the loop re-enters and calls Lock() a second time
// while already holding it -- the same reentrant-mutex self-deadlock
// family as grpc795.go in this directory, just with the double-lock
// spread across loop iterations instead of two lines in a row.

package realworld

import (
	"net"
	"sync"
)

type UDPProxyMoby7559 struct {
	connTrackLock sync.Mutex
}

func (proxy *UDPProxyMoby7559) Run() {
	for i := 0; i < 2; i++ {
		proxy.connTrackLock.Lock()
		_, err := net.DialUDP("udp", nil, nil)
		if err != nil {
			continue
			/// Missing unlock here
		}
		if i == 0 {
			break
		}
	}
	proxy.connTrackLock.Unlock()
}
