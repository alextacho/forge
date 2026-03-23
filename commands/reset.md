---
description: Clear workspace files, keep directory structure and plugin defaults
---

Read `.claude-plugin/plugin.json` and `forge.yaml`. If either is missing, stop: "plugin.json / forge.yaml not found."

## /forge:reset

Clear all paths declared in `reset`, then re-seed scaffold files from their last committed state.

**Key rule: commit what you want to keep before running reset.**
- Scaffold file improved → commit it first → `git checkout --` restores to that commit, your changes survive
- Scaffold file temporarily changed → reset discards back to last commit
- Custom files you want to keep → commit or stash them first

**Requirements:** must be run inside a git repo (`git rev-parse --git-dir`). If not, stop: "forge:reset requires a git repository."

**Steps:**

1. Read `reset`, `scaffold` from `forge.yaml`
2. Expand each reset path — collect all matching files
3. Classify each file:
   - **scaffold** — listed in `scaffold`, restored to last committed version via `git checkout --`
   - **user-added tracked** — in a reset path, tracked by git, not in `scaffold` — will be deleted (recoverable via git)
   - **user-added untracked** — in a reset path, not tracked by git (gitignored or new) — permanently deleted
4. Check scaffold files for uncommitted changes via `git diff --name-only`

**Show pre-flight report — BLOCK if action required:**

If any scaffold files have uncommitted changes, stop and show:
```
⚠  Scaffold files have uncommitted changes — commit or discard before resetting:

   agents/analyzers/swot-analyzer.md   (modified)
   context/signal-weights.md           (modified)

   If you want to KEEP these changes: commit them first, then re-run /forge:reset.
   If you want to DISCARD these changes: run git checkout -- <file>, then re-run /forge:reset.

   Aborting.
```

Otherwise show the full plan:
```
RESET PRE-FLIGHT
────────────────────────────────────────────

Files that will be re-seeded from last commit (scaffold):
  agents/analyzers/0_ANALYZERS.md
  agents/analyzers/swot-analyzer.md
  agents/extractors/0_EXTRACTORS.md
  agents/extractors/changelog-extractor.md
  context/distribution.yaml
  context/signal-weights.md

Files that will be DELETED — recoverable from git (user-added, tracked):
  agents/analyzers/competitive-analyzer.md
  agents/extractors/pricing-extractor.md

Files that will be PERMANENTLY DELETED — not in git (user-added, untracked):
  ⚠  workspace/briefs/brief-2026-03-01.md
  ⚠  workspace/snapshots/run-42.md

────────────────────────────────────────────
Total: 6 re-seeded · 2 git-recoverable · 2 permanently lost

Tip: run /forge:save first to preserve current state as a fixture.
```

5. If untracked files exist: prompt "N files will be permanently deleted (not in git). Continue? [y/N]" — default NO
6. Final confirmation: "Confirm reset? [y/N]"
7. Execute:
   - Delete all files under each reset path, keep directory structure
   - `git checkout -- <path>` for each scaffold file
8. Report: N files deleted, M files restored from git
