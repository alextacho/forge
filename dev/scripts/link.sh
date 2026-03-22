#!/usr/bin/env bash
# link.sh — create .claude/ symlinks for Claude Code discovery
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PLUGIN_JSON="$ROOT/.claude-plugin/plugin.json"

# ── Dependency check ──────────────────────────────────────────────────────────
if ! command -v jq &>/dev/null; then
  echo "ERROR: jq is required. Install with: brew install jq"
  exit 1
fi

if [[ ! -f "$PLUGIN_JSON" ]]; then
  echo "ERROR: plugin.json not found at $PLUGIN_JSON"
  exit 1
fi

PLUGIN_NAME=$(jq -r '.name' "$PLUGIN_JSON")
LINKED=0
SKIPPED=0
WARNED=0

link_files() {
  local source_dir="$1"
  local claude_dir="$2"

  [[ -d "$source_dir" ]] || return 0
  mkdir -p "$claude_dir"

  while IFS= read -r -d '' file; do
    fname=$(basename "$file")
    link="$claude_dir/$fname"
    # Relative path from .claude/<section>/ back to source
    target=$(python3 -c "import os; print(os.path.relpath('$file', '$claude_dir'))" 2>/dev/null || \
             ruby -e "require 'pathname'; print Pathname.new('$file').relative_path_from(Pathname.new('$claude_dir'))")

    if [[ -L "$link" ]]; then
      existing=$(readlink "$link")
      if [[ "$existing" == "$target" ]]; then
        echo "  already linked: $(basename "$claude_dir")/$(basename "$link")"
        ((SKIPPED++)) || true
      else
        echo "  WARN: $(basename "$claude_dir")/$fname points to '$existing' (expected '$target') — skipping"
        ((WARNED++)) || true
      fi
    else
      ln -s "$target" "$link"
      echo "  linked: $(basename "$claude_dir")/$fname"
      ((LINKED++)) || true
    fi
  done < <(find "$source_dir" -maxdepth 1 -name "*.md" -type f -print0)
}

echo "=== forge link: $PLUGIN_NAME ==="
echo ""

# Link plugin skills, agents, commands
SKILLS_DIR=$(jq -r '.skills // "skills/"' "$PLUGIN_JSON")
AGENTS_DIR=$(jq -r '.agents // "agents/"' "$PLUGIN_JSON")
COMMANDS_DIR=$(jq -r '.commands // "commands/"' "$PLUGIN_JSON")

link_files "$ROOT/${SKILLS_DIR%/}"   "$ROOT/.claude/skills"
link_files "$ROOT/${AGENTS_DIR%/}"   "$ROOT/.claude/agents"
link_files "$ROOT/${COMMANDS_DIR%/}" "$ROOT/.claude/commands/$PLUGIN_NAME"

# Link forge dev commands
FORGE_COMMANDS="$ROOT/commands/forge"
if [[ -d "$FORGE_COMMANDS" ]]; then
  echo ""
  echo "  [forge dev commands]"
  link_files "$FORGE_COMMANDS" "$ROOT/.claude/commands/forge"
fi

echo ""
echo "Done: $LINKED linked, $SKIPPED already linked, $WARNED warning(s)"
