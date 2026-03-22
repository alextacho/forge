# forge

A Claude Code skill for developing, testing, and publishing Claude Code plugins.

`forge` manages the full plugin development lifecycle — from scaffolding a new plugin to packaging a distributable release. It lives outside any plugin project and operates on whatever plugin is in your current working directory.

---

## Installation

**Step 1: Add the marketplace**

```bash
claude plugin marketplace add <owner>/forge
```

**Step 2: Install the plugin**

```bash
claude plugin install forge@forge
```

Then invoke it from inside any plugin project directory.

---

## Concepts

### Plugin structure

`forge` works with the official Claude Code plugin structure:

```
my-plugin/
├── .claude-plugin/
│   └── plugin.json       # plugin metadata (name, version, skills, etc.)
├── skills/               # skill definitions
├── agents/               # agent definitions
├── commands/             # slash commands
├── hooks/                # event hooks
├── .mcp.json             # MCP server configs
├── fixtures/             # saved test states (committed)
├── workspace/            # runtime data (gitignored)
└── forge.yaml            # dev config (workspace dirs, fixtures path)
```

`plugin.json` is the official manifest. `forge.yaml` holds dev-only config that shouldn't be part of the published plugin.

### forge.yaml

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

### Fixtures

A **fixture** is a named snapshot of your plugin's runtime state — workspace files and any user-provided context. Fixtures are committed to source control so you can restore a known state for testing.

```
fixtures/
  baseline/
    .fixture.yaml         # description + metadata
    workspace/            # snapshot of workspace contents
    context/              # snapshot of user-provided context files
```

### Snippets

Snippets are pre-built, installable components — MCP server configs, hook patterns, infrastructure scripts. They live inside the `forge` plugin and can be copied into any plugin project with `/forge:add`.

---

## Commands

### `/forge:new`

Interactive scaffold for a new plugin project.

Asks for plugin name, description, and what skills/agents/commands it will have. Generates `plugin.json`, `forge.yaml`, folder structure, and `.gitignore`.

```
/forge:new
```

---

### `/forge:link`

Symlink the plugin's skills, agents, and commands into `.claude/` so Claude Code discovers them immediately without reinstalling. Essential for live development.

```
/forge:link
```

Safe to re-run. Skips symlinks that already point to the right place. Warns if a symlink exists but points elsewhere.

---

### `/forge:unlink`

Remove all `.claude/` symlinks for this plugin. Never touches source files.

```
/forge:unlink
```

---

### `/forge:status`

Full dashboard of current plugin state.

```
Plugin: my-plugin v0.1.0 [draft]

Application:
  skills/
    ✓ analyze.md
    ✗ report.md  (file missing)
  agents: (none)
  commands: (none)

Context:
  shipped defaults:
    ✓ context/prompts.md
  setup required:
    ✓ context/config.md
    ✗ context/credentials.md  — "Add your API key"
  setup optional:
    ✗ context/overrides.md  — not yet created

Workspace: (workspace/)
  profiles/     3 files
  snapshots/    0 files
  runs/         1 file

Fixtures: (fixtures/)
  baseline  — "Clean state, no prior runs"
  with-data — "Three profiles loaded"

Discovery symlinks (.claude/skills/):
  ✓ analyze.md
  ✗ report.md  (not linked — run /forge:link)
```

---

### `/forge:validate`

Full integrity check. Reports all errors before you waste time packaging.

Checks:
- `plugin.json` is valid JSON and has required fields
- Every declared skill/agent/command file exists
- Every declared context file exists
- No declared path leaks into `.claude/` or `workspace/`
- Each skill has valid frontmatter (`name` + `description`)
- Symlinks exist for each declared skill (warning only)

Errors block `/forge:pack`. Warnings are advisory.

```
/forge:validate
```

---

### `/forge:save [name]`

Snapshot current workspace and user-provided context into a named fixture.

```
/forge:save baseline
/forge:save          # prompts for name
```

Saves workspace contents and user context files (those listed in `setup.required`/`setup.optional`). Does not save committed defaults — they're always present. Stages the fixture with `git add` and reminds you to review before committing.

---

### `/forge:load [name]`

Restore a named fixture into the workspace and context.

```
/forge:load baseline
/forge:load          # shows available fixtures, prompts
```

Warns if the workspace has content and asks to confirm overwrite.

---

### `/forge:reset`

Clear workspace files and user-provided context. Keep committed defaults.

```
/forge:reset
```

Previews what will be deleted and asks to confirm. Does not remove the workspace directory structure itself.

Use this to return to a clean state before loading a fixture. The combination of `reset` + `load` is your test setup cycle.

---

### `/forge:snippets [category|keyword]`

Browse available snippets.

```
/forge:snippets              # show all categories
/forge:snippets mcp          # filter by category
/forge:snippets playwright   # search by name or keyword
```

Example output:

```
MCP Servers
  playwright     Browser automation via Playwright MCP
  notion         Notion workspace read/write
  google-drive   Google Drive file access
  github         GitHub repos, PRs, issues
  slack          Slack messaging and search
  linear         Linear issue tracking
  gmail          Gmail read and draft
  gcal           Google Calendar access

Infrastructure
  daily-cron     Daily scheduler — runs a skill on a cron schedule
  weekly-digest  Weekly summary trigger

Hooks
  format-on-save Auto-format files after Write/Edit
  lint-on-save   Run linter after code changes
  post-commit    Hook that runs after each commit

Run /forge:add <name> to install a snippet.
```

---

### `/forge:add <snippet>`

Install a snippet into the plugin project.

```
/forge:add mcp/playwright
/forge:add mcp/notion
/forge:add hooks/format-on-save
/forge:add infra/daily-cron
```

Copies the snippet's files into the appropriate location in your plugin project and shows what was added and what (if anything) needs manual configuration (e.g. env vars, API keys).

---

### `/forge:pack`

Validate and package the plugin into a distributable file.

```
/forge:pack
```

Pre-checks:
1. `/forge:validate` must pass (no errors)
2. Git working tree must be clean
3. `plugin.json` `status` must be `"published"` (set this manually when ready)

Output: `dist/<name>-v<version>.plugin`

Always excludes: `workspace/`, `fixtures/`, `forge.yaml`, `.claude/`, `.dev/`.

---

## Typical workflows

### Starting a new plugin

```
mkdir my-plugin && cd my-plugin
/forge:new
/forge:link
# start writing skills...
```

### Daily development loop

```
# edit skills, test in Claude Code via symlinks
/forge:status        # check what's wired up
/forge:validate      # catch errors early
```

### Testing with fixtures

```
/forge:reset
/forge:load baseline
# run your skills, inspect output
/forge:save after-first-run
```

### Publishing

```
/forge:validate
# bump version in plugin.json, set status: "published"
/forge:pack
```

---

## What forge does not touch

- Files inside `skills/`, `agents/`, `commands/`, `context/` — those are the plugin's responsibility
- Committed defaults (`publish.context` files)
- The workspace directory structure (reset clears contents, not dirs)
