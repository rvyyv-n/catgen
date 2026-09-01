package tui

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"cats/internal/ascii"
	"cats/internal/config"
)

func tempPNG(t *testing.T, w, h int) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "ext.png")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 4), B: 200, A: 255})
		}
	}
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return p
}

func sized(m tea.Model) Model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return next.(Model)
}

func TestOpenImageOverlayLoadsExternalFile(t *testing.T) {
	m := sized(NewModel("images", []string{"a.png", "b.png"}))
	startCount := len(m.Images)

	// Press "o" to open the overlay.
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	m = next.(Model)
	if m.overlay != overlayOpenImage {
		t.Fatal("pressing 'o' did not open the image overlay")
	}

	// Load a real external image through the same path Enter uses.
	p := tempPNG(t, 32, 24)
	m.loadExternal(p)

	if len(m.Images) != startCount+1 {
		t.Fatalf("image list = %d entries, want %d", len(m.Images), startCount+1)
	}
	last := m.Images[len(m.Images)-1]
	if !last.External || last.Path != p {
		t.Fatalf("appended entry = %+v, want External path %s", last, p)
	}
	if m.ImageIdx != len(m.Images)-1 {
		t.Fatalf("selection = %d, want %d", m.ImageIdx, len(m.Images)-1)
	}
	if strings.Contains(m.asciiArt, "Error") || m.asciiArt == "" {
		t.Fatalf("preview not rendered for external image: %q", m.asciiArt)
	}
}

func TestOpenImageOverlayEscCancels(t *testing.T) {
	m := sized(NewModel("images", []string{"a.png"}))
	count := len(m.Images)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)

	if m.overlay != overlayNone {
		t.Error("Esc did not leave the overlay")
	}
	if len(m.Images) != count {
		t.Errorf("image list changed after cancel: %d != %d", len(m.Images), count)
	}
}

func TestLoadExternalBadPathKeepsSelection(t *testing.T) {
	m := sized(NewModel("images", []string{"a.png", "b.png"}))
	m.ImageIdx = 1
	m.loadExternal(filepath.Join(t.TempDir(), "does-not-exist.png"))

	if len(m.Images) != 2 {
		t.Errorf("bad path mutated image list: %d entries", len(m.Images))
	}
	if m.ImageIdx != 1 {
		t.Errorf("bad path moved selection to %d", m.ImageIdx)
	}
	if !strings.HasPrefix(m.statusMsg, "✗") {
		t.Errorf("expected failure status, got %q", m.statusMsg)
	}
}

func TestViewRendersInputBarInOverlay(t *testing.T) {
	m := sized(NewModel("images", nil))
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	m = next.(Model)
	if !strings.Contains(m.View(), "Open image") {
		t.Error("overlay View() missing the open-image prompt")
	}
}

func TestOpenImageOverlayForwardsTypedText(t *testing.T) {
	m := sized(NewModel("images", []string{"a.png"}))
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("o")})
	m = next.(Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("https://e/c.png"), Paste: true})
	m = next.(Model)
	if got := m.input.Value(); got != "https://e/c.png" {
		t.Fatalf("input value = %q, want the pasted URL", got)
	}
}

func TestFitInfoToggle(t *testing.T) {
	m := sized(NewModel("images", nil))
	if m.showFitInfo {
		t.Fatal("fit info should start off")
	}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = next.(Model)
	if !m.showFitInfo {
		t.Fatal("'a' did not enable fit info")
	}
}

func TestSavePresetWritesFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	m := sized(NewModel("images", []string{"a.png"}))
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m = next.(Model)
	if m.overlay != overlaySavePreset {
		t.Fatalf("overlay = %v, want save-preset", m.overlay)
	}
	for _, r := range "mylook" {
		next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(Model)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)

	if m.overlay != overlayNone {
		t.Error("overlay not cleared after save")
	}
	if _, err := os.Stat(filepath.Join(home, ".catgen", "presets", "mylook.json")); err != nil {
		t.Fatalf("preset file not written: %v", err)
	}
}

func TestApplyPresetMapsThemeAndRamp(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	m := sized(NewModel("images", []string{"a.png"}))
	m.applyPreset("matrix") // built-in

	if want := ascii.ThemeByName("matrix"); m.ThemeIdx != want {
		t.Errorf("ThemeIdx = %d, want %d", m.ThemeIdx, want)
	}
	if want := ascii.RampByName("binary"); m.RampIdx != want {
		t.Errorf("RampIdx = %d, want %d", m.RampIdx, want)
	}
}

func TestExportModalWritesPlainText(t *testing.T) {
	m := sized(NewModel("images", nil))
	out := filepath.Join(t.TempDir(), "out.txt")

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = next.(Model)
	if m.overlay != overlayExport {
		t.Fatalf("overlay = %v, want export", m.overlay)
	}
	m.input.SetValue(out)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)

	if m.overlay != overlayNone {
		t.Error("overlay not cleared after export")
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("export not written: %v", err)
	}
}

func TestExportModalWritesPNG(t *testing.T) {
	m := sized(NewModel("images", nil))
	m.loadExternal(tempPNG(t, 48, 36))
	out := filepath.Join(t.TempDir(), "out.png")

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m = next.(Model)

	// Toggle the format field from plain text to image.
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = next.(Model)
	if !m.exportPNG {
		t.Fatal("Right on the format field did not select PNG")
	}

	m.input.SetValue(out)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)

	if m.overlay != overlayNone {
		t.Errorf("overlay not cleared after export: %v", m.overlay)
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("png not written: %v", err)
	}
	defer f.Close()
	cfg, err := png.DecodeConfig(f)
	if err != nil {
		t.Fatalf("output is not a valid PNG: %v", err)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		t.Errorf("degenerate PNG dimensions %dx%d", cfg.Width, cfg.Height)
	}
}

func TestSwapExportExt(t *testing.T) {
	cases := []struct {
		in   string
		png  bool
		want string
	}{
		{"cat_ascii.txt", true, "cat_ascii.png"},
		{"cat_ascii.png", false, "cat_ascii.txt"},
		{"art/my cat.txt", true, "art/my cat.png"},
		{"noext", true, "noext.png"},
		{"", true, defaultExportPath(true)},
		{"", false, defaultExportPath(false)},
	}
	for _, c := range cases {
		if got := swapExportExt(c.in, c.png); got != c.want {
			t.Errorf("swapExportExt(%q, %v) = %q, want %q", c.in, c.png, got, c.want)
		}
	}
}

func TestDefaultExportPathInExportsDir(t *testing.T) {
	if got := defaultExportPath(true); filepath.Dir(got) != exportDir {
		t.Errorf("defaultExportPath(png) = %q, want it under %q", got, exportDir)
	}
	if got := defaultExportPath(false); filepath.Ext(got) != ".txt" {
		t.Errorf("defaultExportPath(text) = %q, want .txt", got)
	}
}

func TestChromeCyclePersists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Cleanup(func() { applyChrome(0) }) // restore package globals for other tests

	m := sized(NewModel("images", nil))
	start := m.ChromeIdx

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	m = next.(Model)

	if m.ChromeIdx != (start+1)%len(chromeSchemes) {
		t.Fatalf("ChromeIdx = %d, want %d", m.ChromeIdx, (start+1)%len(chromeSchemes))
	}
	want := chromeSchemes[m.ChromeIdx]
	if colorBorder != want.Border || colorTeal != want.Accent {
		t.Errorf("applyChrome did not set colours: border=%v accent=%v want %v/%v",
			colorBorder, colorTeal, want.Border, want.Accent)
	}
	if cfg := config.Load(); cfg.Chrome != want.Name {
		t.Errorf("config.Chrome = %q, want %q", cfg.Chrome, want.Name)
	}

	// A fresh model restores the persisted scheme.
	m2 := NewModel("images", nil)
	if m2.ChromeIdx != m.ChromeIdx {
		t.Errorf("reloaded ChromeIdx = %d, want %d", m2.ChromeIdx, m.ChromeIdx)
	}
}
