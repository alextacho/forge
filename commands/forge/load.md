---
description: Restore a named fixture into the workspace
---

Read `.claude-plugin/plugin.json` and `forge.yaml`. If either is missing, stop: "plugin.json / forge.yaml not found."

## /forge:load [name]

Restore a named fixture into the workspace.

**Steps:**

1. Read `workspace.root`, `workspace.structure`, `dev.fixtures_dir` from `forge.yaml`
2. If name omitted: list fixtures in `<fixtures_dir>/` (show name + description from `.fixture.yaml`) and ask which to load
3. If workspace has content: warn and ask to confirm overwrite
4. Copy `<fixtures_dir>/<name>/workspace/` → `<workspace.root>/` (preserve subdir structure)
5. Show `.fixture.yaml` metadata after loading
6. Report: "Workspace ready — fixture '<name>' loaded."

**To start fresh before loading:** run `/forge:reset` first (or run them back to back — that is the standard test setup cycle).
