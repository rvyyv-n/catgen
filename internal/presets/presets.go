// Package presets stores and retrieves CATGEN look presets as JSON files under
// ~/.catgen/presets. A handful of built-in presets are always available even
// when no files exist; saving a preset with a built-in name shadows it on disk.
package presets

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Preset is a named bundle of the TUI's look controls. Theme and Ramp are stored
// by their stable catalog names (not indices) so a preset keeps working if the
// catalog order ever changes.
type Preset struct {
	Name        string  `json:"name"`
	FitMode     int     `json:"fit_mode"`
	CustomWidth int     `json:"custom_width"`
	Theme       string  `json:"theme"`
	Ramp        string  `json:"ramp"`
	Brightness  int     `json:"brightness"`
	Contrast    float64 `json:"contrast"`
	Invert      bool    `json:"invert"`
}

// Builtins are always offered by List and Load. A file of the same name in the
// presets directory takes precedence over the built-in.
var Builtins = []Preset{
	{Name: "matrix", CustomWidth: 45, Theme: "matrix", Ramp: "binary", Contrast: 1.2},
	{Name: "cyberpunk", CustomWidth: 45, Theme: "cyberpunk", Ramp: "blocks", Contrast: 1.1},
	{Name: "amber-crt", CustomWidth: 45, Theme: "amber", Ramp: "standard", Brightness: 5, Contrast: 1.1},
	{Name: "noir", CustomWidth: 45, Theme: "grayscale", Ramp: "detailed", Contrast: 1.3},
}

// Dir returns the presets directory, creating it (and ~/.catgen) if needed.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(home, ".catgen", "presets")
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", err
	}
	return d, nil
}

// sanitize reduces a user-supplied name to a safe, file-system-friendly slug.
func sanitize(name string) string {
	name = strings.TrimSpace(name)
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_' || r == ' ':
			return '-'
		default:
			return -1
		}
	}, name)
	return strings.Trim(name, "-")
}

// List returns the sorted, de-duplicated union of built-in and on-disk preset
// names.
func List() ([]string, error) {
	seen := map[string]bool{}
	var names []string
	for _, b := range Builtins {
		if !seen[b.Name] {
			seen[b.Name] = true
			names = append(names, b.Name)
		}
	}
	if d, err := Dir(); err == nil {
		entries, _ := os.ReadDir(d)
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			n := strings.TrimSuffix(e.Name(), ".json")
			if !seen[n] {
				seen[n] = true
				names = append(names, n)
			}
		}
	}
	sort.Strings(names)
	return names, nil
}

// Load reads a preset by name, preferring an on-disk file over a built-in.
func Load(name string) (Preset, error) {
	name = sanitize(name)
	if name == "" {
		return Preset{}, errors.New("empty preset name")
	}
	if d, err := Dir(); err == nil {
		if b, rerr := os.ReadFile(filepath.Join(d, name+".json")); rerr == nil {
			var p Preset
			if jerr := json.Unmarshal(b, &p); jerr != nil {
				return Preset{}, jerr
			}
			p.Name = name
			return p, nil
		}
	}
	for _, bi := range Builtins {
		if bi.Name == name {
			return bi, nil
		}
	}
	return Preset{}, os.ErrNotExist
}

// Save writes a preset to disk as <name>.json.
func Save(p Preset) error {
	name := sanitize(p.Name)
	if name == "" {
		return errors.New("preset name is empty after sanitizing")
	}
	p.Name = name
	d, err := Dir()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(d, name+".json"), b, 0o644)
}
