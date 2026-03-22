# forge — Implementation Plan

## Overview

Rename and rebuild `skill-dev` → `forge`. The command prefix changes from `/sd:*` to `/forge:*`, the folder structure is updated, dead commands are removed, two new commands are added, and the config source shifts from `MANIFEST.yaml` to the official `plugin.json` + `forge.yaml`.

---

## 1. Folder renames

| Current | New |
|---|---|
| `dev/skills/skill-dev/` | `dev/skills/forge/` |
| `dev/skills/skill-dev/commands/sd/` | `dev/skills/forge/commands/forge/` |

No other top-level directories change. `dev/scripts/`, `dev/fixtures/` stay in place.

---

## 2. Config source change

Current scripts and commands read `MANIFEST.yaml` (custom format). The new forge reads:

| File | Purpose |
|---|---|
| `.claude-plugin/plugin.json` | Official plugin metadata: name, version, status, skills, agents, commands, context |
| `forge.yaml` | Dev-only config: workspace root, workspace subdirs, fixtures dir |

### forge.yaml schema

```yaml
workspace:
  root: workspace/
  structure:
    - profiles
    - snapshots
    - runs

dev:
  fixtures_dir: fixtures/
```

All scripts and command files must be updated to read from these two files instead of `MANIFEST.yaml`.

---

## 3. Command files

### Delete (7 files)

| File | Reason |
|---|---|
| `_preamble.md` | Inline preamble into SKILL.md instead |
| `diff.md` | Cut from v1 |
| `init.md` | Merged into `new` (new creates workspace dirs) |
| `load-fixture.md` | Redundant — `load.md` does everything it does plus more |
| `promote.md` | Absorbed into `pack` as a pre-check gate |
| `qa.md` | Documented as `reset` + `load` — not worth a command |
| `setup.md` | End-user concern, not developer concern — folded into `status` |
| `test.md` | Too vague to be meaningful at this level — cut from v1 |

### Keep and update (8 files)

| Old name | New name | Changes |
|---|---|---|
| `new.md` | `new.md` | Create `plugin.json` + `forge.yaml` instead of `MANIFEST.yaml`; create workspace dirs inline (absorbs init) |
| `link.md` | `link.md` | Read from `plugin.json` instead of MANIFEST; also link forge commands into `.claude/commands/forge/` |
| `unlink.md` | `unlink.md` | Minor update — remove MANIFEST reference |
| `status.md` | `status.md` | Read from `plugin.json` + `forge.yaml`; absorb setup check (show missing required/optional context files inline) |
| `validate.md` | `validate.md` | Validate against `plugin.json` schema instead of MANIFEST; update frontmatter checks |
| `save.md` | `save.md` | Read workspace/fixtures paths from `forge.yaml` instead of MANIFEST |
| `load.md` | `load.md` | Merge `load-fixture` behavior in; read from `forge.yaml` |
| `reset.md` | `reset.md` | Read from `forge.yaml` instead of MANIFEST |
| `publish.md` | `pack.md` | Absorb promote gate (check `status == "published"` in `plugin.json`); update exclusion list to exclude `forge.yaml`, `fixtures/`, `.dev/` |

### Add (3 files)

| File | What it does |
|---|---|
| `snippets.md` | Browse snippets catalog by category or keyword; reads from `dev/snippets/` |
| `add.md` | Install a named snippet into the current plugin project |
| `plan.md` | Read/update `PLAN.md` in the plugin root — show tasks, add, mark done |

---

## 4. PLAN.md convention

`/forge:new` creates a `PLAN.md` stub in the plugin root. It's a plain markdown file — committed to source control, no special format required beyond these three sections:

```markdown
# Plugin Plan

## Design
<!-- What this plugin does, key decisions, open questions -->

## Tasks
- [ ] task description
- [x] completed task

## Done
- Short note on what shipped (date)
```

`/forge:plan` command behavior:
- No args: print current Tasks section + count done
- `add <text>`: append `- [ ] <text>` to Tasks
- `done <text>`: match task by substring, move to Done with today's date
- `design`: print the Design section

forge's SKILL.md will note that `PLAN.md` exists and can be read/updated during any conversation — no separate skill or agent needed.

---

## 5. SKILL.md rewrite

Rewrite `dev/skills/forge/SKILL.md` to:
- Rename all references from `sd` → `forge` and `skill-dev` → `forge`
- Document the 12-command set (remove the 8 deleted commands)
- Replace MANIFEST.yaml config table with `plugin.json` + `forge.yaml` config table
- Add snippets system documentation (catalog structure, `.snippet.yaml` format)
- Note that `PLAN.md` exists in the plugin root and can be read/updated during any conversation
- Update guiding principles

---

## 6. Scripts

Scripts currently parse `MANIFEST.yaml` with `awk`. They need to be updated to read `plugin.json` (JSON) and `forge.yaml` (YAML).

| Script | Changes |
|---|---|
| `link.sh` | Read skills/agents/commands from `plugin.json` using `jq` instead of awk/MANIFEST |
| `unlink.sh` | Minor — remove MANIFEST reference |
| `validate.sh` | Validate `plugin.json` instead of MANIFEST; check `plugin.json` schema |
| `save-fixture.sh` | Read workspace/fixtures paths from `forge.yaml` |
| `load-fixture.sh` | Read workspace/fixtures paths from `forge.yaml` |
| `package.sh` → `pack.sh` | Read from `plugin.json`; remove MANIFEST.yaml from output; add `forge.yaml` to exclusion list; update status check to read from `plugin.json` |

All scripts: add `jq` dependency check at top (for JSON parsing).

---

## 6. Snippets directory

Create `dev/snippets/` with the following structure. Each snippet is a directory containing its files plus a `.snippet.yaml` descriptor.

```
dev/snippets/
  mcp/
    playwright/
      .snippet.yaml
      .mcp.json             # MCP server config block to merge
      README.md             # setup instructions
    notion/
      .snippet.yaml
      .mcp.json
      README.md
    google-drive/
      .snippet.yaml
      .mcp.json
      README.md
    github/
      .snippet.yaml
      .mcp.json
      README.md
    slack/
      .snippet.yaml
      .mcp.json
      README.md
    linear/
      .snippet.yaml
      .mcp.json
      README.md
    gmail/
      .snippet.yaml
      .mcp.json
      README.md
    gcal/
      .snippet.yaml
      .mcp.json
      README.md
  hooks/
    format-on-save/
      .snippet.yaml
      hooks.json            # hook config block to merge
    lint-on-save/
      .snippet.yaml
      hooks.json
    post-commit/
      .snippet.yaml
      hooks.json
  infra/
    daily-cron/
      .snippet.yaml
      scripts/
        daily-cron.sh
    weekly-digest/
      .snippet.yaml
      scripts/
        weekly-digest.sh
```

### .snippet.yaml format

```yaml
name: playwright
category: mcp
description: Browser automation via Playwright MCP
installs:
  - .mcp.json           # merged into target plugin's .mcp.json
requires:
  env:
    - PLAYWRIGHT_MCP_PORT   # optional — shown as note during install
notes: "Run `npx playwright install` before first use."
```

`installs` lists files to copy/merge. For `.mcp.json` and `hooks.json`, the install is a merge (add the block), not a full file replace.

---

## 7. CLAUDE.md update

Update `dev/CLAUDE.md`:
- Replace `skill-dev` references with `forge`
- Update path references: `dev/skills/skill-dev/` → `dev/skills/forge/`

---

## 8. fixtures/README.md update

Update command references from `/skill-dev *` → `/forge:*`.

---

## Execution order

1. Rename folders (`skill-dev/` → `forge/`, `commands/sd/` → `commands/forge/`)
2. Rewrite `SKILL.md`
3. Delete the 8 removed command files
4. Update the 8 kept command files (rename `publish.md` → `pack.md`)
5. Create `snippets.md`, `add.md`, and `plan.md`
6. Update all scripts (`MANIFEST.yaml` → `plugin.json` + `forge.yaml`)
7. Create `dev/snippets/` directory with all snippet stubs
8. Update `CLAUDE.md` and `fixtures/README.md`

---

## What does not change

- `dev/fixtures/` directory and its structure
- `dev/scripts/` directory name
- The concept of workspace, fixtures, and symlink-based live dev
- `.claude-plugin/plugin.json` format (this is official — forge just reads it)
