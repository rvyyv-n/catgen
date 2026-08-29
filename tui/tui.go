package tui

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"cats/ascii"
)

// --- Styles matching the BANGEN reference ---
var (
	// Very subtle borders
	paneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1, 2)

	// Slightly brighter border for left pane
	leftPaneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")). // Cyan/teal
		Bold(true).
		MarginBottom(1)

	sectionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")). // Pink section headers
		MarginBottom(1).
		MarginTop(1)

	labelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")). // Gray
		Width(10)

	valueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(18)

	// Highlighted row style (teal background, black text)
	activeRowStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("86")).
		Foreground(lipgloss.Color("16")). // Black text
		Width(28).                        // Match left pane internal width
		Padding(0, 1)

	inactiveRowStyle = lipgloss.NewStyle().
		Width(28).
		Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1).
		MarginLeft(2)
)

var ramps = []string{
	ascii.RampStandard,
	ascii.RampBlock,
	ascii.RampBraille,
}
var rampNames = []string{"Standard", "Block", "Braille"}

type Model struct {
	ImageDir string
	Images   []string
	ImageIdx int

	Width    int
	Colorize bool
	Invert   bool
	RampIdx  int

	cursor   int
	termW    int
	termH    int
	asciiArt string
	err      error
}

func NewModel(imageDir string, images []string) Model {
	m := Model{
		ImageDir: imageDir,
		Images:   images,
		ImageIdx: 0,
		Width:    45, // Smaller default for laptops
		Colorize: true,
		Invert:   false,
		RampIdx:  0,
		cursor:   0,
	}
	return m
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
			if m.cursor < 4 {
				m.cursor++
			}
		case "left", "h":
			m.adjustValue(-1)
			m.updateAscii()
		case "right", "l", "enter", " ":
			m.adjustValue(1)
			m.updateAscii()
		}
	case tea.WindowSizeMsg:
		m.termW = msg.Width
		m.termH = msg.Height
		// Re-clamp width so it doesn't overflow right pane on resize
		m.clampWidth()
		if m.asciiArt == "" {
			m.updateAscii()
		}
	}
	return m, nil
}

func (m *Model) clampWidth() {
	maxW := m.termW - 40 // Leave 35+ padding for left pane
	if m.Width > maxW && maxW > 10 {
		m.Width = maxW
	}
}

func (m *Model) adjustValue(delta int) {
	switch m.cursor {
	case 0: // Image
		if len(m.Images) == 0 {
			return
		}
		m.ImageIdx += delta
		if m.ImageIdx < 0 {
			m.ImageIdx = len(m.Images) - 1
		} else if m.ImageIdx >= len(m.Images) {
			m.ImageIdx = 0
		}
	case 1: // Width
		m.Width += (delta * 5)
		if m.Width < 10 {
			m.Width = 10
		}
		m.clampWidth()
	case 2: // Colorize
		m.Colorize = !m.Colorize
	case 3: // Ramp
		m.RampIdx += delta
		if m.RampIdx < 0 {
			m.RampIdx = len(ramps) - 1
		} else if m.RampIdx >= len(ramps) {
			m.RampIdx = 0
		}
	case 4: // Invert
		m.Invert = !m.Invert
	}
}

func (m *Model) updateAscii() {
	if len(m.Images) == 0 {
		m.asciiArt = "No images found."
		return
	}

	imgPath := filepath.Join(m.ImageDir, filepath.FromSlash(m.Images[m.ImageIdx]))
	f, err := os.Open(imgPath)
	if err != nil {
		m.err = err
		m.asciiArt = fmt.Sprintf("Error opening image: %v", err)
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		m.err = err
		m.asciiArt = fmt.Sprintf("Error decoding image: %v", err)
		return
	}

	opts := ascii.Options{
		Width:       m.Width,
		Colorize:    m.Colorize,
		Invert:      m.Invert,
		DensityRamp: ramps[m.RampIdx],
	}

	m.asciiArt = ascii.Convert(img, opts)
	m.err = nil
}

func (m Model) View() string {
	if m.termW == 0 {
		return "Initializing..."
	}

	// --- Left Pane (Controls) ---
	var controls strings.Builder
	controls.WriteString(titleStyle.Render("CATS AS A SERVICE") + "\n")
	controls.WriteString(sectionStyle.Render("Controls") + "\n")

	// Helper to render a perfectly aligned row
	renderRow := func(idx int, label string, val string) {
		// Truncate long values cleanly
		if len(val) > 15 {
			val = val[:12] + "..."
		}

		content := lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render(label),
			valueStyle.Render(val),
		)

		if m.cursor == idx {
			// Apply BANGEN style highlight (solid background)
			controls.WriteString(activeRowStyle.Render(content) + "\n")
		} else {
			controls.WriteString(inactiveRowStyle.Render(content) + "\n")
		}
	}

	imgName := "None"
	if len(m.Images) > 0 {
		imgName = filepath.Base(m.Images[m.ImageIdx])
	}

	renderRow(0, "Image", imgName)
	renderRow(1, "Width", fmt.Sprintf("%d", m.Width))
	
	controls.WriteString(sectionStyle.Render("Effects") + "\n")
	renderRow(2, "Color", fmt.Sprintf("%v", m.Colorize))
	renderRow(3, "Style", rampNames[m.RampIdx])
	renderRow(4, "Invert", fmt.Sprintf("%v", m.Invert))

	leftPane := leftPaneStyle.
		Width(34).
		Height(m.termH - 4). // Adjusting height perfectly to terminal size
		Render(controls.String())

	// --- Right Pane (Preview) ---
	rightPaneWidth := m.termW - 40 // 34 for left pane + padding + margins
	if rightPaneWidth < 10 {
		rightPaneWidth = 10
	}

	// Calculate right pane height based on terminal height
	paneHeight := m.termH - 4
	if paneHeight < 5 {
		paneHeight = 5
	}

	rightPane := paneStyle.
		Width(rightPaneWidth).
		Height(paneHeight).
		Render(m.asciiArt)

	// Join panes
	layout := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, "  ", rightPane)

	// Footer (BANGEN style keybinds)
	footer := footerStyle.Render("↑↓ navigate  ↔ adjust  Enter toggle  q quit")

	return lipgloss.JoinVertical(lipgloss.Left, layout, footer)
}
