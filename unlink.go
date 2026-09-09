package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TAbelhaDev/tabelhascaff/ipc"
)

// unlinkResult is the wire format for the unlink method.
type unlinkResult struct {
	Project         string `json:"project"`
	Repo            string `json:"repo"`
	ClaudeDir       string `json:"claude_dir"`
	Restored        bool   `json:"restored"`
	AgentsMdUpdated bool   `json:"agents_md_updated"`
}

// unlink reverses link: the shared directory at ~/agent-memory/<slug>/ is
// left untouched (other worktrees or tools may still depend on it) — only
// this one repo's Claude Code memory symlink is replaced with a real copy
// of its current content, and the AGENTS.md bridge section is removed.
func ipcUnlink(filters map[string]string) int {
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

	shared := sharedMemoryDir(slug)
	claudeDir := claudeMemoryDir(repoAbs)
	result := unlinkResult{Project: slug, Repo: repoAbs, ClaudeDir: claudeDir}

	restored, err := unlinkClaudeDir(claudeDir, shared)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	result.Restored = restored

	updated, err := removeAgentsSection(repoAbs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro atualizando AGENTS.md:", err)
		return 1
	}
	result.AgentsMdUpdated = updated

	return ipc.WriteJSON(result)
}

// unlinkClaudeDir replaces claudeDir (expected to be a symlink to shared)
// with a real directory holding a copy of shared's current content. Returns
// false (no-op, not an error) if claudeDir doesn't exist at all.
func unlinkClaudeDir(claudeDir, shared string) (bool, error) {
	info, err := os.Lstat(claudeDir)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return false, fmt.Errorf("%s não é um symlink — nada a desfazer", claudeDir)
	}

	target, err := symlinkTarget(claudeDir)
	if err != nil {
		return false, err
	}
	if target != shared {
		return false, fmt.Errorf("%s aponta pra %s, não pra %s — recuso mexer", claudeDir, target, shared)
	}

	if err := os.Remove(claudeDir); err != nil {
		return false, err
	}
	if err := copyFlatDir(shared, claudeDir); err != nil {
		return false, err
	}
	return true, nil
}

// copyFlatDir copies every regular file directly under src into a freshly
// created dst directory. Memory stores are flat (no subdirectories), so
// nested entries are skipped rather than recursed into.
func copyFlatDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// removeAgentsSection strips the tamem-owned block from <repo>/AGENTS.md
// (added by ensureAgentsSection), removing the file entirely if that leaves
// it empty. No-op (false, nil) if the file or the block doesn't exist.
func removeAgentsSection(repo string) (bool, error) {
	path := filepath.Join(repo, "AGENTS.md")
	existing, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	text := string(existing)
	startIdx := strings.Index(text, agentsMarkerStart)
	endIdx := strings.Index(text, agentsMarkerEnd)
	if startIdx == -1 || endIdx == -1 {
		return false, nil
	}
	endIdx += len(agentsMarkerEnd)

	prefix := strings.TrimRight(text[:startIdx], "\n")
	suffix := strings.TrimLeft(text[endIdx:], "\n")

	var newText string
	switch {
	case prefix == "" && suffix == "":
		newText = ""
	case prefix == "":
		newText = suffix
	case suffix == "":
		newText = prefix + "\n"
	default:
		newText = prefix + "\n\n" + suffix
	}

	if newText == "" {
		return true, os.Remove(path)
	}
	return true, os.WriteFile(path, []byte(newText), 0o644)
}
