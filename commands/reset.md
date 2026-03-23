---
description: Clear workspace files, keep directory structure and plugin defaults
---

Read `.claude-plugin/plugin.json` and `forge.yaml`. If either is missing, stop: "plugin.json / forge.yaml not found."

## /forge:reset

Clear all workspace files and restore scaffold files to their templates.

**What gets cleared:**
- All files under each directory in `workspace.structure`

**What gets restored:**
- Each file in `scaffold` is overwritten from its template at `.forge/scaffolds/<basename>`

**What is kept:**
- Workspace subdirectory structure (empty dirs remain)
- Plugin source files not declared in `scaffold`
- `.forge/fixtures/` — never touched

**Steps:**

1. Read `workspace`, `scaffold` from `forge.yaml`
2. Collect all files under each workspace subdir + all scaffold paths
3. Preview: "This will delete N workspace files and restore M scaffold files. Continue? [y/N]"
4. If confirmed:
   - Delete all files under workspace subdirs, keep subdirs
   - Copy `.forge/scaffolds/<basename>` → each scaffold path
5. Report: N files deleted, M files restored

**Typical use:** run `/forge:reset` then `/forge:load <fixture>` to get a clean, reproducible test state.
