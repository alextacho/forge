---
description: Symlink plugin skills/agents/commands into .claude/ for live development
---

Read `.claude-plugin/plugin.json`. If not found, stop: "plugin.json not found."

## /forge:link

Create `.claude/` symlinks so Claude Code discovers the plugin's skills, agents, and commands immediately — without reinstalling. Changes to source files are live instantly.

**Steps:**

1. Read `skills`, `agents`, `commands` paths from `plugin.json`
2. Create `.claude/skills/`, `.claude/agents/`, `.claude/commands/<name>/` if they don't exist
3. For each `.md` file under the declared skills path:
   - Compute relative path from `.claude/skills/` back to source
   - If symlink already correct: report "already linked", skip
   - If symlink exists but points elsewhere: warn, skip (do not overwrite)
   - Otherwise: create symlink
4. Repeat for agents (`→ .claude/agents/`) and commands (`→ .claude/commands/<name>/`)
5. For forge dev commands in `dev/skills/forge/commands/forge/`: symlink each `.md` file into `.claude/commands/forge/`
6. Report: linked / already linked / warnings
