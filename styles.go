package main

import "github.com/charmbracelet/lipgloss"

// bridgeGlyph reports the bridge health of a project as a single glyph:
//
//	✓  = both claude and opencode linked (symlink + AGENTS.md section)
//	●  = claude only (symlink, no AGENTS.md section)
//	○  = neither (orphan project with no configured repo)
func bridgeGlyph(claudeLinked, opencodeLinked bool) string {
	switch {
	case claudeLinked && opencodeLinked:
		return theme.Success().Render("✓")
	case claudeLinked:
		return theme.Warning().Render("●")
	default:
		return theme.Dim().Render("○")
	}
}

// statusDot is a colored dot for the bridge panel.
func statusDot(ok bool) string {
	if ok {
		return theme.Success().Render("✓")
	}
	return theme.Error().Render("✗")
}

var dim = theme.Dim()
var bold = lipgloss.NewStyle().Foreground(theme.Text).Background(theme.Base).Bold(true)
