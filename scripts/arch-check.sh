#!/bin/sh
set -e
FAIL=0

check() {
  local layer="$1" dir="$2"; shift 2
  for pkg in "$@"; do
    matches=$(grep -rn --include="*.go" "\"$pkg" "$dir" 2>/dev/null || true)
    if [ -n "$matches" ]; then
      echo "VIOLATION [$layer]: '$pkg' found in $dir/"
      echo "$matches"
      FAIL=1
    fi
  done
}

check "domain"  "internal/domain"  \
  "github.com/dev-au/CodeStream/internal/usecase" \
  "github.com/dev-au/CodeStream/internal/adapter" \
  "github.com/dev-au/CodeStream/internal/infrastructure"

check "usecase" "internal/usecase" \
  "github.com/dev-au/CodeStream/internal/adapter" \
  "github.com/dev-au/CodeStream/internal/infrastructure"

check "adapter" "internal/adapter" \
  "github.com/dev-au/CodeStream/internal/infrastructure"

[ $FAIL -eq 0 ] && echo "✓ Architecture check passed."
exit $FAIL