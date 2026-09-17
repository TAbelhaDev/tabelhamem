package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnlinkClaudeDirMissing(t *testing.T) {
	dir := t.TempDir()
	claude := filepath.Join(dir, "nonexistent")
	shared := filepath.Join(dir, "shared")
	os.Mkdir(shared, 0o755)

	restored, err := unlinkClaudeDir(claude, shared)
	if err != nil {
		t.Fatal(err)
	}
	if restored {
		t.Fatal("expected restored=false for missing claude dir")
	}
}

func TestUnlinkClaudeDirNotSymlink(t *testing.T) {
	dir := t.TempDir()
	claude := filepath.Join(dir, "real-dir")
	shared := filepath.Join(dir, "shared")
	os.Mkdir(claude, 0o755)
	os.Mkdir(shared, 0o755)

	_, err := unlinkClaudeDir(claude, shared)
	if err == nil {
		t.Fatal("expected error for non-symlink")
	}
}

func TestUnlinkClaudeDirWrongTarget(t *testing.T) {
	dir := t.TempDir()
	claude := filepath.Join(dir, "claude-mem")
	wrong := filepath.Join(dir, "wrong")
	shared := filepath.Join(dir, "shared")
	os.Mkdir(wrong, 0o755)
	os.Mkdir(shared, 0o755)
	os.Symlink(wrong, claude)

	_, err := unlinkClaudeDir(claude, shared)
	if err == nil {
		t.Fatal("expected error for wrong target")
	}
}

func TestUnlinkClaudeDirRestoresRealCopy(t *testing.T) {
	dir := t.TempDir()
	claude := filepath.Join(dir, "claude-mem")
	shared := filepath.Join(dir, "shared")
	os.Mkdir(shared, 0o755)
	os.WriteFile(filepath.Join(shared, "mem1.md"), []byte("content1"), 0o644)
	os.WriteFile(filepath.Join(shared, "mem2.md"), []byte("content2"), 0o644)
	os.Symlink(shared, claude)

	restored, err := unlinkClaudeDir(claude, shared)
	if err != nil {
		t.Fatal(err)
	}
	if !restored {
		t.Fatal("expected restored=true")
	}
	// claude should now be a real dir, not a symlink
	fi, err := os.Lstat(claude)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		t.Fatal("claude should be a real dir after unlink")
	}
	// files should be copied
	data, _ := os.ReadFile(filepath.Join(claude, "mem1.md"))
	if string(data) != "content1" {
		t.Fatalf("mem1.md content = %q, want %q", string(data), "content1")
	}
	data, _ = os.ReadFile(filepath.Join(claude, "mem2.md"))
	if string(data) != "content2" {
		t.Fatalf("mem2.md content = %q, want %q", string(data), "content2")
	}
	// shared dir should be untouched
	if _, err := os.Stat(filepath.Join(shared, "mem1.md")); err != nil {
		t.Fatal("shared mem1.md should still exist")
	}
}

func TestRemoveAgentsSectionMissingFile(t *testing.T) {
	dir := t.TempDir()
	changed, err := removeAgentsSection(dir)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("expected changed=false for missing file")
	}
}

func TestRemoveAgentsSectionNoMarkers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	os.WriteFile(path, []byte("# Custom content\n"), 0o644)

	changed, err := removeAgentsSection(dir)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("expected changed=false when no markers present")
	}
}

func TestRemoveAgentsSectionOnlySection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	section := agentsSection("/some/path")
	content := "# AGENTS.md\n\n" + section + "\n"
	os.WriteFile(path, []byte(content), 0o644)

	changed, err := removeAgentsSection(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), agentsMarkerStart) {
		t.Fatal("markers should be removed")
	}
	if !strings.Contains(string(data), "# AGENTS.md") {
		t.Fatal("non-section content should be preserved")
	}
}

func TestRemoveAgentsSectionPreservesOtherContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	section := agentsSection("/some/path")
	content := "# Custom\n\n" + section + "\n\nMore stuff\n"
	os.WriteFile(path, []byte(content), 0o644)

	changed, err := removeAgentsSection(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}
	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), agentsMarkerStart) {
		t.Fatal("markers should be removed")
	}
	if !strings.Contains(string(data), "# Custom") {
		t.Fatal("custom content should be preserved")
	}
	if !strings.Contains(string(data), "More stuff") {
		t.Fatal("trailing content should be preserved")
	}
}

func TestRemoveAgentsSectionOnlyMarkers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "AGENTS.md")
	content := agentsSection("/some/path")
	os.WriteFile(path, []byte(content), 0o644)

	changed, err := removeAgentsSection(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("AGENTS.md should be removed when only markers remain")
	}
}

func TestCopyFlatDir(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	os.Mkdir(src, 0o755)
	os.WriteFile(filepath.Join(src, "a.md"), []byte("aaa"), 0o644)
	os.WriteFile(filepath.Join(src, "b.txt"), []byte("bbb"), 0o644)

	if err := copyFlatDir(src, dst); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dst, "a.md"))
	if string(data) != "aaa" {
		t.Fatalf("a.md = %q, want %q", string(data), "aaa")
	}
	data, _ = os.ReadFile(filepath.Join(dst, "b.txt"))
	if string(data) != "bbb" {
		t.Fatalf("b.txt = %q, want %q", string(data), "bbb")
	}
}
