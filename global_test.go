package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSymlinkAGENTSFileAlreadyCorrect(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "shared", "AGENTS.md")
	link := filepath.Join(dir, "agents.md")
	os.MkdirAll(filepath.Dir(target), 0o755)
	os.WriteFile(target, []byte("content"), 0o644)
	os.Symlink(target, link)

	err := symlinkAGENTSFile(link, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("symlink broken: %v", err)
	}
	if got != target {
		t.Errorf("target=%q, want %q", got, target)
	}
}

func TestSymlinkAGENTSFileCreatesNew(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "shared", "AGENTS.md")
	link := filepath.Join(dir, "agents.md")
	os.MkdirAll(filepath.Dir(target), 0o755)
	os.WriteFile(target, []byte("content"), 0o644)

	err := symlinkAGENTSFile(link, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("symlink broken: %v", err)
	}
	if got != target {
		t.Errorf("target=%q, want %q", got, target)
	}
}

func TestSymlinkAGENTSFileReplacesRealFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "shared", "AGENTS.md")
	link := filepath.Join(dir, "agents.md")
	os.MkdirAll(filepath.Dir(target), 0o755)
	os.WriteFile(target, []byte("shared content"), 0o644)
	os.WriteFile(link, []byte("old content"), 0o644)

	err := symlinkAGENTSFile(link, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("symlink broken: %v", err)
	}
	if got != target {
		t.Errorf("target=%q, want %q", got, target)
	}
	// Timestamped backup should exist.
	matches, _ := filepath.Glob(link + ".bak.*")
	if len(matches) == 0 {
		t.Errorf("backup not created")
	}
}

func TestSymlinkClaudeGlobalAlreadyCorrect(t *testing.T) {
	dir := t.TempDir()
	shared := filepath.Join(dir, "shared")
	claude := filepath.Join(dir, "claude")
	os.MkdirAll(shared, 0o755)
	os.Symlink(shared, claude)

	err := symlinkClaudeGlobal(claude, shared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := os.Readlink(claude)
	if err != nil {
		t.Fatalf("symlink broken: %v", err)
	}
	if got != shared {
		t.Errorf("target=%q, want %q", got, shared)
	}
}

func TestSymlinkClaudeGlobalCreatesNew(t *testing.T) {
	dir := t.TempDir()
	shared := filepath.Join(dir, "shared")
	claude := filepath.Join(dir, "claude")
	os.MkdirAll(shared, 0o755)

	err := symlinkClaudeGlobal(claude, shared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := os.Readlink(claude)
	if err != nil {
		t.Fatalf("symlink broken: %v", err)
	}
	if got != shared {
		t.Errorf("target=%q, want %q", got, shared)
	}
}

func TestSymlinkClaudeGlobalFixesWrongTarget(t *testing.T) {
	dir := t.TempDir()
	shared := filepath.Join(dir, "shared")
	wrong := filepath.Join(dir, "wrong")
	claude := filepath.Join(dir, "claude")
	os.MkdirAll(shared, 0o755)
	os.MkdirAll(wrong, 0o755)
	os.Symlink(wrong, claude)

	err := symlinkClaudeGlobal(claude, shared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := os.Readlink(claude)
	if err != nil {
		t.Fatalf("symlink broken: %v", err)
	}
	if got != shared {
		t.Errorf("target=%q, want %q", got, shared)
	}
}
