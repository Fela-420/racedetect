#!/usr/bin/env bash
# Mutation test for the detector. Usage: scripts/mutate.sh <dir> [<dir>...]
# For each "defer ... Unlock()" line in each non-test .go file, delete that
# one line in a temp copy and check the detector reports a violation.
# A miss means a planted missing-unlock was not noticed.
set -u
DET=${DET:-/tmp/racedetect}
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

if [ $# -eq 0 ]; then echo "usage: $0 <dir> [<dir>...]"; exit 2; fi

total=0; caught=0; missed=()
while IFS= read -r -d '' f; do
  while IFS= read -r ln; do
    [ -z "$ln" ] && continue
    total=$((total+1))
    sed "${ln}d" "$f" > "$TMP/m.go"
    if "$DET" "$TMP/m.go" 2>/dev/null | grep -q VIOLATION; then
      caught=$((caught+1))
    else
      missed+=("$f:$ln")
    fi
  done < <(grep -n 'defer .*Unlock()' "$f" 2>/dev/null | cut -d: -f1)
done < <(find "$@" -type f -name '*.go' ! -name '*_test.go' -print0)

if [ "$total" -eq 0 ]; then
  echo "ERROR: 0 mutants generated. Check the paths passed in, the DET binary, and the grep pattern."
  exit 2
fi

echo "mutants: $total | caught: $caught | missed: $((total-caught))"
for m in "${missed[@]+"${missed[@]}"}"; do echo "  MISSED $m"; done
