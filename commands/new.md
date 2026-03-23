---
description: Scaffold a new plugin project from scratch
---

Read `.claude-plugin/plugin.json` — if it already exists, warn that a plugin is already here and stop.

## /forge:new

Interactive scaffold for a new Claude Code plugin project.

**Steps:**

1. Ask: what is this plugin called? (kebab-case name)
2. Ask: what does it do? (one sentence)
3. Ask: what skills/agents/commands will it have? (list paths, e.g. `skills/analyze.md`)
4. Ask: does it need a runtime workspace? If yes, what subdirectories? (e.g. `profiles`, `snapshots`, `runs`)
5. Ask: any MCP servers, hooks, or infrastructure? (note: can add later with `/forge:add`)

**Generate:**

`.claude-plugin/plugin.json`:
```json
{
  "name": "<name>",
  "description": "<description>",
  "version": "0.1.0",
  "status": "draft",
  "skills": "skills/",
  "agents": "agents/",
  "commands": "commands/"
}
```

`forge.yaml` (only if workspace requested):
```yaml
workspace:
  root: workspace/
  structure:
    - <subdir1>

dev:
  fixtures_dir: fixtures/
```

`PLAN.md`:
```markdown
# Plugin Plan

## Design
<!-- What this plugin does, key decisions, open questions -->

## Tasks
- [ ] write skills

## Done
```

`.gitignore`:
```
workspace/
dist/
.DS_Store
```

Skeleton directories: `skills/`, `agents/`, `commands/` (and `workspace/<subdirs>/`, `fixtures/` if workspace requested).

**Report:** list every file and directory created. Suggest next steps:
- Add skills/agents to their directories
- Run `/forge:link` to wire up live development
- Run `/forge:snippets` to browse available MCP servers and hooks
- Run `/forge:plan` to start tracking work
