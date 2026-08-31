package tui

import (
	"fmt"
	"image"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"cats/internal/ascii"
	"cats/internal/imgsrc"
	"cats/internal/presets"
)

// --- Color Palette (Matching BANGEN Theme) ---
var (
	colorTeal    = lipgloss.Color("86")  // #00d7af (Active Teal)
	colorDark    = lipgloss.Color("16")  // #000000 (Black text for highlight)
	colorPink    = lipgloss.Color("205") // #ff5faf (Section headers)
	colorMuted   = lipgloss.Color("241") // #626262 (Labels & muted text)
	colorBorder  = lipgloss.Color("37")  // #00afaf (Subtle frame border)
	colorText    = lipgloss.Color("252") // #d0d0d0 (Light text)
	colorSuccess = lipgloss.Color("42")  // #00d787 (Success green)

	headerStyle = lipgloss.NewStyle().
			Foreground(colorTeal).
			Bold(true)

	sectionStyle = lipgloss.NewStyle().
			Foreground(colorPink).
			Bold(true)

	footerKeyStyle = lipgloss.NewStyle().
			Foreground(colorTeal).
			Bold(true)

	footerDescStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	msgStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)
)

// themes and ramps mirror the shared catalogs in the ascii package; the TUI
// indexes into them by ThemeIdx / RampIdx.
var themes = ascii.Themes
var ramps = ascii.Ramps

var fitModes = []string{"Auto Fit", "Compact", "Wide", "Max"}

// leftPaneW is the fixed width of the controls pane, including its border.
const leftPaneW = 32

type ItemType int

const (
	ItemProperty ItemType = iota
	ItemThemeRadio
	ItemRampRadio
)

// overlayKind identifies which modal, if any, is intercepting key input.
type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayOpenImage
	overlaySavePreset
	overlayExport
	overlayLoadPreset
)

type MenuItem struct {
	ID        string
	Type      ItemType
	Section   string
	Label     string
	ValueFunc func(m *Model) string
	Index     int
}

// imageEntry is one selectable image. Bundled cats carry a filesystem Path built
// from the image directory; images opened via the "o" key carry an absolute path
// or an http(s) URL and are flagged External.
type imageEntry struct {
	Display  string
	Path     string
	External bool
}

type Model struct {
	ImageDir string
	Images   []imageEntry
	ImageIdx int

	FitModeIdx int
	CustomW    int
	ThemeIdx   int
	RampIdx    int
	Brightness int
	Contrast   float64
	Invert     bool

	cursor    int
	scrollOff int
	menuItems []MenuItem

	termW     int
	termH     int
	asciiArt  string
	statusMsg string
	rng       *rand.Rand

	// Fit-info toggle: when on, the footer shows the resolved output grid size.
	showFitInfo bool

	// Modal overlays share the single text input below.
	overlay overlayKind
	input   textinput.Model

	// Load-preset overlay
	presetNames  []string
	presetCursor int

	// Export overlay
	exportDiscord        bool // false = plain text, true = Discord snippet
	exportField          int  // 0 = format toggle, 1 = output path
	exportPending        bool // path exists; the next Enter overwrites
	userEditedExportPath bool // stop auto-filling the path once the user typed

	// Decoded-image cache so re-renders (and remote URLs) don't reload every keystroke
	curImg    image.Image
	curImgKey string
}

func NewModel(imageDir string, images []string) Model {
	entries := make([]imageEntry, 0, len(images))
	for _, rel := range images {
		entries = append(entries, imageEntry{
			Display: filepath.Base(rel),
			Path:    filepath.Join(imageDir, filepath.FromSlash(rel)),
		})
	}

	ti := textinput.New()
	ti.Placeholder = "local path, ~ path, or https:// URL to an image"
	ti.Prompt = "▸ "
	ti.CharLimit = 1024
	ti.Width = 48

	m := Model{
		ImageDir:   imageDir,
		Images:     entries,
		ImageIdx:   0,
		FitModeIdx: 0, // Auto Fit
		CustomW:    45,
		ThemeIdx:   0, // TrueColor
		RampIdx:    0, // Blocks
		Brightness: 0,
		Contrast:   1.0,
		Invert:     false,
		cursor:     0,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
		input:      ti,
	}
	m.buildMenu()
	return m
}

func (m *Model) buildMenu() {
	var items []MenuItem

	// --- General Section ---
	items = append(items, MenuItem{
		ID:      "image",
		Type:    ItemProperty,
		Section: "General",
		Label:   "Image",
		ValueFunc: func(m *Model) string {
			if len(m.Images) == 0 {
				return "None"
			}
			name := m.Images[m.ImageIdx].Display
			if len(name) > 12 {
				name = name[:9] + "..."
			}
			return name
		},
	})

	items = append(items, MenuItem{
		ID:      "fit",
		Type:    ItemProperty,
		Section: "General",
		Label:   "Fit Mode",
		ValueFunc: func(m *Model) string {
			return fitModes[m.FitModeIdx]
		},
	})

	items = append(items, MenuItem{
		ID:      "width",
		Type:    ItemProperty,
		Section: "General",
		Label:   "Width",
		ValueFunc: func(m *Model) string {
			if m.FitModeIdx == 0 {
				return "Auto"
			}
			return fmt.Sprintf("%d", m.CustomW)
		},
	})

	// --- Palette Section ---
	for i, t := range themes {
		idx := i
		items = append(items, MenuItem{
			ID:      fmt.Sprintf("theme_%d", i),
			Type:    ItemThemeRadio,
			Section: "Palette",
			Label:   t.Label,
			Index:   idx,
		})
	}

	// --- Character Ramp Section ---
	for i, r := range ramps {
		idx := i
		items = append(items, MenuItem{
			ID:      fmt.Sprintf("ramp_%d", i),
			Type:    ItemRampRadio,
			Section: "Character Ramp",
			Label:   r.Label,
			Index:   idx,
		})
	}

	// --- Tuning Section ---
	items = append(items, MenuItem{
		ID:      "brightness",
		Type:    ItemProperty,
		Section: "Tuning",
		Label:   "Brightness",
		ValueFunc: func(m *Model) string {
			if m.Brightness > 0 {
				return fmt.Sprintf("+%d", m.Brightness)
			}
			return fmt.Sprintf("%d", m.Brightness)
		},
	})

	items = append(items, MenuItem{
		ID:      "contrast",
		Type:    ItemProperty,
		Section: "Tuning",
		Label:   "Contrast",
		ValueFunc: func(m *Model) string {
			return fmt.Sprintf("%.1fx", m.Contrast)
		},
	})

	items = append(items, MenuItem{
		ID:      "invert",
		Type:    ItemProperty,
		Section: "Tuning",
		Label:   "Invert",
		ValueFunc: func(m *Model) string {
			if m.Invert {
				return "on"
			}
			return "off"
		},
	})

	m.menuItems = items
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.overlay != overlayNone {
			return m.updateOverlay(msg)
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "o":
			m.overlay = overlayOpenImage
			m.statusMsg = ""
			m.input.Placeholder = "local path, ~ path, or https:// URL to an image"
			m.input.SetValue("")
			m.input.CursorEnd()
			return m, m.input.Focus()

		case "s":
			m.overlay = overlaySavePreset
			m.statusMsg = ""
			m.input.Placeholder = "preset name"
			m.input.SetValue("")
			m.input.CursorEnd()
			return m, m.input.Focus()

		case "p":
			names, _ := presets.List()
			m.presetNames = names
			m.presetCursor = 0
			m.overlay = overlayLoadPreset
			return m, nil

		case "a":
			m.showFitInfo = !m.showFitInfo
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.menuItems)-1 {
				m.cursor++
			}

		case "left", "h":
			m.handleAdjust(-1)
			m.updateAscii()

		case "right", "l":
			m.handleAdjust(1)
			m.updateAscii()

		case "enter", " ":
			m.handleSelect()
			m.updateAscii()

		case "r":
			if len(m.Images) > 0 {
				m.ImageIdx = m.rng.Intn(len(m.Images))
				m.statusMsg = "✓ Loaded random cat!"
				m.updateAscii()
			}

		case "e":
			m.overlay = overlayExport
			m.exportField = 0
			m.exportPending = false
			m.userEditedExportPath = false
			m.statusMsg = ""
			m.input.Placeholder = "output path"
			m.input.SetValue(defaultExportPath(m.exportDiscord))
			m.input.CursorEnd()
			return m, nil

		case "d":
			m.exportToDiscord()
		}

	case tea.WindowSizeMsg:
		m.termW = msg.Width
		m.termH = msg.Height
		m.updateAscii()
	}

	return m, nil
}

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
	}
	return m, nil
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
			m.exportDiscord = !m.exportDiscord
			m.exportPending = false
			if !m.userEditedExportPath {
				m.input.SetValue(defaultExportPath(m.exportDiscord))
				m.input.CursorEnd()
			}
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
		m.userEditedExportPath = true
		m.exportPending = false
		return m, cmd
	}
	return m, nil
}

// defaultExportPath is the suggested filename for each export format.
func defaultExportPath(discord bool) string {
	if discord {
		return "cat_discord.txt"
	}
	return "cat_ascii.txt"
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

	var data string
	if m.exportDiscord {
		img, err := m.currentImage()
		if err != nil {
			m.statusMsg = "✗ " + condenseErr(err)
			return
		}
		colorize := m.ThemeIdx != 1 // colorize unless Grayscale
		data = ascii.ConvertToDiscord(img, colorize, ramps[m.RampIdx].Chars)
	} else {
		data = stripANSI(m.asciiArt)
	}

	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
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

// condenseErr trims a wrapped error down to its last, most specific clause so it
// fits on the status line.
func condenseErr(err error) string {
	s := err.Error()
	if i := strings.LastIndex(s, ": "); i >= 0 && i+2 < len(s) {
		return s[i+2:]
	}
	return s
}

func (m *Model) handleAdjust(delta int) {
	if m.cursor < 0 || m.cursor >= len(m.menuItems) {
		return
	}
	item := m.menuItems[m.cursor]

	switch item.ID {
	case "image":
		if len(m.Images) == 0 {
			return
		}
		m.ImageIdx += delta
		if m.ImageIdx < 0 {
			m.ImageIdx = len(m.Images) - 1
		} else if m.ImageIdx >= len(m.Images) {
			m.ImageIdx = 0
		}

	case "fit":
		m.FitModeIdx += delta
		if m.FitModeIdx < 0 {
			m.FitModeIdx = len(fitModes) - 1
		} else if m.FitModeIdx >= len(fitModes) {
			m.FitModeIdx = 0
		}

	case "width":
		m.CustomW += (delta * 5)
		if m.CustomW < 15 {
			m.CustomW = 15
		}
		if m.CustomW > 120 {
			m.CustomW = 120
		}
		m.FitModeIdx = 1

	case "brightness":
		m.Brightness += (delta * 5)
		if m.Brightness < -40 {
			m.Brightness = -40
		}
		if m.Brightness > 40 {
			m.Brightness = 40
		}

	case "contrast":
		m.Contrast += float64(delta) * 0.1
		if m.Contrast < 0.5 {
			m.Contrast = 0.5
		}
		if m.Contrast > 2.0 {
			m.Contrast = 2.0
		}

	case "invert":
		m.Invert = !m.Invert

	default:
		if item.Type == ItemThemeRadio {
			m.ThemeIdx = item.Index
		} else if item.Type == ItemRampRadio {
			m.RampIdx = item.Index
		}
	}
}

func (m *Model) handleSelect() {
	if m.cursor < 0 || m.cursor >= len(m.menuItems) {
		return
	}
	item := m.menuItems[m.cursor]

	switch item.Type {
	case ItemThemeRadio:
		m.ThemeIdx = item.Index
	case ItemRampRadio:
		m.RampIdx = item.Index
	case ItemProperty:
		m.handleAdjust(1)
	}
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

func (m *Model) exportToDiscord() {
	if len(m.Images) == 0 {
		return
	}
	img, err := m.currentImage()
	if err != nil {
		m.statusMsg = "✗ " + condenseErr(err)
		return
	}

	colorize := (m.ThemeIdx != 1) // Colorize unless Grayscale
	discordSnippet := ascii.ConvertToDiscord(img, colorize, ramps[m.RampIdx].Chars)
	filename := "cat_discord.txt"
	_ = os.WriteFile(filename, []byte(discordSnippet), 0o644)
	m.statusMsg = fmt.Sprintf("✓ Exported Discord snippet to %s!", filename)
}

// currentImage returns the decoded image for the current selection, using the
// cache when the selection has not changed since the last load. Remote URLs are
// therefore fetched once, not on every slider adjustment.
func (m *Model) currentImage() (image.Image, error) {
	if len(m.Images) == 0 {
		return nil, fmt.Errorf("no images available")
	}
	key := m.Images[m.ImageIdx].Path
	if m.curImg != nil && m.curImgKey == key {
		return m.curImg, nil
	}
	img, _, err := imgsrc.LoadImage(key)
	if err != nil {
		return nil, err
	}
	m.curImg = img
	m.curImgKey = key
	return img, nil
}

// renderOpts assembles the ascii.Options for the current controls and terminal
// size. Shared by updateAscii and the fit-info readout so both agree.
func (m *Model) renderOpts() ascii.Options {
	availW := m.termW - leftPaneW - 3
	availH := m.termH - 5

	if availW < 10 {
		availW = 10
	}
	if availH < 5 {
		availH = 5
	}

	targetW := m.CustomW
	autoFit := (m.FitModeIdx == 0)

	switch m.FitModeIdx {
	case 0: // Auto Fit
		targetW = availW
	case 1: // Compact
		targetW = 35
	case 2: // Wide
		targetW = 55
	case 3: // Max
		targetW = availW
	}

	return ascii.Options{
		Width:       targetW,
		MaxWidth:    availW,
		MaxHeight:   availH,
		AutoFit:     autoFit,
		Theme:       themes[m.ThemeIdx].Theme,
		DensityRamp: ramps[m.RampIdx].Chars,
		Brightness:  m.Brightness,
		Contrast:    m.Contrast,
		Invert:      m.Invert,
	}
}

func (m *Model) updateAscii() {
	if len(m.Images) == 0 {
		m.asciiArt = "No images found."
		return
	}

	img, err := m.currentImage()
	if err != nil {
		m.asciiArt = fmt.Sprintf("Error loading image:\n%v", err)
		return
	}

	m.asciiArt = ascii.Convert(img, m.renderOpts())
}

// radioMarker renders the marker for a radio row: a green check for the active
// choice, a hollow circle otherwise. On the cursor row the check is left
// unstyled so it stays legible against the teal selection bar.
func radioMarker(active, onCursor bool) string {
	if !active {
		return "○"
	}
	if onCursor {
		return "✓"
	}
	return lipgloss.NewStyle().Foreground(colorSuccess).Render("✓")
}

// buildFramedBox creates a box with a cleanly centered title in the top border
func buildFramedBox(title string, content string, width int, height int, borderColor lipgloss.Color, titleColor lipgloss.Color) string {
	if width < 6 {
		width = 6
	}
	if height < 3 {
		height = 3
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	headStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(true)

	// Build Centered Top Border
	titleStr := " " + title + " "
	titleLen := len(titleStr)
	innerW := width - 2

	var topBorder string
	if strings.TrimSpace(title) == "" {
		topBorder = borderStyle.Render("┌" + strings.Repeat("─", innerW) + "┐")
	} else if titleLen >= innerW {
		topBorder = borderStyle.Render("┌" + strings.Repeat("─", innerW) + "┐")
	} else {
		leftDashes := (innerW - titleLen) / 2
		rightDashes := innerW - titleLen - leftDashes
		topBorder = borderStyle.Render("┌"+strings.Repeat("─", leftDashes)) +
			headStyle.Render(titleStr) +
			borderStyle.Render(strings.Repeat("─", rightDashes)+"┐")
	}

	// Bottom Border
	bottomBorder := borderStyle.Render("└" + strings.Repeat("─", innerW) + "┘")

	// Side borders and content rows
	contentRows := strings.Split(content, "\n")
	var bodyRows []string

	for i := 0; i < height-2; i++ {
		rowText := ""
		if i < len(contentRows) {
			rowText = contentRows[i]
		}
		// Pad or truncate rowText to innerW
		renderedRow := lipgloss.NewStyle().Width(innerW).Render(rowText)
		bodyRows = append(bodyRows, borderStyle.Render("│")+renderedRow+borderStyle.Render("│"))
	}

	return topBorder + "\n" + strings.Join(bodyRows, "\n") + "\n" + bottomBorder
}

func (m Model) View() string {
	if m.termW == 0 {
		return "Initializing CATGEN..."
	}

	rightPaneW := m.termW - leftPaneW
	if rightPaneW < 15 {
		rightPaneW = 15
	}
	contentH := m.termH - 3
	if contentH < 8 {
		contentH = 8
	}
	innerLeftW := leftPaneW - 2
	innerRightW := rightPaneW - 2
	innerHeight := contentH - 2

	// --- Left Pane (Controls) ---
	type RenderedRow struct {
		isSection bool
		text      string
		menuIdx   int
	}

	var allRows []RenderedRow
	var currentSection string

	// Stylized CATGEN Block Logo Banner
	logoCatStyle := lipgloss.NewStyle().Foreground(colorTeal).Bold(true)
	logoGenStyle := lipgloss.NewStyle().Foreground(colorPink).Bold(true)
	subLogoStyle := lipgloss.NewStyle().Foreground(colorMuted)

	allRows = append(allRows, RenderedRow{
		isSection: true,
		text:      " " + logoCatStyle.Render("█▀▀ ▄▀█ ▀█▀") + " " + logoGenStyle.Render("█▀▀ █▀▀ █▄ █"),
	})
	allRows = append(allRows, RenderedRow{
		isSection: true,
		text:      " " + logoCatStyle.Render("█▄▄ █▀█  █ ") + " " + logoGenStyle.Render("█▄█ ██▄ █ ▀█"),
	})
	allRows = append(allRows, RenderedRow{
		isSection: true,
		text:      subLogoStyle.Render(" ── ASCII Cat Studio ──"),
	})
	allRows = append(allRows, RenderedRow{isSection: true, text: ""})

	for i, item := range m.menuItems {
		if item.Section != currentSection {
			if currentSection != "" {
				allRows = append(allRows, RenderedRow{isSection: true, text: ""})
			}
			currentSection = item.Section
			allRows = append(allRows, RenderedRow{
				isSection: true,
				text:      " " + sectionStyle.Render(currentSection),
			})
		}

		isSelected := (m.cursor == i)
		var prefix string
		if isSelected {
			prefix = "► "
		} else {
			prefix = "  "
		}

		var rowBody string
		switch item.Type {
		case ItemProperty:
			val := item.ValueFunc(&m)
			lbl := fmt.Sprintf("%-10s", item.Label)
			rowBody = fmt.Sprintf("%s%11s", lbl, val)

		case ItemThemeRadio:
			rowBody = fmt.Sprintf("%s %s", radioMarker(m.ThemeIdx == item.Index, isSelected), item.Label)

		case ItemRampRadio:
			rowBody = fmt.Sprintf("%s %s", radioMarker(m.RampIdx == item.Index, isSelected), item.Label)
		}

		rowText := prefix + rowBody

		var renderedText string
		if isSelected {
			rowStyle := lipgloss.NewStyle().
				Background(colorTeal).
				Foreground(colorDark).
				Bold(true).
				Width(innerLeftW)
			renderedText = rowStyle.Render(rowText)
		} else {
			rowStyle := lipgloss.NewStyle().
				Foreground(colorText).
				Width(innerLeftW)
			renderedText = rowStyle.Render(rowText)
		}

		allRows = append(allRows, RenderedRow{
			isSection: false,
			text:      renderedText,
			menuIdx:   i,
		})
	}

	// Find cursor row in allRows
	cursorRow := 0
	for r, row := range allRows {
		if !row.isSection && row.menuIdx == m.cursor {
			cursorRow = r
			break
		}
	}

	// Smooth scrolling for left pane
	if cursorRow < m.scrollOff {
		m.scrollOff = cursorRow
	}
	if cursorRow >= m.scrollOff+innerHeight {
		m.scrollOff = cursorRow - innerHeight + 1
	}
	if m.scrollOff < 0 {
		m.scrollOff = 0
	}

	startIdx := m.scrollOff
	if startIdx > len(allRows) {
		startIdx = len(allRows)
	}
	endIdx := startIdx + innerHeight
	if endIdx > len(allRows) {
		endIdx = len(allRows)
	}

	var visibleTextLines []string
	for _, row := range allRows[startIdx:endIdx] {
		visibleTextLines = append(visibleTextLines, row.text)
	}
	leftBody := strings.Join(visibleTextLines, "\n")

	leftBox := buildFramedBox("", leftBody, leftPaneW, contentH, colorBorder, colorTeal)

	// --- Right Pane (Live Preview) ---
	// Horizontally and vertically center the ASCII Cat
	centeredArt := lipgloss.Place(
		innerRightW,
		innerHeight,
		lipgloss.Center,
		lipgloss.Center,
		m.asciiArt,
	)

	rightBox := buildFramedBox("Live Preview", centeredArt, rightPaneW, contentH, colorBorder, colorTeal)

	// Join the frames flush so the shared edge reads as one divider.
	mainLayout := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)

	totalW := leftPaneW + rightPaneW

	// Full-area modals replace the two-pane body.
	switch m.overlay {
	case overlayExport:
		mainLayout = centeredModal("Export", m.exportModalBody(), totalW, contentH)
	case overlayLoadPreset:
		mainLayout = centeredModal("Load Preset", m.presetModalBody(), totalW, contentH)
	}

	// --- Footer: context hint for the active overlay, keybinds otherwise ---
	var footer string
	switch m.overlay {
	case overlayOpenImage:
		footer = m.inputFooter("Open image", "Enter load · Esc cancel · Ctrl+V paste", totalW)
	case overlaySavePreset:
		footer = m.inputFooter("Save preset", "Enter save · Esc cancel", totalW)
	case overlayExport:
		footer = hintFooter("↑↓ field · ←→ format · Enter write · Esc cancel", m.statusMsg, totalW)
	case overlayLoadPreset:
		footer = hintFooter("↑↓ move · Enter load · Esc cancel", m.statusMsg, totalW)
	default:
		footerItems := lipgloss.JoinHorizontal(lipgloss.Top,
			footerKeyStyle.Render("↑↓"), footerDescStyle.Render(" nav  "),
			footerKeyStyle.Render("↔"), footerDescStyle.Render(" adjust  "),
			footerKeyStyle.Render("⏎"), footerDescStyle.Render(" toggle  "),
			footerKeyStyle.Render("o"), footerDescStyle.Render(" open  "),
			footerKeyStyle.Render("r"), footerDescStyle.Render(" random  "),
			footerKeyStyle.Render("e"), footerDescStyle.Render(" export  "),
			footerKeyStyle.Render("d"), footerDescStyle.Render(" discord  "),
			footerKeyStyle.Render("s"), footerDescStyle.Render(" save  "),
			footerKeyStyle.Render("p"), footerDescStyle.Render(" presets  "),
			footerKeyStyle.Render("a"), footerDescStyle.Render(" info  "),
			footerKeyStyle.Render("q"), footerDescStyle.Render(" quit"),
		)

		var lead string
		if m.statusMsg != "" {
			lead = msgStyle.Render(m.statusMsg)
		} else if m.showFitInfo && m.curImg != nil {
			gw, gh := ascii.Measure(m.curImg, m.renderOpts())
			b := m.curImg.Bounds()
			lead = footerDescStyle.Render(fmt.Sprintf("fit:%s · src %dx%d · grid %dx%d",
				fitModes[m.FitModeIdx], b.Dx(), b.Dy(), gw, gh))
		}
		if lead != "" {
			footerItems = lipgloss.JoinHorizontal(lipgloss.Top, lead, "   ", footerItems)
		}

		footer = lipgloss.NewStyle().
			Width(totalW).
			Align(lipgloss.Right).
			Render(footerItems)
	}

	return lipgloss.JoinVertical(lipgloss.Left, mainLayout, "", footer)
}

// inputFooter renders a labelled text-input line for the open-image and
// save-preset overlays.
func (m Model) inputFooter(label, hint string, totalW int) string {
	l := sectionStyle.Render(label + " ")
	h := footerDescStyle.Render("   " + hint)
	return lipgloss.NewStyle().Width(totalW).Render(l + m.input.View() + h)
}

// hintFooter renders a single dim hint line, optionally prefixed with a status
// message, for the modal overlays.
func hintFooter(hint, status string, totalW int) string {
	line := footerDescStyle.Render(hint)
	if status != "" {
		line = msgStyle.Render(status) + "   " + line
	}
	return lipgloss.NewStyle().Width(totalW).Render(line)
}

// centeredModal frames body under title and centers it in the given area.
func centeredModal(title, body string, areaW, areaH int) string {
	lines := strings.Split(body, "\n")
	boxW := lipgloss.Width(title) + 6
	for _, ln := range lines {
		if w := lipgloss.Width(ln) + 4; w > boxW {
			boxW = w
		}
	}
	if boxW > areaW-2 {
		boxW = areaW - 2
	}
	if boxW < 14 {
		boxW = 14
	}
	boxH := len(lines) + 4
	if boxH > areaH {
		boxH = areaH
	}
	framed := buildFramedBox(title, "\n"+body, boxW, boxH, colorBorder, colorTeal)
	return lipgloss.Place(areaW, areaH, lipgloss.Center, lipgloss.Center, framed)
}

// exportModalBody renders the export modal's two fields: a format toggle and the
// editable output path.
func (m Model) exportModalBody() string {
	plain, discord := "○ Plain text", "○ Discord snippet"
	if m.exportDiscord {
		discord = "◉ Discord snippet"
	} else {
		plain = "◉ Plain text"
	}
	fmtLabel, pathLabel := "  Format:  ", "  Output:  "
	if m.exportField == 0 {
		fmtLabel = "► Format:  "
	} else {
		pathLabel = "► Output:  "
	}
	return fmtLabel + plain + "    " + discord + "\n" + pathLabel + m.input.View()
}

// presetModalBody renders the preset picker list.
func (m Model) presetModalBody() string {
	if len(m.presetNames) == 0 {
		return "  (no presets found)"
	}
	rows := make([]string, len(m.presetNames))
	for i, n := range m.presetNames {
		if i == m.presetCursor {
			rows[i] = lipgloss.NewStyle().Foreground(colorTeal).Bold(true).Render("► " + n)
		} else {
			rows[i] = "  " + n
		}
	}
	return strings.Join(rows, "\n")
}
