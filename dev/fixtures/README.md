# Fixtures

Test fixtures snapshot a specific workspace state so you can replay it reliably during development and testing.

## Structure

Each fixture is a directory under `fixtures/`:

```
fixtures/
  <fixture-name>/
    .fixture.yaml          # metadata: description, created, file counts
    workspace/             # mirrors workspace/ subdirs
      profiles/
      snapshots/
      runs/
      ...
```

## Working with fixtures

| Command | What it does |
|---|---|
| `/forge:save <name>` | Snapshot current workspace into a named fixture |
| `/forge:load <name>` | Restore a fixture into workspace |
| `/forge:reset` | Clear workspace (run before loading a fixture for clean state) |

Or use the scripts directly:

```bash
dev/scripts/save-fixture.sh <name>
dev/scripts/load-fixture.sh <name>
```

## .fixture.yaml format

```yaml
name: <fixture-name>
description: "Short description of what this state represents"
created: 2026-03-21
contents:
  profiles: 3
  snapshots: 6
  runs: 1
```

## Conventions

- Name fixtures after the scenario they represent: `baseline`, `with-data`, `post-run`
- Include enough data to exercise the full plugin workflow
- Fixtures are committed to the repo — keep them small (text only, no binaries)
- Personal or sensitive data should stay in `workspace/` (gitignored), not fixtures
