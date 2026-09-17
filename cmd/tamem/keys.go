package main

import (
	"path/filepath"

	tuiui "github.com/TAbelhaDev/tabelhatuiui"
	"github.com/charmbracelet/bubbles/key"
)

var reg = tuiui.NewKeyRegistry(filepath.Join(tuiui.ConfigDir(), "tabelhamem", "keybindings.json"))

func init() {
	reg.RegisterMany(
		tuiui.Action{ID: "quit", Help: "sair", Keys: []string{"q"}},
		tuiui.Action{ID: "help", Help: "atalhos", Keys: []string{"?"}},
		tuiui.Action{ID: "settings", Help: "rebind keys", Keys: []string{","}},
		tuiui.Action{ID: "refresh", Help: "rescan", Keys: []string{"r"}},
		tuiui.Action{ID: "reload", Help: "recarregar config", Keys: []string{"ctrl+shift+r"}},
		tuiui.Action{ID: "nav", Help: "move focus", Keys: []string{"ctrl+h", "ctrl+l"}, Label: "ctrl+h/l"},
		tuiui.Action{ID: "search", Help: "buscar memória", Keys: []string{"/"}},
		tuiui.Action{ID: "open", Help: "abrir arquivo", Keys: []string{"enter"}},
		tuiui.Action{ID: "link", Help: "ligar projeto", Keys: []string{"l"}},
		tuiui.Action{ID: "unlink", Help: "desligar projeto", Keys: []string{"u"}},
		tuiui.Action{ID: "back", Help: "voltar", Keys: []string{"esc"}},
		tuiui.Action{ID: "scroll", Help: "rolar", Keys: []string{"j", "k", "up", "down"}, Label: "j/k"},
	)
}

func resolve(id string) key.Binding { return reg.Resolve(id) }
