---
description: Clear workspace files, keep directory structure and plugin defaults
---

Read `.claude-plugin/plugin.json` and `forge.yaml`. If either is missing, stop: "plugin.json / forge.yaml not found."

## /forge:reset

Clear all files from the workspace. Keep the directory structure and any committed plugin files.

**What gets cleared:**
- All files under `workspace.root` subdirectories

**What is kept:**
- Workspace subdirectory structure (the empty dirs remain)
- All plugin source files (`skills/`, `agents/`, `commands/`)
- `fixtures/` — never touched

**Steps:**

1. Read `workspace.root`, `workspace.structure` from `forge.yaml`
2. Collect all files under each workspace subdir
3. Preview the list: "This will delete N files from workspace/. Continue? [y/N]"
4. If confirmed: delete files, keep subdirs
5. Report: N files deleted, workspace cleared

**Typical use:** run `/forge:reset` then `/forge:load <fixture>` to get a clean, reproducible test state.
