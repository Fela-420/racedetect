#!/usr/bin/env bash
# Double-lock mutation test. Usage: scripts/mutate_dup_lock.sh <dir> [<dir>...]
# For each plain "x.Lock()" or "x.RLock()" statement, duplicate that line in a
# temp copy so the same lock is taken twice. The detector should report a violation.
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
    sed "${ln}p" "$f" > "$TMP/m.go"
    if "$DET" "$TMP/m.go" 2>/dev/null | grep -qE "re-acquires|re-enters"; then
      caught=$((caught+1))
    else
      missed+=("$f:$ln")
    fi
  done < <(grep -nE '^\s*[A-Za-z0-9_.*]+\.R?Lock\(\)\s*$' "$f" 2>/dev/null | cut -d: -f1)
done < <(find "$@" -type f -name '*.go' ! -name '*_test.go' -print0)

if [ "$total" -eq 0 ]; then
  echo "ERROR: 0 mutants generated."
  exit 2
fi

echo "double-lock mutants: $total | caught: $caught | missed: $((total-caught))"
for m in "${missed[@]+"${missed[@]}"}"; do echo "  MISSED $m"; done
