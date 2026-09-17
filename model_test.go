package main

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	m := newModel()
	if m.width != 0 || m.height != 0 {
		t.Errorf("width=%d height=%d, want 0", m.width, m.height)
	}
	if m.focus != focusProjects {
		t.Errorf("focus=%v, want focusProjects", m.focus)
	}
	if m.mode != modeBrowse {
		t.Errorf("mode=%v, want modeBrowse", m.mode)
	}
}

func TestDiscoverProjectsEmpty(t *testing.T) {
	saved := settings.Projects
	settings.Projects = nil
	defer func() { settings.Projects = saved }()

	projects := discoverProjects()
	if len(projects) != 0 {
		t.Errorf("got %d projects, want 0", len(projects))
	}
}

func TestDiscoverProjectsFromConfig(t *testing.T) {
	saved := settings.Projects
	settings.Projects = []projectEntry{
		{Slug: "test-a", Repo: filepath.Join(t.TempDir(), "a")},
		{Slug: "test-b", Repo: filepath.Join(t.TempDir(), "b")},
	}
	defer func() { settings.Projects = saved }()

	projects := discoverProjects()
	if len(projects) != 2 {
		t.Fatalf("got %d projects, want 2", len(projects))
	}
	if projects[0].slug != "test-a" {
		t.Errorf("slug=%q, want test-a", projects[0].slug)
	}
	if projects[1].slug != "test-b" {
		t.Errorf("slug=%q, want test-b", projects[1].slug)
	}
}

func TestProjectRows(t *testing.T) {
	projects := []projectInfo{
		{slug: "alpha", claudeLinked: true, opencodeLinked: true},
		{slug: "beta", claudeLinked: false, opencodeLinked: false},
	}
	rows := projectRows(projects)
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	if rows[0][1] != "alpha" {
		t.Errorf("row[0][1]=%v, want alpha", rows[0][1])
	}
	if rows[1][1] != "beta" {
		t.Errorf("row[1][1]=%v, want beta", rows[1][1])
	}
}

func TestModelWindowSize(t *testing.T) {
	m := newModel()
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	model, _ := m.Update(msg)
	app := model.(appModel)
	if app.width != 120 {
		t.Errorf("width=%d, want 120", app.width)
	}
	if app.height != 40 {
		t.Errorf("height=%d, want 40", app.height)
	}
}

func TestModelInit(t *testing.T) {
	m := newModel()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestModelQuit(t *testing.T) {
	m := newModel()
	m.width, m.height = 120, 40
	m.layout()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	model, _ := m.Update(msg)
	// quit returns a tea.Quit cmd; the model value is unchanged type-wise.
	if _, ok := model.(appModel); !ok {
		t.Error("expected appModel after quit")
	}
}

func TestRefreshMemoryPanelNoProject(t *testing.T) {
	m := newModel()
	m.width, m.height = 120, 40
	m.layout()

	// Simulate no selected project by clearing the projects list.
	m.projects = nil
	m.refreshProjectTable()
	m.refreshMemoryPanel()

	if len(m.memoryFiles) != 0 {
		t.Errorf("got %d memory files, want 0", len(m.memoryFiles))
	}
	if m.activeFile != "" {
		t.Errorf("activeFile=%q, want empty", m.activeFile)
	}
}
