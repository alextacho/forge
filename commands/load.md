---
description: Restore a named fixture into the workspace
---

Read `.claude-plugin/plugin.json` and `forge.yaml`. If either is missing, stop: "plugin.json / forge.yaml not found."

## /forge:load [name]

Restore a named fixture, writing files back to their original paths.

**Steps:**

1. Read `reset` from `forge.yaml`
2. If name omitted: list fixtures in `.forge/fixtures/` (show name + description from `.fixture.yaml`) and ask which to load
3. Check if any reset paths currently have content — if yes, warn and ask to confirm overwrite
4. For each path in fixture (excluding `.fixture.yaml`): copy → original path, creating directories as needed
5. Show `.fixture.yaml` metadata after loading
6. Report: "Fixture '<name>' loaded — N files restored."

**To start fresh before loading:** run `/forge:reset` first (or run them back to back — that is the standard test setup cycle).
