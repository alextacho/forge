---
description: Add forge infrastructure to an existing plugin project
---

Read `.claude-plugin/plugin.json` — if it does NOT exist, stop: "No plugin found here. Run /forge:new to create one."

Check if `forge.yaml` already exists — if yes, warn: "forge.yaml already exists. Run /forge:status to see current state." and stop.

## /forge:init

Add forge workspace, scaffold, and dev infrastructure to an existing plugin project. Does not touch `plugin.json` or any existing source files.

**Steps:**

1. Ask: does this plugin need a runtime workspace? If yes, what subdirectories? (e.g. `profiles`, `snapshots`, `runs`)
2. Ask: are there any files users will modify that should be resettable to a default? (e.g. `skills/registry.yaml`) — these become scaffold entries
3. For each scaffold path that already exists: offer to snapshot it now as the template (copy to `.forge/scaffolds/<basename>`)
4. For each scaffold path that does not exist: ask for initial template content, write to `.forge/scaffolds/<basename>` and copy to declared path

**Generate:**

`forge.yaml`:
```yaml
workspace:
  root: workspace/
  structure:
    - <subdir1>

scaffold:
  - <path>
```
Omit `workspace` block if none requested. Omit `scaffold` block if none declared.

`.forge/scaffolds/` — created if scaffold entries declared; templates written here

`.forge/fixtures/` — always created (empty)

Workspace subdirectories — created if workspace requested

**Update `.gitignore`** — append any missing entries:
```
workspace/
.forge/fixtures/
```
Do not duplicate entries already present.

**Report:** list every file and directory created or updated. Suggest next steps:
- Run `/forge:save <name>` to snapshot current state as a fixture
- Run `/forge:reset` to verify reset behaviour works as expected
