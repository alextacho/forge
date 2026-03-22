---
description: Full integrity check of plugin structure
---

Read `.claude-plugin/plugin.json`. If not found, stop: "plugin.json not found."

## /forge:validate

Full integrity check. Reports all errors, not just the first. Errors block `/forge:pack`; warnings are advisory.

**Checks:**

1. `plugin.json` is valid JSON and has required field `name`
2. Every `.md` file under declared `skills/` path exists and has valid frontmatter (`description` field)
3. Every `.md` file under declared `agents/` path exists
4. No declared path starts with `workspace/`, `.claude/`, or `dev/`
5. If `.mcp.json` declared: file exists and is valid JSON
6. If hooks declared: hooks file exists and is valid JSON
7. `.claude/skills/` symlinks exist for each declared skill — **warn** if missing (not an error)
8. `forge.yaml` (if present): `workspace.root` and `dev.fixtures_dir` paths are readable

**Output:**
```
Checking my-plugin v0.1.0...

✓ plugin.json valid
✓ skills/analyze.md exists and has valid frontmatter
✗ skills/report.md — file not found
✓ No forbidden paths
⚠ .claude/skills/analyze.md symlink missing — run /forge:link

1 error, 1 warning.
Errors must be resolved before /forge:pack.
```
