# Prior Art & Research Log

Compiled before Stage 3, to avoid building something that already exists
and to ground the ground-truth corpus in real, not just synthetic, bugs.

## Existing static analysis tools (direct overlap)

- **goconcurrencylint** — github.com/sanbricio/goconcurrencylint
  Active, MIT-licensed, built on `golang.org/x/tools/go/analysis` (the
  standard framework). 30+ checks across sync.Mutex, RWMutex, WaitGroup,
  Once, Cond, Pool, and channels. Includes `GCL1001` (lock-without-unlock,
  same rule as our Stage 2), plus rules we don't cover: `GCL1003`
  (defer-before-lock), `GCL1009` (goroutine-started-while-locked deadlock),
  `GCL1012` (lock-order cycles across functions). Ships a golangci-lint
  module plugin, ARCHITECTURE.md, per-check docs, CI.
  ~7 stars / 2 forks as of Oct 2026 — not dominant, actively seeking
  contributions, explicitly requests "comparisons against overlapping
  analyzers" and extra testdata fixtures.

- **GCatch** (academic, USENIX/ASPLOS, UC Davis) — github.com/system-pclub/GCatch
  State-of-the-art static detector. Path-sensitive, inter-procedural,
  models channels + mutexes via a constraint system. Evaluated on 21
  real projects (Docker, Kubernetes, gRPC-go): found 149 blocking
  misuse-of-channel bugs + 119 traditional concurrency bugs.
  Companion tool GFix auto-generates patches for a subset.
  Bar-setting reference, not something to compete with solo.

- **go vet `-copylocks`** — stdlib. Different bug class entirely: flags a
  sync.Mutex/struct-embedding-a-mutex being copied by value (not a
  lock/unlock balance issue). Cheap, complementary rule to add later.

## Real-world bug corpora (for validating the corpus against real bugs, not just hand-written ones)

- **system-pclub/go-concurrency-bugs** (ASPLOS'19, 261 stars)
  171 real, confirmed concurrency bugs with actual before/after fix
  diffs, from Docker, Kubernetes, etcd, CockroachDB, gRPC-go, BoltDB.
  Organized into `blocking-bugs/` and `non-blocking-bugs/`. Each entry
  traces to a real GitHub issue/commit. This is the credibility upgrade
  path for Stage 0 — real bugs instead of only synthetic ones.

- **GoBench** (timmyyuan/gobench, IEEE 2021)
  82 real bugs + 103 simplified "bug kernels" extracted from 9 popular
  Go projects, purpose-built for evaluating detector tools like ours.

## Dynamic / hybrid detectors (adjacent, not competing with static-only tools)

- **GFuzz** — dynamic channel-bug detector via message reordering fuzzing.
- **GoAT** — combined static + dynamic concurrency debugging/visualization.
- **BINGO** — concurrency bug detection via binary analysis (no source needed).

None of the above combine static detection + automated dynamic
confirmation (`go test -race` in an isolated sandbox) + a GitHub-scale
ingestion pipeline in one tool. That combination is racedetect's actual
differentiation — see ARCHITECTURE.md.

## Real-world validation results (Stage 2, run against corpus/real-world/)

| File | Real bug | Stage 2 result (original) | Stage 2b result (after double-lock fix) |
|---|---|---|---|
| cockroach584.go | missing unlock on `break` (x2 funcs) | caught, correct, 2/2 | unchanged, still 2/2 |
| grpc3017.go | missing unlock inside closure | caught, correct | unchanged, still correct |
| grpc795.go | double-lock self-deadlock | caught, wrong reason (2 confusing messages) | **1 clean message, correctly labeled double-lock** |
| moby7559.go | double-lock across loop iterations via `continue` | **MISSED — false negative** | **now caught correctly** |

**Stage 2b fix (implemented):** the walk now asks "does this path hit
Unlock, or does it hit ANOTHER Lock/RLock on the same receiver, first?"
instead of only "does an Unlock exist somewhere downstream?" A
re-acquisition is reported as kind="double-lock" and the walk stops
there (the program deadlocks at that point in practice, so nothing
past it is reachable). A lock site that is itself the target of
someone else's double-lock finding has its own independent
missing-unlock report suppressed, to avoid two messages for one root
cause. Verified against the full synthetic corpus (zero regressions)
and all four real-world files (table above).

**Root cause of the moby7559.go miss:** Stage 2's rule is "does some
Unlock() exist reachable from this Lock()." It does not model that a
second Lock() on the same receiver, reached before any Unlock(), is
itself a deadlock — regardless of whether an Unlock() exists further
downstream. The loop's `continue` path reaches Lock() a second time
without passing through Unlock() first, deadlocking in practice before
the function-level Unlock() after the loop would ever execute. Stage 2
doesn't see that: it just finds *an* Unlock() reachable eventually and
calls the path covered.

**Stage 2b spec, derived directly from this evidence:** flag a
violation if any path reaches a second Lock() call on the same
receiver before passing through a matching Unlock() on that same path
-- independent of whether an Unlock() exists later in the function.
This closes the moby7559.go false negative and gives grpc795.go a
correct "double lock" diagnosis instead of a misleading
missing-unlock one.

## What this means for racedetect (decided Oct 2026)

Two parallel tracks:
1. Contribute the hand-written Stage 0 corpus to goconcurrencylint as
   testdata fixtures (their explicit ask) — fast, real, credible win.
2. Build the piece nothing above does: static flag -> automated dynamic
   confirmation -> scoped GitHub ingestion, validated against the real
   bug corpora above, not just synthetic test files.
