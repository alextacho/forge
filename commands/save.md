---
description: Snapshot workspace and user context into a named fixture
---

Read `.claude-plugin/plugin.json` and `forge.yaml`. If either is missing, stop: "plugin.json / forge.yaml not found."

## /forge:save [name]

Snapshot all reset paths into a named fixture. Fixtures mirror source paths exactly.

**What gets saved:** all files matching each path in `reset` (including user-added files beyond scaffold defaults)

**What is NOT saved:** plugin source files not covered by a reset path — already in git

**Steps:**

1. Read `reset` from `forge.yaml`
2. Prompt for name if not given
3. Warn and confirm if `.forge/fixtures/<name>/` already exists
4. For each reset path: expand glob, copy matched files → `.forge/fixtures/<name>/<original-path>`
5. Auto-generate `.forge/fixtures/<name>/.fixture.yaml`:
   - Count files saved per reset path
   - Prompt for a short description
6. Remind developer to review and commit if fixtures are tracked

**Fixture structure:**
```
.forge/fixtures/<name>/
  .fixture.yaml
  workspace/
    briefs/
    profiles/
    snapshots/
  agents/
    analyzers/
    extractors/
  skills/
    synthesizers/
  context/
```

**.fixture.yaml format:**
```yaml
name: baseline
description: "Clean starting state"
created: 2026-03-21
contents:
  workspace/*: 4
  agents/analyzers/*: 3
  agents/extractors/*: 2
  skills/synthesizers/*: 2
  context/*: 2
```
