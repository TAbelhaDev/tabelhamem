package main

import (
	tuiui "github.com/TAbelhaDev/tabelhatuiui"
)

var theme = tuiui.NewThemeFromEnv("TABELHAMEM")

var (
	colBase     = theme.Base
	colMantle   = theme.Mantle
	colSurface0 = theme.Surface0
	colSurface1 = theme.Surface1
	colOverlay0 = theme.Overlay0
	colOverlay1 = theme.Overlay1
	colText     = theme.Text
	colSubtext0 = theme.Subtext0
	colPrimary  = theme.Primary
	colGreen    = theme.Green
	colYellow   = theme.Yellow
	colRed      = theme.Red
	colBlue     = theme.Blue
)
