---
description: Scaffold a new plugin project from scratch
---

Read `.claude-plugin/plugin.json` — if it already exists, stop: "A plugin already exists here. Run /forge:init to add forge infrastructure to it, or /forge:status to see its current state."

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

`forge.yaml` (only if reset paths or scaffold requested):
```yaml
reset:
  - workspace/*     # all contents of workspace/ recursively
  - <dir>/*         # e.g. agents/extractors/*

scaffold:
  - <path>          # e.g. skills/registry.yaml — seeded from template on reset
```
Omit `reset` block if no managed paths. Omit `scaffold` block if no scaffold files.

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
.forge/fixtures/
.claude/commands/
.claude/skills/
.claude/agents/
dist/
.DS_Store
```
Also gitignore any reset paths that should be ephemeral (e.g. `workspace/`). Scaffold files themselves are committed normally — they are the templates.

Skeleton directories: `skills/`, `agents/`, `commands/` (and directories matching `reset` paths, `.forge/fixtures/` if reset or scaffold requested).

**Report:** list every file and directory created. Suggest next steps:
- Add skills/agents to their directories
- Run `/forge:link` to wire up live development
- Run `/forge:snippets` to browse available MCP servers and hooks
- Run `/forge:plan` to start tracking work
