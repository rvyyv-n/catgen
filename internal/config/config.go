// Package config persists small cross-session TUI settings as JSON at
// ~/.catgen/config.json. Load never fails loudly: a missing or unreadable file
// yields the zero Config so first run and a corrupt file behave the same.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config is the on-disk settings blob. Fields are optional; an empty value
// means "use the built-in default".
type Config struct {
	Chrome string `json:"chrome"` // name of the TUI chrome colour scheme
}

// Path returns ~/.catgen/config.json, creating ~/.catgen if needed.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".catgen")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the config, returning a zero Config on any error.
func Load() Config {
	var c Config
	p, err := Path()
	if err != nil {
		return c
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return c
	}
	_ = json.Unmarshal(b, &c)
	return c
}

// Save writes the config as pretty JSON.
func Save(c Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, b, 0o644)
}
