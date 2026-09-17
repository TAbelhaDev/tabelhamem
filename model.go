package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tuiui "github.com/TAbelhaDev/tabelhatuiui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// panelFocus selects which interactive panel receives key input.
type panelFocus int

const (
	focusProjects panelFocus = iota
	focusMemory
)

// mode is the TUI's top-level operational state.
type mode int

const (
	modeBrowse mode = iota
	modeSearch
	modeLink
	modeUnlink
)

// Vertical overhead, in lines.
const (
	headerLines    = 1
	footerLines    = 1
	boxOverhead    = 2 + 1 // border + title
	minVisibleRows = 3
	minDetailLines = 4
	panelGap       = 1
)

// projectInfo holds the TUI view of a single project.
type projectInfo struct {
	slug           string
	repo           string
	sharedDir      string
	topicCount     int
	claudeLinked   bool
	opencodeLinked bool
	agentsSection  bool
	worktreeCount  int
}

type appModel struct {
	projects []projectInfo
	tbl      table.Model
	mTbl     table.Model // memory files table

	focus          panelFocus
	mode           mode
	contentFocused bool // true = j/k scroll content, false = j/k navigate table

	// content for the memory panel
	memoryFiles   []string
	contentLines  []string
	contentScroll int
	activeFile    string

	// search
	searchInput textinput.Model
	searchQuery string
	results     []searchMatch
	resultIdx   int

	// link/unlink form
	linkSlug   textinput.Model
	linkRepo   textinput.Model
	linkFocus  int // 0=slug, 1=repo
	formStatus string

	width  int
	height int

	status string

	helpModal     *tuiui.HelpModal
	settingsModal *tuiui.SettingsModal
}

func newModel() appModel {
	_ = reg.Load()

	slugInput := textinput.New()
	slugInput.Placeholder = "slug do projeto"
	slugInput.CharLimit = 64
	slugInput.Width = 30

	repoInput := textinput.New()
	repoInput.Placeholder = "caminho do repo"
	repoInput.CharLimit = 256
	repoInput.Width = 50

	searchInput := textinput.New()
	searchInput.Placeholder = "buscar memória..."
	searchInput.CharLimit = 128
	searchInput.Width = 60

	m := appModel{
		linkSlug:    slugInput,
		linkRepo:    repoInput,
		linkFocus:   0,
		searchInput: searchInput,
		helpModal: tuiui.NewHelpModal(tuiui.HelpSection{
			Title:      "Atalhos",
			BindingsFn: reg.Bindings,
		}),
		settingsModal: tuiui.NewSettingsModal(reg),
		focus:         focusProjects,
	}

	m.tbl = table.New(table.WithFocused(true))
	m.mTbl = table.New(table.WithFocused(true))
	m.tbl.SetColumns([]table.Column{{Title: "", Width: 2}, {Title: "Projeto", Width: 30}})
	m.mTbl.SetColumns([]table.Column{{Title: "Arquivos", Width: 40}})
	m.applyStyles()
	m.rescan()
	return m
}

func (m *appModel) applyStyles() {
	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		Foreground(colSubtext0).
		Background(colMantle).
		Bold(true)
	styles.Selected = styles.Selected.
		Foreground(colBase).
		Background(colPrimary).
		Bold(true)
	styles.Cell = styles.Cell.
		Foreground(colText)
	m.tbl.SetStyles(styles)
	m.mTbl.SetStyles(styles)
}

func (m *appModel) rescan() {
	m.projects = discoverProjects()
	// Also add orphan projects not in config.
	configured := make(map[string]bool)
	for _, p := range settings.Projects {
		configured[p.Slug] = true
	}

	agentRoot := filepath.Join(homeDir(), "agent-memory")
	entries, err := os.ReadDir(agentRoot)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() || configured[e.Name()] {
				continue
			}
			dir := filepath.Join(agentRoot, e.Name())
			m.projects = append(m.projects, projectInfo{
				slug:       e.Name(),
				sharedDir:  dir,
				topicCount: len(topicFiles(dir)),
			})
		}
	}

	m.refreshProjectTable()
	m.refreshMemoryPanel()
	m.status = fmt.Sprintf("%d projetos", len(m.projects))
}

// discoverProjects builds the project list from config entries.
func discoverProjects() []projectInfo {
	out := make([]projectInfo, 0, len(settings.Projects))
	for _, pe := range settings.Projects {
		shared := sharedMemoryDir(pe.Slug)
		repoAbs, err := filepath.Abs(pe.Repo)
		if err != nil {
			repoAbs = pe.Repo
		}

		info := projectInfo{
			slug:      pe.Slug,
			repo:      repoAbs,
			sharedDir: shared,
		}

		if dir, err := os.Stat(shared); err == nil && dir.IsDir() {
			info.topicCount = len(topicFiles(shared))
		}

		if pe.Repo != "" {
			wts := gitWorktrees(repoAbs)
			info.worktreeCount = len(wts)
			for _, wt := range wts {
				wtClaude := claudeMemoryDir(wt)
				if target, err := symlinkTarget(wtClaude); err == nil && target == shared {
					info.claudeLinked = true
				}
				if data, err := os.ReadFile(filepath.Join(wt, "AGENTS.md")); err == nil {
					if strings.Contains(string(data), agentsMarkerStart) {
						info.agentsSection = true
						info.opencodeLinked = true
					}
				}
			}
		}
		out = append(out, info)
	}
	return out
}

func projectRows(projects []projectInfo) []table.Row {
	rows := make([]table.Row, len(projects))
	for i, p := range projects {
		glyph := bridgeGlyph(p.claudeLinked, p.opencodeLinked)
		rows[i] = table.Row{glyph, p.slug}
	}
	return rows
}

func (m *appModel) refreshProjectTable() {
	m.tbl.SetColumns([]table.Column{
		{Title: "", Width: 2},
		{Title: "Projeto", Width: 30},
	})
	m.tbl.SetRows(projectRows(m.projects))

	idx := 0
	for i, p := range m.projects {
		if p.slug == m.currentSlug() {
			idx = i
			break
		}
	}
	m.tbl.SetCursor(idx)
}

func (m *appModel) refreshMemoryPanel() {
	m.contentFocused = false
	m.contentScroll = 0
	p := m.current()
	if p == nil {
		m.memoryFiles = nil
		m.activeFile = ""
		m.contentLines = nil
		m.mTbl.SetRows(nil)
		return
	}

	if p.sharedDir != "" {
		if _, err := os.Stat(p.sharedDir); err == nil {
			m.memoryFiles = topicFiles(p.sharedDir)
		}
	}

	rows := make([]table.Row, len(m.memoryFiles))
	for i, f := range m.memoryFiles {
		rows[i] = table.Row{f}
	}
	m.mTbl.SetColumns([]table.Column{{Title: "Arquivos", Width: 40}})
	m.mTbl.SetRows(rows)

	// Load content for the first file if any.
	if len(m.memoryFiles) > 0 && m.activeFile == "" {
		m.loadFileContent(m.memoryFiles[0])
	}
	m.mTbl.SetCursor(0)
}

func (m *appModel) loadFileContent(name string) {
	p := m.current()
	if p == nil {
		return
	}
	data, err := os.ReadFile(filepath.Join(p.sharedDir, name))
	if err != nil {
		m.contentLines = []string{theme.Error().Render("erro lendo " + name)}
		m.activeFile = name
		m.contentScroll = 0
		return
	}
	wrapped := tuiui.WrapText(string(data), m.rightInnerWidth())
	m.contentLines = strings.Split(strings.TrimRight(wrapped, "\n"), "\n")
	m.activeFile = name
	m.contentScroll = 0
}

func (m *appModel) current() *projectInfo {
	idx := m.tbl.Cursor()
	if idx < 0 || idx >= len(m.projects) {
		return nil
	}
	return &m.projects[idx]
}

func (m *appModel) currentSlug() string {
	if p := m.current(); p != nil {
		return p.slug
	}
	return ""
}

func (m *appModel) rightInnerWidth() int {
	total := m.width - panelGap
	if total < 40 {
		total = 40
	}
	share := settings.Layout.SidebarWidthShare + settings.Layout.RightWidthShare
	w := total * settings.Layout.RightWidthShare / share
	if w < 14 {
		w = 14
	}
	return w
}

func (m *appModel) sidebarInnerWidth() int {
	total := m.width - panelGap
	if total < 40 {
		total = 40
	}
	share := settings.Layout.SidebarWidthShare + settings.Layout.RightWidthShare
	w := total * settings.Layout.SidebarWidthShare / share
	if w < 14 {
		w = 14
	}
	return w
}

func (m appModel) Init() tea.Cmd { return nil }

type editorFinishedMsg struct{}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = sizeMsg.Width, sizeMsg.Height
		m.helpModal.SetSize(sizeMsg.Width, sizeMsg.Height)
		m.settingsModal.SetSize(sizeMsg.Width, sizeMsg.Height)
		m.layout()
		return m, nil
	}
	if _, ok := msg.(editorFinishedMsg); ok {
		m.rescan()
		m.layout()
		return m, nil
	}

	if m.settingsModal.Update(msg) {
		return m, nil
	}
	if m.helpModal.Update(msg) {
		return m, nil
	}

	switch m.mode {
	case modeSearch:
		return m.updateSearch(msg)
	case modeLink, modeUnlink:
		return m.updateForm(msg)
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m.forwardToTable(msg)
	}

	switch {
	case key.Matches(keyMsg, resolve("quit")):
		return m, tea.Quit
	case key.Matches(keyMsg, resolve("help")):
		m.helpModal.Toggle()
		return m, nil
	case key.Matches(keyMsg, resolve("settings")):
		m.settingsModal.Toggle()
		return m, nil
	case key.Matches(keyMsg, resolve("refresh")):
		m.rescan()
		m.layout()
		return m, nil
	case key.Matches(keyMsg, resolve("reload")):
		changed, err := reloadSettings()
		m.rescan()
		m.layout()
		switch {
		case err != nil:
			m.status = "config: " + err.Error()
		case changed:
			m.status += " — config recarregada"
		}
		return m, nil
	case key.Matches(keyMsg, resolve("nav")):
		navKeys := resolve("nav").Keys()
		switch {
		case len(navKeys) > 0 && keyMsg.String() == navKeys[0]:
			m.focusLeft()
		case len(navKeys) > 1 && keyMsg.String() == navKeys[1]:
			m.focusRight()
		}
		return m, nil
	case key.Matches(keyMsg, resolve("open")):
		if m.focus == focusMemory && m.activeFile != "" {
			m.contentFocused = !m.contentFocused
			return m, nil
		}
		return m, nil
	case key.Matches(keyMsg, resolve("search")):
		m.mode = modeSearch
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		m.results = nil
		return m, textinput.Blink
	case key.Matches(keyMsg, resolve("link")):
		if m.focus == focusProjects {
			m.mode = modeLink
			m.linkSlug.SetValue("")
			m.linkRepo.SetValue("")
			m.linkSlug.Focus()
			m.linkFocus = 0
			m.formStatus = ""
			return m, textinput.Blink
		}
		return m, nil
	case key.Matches(keyMsg, resolve("unlink")):
		if m.focus == focusProjects && m.current() != nil {
			m.mode = modeUnlink
			m.formStatus = "desligar " + m.currentSlug() + "? (s/n)"
			return m, nil
		}
		return m, nil
	case key.Matches(keyMsg, resolve("back")):
		return m, tea.Quit
	}

	if m.focus == focusMemory {
		return m.forwardToMemory(msg)
	}
	return m.forwardToTable(msg)
}

func (m *appModel) focusLeft() {
	if m.focus == focusMemory {
		m.focus = focusProjects
	}
}

func (m *appModel) focusRight() {
	if m.focus == focusProjects {
		m.focus = focusMemory
	}
}

func (m appModel) forwardToTable(msg tea.Msg) (tea.Model, tea.Cmd) {
	prevIdx := m.tbl.Cursor()
	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	if m.tbl.Cursor() != prevIdx {
		m.contentScroll = 0
		m.refreshMemoryPanel()
	}
	return m, cmd
}

func (m appModel) forwardToMemory(msg tea.Msg) (tea.Model, tea.Cmd) {
	// When content is focused, j/k/up/down scroll the content viewer.
	if m.contentFocused {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "j", "down":
				if m.contentScroll < len(m.contentLines)-1 {
					m.contentScroll++
				}
				return m, nil
			case "k", "up":
				if m.contentScroll > 0 {
					m.contentScroll--
				}
				return m, nil
			case "esc":
				m.contentFocused = false
				return m, nil
			}
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.mTbl, cmd = m.mTbl.Update(msg)
	idx := m.mTbl.Cursor()
	if idx >= 0 && idx < len(m.memoryFiles) && m.memoryFiles[idx] != m.activeFile {
		m.loadFileContent(m.memoryFiles[idx])
	}
	return m, cmd
}

func (m appModel) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	switch keyMsg.String() {
	case "esc":
		m.mode = modeBrowse
		m.searchQuery = ""
		m.results = nil
		m.searchInput.SetValue("")
		return m, nil
	case "enter":
		m.searchQuery = m.searchInput.Value()
		if m.searchQuery != "" {
			m.results = searchMemory(m.searchQuery, "", "")
			m.resultIdx = 0
		}
		return m, nil
	case "down":
		if m.resultIdx < len(m.results)-1 {
			m.resultIdx++
		}
	case "up":
		if m.resultIdx > 0 {
			m.resultIdx--
		}
	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m appModel) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		if m.mode == modeLink {
			if m.linkFocus == 0 {
				var cmd tea.Cmd
				m.linkSlug, cmd = m.linkSlug.Update(msg)
				return m, cmd
			}
			var cmd tea.Cmd
			m.linkRepo, cmd = m.linkRepo.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	if m.mode == modeLink {
		switch keyMsg.String() {
		case "esc":
			m.mode = modeBrowse
			return m, nil
		case "tab":
			m.linkFocus = 1 - m.linkFocus
			if m.linkFocus == 0 {
				m.linkSlug.Focus()
				m.linkRepo.Blur()
			} else {
				m.linkRepo.Focus()
				m.linkSlug.Blur()
			}
			return m, nil
		case "enter":
			slug := strings.TrimSpace(m.linkSlug.Value())
			repo := strings.TrimSpace(m.linkRepo.Value())
			if slug == "" || repo == "" {
				m.formStatus = "slug e repo são obrigatórios"
				return m, nil
			}
			if err := doLink(slug, repo); err != nil {
				m.formStatus = err.Error()
				return m, nil
			}
			_ = appendProject(slug, repo)
			m.mode = modeBrowse
			m.rescan()
			m.layout()
			return m, nil
		default:
			var cmd tea.Cmd
			if m.linkFocus == 0 {
				m.linkSlug, cmd = m.linkSlug.Update(msg)
			} else {
				m.linkRepo, cmd = m.linkRepo.Update(msg)
			}
			return m, cmd
		}
	}

	// modeUnlink
	if keyMsg.String() == "s" || keyMsg.String() == "S" {
		p := m.current()
		if p != nil {
			if p.repo == "" {
				m.formStatus = "sem repo configurado — ligue o projeto antes"
				return m, nil
			}
			if err := doUnlink(p.slug, p.repo); err != nil {
				m.formStatus = err.Error()
				return m, nil
			}
			_ = removeProject(p.slug)
			m.mode = modeBrowse
			m.rescan()
			m.layout()
		}
		return m, nil
	}
	if keyMsg.String() == "n" || keyMsg.String() == "N" || keyMsg.String() == "esc" {
		m.mode = modeBrowse
		return m, nil
	}
	return m, nil
}

// doLink runs the link operation (symlink + AGENTS.md) and returns any error.
func doLink(slug, repo string) error {
	repoAbs, err := filepath.Abs(repo)
	if err != nil {
		return fmt.Errorf("erro: %w", err)
	}
	if info, err := os.Stat(repoAbs); err != nil || !info.IsDir() {
		return fmt.Errorf("repo %q não é um diretório válido", repoAbs)
	}

	shared := sharedMemoryDir(slug)
	if err := os.MkdirAll(shared, 0o755); err != nil {
		return fmt.Errorf("erro criando %s: %w", shared, err)
	}

	worktrees := gitWorktrees(repoAbs)
	for _, wt := range worktrees {
		wtClaude := claudeMemoryDir(wt)
		if _, _, err := linkClaudeDir(wtClaude, shared); err != nil {
			return fmt.Errorf("erro em %s: %w", wt, err)
		}
		if _, err := ensureAgentsSection(wt, shared); err != nil {
			return fmt.Errorf("erro atualizando AGENTS.md em %s: %w", wt, err)
		}
	}
	return nil
}

// doUnlink runs the unlink operation (restore + remove AGENTS.md) and returns any error.
func doUnlink(slug, repo string) error {
	repoAbs, err := filepath.Abs(repo)
	if err != nil {
		return fmt.Errorf("erro: %w", err)
	}
	shared := sharedMemoryDir(slug)

	worktrees := gitWorktrees(repoAbs)
	for _, wt := range worktrees {
		wtClaude := claudeMemoryDir(wt)
		if _, err := unlinkClaudeDir(wtClaude, shared); err != nil {
			return fmt.Errorf("erro em %s: %w", wt, err)
		}
		if _, err := removeAgentsSection(wt); err != nil {
			return fmt.Errorf("erro atualizando AGENTS.md em %s: %w", wt, err)
		}
	}
	return nil
}

// openEditor suspends the TUI to run the editor against the file.
func openEditor(fullPath string) tea.Cmd {
	editor := settings.General.Editor
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "nvim"
	}
	cmd := exec.Command(editor, fullPath)
	cmd.Dir = filepath.Dir(fullPath)
	return tea.ExecProcess(cmd, func(err error) tea.Msg { return editorFinishedMsg{} })
}

func (m *appModel) layout() {
	if m.width == 0 || m.height == 0 {
		return
	}

	// Recompute sidebar column widths.
	sidebarW := m.sidebarInnerWidth()
	if sidebarW < 14 {
		sidebarW = 14
	}
	m.tbl.SetColumns([]table.Column{
		{Title: "", Width: 2},
		{Title: "Projeto", Width: sidebarW - 4},
	})
	m.tbl.SetWidth(sidebarW)

	bodyHeight := m.height - headerLines - footerLines
	minBody := boxOverhead + minVisibleRows + boxOverhead + minDetailLines
	if bodyHeight < minBody {
		bodyHeight = minBody
	}

	sidebarRowsHeight := bodyHeight - boxOverhead
	if sidebarRowsHeight < minVisibleRows {
		sidebarRowsHeight = minVisibleRows
	}
	m.tbl.SetHeight(sidebarRowsHeight)
	m.mTbl.SetHeight(sidebarRowsHeight)
}

func (m appModel) View() string {
	if m.width == 0 {
		return ""
	}

	header := theme.Header(m.width).Render("TAmem — memória compartilhada entre agentes")

	footer := tuiui.NewFooter(reg.Bindings()...).
		Status(m.status).
		Render(m.width, theme)

	// Sidebar
	sidebarInner := m.sidebarInnerWidth()
	sidebarBox := theme.Panel(m.focus == focusProjects).Render(padLines(
		theme.Title().Render("Projetos")+"\n"+m.tbl.View(), sidebarInner,
	))

	// Status panel
	statsBox := theme.Panel(false).Render(padLines(
		theme.Title().Render("Ponte")+"\n"+m.renderStatus(), m.rightInnerWidth(),
	))

	// Memory panel
	var memBody string
	switch m.mode {
	case modeSearch:
		memBody = m.renderSearch()
	case modeLink:
		memBody = m.renderLinkForm()
	case modeUnlink:
		memBody = m.renderUnlinkForm()
	default:
		memBody = m.renderMemory()
	}
	memoryBox := theme.Panel(m.focus == focusMemory).Render(padLines(
		theme.Title().Render("Memória")+"\n"+memBody, m.rightInnerWidth(),
	))

	rightCol := lipgloss.JoinVertical(lipgloss.Left, statsBox, memoryBox)
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarBox, strings.Repeat(" ", panelGap), rightCol)

	view := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)

	if m.settingsModal.Visible() {
		return m.settingsModal.View(theme)
	}
	if m.helpModal.Visible() {
		return m.helpModal.View(theme)
	}
	return view
}

func (m appModel) renderStatus() string {
	p := m.current()
	if p == nil {
		return dim.Render("nenhum projeto selecionado")
	}

	lines := []string{
		fmt.Sprintf("Slug:      %s", p.slug),
		fmt.Sprintf("Store:     %s", dim.Render(p.sharedDir)),
		fmt.Sprintf("Tópicos:   %d", p.topicCount),
	}
	if p.repo != "" {
		lines = append(lines,
			fmt.Sprintf("Repo:      %s", dim.Render(p.repo)),
			fmt.Sprintf("Claude:    %s", statusDot(p.claudeLinked)),
			fmt.Sprintf("OpenCode:  %s", statusDot(p.opencodeLinked)),
			fmt.Sprintf("AGENTS.md: %s", statusDot(p.agentsSection)),
		)
		if p.worktreeCount > 0 {
			lines = append(lines, fmt.Sprintf("Worktrees: %d", p.worktreeCount))
		}
	} else {
		lines = append(lines, dim.Render("(sem repo configurado)"))
	}

	return strings.Join(lines, "\n")
}

func (m appModel) renderMemory() string {
	if len(m.memoryFiles) == 0 {
		return dim.Render("nenhum arquivo de memória")
	}

	// Show file list + content split.
	var b strings.Builder
	b.WriteString(m.mTbl.View())

	if m.activeFile != "" && len(m.contentLines) > 0 {
		b.WriteString("\n" + bold.Render(m.activeFile) + "\n")
		scroll := m.contentScroll
		maxScroll := 0
		if n := len(m.contentLines) - 20; n > 0 {
			maxScroll = n
		}
		if scroll > maxScroll {
			scroll = maxScroll
		}
		end := scroll + 20
		if end > len(m.contentLines) {
			end = len(m.contentLines)
		}
		for _, line := range m.contentLines[scroll:end] {
			b.WriteString(line + "\n")
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func (m appModel) renderSearch() string {
	var b strings.Builder
	b.WriteString(m.searchInput.View() + "\n")
	if len(m.results) == 0 && m.searchQuery != "" {
		b.WriteString(dim.Render("nenhum resultado"))
	} else {
		for i, r := range m.results {
			glyph := "  "
			if i == m.resultIdx {
				glyph = theme.Success().Render("▸ ")
			}
			label := r.Project + "/" + r.File
			if r.Name != "" {
				label = r.Project + "/" + r.Name
			}
			b.WriteString(glyph + label + "\n")
			if i == m.resultIdx && r.Snippet != "" {
				b.WriteString("  " + dim.Render(r.Snippet) + "\n")
			}
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m appModel) renderLinkForm() string {
	var b strings.Builder
	b.WriteString("Slug do projeto:\n")
	b.WriteString(m.linkSlug.View() + "\n\n")
	b.WriteString("Caminho do repo:\n")
	b.WriteString(m.linkRepo.View() + "\n\n")
	if m.formStatus != "" {
		b.WriteString(theme.Error().Render(m.formStatus) + "\n")
	}
	b.WriteString(dim.Render("tab muda campo · enter confirma · esc cancela"))
	return b.String()
}

func (m appModel) renderUnlinkForm() string {
	var b strings.Builder
	p := m.current()
	if p != nil {
		b.WriteString(fmt.Sprintf("Desligar %s?\n", bold.Render(p.slug)))
		b.WriteString(dim.Render(p.repo) + "\n\n")
		b.WriteString(m.formStatus + "\n")
	}
	return b.String()
}

// padLines pads every line of s to width (same as tuiui.PadLines).
func padLines(s string, width int) string {
	return tuiui.PadLines(s, width)
}
