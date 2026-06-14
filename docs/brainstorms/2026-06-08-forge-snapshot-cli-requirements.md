---
date: 2026-06-08
topic: forge-snapshot-cli
---

# Forge Snapshot CLI Requirements

## Summary

Build a standalone CLI for saving, loading, and resetting configured project state. It uses project-local snapshots and templates, requires confirmation before every operation, and has no dependency on Claude skills or plugin infrastructure.

## Key Decisions

- **One managed path list.** The same configured files and directories participate in save, load, and reset.
- **Exact restoration.** Loading clears all managed paths before restoring a snapshot, preventing stale files from surviving.
- **Project-local state.** Configuration, snapshots, and templates live under `.forge/`.
- **Templates define reset state.** Reset clears managed paths and restores matching content from committed templates.
- **Overwrite by default.** Saving to an existing snapshot replaces it. Users may opt into collision failure with `--no-clobber`.
- **Interactive safety.** Every command previews its effects and requires approval unless `--yes` is supplied.

## Requirements

**Project and configuration**

- R1. The CLI searches upward from the current directory until it finds `.forge/config.yaml`.
- R2. The directory containing `.forge/config.yaml` is the project root for path resolution and storage.
- R3. Configuration contains one list of managed paths shared by save, load, and reset.
- R4. Managed paths are literal project-relative files or directories; glob patterns and absolute paths are unsupported.
- R5. The CLI rejects paths that resolve outside the project root, including traversal through `..`.

**Snapshots**

- R6. Snapshots are stored under `.forge/snapshots/`.
- R7. Running `forge save` writes the snapshot named `default`.
- R8. Running `forge save <name>` writes the specified named snapshot.
- R9. Saving replaces an existing snapshot with the same name unless `--no-clobber` is supplied.
- R10. A snapshot records each managed path as either present or absent at save time.
- R11. Present files and directories are copied without following symlinks; symlinks are preserved as symlinks.
- R12. Snapshot replacement is atomic enough that an interrupted save does not leave the prior snapshot partially overwritten.

**Loading**

- R13. Running `forge load` loads the `default` snapshot.
- R14. Running `forge load <name>` loads the specified snapshot, including `default` when named explicitly.
- R15. Load fails without modifying managed paths when the requested snapshot does not exist or is invalid.
- R16. Load clears every managed path before restoring the snapshot.
- R17. Paths recorded as present are restored exactly from the snapshot.
- R18. Paths recorded as absent remain absent after loading.
- R19. Load preserves stored symlinks without traversing their targets.

**Reset**

- R20. Reset clears every managed path.
- R21. Reset restores matching files, directories, and symlinks from `.forge/templates/`, using paths relative to the project root.
- R22. Managed paths without matching template content remain absent after reset.
- R23. `.forge/templates/` is intended to be committed to version control.
- R24. `.forge/snapshots/` is intended to be ignored by version control.

**Confirmation and reporting**

- R25. Save, load, and reset show a concise preview before changing project or snapshot state.
- R26. The preview identifies the command, snapshot when applicable, managed paths affected, and whether content will be replaced, removed, or restored.
- R27. Each command requires affirmative user confirmation and otherwise exits without changes.
- R28. `--yes` skips interactive confirmation for scripting and CI.
- R29. Commands fail without changes when confirmation is required but no interactive input is available.
- R30. Each successful command reports a concise count of saved, restored, removed, absent, or replaced paths as applicable.

## Key Flows

- F1. Save project state
  - **Trigger:** The user runs `forge save [name]`.
  - **Steps:** Locate the project, validate configuration and paths, inspect current state, preview the snapshot replacement, request approval, then write the complete snapshot.
  - **Outcome:** The selected snapshot exactly represents the present or absent state of every managed path.
  - **Covered by:** R1-R12, R25-R30.

- F2. Load project state
  - **Trigger:** The user runs `forge load [name]`.
  - **Steps:** Locate and validate the snapshot, preview removals and restorations, request approval, clear all managed paths, then restore present entries.
  - **Outcome:** Managed paths exactly match the selected snapshot.
  - **Covered by:** R13-R19, R25-R30.

- F3. Reset project state
  - **Trigger:** The user runs `forge reset`.
  - **Steps:** Validate configured paths and template content, preview removals and restorations, request approval, clear all managed paths, then restore matching templates.
  - **Outcome:** Managed paths match the template baseline, with untemplated paths absent.
  - **Covered by:** R20-R30.

## Acceptance Examples

- AE1. **Covers R7, R9.** Given an existing `default` snapshot, when the user runs `forge save` and approves, then the snapshot is replaced with the current managed state.
- AE2. **Covers R9.** Given an existing named snapshot, when the user runs `forge save baseline --no-clobber`, then the command fails before changing that snapshot.
- AE3. **Covers R10, R18.** Given a configured file that is absent during save but exists later, when that snapshot is loaded, then the later file is removed.
- AE4. **Covers R16-R18.** Given a managed directory containing extra files not present in a snapshot, when the snapshot is loaded, then the extra files do not survive.
- AE5. **Covers R20-R22.** Given one managed path with a matching template and another without one, when reset completes, then the first is restored and the second remains absent.
- AE6. **Covers R25-R29.** Given a destructive load, when the user declines the preview or confirmation cannot be obtained, then no managed path changes.
- AE7. **Covers R28.** Given a valid non-interactive environment, when a command is run with `--yes`, then it proceeds without reading from standard input.
- AE8. **Covers R5, R11, R19.** Given a configured escape path or a symlink targeting content outside the project, then the escape path is rejected and the symlink target is never traversed or modified.

## Scope Boundaries

- No Claude Code skills, slash commands, plugin manifests, marketplace release flow, or plugin packaging.
- No glob patterns, external managed paths, or symlink traversal.
- No overlay-style loading; load always restores exact managed state.
- No global snapshot store or cross-project snapshot sharing.
- No requirement for snapshots to be committed to version control.

## Dependencies and Assumptions

- The project can write within `.forge/` and all configured managed paths.
- Snapshot names will be validated as safe single names rather than arbitrary filesystem paths.
- Configuration and templates are developer-authored project assets; snapshots are local runtime data.
