package tui

import (
	"fmt"
	"image"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/rvyyv-n/catgen/internal/ascii"
	"github.com/rvyyv-n/catgen/internal/config"
	"github.com/rvyyv-n/catgen/internal/imgsrc"
	"github.com/rvyyv-n/catgen/internal/presets"
	"github.com/rvyyv-n/catgen/internal/samples"
)

// --- Color Palette ---
// colorMuted, colorDark and colorSuccess are fixed. The other four are set by
// the active chrome scheme (see applyChrome) and default to the "teal" look.
var (
	colorTeal   = lipgloss.Color("86")  // accent: selection, logo, footer keys
	colorPink   = lipgloss.Color("205") // section headers, logo tail
	colorBorder = lipgloss.Color("37")  // frame borders
	colorText   = lipgloss.Color("252") // unselected row text

	colorDark    = lipgloss.Color("16")  // black text on the selection highlight
	colorMuted   = lipgloss.Color("241") // labels & muted hint text
	colorSuccess = lipgloss.Color("42")  // success green (status ticks)

	sectionStyle = lipgloss.NewStyle().Foreground(colorPink).Bold(true)

	footerKeyStyle = lipgloss.NewStyle().Foreground(colorTeal).Bold(true)

	footerDescStyle = lipgloss.NewStyle().Foreground(colorMuted)

	msgStyle = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
)

// chromeScheme is a named set of the four themeable chrome colours. It is
// separate from the ASCII art palettes in the ascii package.
type chromeScheme struct {
	Name    string
	Accent  lipgloss.Color
	Section lipgloss.Color
	Border  lipgloss.Color
	Text    lipgloss.Color
}

// chromeSchemes is the built-in chrome catalog, cycled by the `t` key. Index 0
// is the historical look. Fields: Name, Accent, Section, Border, Text.
var chromeSchemes = []chromeScheme{
	{"teal", "86", "205", "37", "252"},
	{"amber", "214", "208", "94", "223"},
	{"magenta", "213", "212", "89", "255"},
	{"green", "83", "40", "28", "252"},
	{"ocean", "39", "75", "24", "252"},
	{"violet", "141", "177", "54", "253"},
	{"mono", "252", "245", "240", "252"},
}

// chromeByName returns the catalog index of a scheme name, or 0 if unknown.
func chromeByName(name string) int {
	for i, c := range chromeSchemes {
		if c.Name == name {
			return i
		}
	}
	return 0
}

// applyChrome sets the themeable colour vars and rebuilds the styles that
// capture them. Safe to call repeatedly; the TUI's Update/View are single
// goroutine so the package-level mutation does not race.
func applyChrome(idx int) {
	if idx < 0 || idx >= len(chromeSchemes) {
		idx = 0
	}
	c := chromeSchemes[idx]
	colorTeal = c.Accent
	colorPink = c.Section
	colorBorder = c.Border
	colorText = c.Text
	sectionStyle = lipgloss.NewStyle().Foreground(colorPink).Bold(true)
	footerKeyStyle = lipgloss.NewStyle().Foreground(colorTeal).Bold(true)
}

// themes and ramps mirror the shared catalogs in the ascii package; the TUI
// indexes into them by ThemeIdx / RampIdx.
var themes = ascii.Themes
var ramps = ascii.Ramps

var fitModes = []string{"Auto", "Compact", "Wide", "Max", "Custom"}

const fitCustom = 4 // index of the manual-width mode in fitModes

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
	overlayExports
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
	ChromeIdx  int

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
	exportPNG     bool // false = plain text (.txt), true = image (.png)
	exportField   int  // 0 = format toggle, 1 = output path
	exportPending bool // path exists; the next Enter overwrites

	// Exports browser overlay
	exportList        []exportEntry
	exportListCursor  int
	exportListPending bool // the highlighted row's next d/Backspace deletes it

	// Decoded-image cache so re-renders (and remote URLs) don't reload every keystroke
	curImg    image.Image
	curImgKey string
}

func NewModel(imageDir string, images []string) Model {
	entries := make([]imageEntry, 0, len(images))
	for _, rel := range images {
		// Embedded sample refs are "embedded:<name>" so imgsrc can still
		// resolve them from the compiled-in pool; strip that prefix for
		// display so the menu shows the original filename, not "embedded:...".
		entries = append(entries, imageEntry{
			Display: filepath.Base(strings.TrimPrefix(rel, samples.Prefix)),
			Path:    filepath.Join(imageDir, filepath.FromSlash(rel)),
		})
	}

	ti := textinput.New()
	ti.Placeholder = "local path, ~ path, or https:// URL to an image"
	ti.Prompt = "▸ "
	ti.CharLimit = 1024
	ti.Width = 48

	chromeIdx := chromeByName(config.Load().Chrome)
	applyChrome(chromeIdx)

	m := Model{
		ImageDir:   imageDir,
		Images:     entries,
		ImageIdx:   0,
		FitModeIdx: 0, // Auto
		CustomW:    45,
		ThemeIdx:   0, // TrueColor
		RampIdx:    0, // Blocks
		Brightness: 0,
		Contrast:   1.0,
		Invert:     false,
		ChromeIdx:  chromeIdx,
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
			switch m.FitModeIdx {
			case 1:
				return "35"
			case 2:
				return "55"
			case fitCustom:
				return fmt.Sprintf("%d", m.CustomW)
			default: // Auto, Max
				return "Auto"
			}
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

		case "x":
			m.exportList = listExports()
			m.exportListCursor = 0
			m.exportListPending = false
			m.overlay = overlayExports
			m.statusMsg = ""
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
			m.statusMsg = ""
			m.input.Placeholder = "output path"
			m.input.SetValue(defaultExportPath(m.exportPNG))
			m.input.CursorEnd()
			return m, nil

		case "t":
			m.ChromeIdx = (m.ChromeIdx + 1) % len(chromeSchemes)
			applyChrome(m.ChromeIdx)
			name := chromeSchemes[m.ChromeIdx].Name
			if err := config.Save(config.Config{Chrome: name}); err != nil {
				m.statusMsg = "✗ " + condenseErr(err)
			} else {
				m.statusMsg = "✓ Theme: " + name
			}
		}

	case tea.WindowSizeMsg:
		m.termW = msg.Width
		m.termH = msg.Height
		m.updateAscii()
	}

	return m, nil
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
		// Adjusting Width switches to the manual-width mode and moves it.
		if m.FitModeIdx != fitCustom {
			m.FitModeIdx = fitCustom
		}
		m.CustomW += (delta * 5)
		if m.CustomW < 15 {
			m.CustomW = 15
		}
		if m.CustomW > 200 {
			m.CustomW = 200
		}

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

	autoFit := (m.FitModeIdx == 0)

	var targetW int
	switch m.FitModeIdx {
	case 0, 3: // Auto, Max
		targetW = availW
	case 1: // Compact
		targetW = 35
	case 2: // Wide
		targetW = 55
	case fitCustom:
		targetW = m.CustomW
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
