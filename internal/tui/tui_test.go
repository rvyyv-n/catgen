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
	if !m.inputMode {
		t.Fatal("pressing 'o' did not enter input mode")
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

	if m.inputMode {
		t.Error("Esc did not leave input mode")
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
