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

	"cats/internal/ascii"
)

// --- Color Palette (Matching BANGEN Theme) ---
var (
	colorTeal     = lipgloss.Color("86")  // #00d7af (Active Teal)
	colorDark     = lipgloss.Color("16")  // #000000 (Black text for highlight)
	colorPink     = lipgloss.Color("205") // #ff5faf (Section headers)
	colorMuted    = lipgloss.Color("241") // #626262 (Labels & muted text)
	colorBorder   = lipgloss.Color("37")  // #00afaf (Subtle frame border)
	colorText     = lipgloss.Color("252") // #d0d0d0 (Light text)
	colorSuccess  = lipgloss.Color("42")  // #00d787 (Success green)

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

type ItemType int

const (
	ItemProperty ItemType = iota
	ItemThemeRadio
	ItemRampRadio
)

type MenuItem struct {
	ID        string
	Type      ItemType
	Section   string
	Label     string
	ValueFunc func(m *Model) string
	Index     int
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
	rng        *rand.Rand
}

func NewModel(imageDir string, images []string) Model {
	m := Model{
		ImageDir:   imageDir,
		Images:     images,
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
			Label:   t.Name,
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
			Label:   r.Name,
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
			if len(m.Images) > 0 {
				m.ImageIdx = m.rng.Intn(len(m.Images))
				m.statusMsg = "✓ Loaded random cat!"
				m.updateAscii()
			}

		case "e":
			m.exportToFile()

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

func (m *Model) exportToFile() {
	if m.asciiArt == "" {
		return
	}
	filename := "cat_ascii.txt"
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

func (m *Model) exportToDiscord() {
	if len(m.Images) == 0 {
		return
	}
	imgPath := filepath.Join(m.ImageDir, filepath.FromSlash(m.Images[m.ImageIdx]))
	f, err := os.Open(imgPath)
	if err != nil {
		m.statusMsg = "Error reading image for Discord export"
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		m.statusMsg = "Error decoding image for Discord export"
		return
	}

	colorize := (m.ThemeIdx != 1) // Colorize unless Grayscale
	discordSnippet := ascii.ConvertToDiscord(img, colorize, ramps[m.RampIdx].Ramp)
	filename := "cat_discord.txt"
	_ = os.WriteFile(filename, []byte(discordSnippet), 0644)
	m.statusMsg = fmt.Sprintf("✓ Exported Discord snippet to %s!", filename)
}

func (m *Model) updateAscii() {
	if len(m.Images) == 0 {
		m.asciiArt = "No images found."
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

	leftPaneW := 28
	availW := m.termW - leftPaneW - 5
	availH := m.termH - 4

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
	if titleLen >= innerW {
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

	leftPaneW := 28
	rightPaneW := m.termW - leftPaneW - 2
	if rightPaneW < 15 {
		rightPaneW = 15
	}
	contentH := m.termH - 2
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
			radio := "○"
			if m.ThemeIdx == item.Index {
				radio = "●"
			}
			rowBody = fmt.Sprintf("%s %s", radio, item.Label)

		case ItemRampRadio:
			radio := "○"
			if m.RampIdx == item.Index {
				radio = "●"
			}
			rowBody = fmt.Sprintf("%s %s", radio, item.Label)
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

	leftBox := buildFramedBox("CATGEN", leftBody, leftPaneW, contentH, colorBorder, colorTeal)

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

	// Seamlessly join frames horizontally
	mainLayout := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, " ", rightBox)

	// --- Footer (Right Aligned across total width) ---
	footerItems := lipgloss.JoinHorizontal(lipgloss.Top,
		footerKeyStyle.Render("↑↓"), footerDescStyle.Render(" navigate  "),
		footerKeyStyle.Render("↔"), footerDescStyle.Render(" adjust  "),
		footerKeyStyle.Render("Enter"), footerDescStyle.Render(" toggle  "),
		footerKeyStyle.Render("r"), footerDescStyle.Render(" random  "),
		footerKeyStyle.Render("e"), footerDescStyle.Render(" export  "),
		footerKeyStyle.Render("d"), footerDescStyle.Render(" discord  "),
		footerKeyStyle.Render("q"), footerDescStyle.Render(" quit"),
	)

	if m.statusMsg != "" {
		footerItems = lipgloss.JoinHorizontal(lipgloss.Top, msgStyle.Render(m.statusMsg), "   ", footerItems)
	}

	totalW := leftPaneW + rightPaneW + 1
	footer := lipgloss.NewStyle().
		Width(totalW).
		Align(lipgloss.Right).
		Render(footerItems)

	return lipgloss.JoinVertical(lipgloss.Left, mainLayout, footer)
}
