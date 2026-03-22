#!/usr/bin/env bash
# daily-cron.sh — run a Claude Code skill on a daily schedule
#
# SETUP (macOS launchd):
#   1. Edit SKILL_NAME below to the skill you want to run (e.g. "my-plugin:daily-report")
#   2. Edit PLUGIN_DIR to the absolute path of your plugin project
#   3. Install: cp scripts/daily-cron.sh ~/Library/LaunchAgents/com.my-plugin.daily.sh
#              chmod +x ~/Library/LaunchAgents/com.my-plugin.daily.sh
#   4. Create a launchd plist (see example below) and load with launchctl
#
# SETUP (cron):
#   Add to crontab: 0 8 * * * /path/to/scripts/daily-cron.sh >> /tmp/my-plugin-cron.log 2>&1
#
# LAUNCHD PLIST EXAMPLE (~/.launchd/com.my-plugin.daily.plist):
#   <?xml version="1.0" encoding="UTF-8"?>
#   <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "...">
#   <plist version="1.0"><dict>
#     <key>Label</key><string>com.my-plugin.daily</string>
#     <key>ProgramArguments</key><array>
#       <string>/bin/bash</string>
#       <string>/path/to/scripts/daily-cron.sh</string>
#     </array>
#     <key>StartCalendarInterval</key><dict>
#       <key>Hour</key><integer>8</integer>
#       <key>Minute</key><integer>0</integer>
#     </dict>
#   </dict></plist>

set -euo pipefail

SKILL_NAME="my-plugin:daily-skill"   # ← edit this
PLUGIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_FILE="$PLUGIN_DIR/workspace/runs/daily-$(date +%Y-%m-%d).log"

mkdir -p "$(dirname "$LOG_FILE")"

echo "[$(date)] Starting daily run of $SKILL_NAME" | tee -a "$LOG_FILE"

cd "$PLUGIN_DIR"
claude --skill "$SKILL_NAME" >> "$LOG_FILE" 2>&1

echo "[$(date)] Done." | tee -a "$LOG_FILE"
