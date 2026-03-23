---
description: Validate and package the plugin into a distributable file
---

Read `.claude-plugin/plugin.json`. If not found, stop: "plugin.json not found."

## /forge:pack

Validate then package the plugin into `dist/<name>-v<version>.plugin`.

**Pre-checks (abort if any fail):**

1. Run `/forge:validate` — must pass with no errors
2. Git working tree must be clean (`git status --porcelain` returns empty)
3. `plugin.json` `status` must be `"published"` — set this manually when ready to ship

**Packaging:**

4. Read `name` and `version` from `plugin.json`
5. Create `dist/` if needed
6. Copy into a temp directory:
   - All `.md` files under declared `skills/`, `agents/`, `commands/` paths
   - `.mcp.json` if present
   - `hooks/` directory if present
   - `.claude-plugin/plugin.json`
7. Always exclude: `workspace/`, `.forge/`, `forge.yaml`, `.claude/`, `dev/`, `dist/`, `.DS_Store`
8. Zip into `dist/<name>-v<version>.plugin`
9. Report: output path + count of included files

**To mark as published:** edit `plugin.json`, set `"status": "published"`, commit, then run `/forge:pack`.
