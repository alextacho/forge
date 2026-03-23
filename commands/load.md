---
description: Restore a named fixture into the workspace
---

Read `.claude-plugin/plugin.json` and `forge.yaml`. If either is missing, stop: "plugin.json / forge.yaml not found."

## /forge:load [name]

Restore a named fixture into the workspace and scaffold paths.

**Steps:**

1. Read `workspace`, `scaffold` from `forge.yaml`
2. If name omitted: list fixtures in `.forge/fixtures/` (show name + description from `.fixture.yaml`) and ask which to load
3. If workspace has content: warn and ask to confirm overwrite
4. For each workspace directory: copy `fixture/<name>/<dir>/` → project directory (preserve subdir structure)
5. For each scaffold entry present in fixture: copy `fixture/<name>/<path>` → its declared path
6. Show `.fixture.yaml` metadata after loading
7. Report: "Fixture '<name>' loaded."

**To start fresh before loading:** run `/forge:reset` first (or run them back to back — that is the standard test setup cycle).
