package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTopicFilesExcludesDirs(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, "feedback_x.md"), []byte(""), 0o644)
	os.Mkdir(filepath.Join(dir, "subdir"), 0o755)

	got := topicFiles(dir)
	if len(got) != 1 || got[0] != "feedback_x.md" {
		t.Fatalf("topicFiles = %v, want [feedback_x.md]", got)
	}
}

func TestTopicFilesNoMd(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(dir, "data.json"), []byte(""), 0o644)

	got := topicFiles(dir)
	if len(got) != 0 {
		t.Fatalf("topicFiles = %v, want empty", got)
	}
}

func TestTopicFilesMultipleMd(t *testing.T) {
	dir := t.TempDir()
	names := []string{"a.md", "b.md", "c.md", "MEMORY.md"}
	for _, n := range names {
		os.WriteFile(filepath.Join(dir, n), []byte(""), 0o644)
	}

	got := topicFiles(dir)
	if len(got) != 3 {
		t.Fatalf("topicFiles = %v (len %d), want 3 entries", got, len(got))
	}
}
