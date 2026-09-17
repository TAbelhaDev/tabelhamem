# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `global` IPC method: idempotent setup of the shared global memory store
  at `~/agent-memory/global/`. Migrates existing `~/.config/opencode/AGENTS.md`
  and `~/.claude/memory/` content into the store, symlinks both to the
  shared store.
- Interactive TUI (Bubble Tea + tabelhatuiui v0.6.0): 3-panel layout
  (projects sidebar, bridge status, memory browser) with integrated
  search, link/unlink forms, inline scrollable file viewer, and
  rebindable keybindings.
- `config.toml` at `~/.config/tabelhamem/config.toml`: `[[projects]]`
  slug-to-repo mapping, `[layout]` panel shares, `[general] editor`.
- `searchMemory()` extracted as a reusable core shared by both IPC and TUI
  search modes.

### Changed
- `tamem` (no args) now launches the TUI instead of printing usage to
  stderr and exiting with code 1.
- Dependencies: added bubbletea v1.3.10, bubbles v1.0.0, lipgloss v1.1.0,
  tabelhatuiui v0.6.0, BurntSushi/toml v1.6.0.

## [0.3.0] - 2026-09-14

### Added
- Automatic git worktree detection: `link`, `unlink`, and `status` now cover
  all worktrees of a repo in a single invocation via `git worktree list`.
- `linkResult`, `unlinkResult`, and `statusResult` include a `worktrees`
  array with per-worktree bridge state.
- `gitWorktrees()` helper in store with graceful fallback when `git` is
  unavailable.
- Test suite: 42 tests covering store helpers, frontmatter parsing,
  snippet extraction, link/unlink flows, and AGENTS.md section management.

### Changed
- Removed the limitation that `link` must be re-run per worktree.

## [0.2.0] - 2026-09-09

### Added
- `search` method: full-text search across all bridged projects, with
  optional `type=` and `project=` filters.
- `unlink` method: reverses `link` for one repo, restoring a real directory
  and removing the AGENTS.md section.

### Fixed
- Build now depends on the published `tabelhascaff` v0.7.1 instead of a
  local `replace` directive.

## [0.1.0] - 2026-09-09

### Added
- Initial release: `link`, `status`, and `list` IPC methods.
- Bilingual README (English + Portuguese).
- CI and release workflows.
