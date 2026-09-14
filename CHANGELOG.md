# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
