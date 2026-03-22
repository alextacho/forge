#!/usr/bin/env bash
# unlink.sh — remove all symlinks from .claude/skills/, .claude/agents/, .claude/commands/
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

REMOVED=0

remove_section() {
  local section="$1"
  local claude_dir="$ROOT/.claude/${section}"

  [[ -d "$claude_dir" ]] || return 0

  while IFS= read -r -d '' link; do
    if [[ -L "$link" ]]; then
      rm "$link"
      echo "  removed: .claude/${section}/$(basename "$link")"
      ((REMOVED++)) || true
    fi
  done < <(find "$claude_dir" -maxdepth 2 -type l -print0)
}

echo "=== forge unlink ==="
echo ""

remove_section "skills"
remove_section "agents"
remove_section "commands"

echo ""
echo "Done: $REMOVED symlink(s) removed. Source files untouched."
