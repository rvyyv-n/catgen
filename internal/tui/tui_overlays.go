package tui

import (
	"bytes"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/rvyyv-n/catgen/internal/ascii"
	"github.com/rvyyv-n/catgen/internal/imgsrc"
	"github.com/rvyyv-n/catgen/internal/presets"
)

// updateOverlay routes key input to the handler for the active modal.
func (m Model) updateOverlay(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.overlay {
	case overlayOpenImage:
		return m.updateOpenImage(msg)
	case overlaySavePreset:
		return m.updateSavePreset(msg)
	case overlayLoadPreset:
		return m.updateLoadPreset(msg)
	case overlayExport:
		return m.updateExport(msg)
	case overlayExports:
		return m.updateExports(msg)
	}
	return m, nil
}

// exportEntry is one file listed by the exports browser.
type exportEntry struct {
	Name    string
	Path    string
	Size    int64
	ModTime time.Time
}

// listExports returns the regular files in the exports directory, newest first.
func listExports() []exportEntry {
	dirEntries, err := os.ReadDir(exportDir)
	if err != nil {
		return nil
	}
	out := make([]exportEntry, 0, len(dirEntries))
	for _, de := range dirEntries {
		if de.IsDir() {
			continue
		}
		info, err := de.Info()
		if err != nil {
			continue
		}
		out = append(out, exportEntry{
			Name:    de.Name(),
			Path:    filepath.Join(exportDir, de.Name()),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ModTime.After(out[j].ModTime) })
	return out
}

// openInViewer hands a path to the OS's default handler.
func openInViewer(path string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/c", "start", "", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}

// updateExports drives the exports browser: navigate the list, Enter opens the
// file, c copies its path, d / Backspace deletes it after a confirm.
func (m Model) updateExports(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c", "q", "x":
		m.overlay = overlayNone
		return m, nil

	case "up", "k":
		if m.exportListCursor > 0 {
			m.exportListCursor--
		}
		m.exportListPending = false

	case "down", "j":
		if m.exportListCursor < len(m.exportList)-1 {
			m.exportListCursor++
		}
		m.exportListPending = false

	case "enter", " ":
		if e, ok := m.currentExport(); ok {
			if err := openInViewer(e.Path); err != nil {
				m.statusMsg = "✗ " + condenseErr(err)
			} else {
				m.statusMsg = "✓ Opened " + e.Name
			}
		}

	case "c":
		if e, ok := m.currentExport(); ok {
			abs, err := filepath.Abs(e.Path)
			if err != nil {
				abs = e.Path
			}
			if err := clipboard.WriteAll(abs); err != nil {
				m.statusMsg = "✗ " + condenseErr(err)
			} else {
				m.statusMsg = "✓ Copied path to clipboard"
			}
		}

	case "d", "backspace":
		e, ok := m.currentExport()
		if !ok {
			return m, nil
		}
		if !m.exportListPending {
			m.exportListPending = true
			m.statusMsg = "Delete " + e.Name + "? press d again"
			return m, nil
		}
		if err := os.Remove(e.Path); err != nil {
			m.statusMsg = "✗ " + condenseErr(err)
		} else {
			m.statusMsg = "✓ Deleted " + e.Name
		}
		m.exportListPending = false
		m.exportList = listExports()
		if m.exportListCursor >= len(m.exportList) {
			m.exportListCursor = len(m.exportList) - 1
		}
		if m.exportListCursor < 0 {
			m.exportListCursor = 0
		}
	}
	return m, nil
}

// currentExport returns the highlighted entry, if any.
func (m Model) currentExport() (exportEntry, bool) {
	if m.exportListCursor < 0 || m.exportListCursor >= len(m.exportList) {
		return exportEntry{}, false
	}
	return m.exportList[m.exportListCursor], true
}

// updateOpenImage drives the open-image overlay: Enter loads the typed
// reference, Esc dismisses it, everything else feeds the text field.
func (m Model) updateOpenImage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.overlay = overlayNone
		m.input.Blur()
		return m, nil

	case "enter":
		ref := m.input.Value()
		m.overlay = overlayNone
		m.input.Blur()
		m.loadExternal(ref)
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// updateSavePreset captures a preset name and writes the current look to disk.
func (m Model) updateSavePreset(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.overlay = overlayNone
		m.input.Blur()
		return m, nil

	case "enter":
		name := m.input.Value()
		m.overlay = overlayNone
		m.input.Blur()
		m.savePreset(name)
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// updateLoadPreset drives the preset picker list.
func (m Model) updateLoadPreset(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c", "q":
		m.overlay = overlayNone

	case "up", "k":
		if m.presetCursor > 0 {
			m.presetCursor--
		}

	case "down", "j":
		if m.presetCursor < len(m.presetNames)-1 {
			m.presetCursor++
		}

	case "enter", " ":
		if m.presetCursor >= 0 && m.presetCursor < len(m.presetNames) {
			m.applyPreset(m.presetNames[m.presetCursor])
		}
		m.overlay = overlayNone
	}
	return m, nil
}

// updateExport drives the export modal: a format toggle and an editable path.
func (m Model) updateExport(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.overlay = overlayNone
		m.input.Blur()
		return m, nil

	case "up", "down", "tab":
		m.exportField = 1 - m.exportField
		if m.exportField == 1 {
			return m, m.input.Focus()
		}
		m.input.Blur()
		return m, nil

	case "left", "right":
		if m.exportField == 0 {
			m.exportPNG = !m.exportPNG
			m.exportPending = false
			m.input.SetValue(swapExportExt(m.input.Value(), m.exportPNG))
			m.input.CursorEnd()
			return m, nil
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd

	case "enter":
		m.doExport()
		return m, nil
	}

	if m.exportField == 1 {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		m.exportPending = false
		return m, cmd
	}
	return m, nil
}

// exportDir is where exports land by default, kept separate from the source
// tree so a user's output is easy to find.
const exportDir = "exports"

// defaultExportPath is the suggested path for each export format.
func defaultExportPath(png bool) string {
	if png {
		return filepath.Join(exportDir, "cat_ascii.png")
	}
	return filepath.Join(exportDir, "cat_ascii.txt")
}

// swapExportExt keeps the user's chosen stem but swaps the extension to match
// the selected format. An empty path falls back to the default name.
func swapExportExt(path string, png bool) string {
	want := ".txt"
	if png {
		want = ".png"
	}
	if strings.TrimSpace(path) == "" {
		return defaultExportPath(png)
	}
	if ext := filepath.Ext(path); ext != "" {
		return strings.TrimSuffix(path, ext) + want
	}
	return path + want
}

// savePreset serializes the current look under the given name.
func (m *Model) savePreset(name string) {
	if strings.TrimSpace(name) == "" {
		m.statusMsg = "Preset name required"
		return
	}
	p := presets.Preset{
		Name:        name,
		FitMode:     m.FitModeIdx,
		CustomWidth: m.CustomW,
		Theme:       themes[m.ThemeIdx].Name,
		Ramp:        ramps[m.RampIdx].Name,
		Brightness:  m.Brightness,
		Contrast:    m.Contrast,
		Invert:      m.Invert,
	}
	if err := presets.Save(p); err != nil {
		m.statusMsg = "✗ " + condenseErr(err)
		return
	}
	m.statusMsg = "✓ Saved preset " + p.Name
}

// applyPreset loads a preset by name and maps it onto the model's controls.
// Unknown theme/ramp names leave the current selection untouched.
func (m *Model) applyPreset(name string) {
	p, err := presets.Load(name)
	if err != nil {
		m.statusMsg = "✗ " + condenseErr(err)
		return
	}
	if p.FitMode >= 0 && p.FitMode < len(fitModes) {
		m.FitModeIdx = p.FitMode
	}
	if p.CustomWidth > 0 {
		m.CustomW = p.CustomWidth
	}
	if i := ascii.ThemeByName(p.Theme); i >= 0 {
		m.ThemeIdx = i
	}
	if i := ascii.RampByName(p.Ramp); i >= 0 {
		m.RampIdx = i
	}
	m.Brightness = p.Brightness
	if p.Contrast > 0 {
		m.Contrast = p.Contrast
	}
	m.Invert = p.Invert
	m.statusMsg = "✓ Loaded preset " + name
	m.updateAscii()
}

// doExport writes the current art to the modal's path, asking once before it
// overwrites an existing file.
func (m *Model) doExport() {
	path := strings.TrimSpace(m.input.Value())
	if path == "" {
		m.statusMsg = "Output path required"
		return
	}
	if _, err := os.Stat(path); err == nil && !m.exportPending {
		m.exportPending = true
		m.statusMsg = "File exists — press Enter again to overwrite"
		return
	}

	var data []byte
	if m.exportPNG {
		img, err := m.currentImage()
		if err != nil {
			m.statusMsg = "✗ " + condenseErr(err)
			return
		}
		grid := ascii.ConvertGrid(img, m.renderOpts())
		pngImg, err := ascii.RenderPNG(grid, 2)
		if err != nil {
			m.statusMsg = "✗ " + condenseErr(err)
			return
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, pngImg); err != nil {
			m.statusMsg = "✗ " + condenseErr(err)
			return
		}
		data = buf.Bytes()
	} else {
		data = []byte(stripANSI(m.asciiArt))
	}

	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			m.statusMsg = "✗ " + condenseErr(err)
			return
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		m.statusMsg = "✗ " + condenseErr(err)
		return
	}
	m.overlay = overlayNone
	m.input.Blur()
	m.exportPending = false
	m.statusMsg = "✓ Exported to " + path
}

// loadExternal resolves a user-supplied path or URL, appends it to the image
// list, and selects it. Failures surface in the status line without disturbing
// the current selection.
func (m *Model) loadExternal(ref string) {
	clean := imgsrc.Ref(ref)
	if clean == "" {
		m.statusMsg = "No image path entered"
		return
	}

	img, _, err := imgsrc.LoadImage(clean)
	if err != nil {
		m.statusMsg = "✗ " + condenseErr(err)
		return
	}

	name := filepath.Base(clean)
	if imgsrc.IsURL(clean) {
		if i := strings.IndexAny(name, "?#"); i >= 0 {
			name = name[:i]
		}
	}
	if name == "" || name == "." || name == "/" {
		name = "image"
	}

	m.Images = append(m.Images, imageEntry{Display: name, Path: clean, External: true})
	m.ImageIdx = len(m.Images) - 1
	m.curImg = img
	m.curImgKey = clean
	m.statusMsg = "✓ Loaded " + name
	m.updateAscii()
}

// stripANSI removes SGR escape sequences so exported plain text carries no color codes.
func stripANSI(s string) string {
	var clean strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		clean.WriteRune(r)
	}
	return clean.String()
}
