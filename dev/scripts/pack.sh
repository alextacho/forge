#!/usr/bin/env bash
# pack.sh — validate and package plugin into dist/<name>-v<version>.plugin
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PLUGIN_JSON="$ROOT/.claude-plugin/plugin.json"
DIST_DIR="$ROOT/dist"

# ── Dependency check ──────────────────────────────────────────────────────────
if ! command -v jq &>/dev/null; then
  echo "ERROR: jq is required. Install with: brew install jq"
  exit 1
fi

# ── 1. Run validate.sh ────────────────────────────────────────────────────────
echo "=== Running validate.sh ==="
if ! bash "$SCRIPT_DIR/validate.sh"; then
  echo ""
  echo "ERROR: Validation failed. Fix all errors before packaging."
  exit 1
fi
echo ""

# ── 2. Check git is clean ─────────────────────────────────────────────────────
if ! git -C "$ROOT" diff --quiet || ! git -C "$ROOT" diff --cached --quiet; then
  echo "ERROR: Git working tree is not clean. Commit or stash changes before packaging."
  git -C "$ROOT" status --short
  exit 1
fi

# ── 3. Check status == published ─────────────────────────────────────────────
STATUS=$(jq -r '.status // "draft"' "$PLUGIN_JSON")
if [[ "$STATUS" != "published" ]]; then
  echo "ERROR: plugin.json status is '$STATUS'. Must be 'published' to package."
  echo "       Edit plugin.json, set \"status\": \"published\", commit, then re-run."
  exit 1
fi

# ── 4. Read name and version ──────────────────────────────────────────────────
NAME=$(jq -r '.name' "$PLUGIN_JSON")
VERSION=$(jq -r '.version' "$PLUGIN_JSON")
OUTPUT="$DIST_DIR/${NAME}-v${VERSION}.plugin"

echo "=== Packaging $NAME v$VERSION ==="
echo ""

# ── 5. Build in temp directory ────────────────────────────────────────────────
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

INCLUDED=()

copy_dir() {
  local key="$1"
  local default="$2"
  local dir
  dir=$(jq -r ".${key} // \"${default}\"" "$PLUGIN_JSON")
  dir="${dir%/}"
  local src="$ROOT/$dir"

  # Hard exclude forbidden paths
  if [[ "$dir" == workspace* || "$dir" == .claude* || "$dir" == dev/* ]]; then
    echo "  SKIP (forbidden): $dir"
    return
  fi

  [[ -d "$src" ]] || return 0

  while IFS= read -r file; do
    rel="${file#$ROOT/}"
    dst="$TMPDIR/$rel"
    mkdir -p "$(dirname "$dst")"
    cp "$file" "$dst"
    echo "  + $rel"
    INCLUDED+=("$rel")
  done < <(find "$src" -name "*.md" -type f 2>/dev/null)
}

copy_dir "skills"   "skills"
copy_dir "agents"   "agents"
copy_dir "commands" "commands"

# Copy .mcp.json if present
if [[ -f "$ROOT/.mcp.json" ]]; then
  cp "$ROOT/.mcp.json" "$TMPDIR/.mcp.json"
  echo "  + .mcp.json"
  INCLUDED+=(".mcp.json")
fi

# Copy hooks/ if present
if [[ -d "$ROOT/hooks" ]]; then
  cp -r "$ROOT/hooks" "$TMPDIR/hooks"
  echo "  + hooks/"
  INCLUDED+=("hooks/")
fi

# Always include plugin.json
mkdir -p "$TMPDIR/.claude-plugin"
cp "$PLUGIN_JSON" "$TMPDIR/.claude-plugin/plugin.json"
echo "  + .claude-plugin/plugin.json"

# ── 6. Zip into dist/ ─────────────────────────────────────────────────────────
mkdir -p "$DIST_DIR"
(cd "$TMPDIR" && zip -r "$OUTPUT" . -x "*.DS_Store")

echo ""
echo "=== Done ==="
echo "Output: $OUTPUT"
echo "Included ${#INCLUDED[@]} file(s) + plugin.json"
