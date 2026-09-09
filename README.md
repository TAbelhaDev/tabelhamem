<div align="center">

# tabelhamem

**Shared agent memory bridge between Claude Code and OpenCode.**

**English** · [Português](README.pt-BR.md)

[![Go Version](https://img.shields.io/github/go-mod/go-version/TAbelhaDev/tabelhamem?style=flat-square&logo=go&logoColor=white&color=00ADD8)](go.mod)
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
directory, with no way to point it elsewhere. OpenCode has no equivalent
automatic feature: only a global, hand-maintained `memory.md` journal and
per-repo `AGENTS.md` instruction files.

`tamem` bridges the two by making both point at the same plain-markdown
store: `~/agent-memory/<project>/`, sibling to `~/jobs` (the user's own
automation state, not owned by any single tool). Claude Code's per-project
memory directory becomes a symlink into it (transparent — Claude just does
normal file I/O); OpenCode is taught to read/write the same location through
an instruction block `tamem` maintains in the repo's `AGENTS.md`.

This also incidentally fixes a Claude Code quirk where every git worktree of
the same repo gets its own disconnected memory directory — running `tamem
run link` again from a worktree points it at the same shared bucket.

## Install

```bash
go install github.com/TAbelhaDev/tabelhamem@latest
```

## Usage

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
| `link` | `project=`, `repo=` | Creates/updates the bridge for a project: migrate + symlink + AGENTS.md section |
| `unlink` | `project=`, `repo=` | Reverses `link` for one repo: restores a real directory, removes the AGENTS.md section, leaves the shared dir alone |
| `status` | `project=`, `repo=` (optional) | Reports whether the symlink and AGENTS.md section are in place |
| `list` | (none) | Lists every project under `~/agent-memory/` |
| `search` | `query=`, `type=` (optional), `project=` (optional) | Full-text search across every bridged project's memory files |

## Limitations

- OpenCode has no built-in mechanism to auto-read `AGENTS.md` instructions
  the way Claude Code auto-loads `MEMORY.md` — the bridge only works as
  reliably as the model follows that instruction each session.
- `link` must be re-run per git worktree; a worktree's Claude Code memory
  directory is not touched automatically just because the main checkout was
  linked.
- After `unlink`, re-running `link` will refuse to overwrite if the local
  copy has since diverged from the shared store (same safety check as a
  fresh, never-linked directory) — resolve by hand (diff the two, then
  remove the local copy once you're sure nothing would be lost).
