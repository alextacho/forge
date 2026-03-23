---
description: Snapshot workspace and user context into a named fixture
---

Read `.claude-plugin/plugin.json` and `forge.yaml`. If either is missing, stop: "plugin.json / forge.yaml not found."

## /forge:save [name]

Snapshot the current workspace state into a named fixture. Fixtures are committed to source control.

**What gets saved:**
- All files under `workspace.root` subdirectories
- User-provided context files that exist (those not shipped as plugin defaults)

**What is NOT saved:**
- Plugin source files (`skills/`, `agents/`, `commands/`) — already in git
- The workspace directory structure itself — recreated on load

**Steps:**

1. Read `workspace.root`, `workspace.structure`, `dev.fixtures_dir` from `forge.yaml`
2. Prompt for name if not given
3. Warn and confirm if `<fixtures_dir>/<name>/` already exists
4. Copy workspace contents → `<fixtures_dir>/<name>/workspace/`
5. Auto-generate `.fixture.yaml`:
   - Count files per workspace subdir
   - Prompt for a short description
6. `git add -f <fixtures_dir>/<name>/` (force needed — `workspace/` is gitignored but `fixtures/` is committed)
7. Remind developer to review and commit

**Fixture structure:**
```
fixtures/<name>/
  .fixture.yaml        # description + file counts
  workspace/
    profiles/
    snapshots/
    ...
```

**.fixture.yaml format:**
```yaml
name: baseline
description: "Clean starting state"
created: 2026-03-21
contents:
  profiles: 2
  snapshots: 0
  runs: 1
```
