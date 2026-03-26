#!/usr/bin/env bash
set -euo pipefail

OLD_VER="${1:-v0.5.9}"
NEW_VER="${2:-v0.5.10}"
ROOT="${3:-.}"

# Escape for sed (dots in v0.5.9)
OLD_ESC="${OLD_VER//./\\.}"

cd "$ROOT"

mapfile -d '' files < <(find . \( -path './.git' -o -path './vendor' \) -prune -o -name go.mod -print0)

for f in "${files[@]}"; do
  if grep -q 'github\.com/lariv-in/lago' "$f" && grep -qF "$OLD_VER" "$f"; then
    sed -i "/github\\.com\\/lariv-in\\/lago/s/${OLD_ESC}/${NEW_VER}/g" "$f"
    echo "updated: $f"
  fi
done

echo "Done. Run 'go mod tidy' in each touched module (or from repo root if you use a workspace)."