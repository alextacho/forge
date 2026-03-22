---
description: Install a snippet into the current plugin project
---

Read `.claude-plugin/plugin.json`. If not found, stop: "plugin.json not found."

## /forge:add <snippet>

Install a snippet from the forge catalog into the current plugin project.

**Usage:**
```
/forge:add playwright
/forge:add mcp/notion
/forge:add hooks/format-on-save
/forge:add infra/daily-cron
```

Name can be bare (`playwright`) or path-qualified (`mcp/playwright`). If bare name matches multiple snippets, list them and ask to clarify.

**Steps:**

1. Locate snippet in `dev/snippets/` — search by name, then by `<category>/<name>`
2. If not found: show available snippets and stop
3. Read `.snippet.yaml` — get `installs` list and `notes`
4. For each file in `installs`:
   - `.mcp.json`: merge the snippet's server block into the plugin's `.mcp.json` (create file if absent)
   - `hooks.json`: merge the snippet's hook block into `hooks/hooks.json` (create file if absent)
   - Other files (scripts, etc.): copy to declared destination path, create parent dirs as needed
5. Show what was installed
6. If `notes` is set: print it (setup instructions, env vars to configure, etc.)

**Example output:**
```
Installing playwright (MCP server)...
  ✓ merged server block into .mcp.json

Note: Run `npx playwright install` before first use.
      Set PLAYWRIGHT_MCP_PORT in your environment if needed.
```

**Merge behavior for .mcp.json:**

The snippet's `.mcp.json` contains a single server block keyed by server name:
```json
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": ["@playwright/mcp"]
    }
  }
}
```

`/forge:add` reads the plugin's existing `.mcp.json`, adds the new server block under `mcpServers`, and writes it back. Existing entries are untouched. If the key already exists, warn and skip.
