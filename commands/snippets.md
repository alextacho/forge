---
description: Browse available snippets by category or keyword
---

## /forge:snippets [category|keyword]

Browse the forge snippet catalog. Snippets are pre-built, installable components — MCP server configs, hook patterns, and infrastructure scripts.

**Usage:**
```
/forge:snippets              # show all categories
/forge:snippets mcp          # filter by category
/forge:snippets playwright   # search by name or keyword
```

**Steps:**

1. Locate the `dev/snippets/` directory inside the forge plugin installation
2. If a category or keyword is given: filter to matching snippets (match against name, category, description)
3. For each matching snippet: read its `.snippet.yaml` and display name, category, description
4. If no argument: group by category

**Output format:**
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

Hooks
  format-on-save   Auto-format files after Write/Edit
  lint-on-save     Run linter after code changes
  post-commit      Hook that runs after each commit

Infrastructure
  daily-cron       Run a skill on a daily schedule
  weekly-digest    Weekly summary trigger

Run /forge:add <name> to install any snippet.
```

If a keyword is given and matches only one snippet, show its full `.snippet.yaml` details including `notes` and any required env vars.
