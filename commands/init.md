---
description: Add forge infrastructure to an existing plugin project
---

Read `.claude-plugin/plugin.json` — if it does NOT exist, stop: "No plugin found here. Run /forge:new to create one."

Check if `forge.yaml` already exists — if yes, warn: "forge.yaml already exists. Run /forge:status to see current state." and stop.

## /forge:init

Add forge workspace, scaffold, and dev infrastructure to an existing plugin project. Does not touch `plugin.json` or any existing source files.

**Steps:**

1. Ask: which directories should be cleared on reset? (e.g. `workspace/*`, `agents/extractors/*`) — these become `reset` entries
2. Ask: are there any files that should be re-seeded to a default after reset? (e.g. `skills/registry.yaml`) — these become `scaffold` entries
3. Remind the developer: scaffold files are restored from git on reset — make sure they are committed before using `/forge:reset`

**Generate:**

`forge.yaml`:
```yaml
reset:
  - <dir>/*

scaffold:
  - <path>
```
Omit `reset` block if none declared. Omit `scaffold` block if none declared.

`.forge/fixtures/` — always created (empty)

Directories matching reset paths — created if not present

**Update `.gitignore`** — append any missing entries:
```
.forge/fixtures/
```
Also add any reset paths that should be ephemeral (ask the developer). Do not duplicate entries already present.

**Report:** list every file and directory created or updated. Suggest next steps:
- Run `/forge:save <name>` to snapshot current state as a fixture
- Run `/forge:reset` to verify reset behaviour works as expected
