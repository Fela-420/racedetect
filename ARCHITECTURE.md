# racedetect — Architecture

## Positioning (read docs/SOURCES.md first)

Pure static lock-balance checking is already well covered by
goconcurrencylint. racedetect's reason to exist is the pipeline nothing
else combines:

    static flag  -->  automated dynamic confirmation  -->  scoped ingestion

A static hit alone is a lead, not a finding. This tool's job is to turn
leads into confirmed, reproducible bugs automatically, at a scale a
human reviewing GitHub by hand can't match — then stop, because
"thousands of repos scanned" is not a credible claim (see roadmap
critique on GitHub Code Search rate limits).

## Pipeline

| Stage | Component | Status |
|---|---|---|
| 0 | Ground-truth corpus (synthetic + real-world extracted) | synthetic done; real-world pending |
| 1 | AST parser scaffold | done |
| 2 | CFG-based static reachability (lock-without-unlock) | done, verified against corpus |
| 2b | Defer-before-lock, goroutine-held-lock-deadlock rules | not started |
| 3 | Dynamic verifier: generates a `go test -race` harness per static hit | not started |
| 4 | Docker sandbox for untrusted repo execution | not started |
| 5 | Scoped GitHub ingestion (curated repo list, not mass crawl) | not started |

## Validation strategy

Every stage is checked against two corpora, not one:
- `corpus/` — hand-written, exact bug/no-bug pairs, fast iteration.
- `corpus/real-world/` — extracted from system-pclub/go-concurrency-bugs
  and GoBench, real confirmed bugs with real fix commits. This is what
  justifies any claim stronger than "passes my own tests."

## Why dynamic confirmation matters

A static `[VIOLATION]` from Stage 2 is a claim about reachability in the
AST, not proof the bug fires at runtime. Stage 3 closes that gap: for
every static hit, auto-generate a minimal `_test.go` harness that
spawns concurrent callers of the flagged function, run it under
`go test -race -count=50` inside the Stage 4 sandbox, and only report a
finding as CONFIRMED if the race detector actually triggers. Unconfirmed
static hits are kept as LEADS, not reported as bugs — this is the
precision discipline that keeps this tool useful instead of noisy.
