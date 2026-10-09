# Start Here — What racedetect Is, In Plain Language

If you're lost, read this file only. Everything else can wait.

## The one-sentence version

racedetect is a tool that reads Go programming code and finds one
specific kind of bug: a program "locks a door" (to stop two parts of
the program from touching the same data at once) and then forgets to
"unlock" it on some path through the code — which makes the program
freeze forever.

## Why this bug matters

Imagine a bathroom with one key. Whoever has the key locks the door
while inside. If someone takes the key in, uses the bathroom, but
forgets to put the key back under some circumstances (say, if they
get a phone call and leave early) — everyone else is now locked out
forever, with no way to get the key back. That's the bug, in code form.
It's called a "concurrency bug" and it's genuinely hard to spot by eye
because it only shows up on certain paths through the code, not every time.

## What we've actually built so far (all of it, no jargon)

1. **A reader.** A small Go program that can open any other Go file and
   understand its structure — which functions exist, where the "lock"
   and "unlock" calls are.

2. **A checker.** This is the real tool. For every "lock," it traces
   every possible route the code could take afterward, and checks: does
   every single route eventually hit "unlock"? If even one route
   doesn't, that's a real bug, and it tells you the exact line number.

3. **A test folder (`corpus/`).** Files we wrote on purpose — some with
   the bug, some without — so we could check the checker gives the
   right answer every time before trusting it on anything real.

4. **Four real bugs from real, famous software** (CockroachDB, gRPC,
   Docker/Moby) — pulled from an academic research dataset, not made up
   — to prove the checker works on code we didn't write ourselves. It
   caught 2 cleanly, caught 1 by half-accident, and **missed 1
   completely** — which actually taught us something important about
   what the checker still can't do (see below).

## Where this came from — the bigger picture

Before building more, we stopped and researched: does a tool like this
already exist? Yes, partly — a small open-source project called
`goconcurrencylint` already checks for this bug and more. So the plan
split in two:
- **Give back:** share our hand-written test files with that project.
- **Build something new:** combine this kind of checking with actually
  *running* the code to prove a bug is real, then eventually scanning
  real GitHub projects for bugs — something no existing tool does.

## What "missed a bug" taught us — UPDATE: now fixed

One of the four real bugs slipped past the checker (`moby7559.go`). Not
a failure — figuring out *why* it was missed told us exactly what to
build next, and it's now built and verified:

**The fix:** the checker used to only ask "does an Unlock() exist
somewhere down the line?" Now it also asks "does the code grab the
SAME lock a second time before releasing it first?" — because that
freezes the program immediately, before it ever gets the chance to
reach that later Unlock(). Both real bugs that exposed this gap
(`grpc795.go` and `moby7559.go`) are now caught correctly, with clear
messages, and nothing that used to work correctly broke in the process
(checked against all 10 hand-written test files plus all 4 real-world
bugs — every single one still gives the right answer).

## What happens next, your choice

Nothing moves forward until you say so. The next real steps are:
1. Get everything we've built onto your actual computer and a real
   GitHub repo (nothing is backed up anywhere right now).
2. Then, whenever you're ready: build the "double-lock" rule the
   missed bug pointed us toward.

## If you forget everything else

You are not behind. You have a working tool that already caught real
bugs in real, famous software, on day two. That's further than most
people get.
