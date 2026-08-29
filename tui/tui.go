package tui

import (
	"fmt"
	"image"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"cats/ascii"
)

// --- Color Palette (Matching BANGEN Theme) ---
var (
	colorTeal      = lipgloss.Color("86")  // #00d7af (Active Teal)
	colorDark      = lipgloss.Color("16")  // #000000 (Black text for highlight)
	colorPink      = lipgloss.Color("205") // #ff5faf (Section headers)
	colorMuted     = lipgloss.Color("241") // #626262 (Labels & borders)
	colorBorder    = lipgloss.Color("37")  // #00afaf (Subtle frame border)
	colorText      = lipgloss.Color("252") // #d0d0d0 (Light text)
	colorDim       = lipgloss.Color("244") // #808080 (Dim text)
	colorSuccess   = lipgloss.Color("42")  // #00d787 (Success green)

	// Border Styles
	frameBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
	}

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

type ThemeOption struct {
	Name  string
	Theme ascii.Theme
}

var themes = []ThemeOption{
	{"TrueColor (RGB)", ascii.ThemeTrueColor},
	{"Grayscale", ascii.ThemeGrayscale},
	{"Matrix Glow", ascii.ThemeMatrix},
	{"Cyberpunk", ascii.ThemeCyberpunk},
	{"Amber Phosphor", ascii.ThemeAmber},
	{"Ice Blue", ascii.ThemeIceBlue},
}

type RampOption struct {
	Name string
	Ramp string
}

var ramps = []RampOption{
	{"Blocks (░▒▓█)", ascii.RampBlocks},
	{"Standard (.-=+*)", ascii.RampStandard},
	{"Braille (⠋⠙⠹)", ascii.RampBraille},
	{"Detailed ASCII", ascii.RampDetailed},
	{"Binary Matrix (01)", ascii.RampBinary},
	{"Minimal ( .oO@)", ascii.RampMinimal},
}

var fitModes = []string{"Auto Fit", "Compact", "Wide", "Max"}

// Navigation items in the Left Pane
type ItemType int

const (
	ItemProperty ItemType = iota
	ItemThemeRadio
	ItemRampRadio
	ItemAction
)

type MenuItem struct {
	ID        string
	Type      ItemType
	Section   string
	Label     string
	ValueFunc func(m *Model) string
	Index     int // Index within category
}

type Model struct {
	ImageDir   string
	Images     []string
	ImageIdx   int

	FitModeIdx int
	CustomW    int
	ThemeIdx   int
	RampIdx    int
	Brightness int
	Contrast   float64
	Invert     bool

	cursor     int
	scrollOff  int
	menuItems  []MenuItem

	termW      int
	termH      int
	asciiArt   string
	statusMsg  string
	msgTimer   int
	rng        *rand.Rand
}

func NewModel(imageDir string, images []string) Model {
	m := Model{
		ImageDir:   imageDir,
		Images:     images,
		ImageIdx:   0,
		FitModeIdx: 0, // Auto Fit default
		CustomW:    45,
		ThemeIdx:   0, // TrueColor
		RampIdx:    0, // Blocks
		Brightness: 0,
		Contrast:   1.0,
		Invert:     false,
		cursor:     0,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
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
			name := filepath.Base(m.Images[m.ImageIdx])
			if len(name) > 14 {
				name = name[:11] + "..."
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

	// --- Palette (Themes) ---
	for i, t := range themes {
		idx := i
		items = append(items, MenuItem{
			ID:      fmt.Sprintf("theme_%d", i),
			Type:    ItemThemeRadio,
			Section: "Palette",
			Label:   t.Name,
			Index:   idx,
		})
	}

	// --- Character Ramp ---
	for i, r := range ramps {
		idx := i
		items = append(items, MenuItem{
			ID:      fmt.Sprintf("ramp_%d", i),
			Type:    ItemRampRadio,
			Section: "Character Ramp",
			Label:   r.Name,
			Index:   idx,
		})
	}

	// --- Tuning ---
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
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

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
			// Random Cat
			if len(m.Images) > 0 {
				m.ImageIdx = m.rng.Intn(len(m.Images))
				m.statusMsg = "✓ Loaded random cat!"
				m.updateAscii()
			}

		case "e":
			// Export ASCII to file
			m.exportToFile()
		}

	case tea.WindowSizeMsg:
		m.termW = msg.Width
		m.termH = msg.Height
		m.updateAscii()
	}

	return m, nil
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
		m.FitModeIdx = 1 // Switch to custom width

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

func (m *Model) exportToFile() {
	if m.asciiArt == "" {
		return
	}
	filename := "cat_ascii.txt"
	// Strip ANSI color codes for clean text export
	var clean strings.Builder
	inEscape := false
	for _, r := range m.asciiArt {
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

	_ = os.WriteFile(filename, []byte(clean.String()), 0644)
	m.statusMsg = fmt.Sprintf("✓ Exported to %s!", filename)
}

func (m *Model) updateAscii() {
	if len(m.Images) == 0 {
		m.asciiArt = "No images found in directory."
		return
	}

	imgPath := filepath.Join(m.ImageDir, filepath.FromSlash(m.Images[m.ImageIdx]))
	f, err := os.Open(imgPath)
	if err != nil {
		m.asciiArt = fmt.Sprintf("Error opening image:\n%v", err)
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		m.asciiArt = fmt.Sprintf("Error decoding image:\n%v", err)
		return
	}

	// Calculate exact bounds for Live Preview Pane
	leftPaneW := 33
	availW := m.termW - leftPaneW - 8
	availH := m.termH - 6 // Top frame, bottom frame, header, footer

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
		targetW = 60
	case 3: // Max
		targetW = availW
	}

	opts := ascii.Options{
		Width:       targetW,
		MaxWidth:    availW,
		MaxHeight:   availH,
		AutoFit:     autoFit,
		Theme:       themes[m.ThemeIdx].Theme,
		DensityRamp: ramps[m.RampIdx].Ramp,
		Brightness:  m.Brightness,
		Contrast:    m.Contrast,
		Invert:      m.Invert,
	}

	m.asciiArt = ascii.Convert(img, opts)
}

func (m Model) View() string {
	if m.termW == 0 {
		return "Initializing CATGEN..."
	}

	leftPaneW := 33
	rightPaneW := m.termW - leftPaneW - 4
	if rightPaneW < 15 {
		rightPaneW = 15
	}
	contentH := m.termH - 5
	if contentH < 8 {
		contentH = 8
	}

	// --- Left Pane (Controls) ---
	var lines []string
	var currentSection string

	visibleCount := contentH
	if m.cursor < m.scrollOff {
		m.scrollOff = m.cursor
	}
	if m.cursor >= m.scrollOff+visibleCount-2 {
		m.scrollOff = m.cursor - visibleCount + 3
	}
	if m.scrollOff < 0 {
		m.scrollOff = 0
	}

	for i, item := range m.menuItems {
		// Section Header
		if item.Section != currentSection {
			currentSection = item.Section
			lines = append(lines, sectionStyle.Render(currentSection))
		}

		isSelected := (m.cursor == i)
		var rowContent string

		switch item.Type {
		case ItemProperty:
			val := item.ValueFunc(&m)
			lbl := fmt.Sprintf("%-12s", item.Label)
			rowContent = fmt.Sprintf("%s %14s", lbl, val)

		case ItemThemeRadio:
			radio := "○"
			if m.ThemeIdx == item.Index {
				radio = "●"
			}
			rowContent = fmt.Sprintf(" %s %s", radio, item.Label)

		case ItemRampRadio:
			radio := "○"
			if m.RampIdx == item.Index {
				radio = "●"
			}
			rowContent = fmt.Sprintf(" %s %s", radio, item.Label)
		}

		if isSelected {
			// Solid teal highlight across the row with black text
			rowStyle := lipgloss.NewStyle().
				Background(colorTeal).
				Foreground(colorDark).
				Bold(true).
				Width(leftPaneW - 4)

			lines = append(lines, rowStyle.Render("►"+rowContent[1:]))
		} else {
			rowStyle := lipgloss.NewStyle().
				Foreground(colorText).
				Width(leftPaneW - 4)

			lines = append(lines, rowStyle.Render(" "+rowContent))
		}
	}

	// Slice visible menu lines for scroll window
	var visibleMenu []string
	startIdx := m.scrollOff
	if startIdx > len(lines) {
		startIdx = len(lines)
	}
	endIdx := startIdx + contentH
	if endIdx > len(lines) {
		endIdx = len(lines)
	}
	visibleMenu = lines[startIdx:endIdx]

	controlsBody := strings.Join(visibleMenu, "\n")

	// Render Left Frame with Title in Border
	leftBox := lipgloss.NewStyle().
		Border(frameBorder).
		BorderForeground(colorBorder).
		Width(leftPaneW).
		Height(contentH).
		Render(controlsBody)

	// Embed "─── Controls ───" header onto top border
	leftBoxLines := strings.Split(leftBox, "\n")
	if len(leftBoxLines) > 0 {
		topBar := leftBoxLines[0]
		header := "─ Controls "
		if len(topBar) > len(header)+3 {
			leftBoxLines[0] = topBar[:3] + headerStyle.Render(header) + topBar[3+len(header):]
		}
		leftBox = strings.Join(leftBoxLines, "\n")
	}

	// --- Right Pane (Live Preview) ---
	rightBox := lipgloss.NewStyle().
		Border(frameBorder).
		BorderForeground(colorBorder).
		Width(rightPaneW).
		Height(contentH).
		Render(m.asciiArt)

	// Embed "─── Live Preview ───" header onto top border
	rightBoxLines := strings.Split(rightBox, "\n")
	if len(rightBoxLines) > 0 {
		topBar := rightBoxLines[0]
		header := "─ Live Preview ─ CATGEN "
		if len(topBar) > len(header)+3 {
			rightBoxLines[0] = topBar[:3] + headerStyle.Render(header) + topBar[3+len(header):]
		}
		rightBox = strings.Join(rightBoxLines, "\n")
	}

	// Join Panes Horizontally
	mainLayout := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, " ", rightBox)

	// --- Footer (BANGEN Style) ---
	footerLeft := lipgloss.JoinHorizontal(lipgloss.Top,
		footerKeyStyle.Render("↑↓"), footerDescStyle.Render(" navigate  "),
		footerKeyStyle.Render("↔"), footerDescStyle.Render(" adjust  "),
		footerKeyStyle.Render("Enter"), footerDescStyle.Render(" edit/toggle  "),
		footerKeyStyle.Render("r"), footerDescStyle.Render(" random  "),
		footerKeyStyle.Render("e"), footerDescStyle.Render(" export  "),
		footerKeyStyle.Render("q"), footerDescStyle.Render(" quit"),
	)

	status := ""
	if m.statusMsg != "" {
		status = msgStyle.Render(m.statusMsg)
	}

	footer := lipgloss.JoinHorizontal(lipgloss.Top, footerLeft, "    ", status)

	return lipgloss.JoinVertical(lipgloss.Left, mainLayout, footer)
}
