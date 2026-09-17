<div align="center">

# tabelhamem

**Shared agent memory bridge between Claude Code and OpenCode.**

**English** · [Português](README.pt-BR.md)

[![Go Version](https://img.shields.io/github/go-mod/go-version/TAbelhaDev/tabelhamem?style=flat-square&logo=go&logoColor=white&color=00ADD8)](go.mod)
[![Built with Bubble Tea](https://img.shields.io/badge/built%20with-Bubble%20Tea-ff69b4?style=flat-square)](https://github.com/charmbracelet/bubbletea)
[![Powered by tabelhatuiui](https://img.shields.io/badge/theme-tabelhatuiui-d6b4f7?style=flat-square)](https://github.com/TAbelhaDev/tabelhatuiui)
[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue?style=flat-square)](LICENSE)

[![ko-fi](https://ko-fi.com/img/githubbutton_sm.svg)](https://ko-fi.com/ianptkcs)

</div>

---

## Why

Claude Code has a built-in, automatic memory system: it injects a
project-scoped `MEMORY.md` index plus typed topic files
(`feedback_*.md`/`project_*.md`/`reference_*.md`/`user_*.md`, YAML
frontmatter) at the start of every session, stored at
`~/.claude/projects/<escaped-cwd>/memory/` — one directory per exact working
directory, with no way to point it elsewhere. OpenCode has no per-project
memory dir of its own, but it auto-loads `AGENTS.md` at session start —
tamem uses that to teach it the same store.

`tamem` bridges the two by making both point at the same plain-markdown
store: `~/agent-memory/<project>/`, sibling to `~/jobs` (the user's own
automation state, not owned by any single tool). Claude Code's per-project
memory directory becomes a symlink into it (transparent — Claude just does
normal file I/O); OpenCode reads/writes the same location through an
instruction block `tamem` maintains in the repo's `AGENTS.md`, auto-loaded
by opencode each session.

When the repo is a git repository, `tamem` automatically detects all git
worktrees via `git worktree list` and links, unlinks, or checks the bridge
health for every one of them in a single invocation — no need to re-run
per worktree.

## Install

```bash
go install github.com/TAbelhaDev/tabelhamem/cmd/tamem@latest
```

### Local development

A `post-commit` hook in `.githooks/` rebuilds and reinstalls `tamem` to
`~/.local/bin/tamem` after every commit, so the local command never goes
stale. Git doesn't enable a repo's `.githooks/` automatically on clone — run
this once per clone:

```bash
git config core.hooksPath .githooks
```

## Usage

Running `tamem` with no arguments launches the interactive TUI for browsing,
searching, linking, and unlinking projects. The `ipc` subcommand remains
available for scripting.

### TUI

```bash
# Launch the interactive TUI
tamem

# Configure your projects in ~/.config/tabelhamem/config.toml
cat <<'EOF'
[[projects]]
slug = "tabelharadar"
repo = "/home/ianptkcs/codigo/tabelhadev/tabelharadar"

[layout]
sidebar_width_share = 1
right_width_share = 4
stats_height_share = 1
memory_height_share = 4

[general]
editor = ""
EOF
```

Keybindings (rebindable in `~/.config/tabelhamem/keybindings.json`):

| Key | Action |
|---|---|
| `q` | Quit |
| `?` | Help |
| `,` | Rebind keys |
| `r` | Rescan projects |
| `ctrl+shift+r` | Reload config |
| `ctrl+h` / `ctrl+l` | Navigate panels |
| `j` / `k` | Move cursor / scroll content |
| `enter` | Open file in editor |
| `/` | Search memory |
| `l` | Link project |
| `u` | Unlink project |
| `esc` | Back / quit |

### IPC (scriptable JSON)

```bash
# Link a project's memory: migrates existing Claude Code memory files into
# ~/agent-memory/<project>/, symlinks Claude's own memory dir to it, and
# writes/updates the bridge instructions in <repo>/AGENTS.md. Idempotent.
tamem ipc link project=tabelharadar repo=/home/ianptkcs/codigo/tabelhadev/tabelharadar --json

# Undo it: <repo>'s Claude Code memory dir gets a real copy of the current
# shared content back (the shared dir itself is left alone), and the
# AGENTS.md section is removed.
tamem ipc unlink project=tabelharadar repo=/home/ianptkcs/codigo/tabelhadev/tabelharadar --json

# Check the bridge's health for one project
tamem ipc status project=tabelharadar repo=/home/ianptkcs/codigo/tabelhadev/tabelharadar --json

# List every project currently bridged
tamem ipc list --json

# Search memory across every bridged project
tamem ipc search query=worktree type=feedback --json
```

## IPC Methods

| Method | Filters | Description |
|---|---|---|
| `global` | (none) | Sets up the shared global memory store: creates `~/agent-memory/global/`, migrates existing `AGENTS.md`, symlinks Claude Code, and updates the OpenCode pointer |
| `link` | `project=`, `repo=` | Creates/updates the bridge for a project: migrate + symlink + AGENTS.md section |
| `unlink` | `project=`, `repo=` | Reverses `link` for one repo: restores a real directory, removes the AGENTS.md section, leaves the shared dir alone |
| `status` | `project=`, `repo=` (optional) | Reports whether the symlink and AGENTS.md section are in place |
| `list` | (none) | Lists every project under `~/agent-memory/` |
| `search` | `query=`, `type=` (optional), `project=` (optional) | Full-text search across every bridged project's memory files |

## Limitations

- The OpenCode side of the bridge is instruction-driven: opencode auto-loads
  the `AGENTS.md` block, but actually reading/writing the shared store still
  depends on the model following that instruction each session.
- Worktree detection requires `git` on `$PATH`. If `git` is unavailable,
  `tamem` falls back to operating on a single directory (the `repo=` path).
- After `unlink`, re-running `link` will refuse to overwrite if the local
  copy has since diverged from the shared store (same safety check as a
  fresh, never-linked directory) — resolve by hand (diff the two, then
  remove the local copy once you're sure nothing would be lost).
