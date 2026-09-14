package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinkClaudeDirNonexistent(t *testing.T) {
	dir := t.TempDir()
	claude := filepath.Join(dir, "claude-mem")
	shared := filepath.Join(dir, "shared")
	os.Mkdir(shared, 0o755)

	migrated, alreadyLinked, err := linkClaudeDir(claude, shared)
	if err != nil {
		t.Fatal(err)
	}
	if migrated != nil {
		t.Fatalf("expected nil migrated, got %v", migrated)
	}
	if alreadyLinked {
		t.Fatal("expected alreadyLinked=false")
	}
	if _, err := os.Lstat(claude); err != nil {
		t.Fatalf("claude dir should be a symlink, err: %v", err)
	}
	target, _ := os.Readlink(claude)
	if target != shared {
		t.Fatalf("symlink target = %q, want %q", target, shared)
	}
}

func TestLinkClaudeDirAlreadyLinked(t *testing.T) {
	dir := t.TempDir()
	claude := filepath.Join(dir, "claude-mem")
	shared := filepath.Join(dir, "shared")
	os.Mkdir(shared, 0o755)
	os.Symlink(shared, claude)

	migrated, alreadyLinked, err := linkClaudeDir(claude, shared)
	if err != nil {
		t.Fatal(err)
	}
	if migrated != nil {
		t.Fatalf("expected nil migrated, got %v", migrated)
	}
	if !alreadyLinked {
		t.Fatal("expected alreadyLinked=true")
	}
}

func TestLinkClaudeDirWrongTarget(t *testing.T) {
	dir := t.TempDir()
	claude := filepath.Join(dir, "claude-mem")
	wrong := filepath.Join(dir, "wrong")
	os.Mkdir(wrong, 0o755)
	os.Symlink(wrong, claude)

	shared := filepath.Join(dir, "shared")
	os.Mkdir(shared, 0o755)

	_, _, err := linkClaudeDir(claude, shared)
	if err == nil {
		t.Fatal("expected error for wrong symlink target")
	}
}

func TestLinkClaudeDirMigratesRealDir(t *testing.T) {
	dir := t.TempDir()
	claude := filepath.Join(dir, "claude-mem")
	shared := filepath.Join(dir, "shared")
	os.Mkdir(shared, 0o755)
	os.Mkdir(claude, 0o755)
	os.WriteFile(filepath.Join(claude, "file1.md"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(claude, "file2.md"), []byte("b"), 0o644)

	migrated, alreadyLinked, err := linkClaudeDir(claude, shared)
	if err != nil {
		t.Fatal(err)
	}
	if alreadyLinked {
		t.Fatal("expected alreadyLinked=false")
	}
	if len(migrated) != 2 {
		t.Fatalf("expected 2 migrated files, got %d", len(migrated))
	}
	// claude should now be a symlink
	if _, err := os.Lstat(claude); err != nil {
		t.Fatalf("claude should exist, err: %v", err)
	}
	fi, _ := os.Lstat(claude)
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatal("claude should be a symlink after migration")
	}
	// files should be in shared
	if _, err := os.Stat(filepath.Join(shared, "file1.md")); err != nil {
		t.Fatal("file1.md should be in shared")
	}
	if _, err := os.Stat(filepath.Join(shared, "file2.md")); err != nil {
		t.Fatal("file2.md should be in shared")
	}
}

func TestLinkClaudeDirConflict(t *testing.T) {
	dir := t.TempDir()
	claude := filepath.Join(dir, "claude-mem")
	shared := filepath.Join(dir, "shared")
	os.Mkdir(shared, 0o755)
	os.WriteFile(filepath.Join(shared, "conflict.md"), []byte("exists"), 0o644)
	os.Mkdir(claude, 0o755)
	os.WriteFile(filepath.Join(claude, "conflict.md"), []byte("new"), 0o644)

	_, _, err := linkClaudeDir(claude, shared)
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(err.Error(), "recuso sobrescrever") {
		t.Fatalf("expected conflict message, got: %v", err)
	}
}

func TestAgentsSectionContainsMarkers(t *testing.T) {
	section := agentsSection("/some/shared/path")
	if !strings.Contains(section, agentsMarkerStart) {
		t.Fatal("missing marker start")
	}
	if !strings.Contains(section, agentsMarkerEnd) {
		t.Fatal("missing marker end")
	}
	if !strings.Contains(section, "/some/shared/path") {
		t.Fatal("missing shared path")
	}
}

func TestEnsureAgentsSectionCreatesNew(t *testing.T) {
	dir := t.TempDir()
	shared := filepath.Join(dir, "shared")
	os.Mkdir(shared, 0o755)

	changed, err := ensureAgentsSection(dir, shared)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed=true for new file")
	}
	data, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(string(data), agentsMarkerStart) {
		t.Fatal("AGENTS.md should contain marker start")
	}
}

func TestEnsureAgentsSectionIdempotent(t *testing.T) {
	dir := t.TempDir()
	shared := filepath.Join(dir, "shared")
	os.Mkdir(shared, 0o755)

	ensureAgentsSection(dir, shared)
	changed, err := ensureAgentsSection(dir, shared)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("expected changed=false for idempotent call")
	}
}

func TestEnsureAgentsSectionReplacesStale(t *testing.T) {
	dir := t.TempDir()
	shared := filepath.Join(dir, "shared")
	os.Mkdir(shared, 0o755)
	path := filepath.Join(dir, "AGENTS.md")

	// Write section with a different shared path
	content := "# AGENTS.md\n\n" + agentsSection("/old/path") + "\n"
	os.WriteFile(path, []byte(content), 0o644)

	changed, err := ensureAgentsSection(dir, shared)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed=true for stale path")
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "/old/path") {
		t.Fatal("stale path should have been replaced")
	}
	if !strings.Contains(string(data), shared) {
		t.Fatal("new shared path should be present")
	}
}

func TestEnsureAgentsSectionPreservesOtherContent(t *testing.T) {
	dir := t.TempDir()
	shared := filepath.Join(dir, "shared")
	os.Mkdir(shared, 0o755)
	path := filepath.Join(dir, "AGENTS.md")

	otherContent := "# My custom AGENTS.md\n\nSome other instructions.\n"
	os.WriteFile(path, []byte(otherContent), 0o644)

	changed, err := ensureAgentsSection(dir, shared)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "My custom AGENTS.md") {
		t.Fatal("other content should be preserved")
	}
	if !strings.Contains(string(data), agentsMarkerStart) {
		t.Fatal("marker should be appended")
	}
}
