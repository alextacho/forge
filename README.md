# forge

`forge` is a small CLI for saving, loading, and resetting project-local state.

It manages one configured list of files and directories:

- `save` snapshots their current state
- `load` restores an exact snapshot
- `reset` restores committed template content

Every operation shows a preview and requires confirmation unless `--yes` is supplied.
Forge also includes agent-facing discovery commands and a read-only MCP server.

## Requirements

- Go 1.25 or newer

## Install

Install the latest tagged release with Go:

```bash
go install github.com/alextacho/forge/cmd/forge@latest
```

Install a specific version:

```bash
go install github.com/alextacho/forge/cmd/forge@v0.1.0
```

Or download a prebuilt binary from the
[GitHub releases page](https://github.com/alextacho/forge/releases). Archives are
published for macOS, Linux, and Windows on amd64 and arm64. Verify a downloaded
archive with `checksums.txt`:

```bash
sha256sum forge_v0.1.0_linux_amd64.tar.gz
# macOS: shasum -a 256 forge_v0.1.0_darwin_arm64.tar.gz
```

For local development from a checkout:

```bash
go build -o forge ./cmd/forge
```

Confirm the installed binary:

```bash
forge version
```

## Upgrade

Upgrade to the latest tagged release:

```bash
go install github.com/alextacho/forge/cmd/forge@latest
```

Upgrade or downgrade to a specific version:

```bash
go install github.com/alextacho/forge/cmd/forge@v0.1.0
```

If you installed from a release archive, replace the old `forge` binary with the
new one from the matching archive for your platform.

## Project setup

Create `.forge/config.yaml` in the project root:

```yaml
paths:
  - workspace/
  - config/generated.yaml
```

Managed paths must be literal project-relative files or directories. Absolute paths,
glob patterns, `..`, overlapping entries, and anything under `.forge/` are rejected.

The CLI searches upward from the current directory for this configuration, so commands
can run from nested directories.

Recommended layout:

```text
project/
├── .forge/
│   ├── config.yaml       # commit
│   ├── templates/        # commit
│   └── snapshots/        # ignore
├── workspace/
└── config/
    └── generated.yaml
```

Add this to the project `.gitignore`:

```gitignore
.forge/snapshots/
```

## Commands

```bash
forge help
forge version
forge --version
forge instructions
forge status
forge save [name] [--yes] [--no-clobber]
forge load [name] [--yes]
forge reset [--yes]
forge mcp
```

### Version

```bash
forge version
forge --version
```

Prints the release version, commit, and build date. Development builds print `dev`
unless version metadata or Go module build information is available.

### Instructions

```bash
forge instructions
```

Prints stable, agent-oriented instructions for project discovery, configuration,
snapshotting, reset behavior, and safety constraints. Use this when an agent needs
to learn how Forge should be used inside a project.

### Status

```bash
forge status
```

Inspects the current Forge project without changing files. It reports the project
root, config path, configured managed paths and their current state, available
snapshots, and configured template paths that are present.

### Save

```bash
forge save
forge save baseline
```

Without a name, `save` uses the snapshot name `default`. Saving the same name again
replaces the previous snapshot.

Use `--no-clobber` to fail when the snapshot already exists:

```bash
forge save baseline --no-clobber
```

Each managed path is recorded as present or absent. Loading the snapshot later removes
a path that was absent when saved.

### Load

```bash
forge load
forge load baseline
forge load default
```

Without a name, `load` uses `default`. Loading is exact: all managed paths are cleared
before present snapshot entries are restored. Extra files do not survive.

### Reset

```bash
forge reset
```

Reset clears every managed path, then restores matching content from
`.forge/templates/`. Templates mirror project-relative paths:

```text
.forge/templates/
├── workspace/
│   └── seed.txt
└── config/
    └── generated.yaml
```

A managed path with no matching template remains absent.

### MCP

```bash
forge mcp
```

Runs a read-only MCP server over stdio for agent discoverability. The server exposes:

- `forge_instructions`
- `forge_config_schema`
- `forge_project_status`
- `forge_list_snapshots`

It also exposes resources:

- `forge://instructions`
- `forge://config-schema`
- `forge://status`

The MCP server does not expose mutating tools. Agents should use it to discover how
Forge works and inspect project state, then call the CLI deliberately for `save`,
`load`, or `reset`.

Example MCP server configuration:

```json
{
  "mcpServers": {
    "forge": {
      "command": "forge",
      "args": ["mcp"]
    }
  }
}
```

## Confirmation

All commands print a short summary and ask for confirmation:

```text
Load snapshot "baseline": remove 2 present paths, restore 1, leave 1 absent.
Continue? [y/N]
```

Use `--yes` for scripts and CI:

```bash
forge load baseline --yes
forge reset --yes
```

`--yes` bypasses approval only. Configuration, snapshot, and filesystem validation
still run.

## Filesystem behavior

`forge` preserves:

- file contents
- directory structure
- permission bits
- symlink text

It does not preserve ownership, ACLs, extended attributes, or timestamps.

Symlinks are copied as symlinks and never traversed. Unsupported filesystem objects,
including sockets, devices, and named pipes, fail validation.

Snapshot replacement is staged before publication. Load and reset validate all source
content before clearing managed paths, but an unexpected write or permission failure
during restoration can still leave managed paths partially restored. The command exits
non-zero and reports this condition.

Do not run mutating `forge` commands concurrently within the same project.

## Development

```bash
go test ./...
go vet ./...
```

Build a local binary with explicit version metadata:

```bash
go build \
  -ldflags "-X github.com/alextacho/forge/internal/version.Version=dev -X github.com/alextacho/forge/internal/version.Commit=$(git rev-parse --short=12 HEAD) -X github.com/alextacho/forge/internal/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o forge ./cmd/forge
```

## Versioning and releases

Forge uses semantic version tags:

```text
vMAJOR.MINOR.PATCH
```

Release tags trigger the GitHub Actions release workflow. The workflow runs CI,
builds platform archives, stamps `forge version`, generates SHA256 checksums, and
publishes a GitHub Release.

Create a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Use version increments as follows:

- Patch: compatible bug fixes and documentation-only changes.
- Minor: backward-compatible commands, flags, MCP tools, or behavior.
- Major: breaking command syntax, config format, snapshot format, or MCP contract.

## License

MIT
