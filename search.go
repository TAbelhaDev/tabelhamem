package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TAbelhaDev/tabelhascaff/ipc"
)

// searchMatch is one row of the search method's output.
type searchMatch struct {
	Project     string `json:"project"`
	File        string `json:"file"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Snippet     string `json:"snippet,omitempty"`
}

// frontmatter holds the subset of a memory file's YAML frontmatter tamem
// understands (name, description, metadata.type). Parsed by hand rather
// than pulling in a YAML dependency — the schema is fixed and simple.
type frontmatter struct {
	Name        string
	Description string
	Type        string
}

func parseFrontmatter(content string) frontmatter {
	var fm frontmatter
	if !strings.HasPrefix(content, "---\n") {
		return fm
	}
	rest := content[4:]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		return fm
	}
	inMetadata := false
	for _, line := range strings.Split(rest[:end], "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "metadata:" {
			inMetadata = true
			continue
		}
		if !strings.HasPrefix(line, " ") && trimmed != "" {
			inMetadata = false
		}
		switch {
		case strings.HasPrefix(trimmed, "name:"):
			fm.Name = unquote(strings.TrimSpace(strings.TrimPrefix(trimmed, "name:")))
		case strings.HasPrefix(trimmed, "description:"):
			fm.Description = unquote(strings.TrimSpace(strings.TrimPrefix(trimmed, "description:")))
		case inMetadata && strings.HasPrefix(trimmed, "type:"):
			fm.Type = unquote(strings.TrimSpace(strings.TrimPrefix(trimmed, "type:")))
		}
	}
	return fm
}

func unquote(s string) string {
	return strings.Trim(s, `"'`)
}

// searchMemory performs a full-text search across all bridged projects'
// memory files, returning the matches. It is the reusable core shared by the
// IPC search method and the TUI search mode.
func searchMemory(query, typeFilter, projectFilter string) []searchMatch {
	queryLower := strings.ToLower(query)

	root := filepath.Join(homeDir(), "agent-memory")
	entries, err := os.ReadDir(root)
	if err != nil {
		return []searchMatch{}
	}

	matches := make([]searchMatch, 0)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if projectFilter != "" && e.Name() != projectFilter {
			continue
		}
		dir := filepath.Join(root, e.Name())
		for _, f := range topicFiles(dir) {
			data, err := os.ReadFile(filepath.Join(dir, f))
			if err != nil {
				continue
			}
			content := string(data)
			fm := parseFrontmatter(content)
			if typeFilter != "" && fm.Type != typeFilter {
				continue
			}
			if !strings.Contains(strings.ToLower(content), queryLower) {
				continue
			}
			matches = append(matches, searchMatch{
				Project:     e.Name(),
				File:        f,
				Name:        fm.Name,
				Description: fm.Description,
				Type:        fm.Type,
				Snippet:     snippetAround(content, queryLower),
			})
		}
	}
	return matches
}

func ipcSearch(filters map[string]string) int {
	query := filters["query"]
	if query == "" {
		fmt.Fprintln(os.Stderr, "filtro query= é obrigatório")
		return 1
	}
	return ipc.WriteJSON(searchMemory(query, filters["type"], filters["project"]))
}

// searchDigestOut is the wire format for the search-digest method.
type searchDigestOut struct {
	Markdown string        `json:"markdown"`
	Matches  []searchMatch `json:"matches"`
	Count    int           `json:"count"`
}

func ipcSearchDigest(filters map[string]string) int {
	query := filters["query"]
	if query == "" {
		fmt.Fprintln(os.Stderr, "filtro query= é obrigatório")
		return 1
	}

	matches := searchMemory(query, filters["type"], filters["project"])

	var b strings.Builder
	b.WriteString("# Resultados da busca na memória\n\n")
	fmt.Fprintf(&b, "**Query:** %s\n\n", query)
	if filters["type"] != "" {
		fmt.Fprintf(&b, "**Tipo:** %s\n\n", filters["type"])
	}
	fmt.Fprintf(&b, "**Encontrados:** %d resultado(s)\n\n", len(matches))

	if len(matches) == 0 {
		b.WriteString("Nenhum resultado encontrado.\n")
	} else {
		for _, m := range matches {
			fmt.Fprintf(&b, "## %s / %s\n\n", m.Project, m.File)
			if m.Description != "" {
				fmt.Fprintf(&b, "%s\n\n", m.Description)
			}
			if m.Snippet != "" {
				fmt.Fprintf(&b, "> %s\n\n", m.Snippet)
			}
		}
	}

	return ipc.WriteJSON(searchDigestOut{
		Markdown: b.String(),
		Matches:  matches,
		Count:    len(matches),
	})
}

// snippetAround returns ~80 chars of context around the first occurrence of
// queryLower in content, newlines flattened to spaces.
func snippetAround(content, queryLower string) string {
	idx := strings.Index(strings.ToLower(content), queryLower)
	if idx == -1 {
		return ""
	}
	start := max(idx-40, 0)
	end := min(idx+len(queryLower)+40, len(content))
	return strings.TrimSpace(strings.ReplaceAll(content[start:end], "\n", " "))
}
