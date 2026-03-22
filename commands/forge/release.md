---
description: Bump version, commit, and push a new plugin release
---

Read `.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json`. If either is not found, stop and report which is missing.

## /forge:release [patch|minor|major]

Bump the version across all manifest files, commit, and push so the marketplace can serve the new version directly from GitHub.

**Default:** `patch` if no argument given.

**Pre-checks (abort if any fail):**

1. Run `/forge:validate` — must pass with no errors
2. Git working tree must be clean (`git status --porcelain` returns empty) — commit or stash changes first

**Steps:**

3. Read current `version` from `plugin.json`
4. Compute new version by incrementing the appropriate semver segment
5. Update `version` in `.claude-plugin/plugin.json`
6. Update `version` in `.claude-plugin/marketplace.json`
7. Set `status` to `"published"` in `plugin.json`
8. Commit: `git add .claude-plugin/plugin.json .claude-plugin/marketplace.json && git commit -m "release: v<new_version>"`
9. Push: `git push`
10. Report: `Released v<new_version> — marketplace will serve the update from GitHub.`

**Note:** This is the complete release flow for marketplace-distributed plugins. Claude Code pulls directly from GitHub — no zip packaging needed. Use `/forge:pack` only if you need a distributable file for manual or offline installation.
