package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	tuiui "github.com/TAbelhaDev/tabelhatuiui"
)

// projectEntry maps a memory slug to its repo path, enabling full bridge
// health checks and link/unlink actions from the TUI.
type projectEntry struct {
	Slug string `toml:"slug"`
	Repo string `toml:"repo"`
}

type config struct {
	Projects []projectEntry `toml:"projects"`
	Layout   layoutConfig   `toml:"layout"`
	General  generalConfig  `toml:"general"`
}

type layoutConfig struct {
	SidebarWidthShare int `toml:"sidebar_width_share"`
	RightWidthShare   int `toml:"right_width_share"`
	StatsHeightShare  int `toml:"stats_height_share"`
	MemoryHeightShare int `toml:"memory_height_share"`
}

type generalConfig struct {
	Editor string `toml:"editor"`
}

func defaultConfig() config {
	return config{
		Layout: layoutConfig{
			SidebarWidthShare: 1,
			RightWidthShare:   4,
			StatsHeightShare:  1,
			MemoryHeightShare: 4,
		},
		General: generalConfig{Editor: ""},
	}
}

func normalize(c config) config {
	d := defaultConfig()
	if c.Layout.SidebarWidthShare < 1 {
		c.Layout.SidebarWidthShare = d.Layout.SidebarWidthShare
	}
	if c.Layout.RightWidthShare < 1 {
		c.Layout.RightWidthShare = d.Layout.RightWidthShare
	}
	if c.Layout.StatsHeightShare < 1 {
		c.Layout.StatsHeightShare = d.Layout.StatsHeightShare
	}
	if c.Layout.MemoryHeightShare < 1 {
		c.Layout.MemoryHeightShare = d.Layout.MemoryHeightShare
	}
	return c
}

func configPath() string {
	return tuiui.ConfigPath("tabelhamem", "config.toml")
}

var cfg *tuiui.Config[config]

var settings = defaultConfig()

func loadSettings() error {
	cfg = tuiui.NewConfig(configPath(), defaultConfig())
	err := cfg.Load()
	settings = normalize(cfg.Get())
	return err
}

func reloadSettings() (bool, error) {
	if cfg == nil {
		return true, loadSettings()
	}
	changed, err := cfg.Reload()
	settings = normalize(cfg.Get())
	return changed, err
}

// appendProject adds a slug/repo mapping to the config and persists it.
func appendProject(slug, repo string) error {
	c := cfg.Get()
	c.Projects = append(c.Projects, projectEntry{Slug: slug, Repo: repo})
	if err := writeConfig(c); err != nil {
		return err
	}
	settings = normalize(c)
	return nil
}

// removeProject removes a slug mapping from the config and persists it.
func removeProject(slug string) error {
	c := cfg.Get()
	var updated []projectEntry
	for _, p := range c.Projects {
		if p.Slug != slug {
			updated = append(updated, p)
		}
	}
	c.Projects = updated
	if err := writeConfig(c); err != nil {
		return err
	}
	settings = normalize(c)
	return nil
}

// writeConfig marshals the config to TOML and writes it to the config path.
func writeConfig(c config) error {
	path := configPath()
	if err := os.MkdirAll(tuiui.ConfigDir()+"/tabelhamem", 0o755); err != nil {
		return fmt.Errorf("criando diretório de config: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("criando config.toml: %w", err)
	}
	defer f.Close()
	enc := toml.NewEncoder(f)
	if err := enc.Encode(c); err != nil {
		return fmt.Errorf("salvando config.toml: %w", err)
	}
	return nil
}

func init() {
	if err := loadSettings(); err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "aviso: config.toml:", err)
		}
	}
}
