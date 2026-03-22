---
description: Read and update PLAN.md — design notes, task checklist, done log
---

Read `.claude-plugin/plugin.json`. If not found, stop: "plugin.json not found."

## /forge:plan [subcommand]

Read and update `PLAN.md` in the plugin root. Simple task tracking — no phases, no state files, just markdown.

**PLAN.md structure:**
```markdown
# Plugin Plan

## Design
<!-- What this plugin does, key decisions, open questions -->

## Tasks
- [ ] task description
- [x] completed task

## Done
- Short note on what shipped (2026-03-21)
```

---

### /forge:plan

No subcommand — show current status:
- Print all open tasks (`- [ ]` lines) from the Tasks section
- Print count: "N open, M done"
- Print the last 3 Done entries

If `PLAN.md` doesn't exist: offer to create it with the default stub.

---

### /forge:plan add <text>

Append a new task to the Tasks section:
```
- [ ] <text>
```

---

### /forge:plan done <text>

Mark a task as done:
1. Find the task in Tasks section by substring match
2. If multiple matches: list them and ask to confirm
3. Remove the line from Tasks
4. Append to Done section: `- <text> (<today's date>)`

---

### /forge:plan design

Print the full Design section. Use this to review decisions and open questions during development.

---

**Error handling:**
- `PLAN.md` not found with no subcommand: offer to create it
- `PLAN.md` not found with `add`/`done`/`design`: create it first with default stub, then apply the operation
- No matching task for `done`: list all open tasks and ask which was meant
