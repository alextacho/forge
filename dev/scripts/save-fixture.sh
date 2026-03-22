#!/usr/bin/env bash
# save-fixture.sh — snapshot workspace/ into a named fixture
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
FORGE_YAML="$ROOT/forge.yaml"

# ── Read forge.yaml ───────────────────────────────────────────────────────────
if [[ ! -f "$FORGE_YAML" ]]; then
  echo "ERROR: forge.yaml not found at $FORGE_YAML"
  exit 1
fi

# Parse forge.yaml with ruby (built-in on macOS) or basic grep
read_yaml() {
  local key="$1"
  if command -v ruby &>/dev/null; then
    ruby -e "require 'yaml'; d=YAML.safe_load(File.read('$FORGE_YAML')); print d.dig(*'$key'.split('.')) || ''"
  else
    grep -A1 "${key##*.}:" "$FORGE_YAML" | tail -1 | tr -d ' '
  fi
}

WORKSPACE=$(read_yaml "workspace.root")
FIXTURES_DIR=$(read_yaml "dev.fixtures_dir")
WORKSPACE="${WORKSPACE:-workspace}"
FIXTURES_DIR="${FIXTURES_DIR:-fixtures}"
WORKSPACE="$ROOT/${WORKSPACE%/}"
FIXTURES_DIR="$ROOT/${FIXTURES_DIR%/}"

# ── Get fixture name ──────────────────────────────────────────────────────────
FIXTURE_NAME="${1:-}"
if [[ -z "$FIXTURE_NAME" ]]; then
  printf "Fixture name: "
  read -r FIXTURE_NAME
fi
if [[ -z "$FIXTURE_NAME" ]]; then
  echo "ERROR: fixture name is required"
  exit 1
fi

FIXTURE_DIR="$FIXTURES_DIR/$FIXTURE_NAME"

# ── Confirm overwrite ─────────────────────────────────────────────────────────
if [[ -d "$FIXTURE_DIR" ]]; then
  echo "WARNING: fixture '$FIXTURE_NAME' already exists"
  printf "Overwrite? [y/N] "
  read -r confirm
  if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
    echo "Aborted."
    exit 0
  fi
  rm -rf "$FIXTURE_DIR"
fi

mkdir -p "$FIXTURE_DIR"

# ── Copy workspace → fixture ──────────────────────────────────────────────────
echo "Saving workspace/ → fixtures/$FIXTURE_NAME/ ..."
rsync -a --exclude=".gitkeep" "$WORKSPACE/" "$FIXTURE_DIR/workspace/"

# ── Generate .fixture.yaml ────────────────────────────────────────────────────
echo ""
printf "Short description for this fixture: "
read -r DESCRIPTION

{
  echo "name: $FIXTURE_NAME"
  echo "description: \"${DESCRIPTION}\""
  echo "created: $(date -u +"%Y-%m-%d")"
  echo "contents:"
  for subdir in "$WORKSPACE"/*/; do
    [[ -d "$subdir" ]] || continue
    name=$(basename "$subdir")
    count=$(find "$subdir" -maxdepth 1 -type f -not -name ".gitkeep" 2>/dev/null | wc -l | tr -d ' ')
    echo "  $name: $count"
  done
} > "$FIXTURE_DIR/.fixture.yaml"

echo ""
echo "Saved: fixtures/$FIXTURE_NAME/"
echo ""
sed 's/^/  /' "$FIXTURE_DIR/.fixture.yaml"

# ── Stage with git ────────────────────────────────────────────────────────────
echo ""
echo "Staging fixture files..."
cd "$ROOT"
git add -f "$FIXTURES_DIR/$FIXTURE_NAME/"
echo "Staged. Review with 'git diff --cached' before committing."
