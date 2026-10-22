#!/usr/bin/env bash
# Mutation test for the detector. For each deferred Unlock in each file,
# delete that one line in a temp copy and check the detector reports a
# violation. A miss means the detector did not notice a planted missing unlock.
set -u
DET=${DET:-/tmp/racedetect}
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

caught=0; total=0
missed=()
for f in "$@"; do
  mapfile -t lines < <(grep -n 'defer .*Unlock()' "$f" | cut -d: -f1)
  for ln in "${lines[@]}"; do
    total=$((total+1))
    sed "${ln}d" "$f" > "$TMP/m.go"
    if "$DET" "$TMP/m.go" | grep -q VIOLATION; then
      caught=$((caught+1))
    else
      missed+=("$f:$ln")
    fi
  done
done

echo "mutants: $total | caught: $caught | missed: $((total-caught))"
for m in "${missed[@]}"; do echo "  MISSED $m"; done
