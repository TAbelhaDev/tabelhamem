package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TAbelhaDev/tabelhascaff/ipc"
)

// linkResult is the wire format for the link method.
type linkResult struct {
	Project         string   `json:"project"`
	Repo            string   `json:"repo"`
	SharedDir       string   `json:"shared_dir"`
	ClaudeDir       string   `json:"claude_dir"`
	MigratedFiles   []string `json:"migrated_files,omitempty"`
	AlreadyLinked   bool     `json:"already_linked"`
	AgentsMdUpdated bool     `json:"agents_md_updated"`
}

func ipcLink(filters map[string]string) int {
	slug := filters["project"]
	repo := filters["repo"]
	if slug == "" || repo == "" {
		fmt.Fprintln(os.Stderr, "filtros project= e repo= são obrigatórios")
		return 1
	}

	repoAbs, err := filepath.Abs(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	if info, err := os.Stat(repoAbs); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "repo %q não é um diretório válido\n", repoAbs)
		return 1
	}

	shared := sharedMemoryDir(slug)
	claudeDir := claudeMemoryDir(repoAbs)

	result := linkResult{Project: slug, Repo: repoAbs, SharedDir: shared, ClaudeDir: claudeDir}

	if err := os.MkdirAll(shared, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "erro criando", shared, ":", err)
		return 1
	}

	migrated, alreadyLinked, err := linkClaudeDir(claudeDir, shared)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	result.MigratedFiles = migrated
	result.AlreadyLinked = alreadyLinked

	updated, err := ensureAgentsSection(repoAbs, shared)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro atualizando AGENTS.md:", err)
		return 1
	}
	result.AgentsMdUpdated = updated

	return ipc.WriteJSON(result)
}

// linkClaudeDir makes claudeDir a symlink to shared, migrating any existing
// files into shared first. Returns the migrated file names (nil if none),
// and whether claudeDir was already correctly linked (a no-op).
func linkClaudeDir(claudeDir, shared string) ([]string, bool, error) {
	info, err := os.Lstat(claudeDir)

	switch {
	case errors.Is(err, os.ErrNotExist):
		if err := os.MkdirAll(filepath.Dir(claudeDir), 0o755); err != nil {
			return nil, false, err
		}
		return nil, false, os.Symlink(shared, claudeDir)

	case err != nil:
		return nil, false, err

	case info.Mode()&os.ModeSymlink != 0:
		target, err := symlinkTarget(claudeDir)
		if err != nil {
			return nil, false, fmt.Errorf("lendo symlink existente em %s: %w", claudeDir, err)
		}
		if target != shared {
			return nil, false, fmt.Errorf("%s já é um symlink pra %s (esperado %s) — recuso sobrescrever", claudeDir, target, shared)
		}
		return nil, true, nil

	case info.IsDir():
		entries, err := os.ReadDir(claudeDir)
		if err != nil {
			return nil, false, err
		}
		var migrated []string
		for _, e := range entries {
			src := filepath.Join(claudeDir, e.Name())
			dst := filepath.Join(shared, e.Name())
			if _, err := os.Stat(dst); err == nil {
				return migrated, false, fmt.Errorf("%s já existe em %s — recuso sobrescrever, resolva manualmente", e.Name(), shared)
			}
			if err := moveFile(src, dst); err != nil {
				return migrated, false, fmt.Errorf("movendo %s: %w", e.Name(), err)
			}
			migrated = append(migrated, e.Name())
		}
		if err := os.Remove(claudeDir); err != nil {
			return migrated, false, err
		}
		return migrated, false, os.Symlink(shared, claudeDir)

	default:
		return nil, false, fmt.Errorf("%s existe mas não é diretório nem symlink — não sei lidar", claudeDir)
	}
}

const (
	agentsMarkerStart = "<!-- tamem:memory-bridge:start -->"
	agentsMarkerEnd   = "<!-- tamem:memory-bridge:end -->"
)

// agentsSection is the block tamem keeps in a repo's AGENTS.md, teaching an
// agent with no built-in memory feature (like OpenCode) to read/write the
// same shared store Claude Code reads/writes automatically.
func agentsSection(shared string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", agentsMarkerStart)
	fmt.Fprintf(&b, "## Shared agent memory\n\n")
	fmt.Fprintf(&b, "This project's persistent notes live at `%s`, shared\n", shared)
	fmt.Fprintf(&b, "between Claude Code (which reads/writes it automatically through a\n")
	fmt.Fprintf(&b, "symlink at its own `~/.claude/projects/<escaped-cwd>/memory/`) and any\n")
	fmt.Fprintf(&b, "other agent with no built-in memory feature — that has to be done\n")
	fmt.Fprintf(&b, "manually, following the convention below.\n\n")
	fmt.Fprintf(&b, "Before starting non-trivial work, read `%s/MEMORY.md`\n", shared)
	fmt.Fprintf(&b, "(an index) and any linked topic file relevant to the task.\n\n")
	fmt.Fprintf(&b, "When you learn something worth remembering — not code patterns or git\n")
	fmt.Fprintf(&b, "history (derivable from the repo), but user preferences, durable\n")
	fmt.Fprintf(&b, "feedback, project state/decisions, or pointers to external systems —\n")
	fmt.Fprintf(&b, "write a new `.md` file in that directory and add a one-line entry to\n")
	fmt.Fprintf(&b, "`MEMORY.md`. Each topic file needs YAML frontmatter:\n\n")
	fmt.Fprintf(&b, "```yaml\n---\nname: short-kebab-case-slug\ndescription: one-line summary used to judge relevance later\nmetadata:\n  type: user | feedback | project | reference\n---\n```\n\n")
	fmt.Fprintf(&b, "- **user**: the user's role, goals, expertise.\n")
	fmt.Fprintf(&b, "- **feedback**: guidance the user gave about how to approach work\n")
	fmt.Fprintf(&b, "  (what to avoid, what worked) — include *why*.\n")
	fmt.Fprintf(&b, "- **project**: ongoing work/decisions not derivable from the code —\n")
	fmt.Fprintf(&b, "  include *why* and *how to apply*.\n")
	fmt.Fprintf(&b, "- **reference**: pointers to external systems (trackers, dashboards, docs).\n\n")
	fmt.Fprintf(&b, "Link related entries with `[[other-file-name]]` (without the `.md`\n")
	fmt.Fprintf(&b, "extension). Don't duplicate an existing entry — update it instead.\n")
	fmt.Fprintf(&b, "%s", agentsMarkerEnd)
	return b.String()
}

// ensureAgentsSection writes or replaces the tamem-owned block in
// <repo>/AGENTS.md, creating the file if it doesn't exist yet. Returns
// whether the file changed.
func ensureAgentsSection(repo, shared string) (bool, error) {
	path := filepath.Join(repo, "AGENTS.md")
	section := agentsSection(shared)

	existing, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		content := "# AGENTS.md\n\n" + section + "\n"
		return true, os.WriteFile(path, []byte(content), 0o644)
	}
	if err != nil {
		return false, err
	}

	text := string(existing)
	startIdx := strings.Index(text, agentsMarkerStart)
	endIdx := strings.Index(text, agentsMarkerEnd)
	if startIdx == -1 || endIdx == -1 {
		newText := strings.TrimRight(text, "\n") + "\n\n" + section + "\n"
		return true, os.WriteFile(path, []byte(newText), 0o644)
	}

	endIdx += len(agentsMarkerEnd)
	newText := text[:startIdx] + section + text[endIdx:]
	if newText == text {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(newText), 0o644)
}
