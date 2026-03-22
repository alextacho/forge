#!/usr/bin/env bash
# validate.sh — check plugin.json integrity, declared file existence, frontmatter, symlinks
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
PLUGIN_JSON="$ROOT/.claude-plugin/plugin.json"

ERRORS=0
WARNINGS=0

err()  { echo "  ERROR: $*"; ((ERRORS++));   true; }
warn() { echo "  WARN:  $*"; ((WARNINGS++)); true; }
ok()   { echo "  OK:    $*"; }

# ── Dependency check ──────────────────────────────────────────────────────────
if ! command -v jq &>/dev/null; then
  echo "ERROR: jq is required. Install with: brew install jq"
  exit 1
fi

echo ""
echo "=== forge validate: $ROOT ==="
echo ""

# ── 1. plugin.json exists and is valid JSON ───────────────────────────────────
echo "── 1. plugin.json"
if [[ ! -f "$PLUGIN_JSON" ]]; then
  err "plugin.json not found at $PLUGIN_JSON"
  echo ""
  echo "FAILED: plugin.json missing — cannot continue."
  exit 1
fi

PLUGIN_NAME=$(jq -r '.name // empty' "$PLUGIN_JSON" 2>/dev/null || true)
PLUGIN_VERSION=$(jq -r '.version // empty' "$PLUGIN_JSON" 2>/dev/null || true)
PLUGIN_STATUS=$(jq -r '.status // "draft"' "$PLUGIN_JSON" 2>/dev/null || true)

if [[ -z "$PLUGIN_NAME" ]]; then
  err "plugin.json missing required field: name"
else
  ok "plugin.json valid — $PLUGIN_NAME v$PLUGIN_VERSION [$PLUGIN_STATUS]"
fi

# ── 2. Declared directories exist and contain no forbidden paths ──────────────
echo ""
echo "── 2. Declared paths"

check_dir() {
  local key="$1"
  local default="$2"
  local dir
  dir=$(jq -r ".${key} // \"${default}\"" "$PLUGIN_JSON")
  dir="${dir%/}"  # strip trailing slash
  local full="$ROOT/$dir"

  if [[ "$dir" == workspace* || "$dir" == .claude* || "$dir" == dev/* ]]; then
    err "$key path '$dir' must not start with workspace/, .claude/, or dev/"
    return
  fi

  if [[ -d "$full" ]]; then
    local count
    count=$(find "$full" -name "*.md" -type f 2>/dev/null | wc -l | tr -d ' ')
    ok "$dir/ ($count .md files)"
  else
    warn "$dir/ does not exist (create it or add skills/agents/commands)"
  fi
}

check_dir "skills"   "skills"
check_dir "agents"   "agents"
check_dir "commands" "commands"

# Check .mcp.json if referenced
MCP_FILE=$(jq -r '.mcpServers // empty' "$PLUGIN_JSON" 2>/dev/null || true)
if [[ -n "$MCP_FILE" && "$MCP_FILE" != "null" ]]; then
  if [[ -f "$ROOT/$MCP_FILE" ]]; then
    if jq . "$ROOT/$MCP_FILE" &>/dev/null; then
      ok "$MCP_FILE (valid JSON)"
    else
      err "$MCP_FILE is not valid JSON"
    fi
  else
    err "$MCP_FILE declared but not found"
  fi
fi

# Check hooks if referenced
HOOKS_FILE=$(jq -r '.hooks // empty' "$PLUGIN_JSON" 2>/dev/null || true)
if [[ -n "$HOOKS_FILE" && "$HOOKS_FILE" != "null" ]]; then
  if [[ -f "$ROOT/$HOOKS_FILE" ]]; then
    if jq . "$ROOT/$HOOKS_FILE" &>/dev/null; then
      ok "$HOOKS_FILE (valid JSON)"
    else
      err "$HOOKS_FILE is not valid JSON"
    fi
  else
    err "$HOOKS_FILE declared but not found"
  fi
fi

# ── 3. Skill frontmatter ──────────────────────────────────────────────────────
echo ""
echo "── 3. Skill frontmatter"

SKILLS_DIR=$(jq -r '.skills // "skills/"' "$PLUGIN_JSON")
SKILLS_FULL="$ROOT/${SKILLS_DIR%/}"

if [[ -d "$SKILLS_FULL" ]]; then
  SKILL_COUNT=0
  while IFS= read -r skill_file; do
    SKILL_COUNT=$((SKILL_COUNT + 1))
    rel="${skill_file#$ROOT/}"
    first_line=$(head -1 "$skill_file")
    if [[ "$first_line" != "---" ]]; then
      err "$rel: missing frontmatter (must start with ---)"
      continue
    fi
    has_desc=$(awk '/^---/{found++; if(found==2) exit; next} found==1 && /^description:/{print; exit}' "$skill_file")
    if [[ -z "$has_desc" ]]; then
      err "$rel: frontmatter missing 'description' field"
    else
      ok "$rel (frontmatter valid)"
    fi
  done < <(find "$SKILLS_FULL" -name "*.md" -type f 2>/dev/null)

  [[ $SKILL_COUNT -eq 0 ]] && warn "No .md files found in $SKILLS_DIR"
else
  warn "$SKILLS_DIR/ not found — skipping frontmatter check"
fi

# ── 4. Symlink check ──────────────────────────────────────────────────────────
echo ""
echo "── 4. Discovery symlinks"

CLAUDE_SKILLS="$ROOT/.claude/skills"
if [[ ! -d "$CLAUDE_SKILLS" ]]; then
  warn ".claude/skills/ does not exist — run /forge:link"
elif [[ -d "$SKILLS_FULL" ]]; then
  while IFS= read -r skill_file; do
    fname=$(basename "$skill_file")
    link_path="$CLAUDE_SKILLS/$fname"
    if [[ -L "$link_path" ]]; then
      ok "symlink: .claude/skills/$fname"
    else
      warn "symlink missing for $fname — run /forge:link"
    fi
  done < <(find "$SKILLS_FULL" -maxdepth 1 -name "*.md" -type f 2>/dev/null)
fi

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "── Summary"
if [[ $ERRORS -gt 0 ]]; then
  echo "  FAILED: $ERRORS error(s), $WARNINGS warning(s)"
  echo "  Errors must be resolved before /forge:pack"
  exit 1
else
  echo "  PASSED: 0 errors, $WARNINGS warning(s)"
  exit 0
fi
