---
description: Remove all .claude/ symlinks for this plugin
---

Read `.claude-plugin/plugin.json`. If not found, stop: "plugin.json not found."

## /forge:unlink

Remove all symlinks from `.claude/skills/`, `.claude/agents/`, `.claude/commands/`. Never touches source files.

**Steps:**

1. For each of `.claude/skills/`, `.claude/agents/`, `.claude/commands/`: find all symlinks, remove them, leave directory in place
2. Report: how many removed per directory
