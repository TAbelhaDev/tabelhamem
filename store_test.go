package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeMemoryDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := claudeMemoryDir("/home/user/my.project/repo")
	want := filepath.Join(home, ".claude", "projects", "-home-user-my-project-repo", "memory")
	if got != want {
		t.Fatalf("claudeMemoryDir = %q, want %q", got, want)
	}
}

func TestClaudeMemoryDirNestedDots(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := claudeMemoryDir("/a.b/c.d")
	want := filepath.Join(home, ".claude", "projects", "-a-b-c-d", "memory")
	if got != want {
		t.Fatalf("claudeMemoryDir = %q, want %q", got, want)
	}
}

func TestSharedMemoryDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := sharedMemoryDir("my-project")
	want := filepath.Join(home, "agent-memory", "my-project")
	if got != want {
		t.Fatalf("sharedMemoryDir = %q, want %q", got, want)
	}
}

func TestSymlinkTargetAbs(t *testing.T) {
	dir := t.TempDir()
	link := filepath.Join(dir, "link")
	target := filepath.Join(dir, "target")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	got, err := symlinkTarget(link)
	if err != nil {
		t.Fatal(err)
	}
	if got != target {
		t.Fatalf("symlinkTarget = %q, want %q", got, target)
	}
}

func TestSymlinkTargetRelative(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "link")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("target", link); err != nil {
		t.Fatal(err)
	}
	got, err := symlinkTarget(link)
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(dir, "target") {
		t.Fatalf("symlinkTarget = %q, want %q", got, filepath.Join(dir, "target"))
	}
}

func TestMoveFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.txt")
	dst := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := moveFile(src, dst); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("moved content = %q, want %q", string(data), "hello")
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("src should be gone, err = %v", err)
	}
}

func TestTopicFiles(t *testing.T) {
	dir := t.TempDir()
	files := []string{"MEMORY.md", "feedback_x.md", "project_a.md", "not-md.txt", "reference_b.md"}
	for _, f := range files {
		os.WriteFile(filepath.Join(dir, f), []byte("content"), 0o644)
	}
	got := topicFiles(dir)
	want := []string{"feedback_x.md", "project_a.md", "reference_b.md"}
	if len(got) != len(want) {
		t.Fatalf("topicFiles = %v (len %d), want %v (len %d)", got, len(got), want, len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("topicFiles[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestTopicFilesEmpty(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte("index"), 0o644)
	got := topicFiles(dir)
	if len(got) != 0 {
		t.Fatalf("topicFiles on MEMORY.md-only dir = %v, want empty", got)
	}
}

func TestGitWorktreesFallback(t *testing.T) {
	dir := t.TempDir()
	got := gitWorktrees(dir)
	if len(got) != 1 || got[0] != dir {
		t.Fatalf("gitWorktrees on non-repo = %v, want [%s]", got, dir)
	}
}
