#!/usr/bin/env bash
# load-fixture.sh — copy a named fixture into workspace/
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
FORGE_YAML="$ROOT/forge.yaml"

# ── Read forge.yaml ───────────────────────────────────────────────────────────
if [[ ! -f "$FORGE_YAML" ]]; then
  echo "ERROR: forge.yaml not found at $FORGE_YAML"
  exit 1
fi

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
  echo "Available fixtures:"
  for d in "$FIXTURES_DIR"/*/; do
    if [[ -d "$d" ]]; then
      name=$(basename "$d")
      desc=""
      [[ -f "$d/.fixture.yaml" ]] && desc=$(grep "^description:" "$d/.fixture.yaml" | sed 's/description: *//' | tr -d '"')
      printf "  %-20s %s\n" "$name" "$desc"
    fi
  done
  echo ""
  printf "Fixture name: "
  read -r FIXTURE_NAME
fi

FIXTURE_DIR="$FIXTURES_DIR/$FIXTURE_NAME"
if [[ ! -d "$FIXTURE_DIR" ]]; then
  echo "ERROR: fixture '$FIXTURE_NAME' not found"
  exit 1
fi

# ── Check workspace for existing content ─────────────────────────────────────
WORKSPACE_FILES=$(find "$WORKSPACE" -mindepth 2 -type f -not -name ".gitkeep" 2>/dev/null | wc -l | tr -d ' ')
if [[ "$WORKSPACE_FILES" -gt 0 ]]; then
  echo "WARNING: workspace/ has $WORKSPACE_FILES file(s)."
  printf "Overwrite with fixture '%s'? [y/N] " "$FIXTURE_NAME"
  read -r confirm
  if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
    echo "Aborted."
    exit 0
  fi
  find "$WORKSPACE" -mindepth 2 -type f -not -name ".gitkeep" -delete 2>/dev/null || true
  echo "Cleared existing workspace contents."
fi

# ── Copy fixture → workspace ──────────────────────────────────────────────────
echo "Loading fixture '$FIXTURE_NAME' into workspace/..."
if [[ -d "$FIXTURE_DIR/workspace" ]]; then
  cp -r "$FIXTURE_DIR/workspace/." "$WORKSPACE/"
fi

echo ""
echo "Done. Fixture '$FIXTURE_NAME' loaded."
if [[ -f "$FIXTURE_DIR/.fixture.yaml" ]]; then
  echo ""
  sed 's/^/  /' "$FIXTURE_DIR/.fixture.yaml"
fi
