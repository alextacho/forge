---
description: Show full plugin state dashboard
---

Read `.claude-plugin/plugin.json` and `forge.yaml` (if present). If `plugin.json` not found, stop: "plugin.json not found."

## /forge:status

Show a full dashboard of current plugin state.

**Steps:**

1. Read `plugin.json` — name, version, status
2. **Application layer**: scan declared `skills/`, `agents/`, `commands/` dirs, list each `.md` file, check it exists (✓/✗)
3. **Workspace** (if `forge.yaml` present): list `workspace.structure` subdirs with file counts
4. **Fixtures** (if `forge.yaml` present): scan `dev.fixtures_dir` for subdirectories, show name + description from `.fixture.yaml`
5. **Discovery symlinks**: check `.claude/skills/` symlink state for each declared skill (✓/✗)
6. **PLAN.md**: if present, show open task count and last Done entry

**Output format:**
```
Plugin: my-plugin v0.1.0 [draft]

Application:
  skills/
    ✓ analyze.md
    ✗ report.md  (file missing)
  agents: (none)
  commands: (none)

Workspace: (workspace/)
  profiles/     3 files
  snapshots/    0 files
  runs/         1 file

Fixtures: (fixtures/)
  baseline  — "Clean starting state"
  with-data — "Three profiles loaded"

Discovery symlinks (.claude/skills/):
  ✓ analyze.md
  ✗ report.md  (not linked — run /forge:link)

Plan: 4 open tasks · last done: "Created initial structure (2026-03-21)"
```

Omit Workspace and Fixtures sections if `forge.yaml` is absent.
