package main

import (
	"testing"
)

func TestParseFrontmatterFull(t *testing.T) {
	content := `---
name: my-slug
description: a description
metadata:
  type: feedback
---
Some body text.`
	fm := parseFrontmatter(content)
	if fm.Name != "my-slug" {
		t.Fatalf("Name = %q, want %q", fm.Name, "my-slug")
	}
	if fm.Description != "a description" {
		t.Fatalf("Description = %q, want %q", fm.Description, "a description")
	}
	if fm.Type != "feedback" {
		t.Fatalf("Type = %q, want %q", fm.Type, "feedback")
	}
}

func TestParseFrontmatterNoFrontmatter(t *testing.T) {
	fm := parseFrontmatter("just plain text")
	if fm.Name != "" || fm.Description != "" || fm.Type != "" {
		t.Fatalf("expected empty frontmatter, got %+v", fm)
	}
}

func TestParseFrontmatterMissingEnd(t *testing.T) {
	fm := parseFrontmatter("---\nname: foo\ndescription: bar\n")
	if fm.Name != "" || fm.Description != "" {
		t.Fatalf("expected empty frontmatter for missing end marker, got %+v", fm)
	}
}

func TestParseFrontmatterQuotedValues(t *testing.T) {
	content := `---
name: "quoted-name"
description: 'single-quoted'
metadata:
  type: "project"
---`
	fm := parseFrontmatter(content)
	if fm.Name != "quoted-name" {
		t.Fatalf("Name = %q, want %q", fm.Name, "quoted-name")
	}
	if fm.Description != "single-quoted" {
		t.Fatalf("Description = %q, want %q", fm.Description, "single-quoted")
	}
	if fm.Type != "project" {
		t.Fatalf("Type = %q, want %q", fm.Type, "project")
	}
}

func TestParseFrontmatterNoMetadata(t *testing.T) {
	content := `---
name: slug
description: desc
---`
	fm := parseFrontmatter(content)
	if fm.Name != "slug" {
		t.Fatalf("Name = %q, want %q", fm.Name, "slug")
	}
	if fm.Type != "" {
		t.Fatalf("Type should be empty without metadata block, got %q", fm.Type)
	}
}

func TestUnquote(t *testing.T) {
	tests := []struct{ in, want string }{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{"hello", "hello"},
		{`"it's"`, "it's"},
	}
	for _, tc := range tests {
		got := unquote(tc.in)
		if got != tc.want {
			t.Fatalf("unquote(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSnippetAroundFound(t *testing.T) {
	content := "this is some text with worktree mentioned in it and more words here for context"
	got := snippetAround(content, "worktree")
	if len(got) == 0 {
		t.Fatal("expected non-empty snippet")
	}
}

func TestSnippetAroundNotFound(t *testing.T) {
	got := snippetAround("hello world", "xyz")
	if got != "" {
		t.Fatalf("expected empty snippet, got %q", got)
	}
}

func TestSnippetAroundNewlinesFlattened(t *testing.T) {
	content := "line one\nline two\nworktree here\nline four\nline five"
	got := snippetAround(content, "worktree")
	for _, r := range got {
		if r == '\n' {
			t.Fatalf("snippet should have newlines flattened, got %q", got)
		}
	}
}

func TestSnippetAroundBoundaryStart(t *testing.T) {
	content := "worktree at start"
	got := snippetAround(content, "worktree")
	if len(got) == 0 {
		t.Fatal("expected non-empty snippet")
	}
}

func TestSnippetAroundBoundaryEnd(t *testing.T) {
	content := "lots of text before the target area worktree"
	got := snippetAround(content, "worktree")
	if len(got) == 0 {
		t.Fatal("expected non-empty snippet")
	}
}
