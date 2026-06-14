package guide

const Usage = `Usage:
  forge help
  forge version
  forge --version
  forge instructions
  forge init <path> [path...]
  forge status
  forge save [name] [--yes] [--no-clobber]
  forge load [name] [--yes]
  forge reset [--yes]
  forge mcp

Commands:
  help          Show command syntax.
  version       Show the Forge version.
  instructions  Show agent-oriented setup and snapshot guidance.
  init          Create .forge/config.yaml for the current project.
  status        Inspect the current Forge project without changing files.
  save          Snapshot configured paths. Defaults to snapshot "default".
  load          Restore an exact snapshot. Defaults to snapshot "default".
  reset         Restore committed template content from .forge/templates.
  mcp           Run the read-only Forge MCP server over stdio.

Mutating commands preview their effect and require confirmation unless --yes is supplied.`

const Instructions = `Forge is a project-local state snapshot CLI.

Project discovery:
  - Run forge from the project root or any nested directory.
  - Forge searches upward for .forge/config.yaml.
  - The directory containing .forge/config.yaml is the project root.

Configuration:
  - Run forge init <path> [path...] to create .forge/config.yaml.
  - Commit .forge/config.yaml and .forge/templates/ when reset should restore baseline content.
  - Configure one literal project-relative path list:

    paths:
      - workspace/
      - config/generated.yaml

  - Paths may name files or directories.
  - Absolute paths, glob patterns, .. traversal, overlapping paths, and anything under .forge/ are rejected.
  - forge init also creates .forge/templates/ and adds .forge/snapshots/ to .gitignore.

Snapshot workflow:
  - forge init workspace config/generated.yaml
      Initialize Forge config for the current project.
  - forge save --yes
      Save the current configured state as the default snapshot.
  - forge save baseline --yes
      Save a named snapshot.
  - forge save baseline --no-clobber --yes
      Save only if the named snapshot does not already exist.
  - forge load --yes
      Restore the default snapshot exactly.
  - forge load baseline --yes
      Restore a named snapshot exactly.
  - forge reset --yes
      Clear all configured paths and restore matching files from .forge/templates/.

Safety model:
  - save, load, and reset print a preview before changing state.
  - Without --yes, they require an interactive terminal confirmation.
  - load and reset clear managed paths before restoring content, so extra files inside managed directories do not survive.
  - Snapshots record present and absent paths. Loading a snapshot removes paths that were absent when saved.
  - Symlinks are copied as symlinks and are not traversed.
  - Do not run mutating forge commands concurrently in the same project.

Agent usage:
  - Prefer forge status before choosing an operation.
  - Use forge instructions when the project needs setup guidance.
  - Use --yes only after you have decided the previewed operation is intended.
  - Treat load and reset as destructive because they remove current managed content.`

const ConfigSchema = `Forge configuration is YAML at .forge/config.yaml.

Required shape:

paths:
  - workspace/
  - config/generated.yaml

Rules:
  - paths is required and must contain at least one entry.
  - Entries are literal project-relative paths.
  - Files and directories are both supported.
  - Absolute paths are rejected.
  - .. traversal is rejected.
  - Glob patterns are not expanded.
  - Duplicate and overlapping paths are rejected.
  - .forge and anything below .forge/ are reserved.`
