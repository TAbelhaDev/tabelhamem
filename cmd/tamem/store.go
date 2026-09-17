package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// homeDir resolves $HOME.
func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

// claudeMemoryDir applies Claude Code's own escaping rule (every '/' and '.'
// in the absolute repo path becomes '-') to find where it keeps that
// project's memory: ~/.claude/projects/<encoded>/memory.
func claudeMemoryDir(repoAbs string) string {
	encoded := strings.NewReplacer("/", "-", ".", "-").Replace(repoAbs)
	return filepath.Join(homeDir(), ".claude", "projects", encoded, "memory")
}

// sharedMemoryDir is the canonical, tool-agnostic memory store for a logical
// project, shared between Claude Code (via symlink) and OpenCode (via an
// AGENTS.md instruction).
func sharedMemoryDir(slug string) string {
	return filepath.Join(homeDir(), "agent-memory", slug)
}

// symlinkTarget resolves what path an existing symlink at p points to,
// returning it as an absolute path.
func symlinkTarget(p string) (string, error) {
	target, err := os.Readlink(p)
	if err != nil {
		return "", err
	}
	if filepath.IsAbs(target) {
		return target, nil
	}
	return filepath.Join(filepath.Dir(p), target), nil
}

// moveFile renames src to dst, falling back to copy+remove when they live on
// different filesystems (os.Rename returns EXDEV in that case).
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	if info, err := os.Stat(src); err == nil {
		_ = os.Chmod(dst, info.Mode())
	}
	return os.Remove(src)
}

// gitWorktrees returns every worktree path (including the main one) for a
// git repo at repo. If git is unavailable or repo is not a git repo it
// falls back to returning []string{repo} so the caller always gets at least
// the original path.
func gitWorktrees(repo string) []string {
	out, err := exec.Command("git", "-C", repo, "worktree", "list", "--porcelain").Output()
	if err != nil {
		return []string{repo}
	}
	var paths []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "worktree ") {
			p := strings.TrimPrefix(line, "worktree ")
			if p != "" {
				paths = append(paths, p)
			}
		}
	}
	if len(paths) == 0 {
		return []string{repo}
	}
	return paths
}
