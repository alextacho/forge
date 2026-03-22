#!/usr/bin/env bash
# weekly-digest.sh — run a Claude Code skill on a weekly schedule
#
# SETUP:
#   Edit SKILL_NAME and PLUGIN_DIR, then install via cron or launchd.
#
# CRON (every Monday at 8am):
#   0 8 * * 1 /path/to/scripts/weekly-digest.sh >> /tmp/my-plugin-weekly.log 2>&1

set -euo pipefail

SKILL_NAME="my-plugin:weekly-digest"   # ← edit this
PLUGIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_FILE="$PLUGIN_DIR/workspace/runs/weekly-$(date +%Y-%W).log"

mkdir -p "$(dirname "$LOG_FILE")"

echo "[$(date)] Starting weekly digest: $SKILL_NAME" | tee -a "$LOG_FILE"

cd "$PLUGIN_DIR"
claude --skill "$SKILL_NAME" >> "$LOG_FILE" 2>&1

echo "[$(date)] Done." | tee -a "$LOG_FILE"
