package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TAbelhaDev/tabelhascaff/ipc"
)

// statusResult is the wire format for the status method.
type statusResult struct {
	Project            string   `json:"project"`
	SharedDir          string   `json:"shared_dir"`
	SharedExists       bool     `json:"shared_exists"`
	TopicFiles         []string `json:"topic_files,omitempty"`
	ClaudeDir          string   `json:"claude_dir,omitempty"`
	ClaudeLinked       bool     `json:"claude_linked"`
	AgentsMdHasSection bool     `json:"agents_md_has_section"`
}

func ipcStatus(filters map[string]string) int {
	slug := filters["project"]
	if slug == "" {
		fmt.Fprintln(os.Stderr, "filtro project= é obrigatório")
		return 1
	}

	shared := sharedMemoryDir(slug)
	result := statusResult{Project: slug, SharedDir: shared}

	if info, err := os.Stat(shared); err == nil && info.IsDir() {
		result.SharedExists = true
		result.TopicFiles = topicFiles(shared)
	}

	if repo := filters["repo"]; repo != "" {
		repoAbs, err := filepath.Abs(repo)
		if err != nil {
			fmt.Fprintln(os.Stderr, "erro:", err)
			return 1
		}
		claudeDir := claudeMemoryDir(repoAbs)
		result.ClaudeDir = claudeDir
		if target, err := symlinkTarget(claudeDir); err == nil {
			result.ClaudeLinked = target == shared
		}
		if data, err := os.ReadFile(filepath.Join(repoAbs, "AGENTS.md")); err == nil {
			result.AgentsMdHasSection = strings.Contains(string(data), agentsMarkerStart)
		}
	}

	return ipc.WriteJSON(result)
}

// listEntry is one row of the list method's output.
type listEntry struct {
	Project    string `json:"project"`
	SharedDir  string `json:"shared_dir"`
	TopicCount int    `json:"topic_count"`
}

func ipcList(_ map[string]string) int {
	root := filepath.Join(homeDir(), "agent-memory")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return ipc.WriteJSON([]listEntry{})
		}
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}

	out := make([]listEntry, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		out = append(out, listEntry{
			Project:    e.Name(),
			SharedDir:  dir,
			TopicCount: len(topicFiles(dir)),
		})
	}
	return ipc.WriteJSON(out)
}

// topicFiles lists the memory topic files in dir (every .md file except the
// MEMORY.md index itself).
func topicFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.Name() != "MEMORY.md" && strings.HasSuffix(e.Name(), ".md") {
			out = append(out, e.Name())
		}
	}
	return out
}
