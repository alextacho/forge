---
description: Snapshot workspace and user context into a named fixture
---

Read `.claude-plugin/plugin.json` and `forge.yaml`. If either is missing, stop: "plugin.json / forge.yaml not found."

## /forge:save [name]

Snapshot the current workspace state into a named fixture. Fixtures mirror source paths exactly.

**What gets saved:**
- All files under each directory listed in `workspace.structure`
- All files listed in `scaffold` (user-modified copies of scaffold templates)

**What is NOT saved:**
- Plugin source files not declared in `scaffold` — already in git
- Directory structure itself — recreated on load

**Steps:**

1. Read `workspace`, `scaffold` from `forge.yaml`
2. Prompt for name if not given
3. Warn and confirm if `.forge/fixtures/<name>/` already exists
4. For each workspace directory: copy contents → `.forge/fixtures/<name>/<dir>/`
5. For each scaffold entry: copy file → `.forge/fixtures/<name>/<path>`
6. Auto-generate `.forge/fixtures/<name>/.fixture.yaml`:
   - Count files saved per source path
   - Prompt for a short description
7. Remind developer to review and commit if fixtures are tracked

**Fixture structure:**
```
.forge/fixtures/<name>/
  .fixture.yaml
  workspace/
    profiles/
    snapshots/
  skills/
    registry.yaml
```

**.fixture.yaml format:**
```yaml
name: baseline
description: "Clean starting state"
created: 2026-03-21
contents:
  workspace/profiles: 2
  workspace/snapshots: 0
  skills/registry.yaml: 1
```
