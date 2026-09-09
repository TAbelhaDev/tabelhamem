package main

import (
	"io"
	"os"
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
