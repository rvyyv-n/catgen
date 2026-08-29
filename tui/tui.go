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

var (
	paneStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	activePaneStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("86")).
		Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	selectedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	footerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)
)

var ramps = []string{
	ascii.RampStandard,
	ascii.RampBlock,
	ascii.RampBraille,
}
var rampNames = []string{"Standard", "Block", "Braille"}

type Model struct {
	ImageDir  string
	Images    []string
	ImageIdx  int
	
	Width     int
	Colorize  bool
	Invert    bool
	RampIdx   int

	cursor    int
	termW     int
	termH     int
	asciiArt  string
	err       error
}

func NewModel(imageDir string, images []string) Model {
	m := Model{
		ImageDir: imageDir,
		Images:   images,
		ImageIdx: 0,
		Width:    60,
		Colorize: true,
		Invert:   false,
		RampIdx:  0,
		cursor:   0,
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return nil // Triggered externally or on first size msg
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
		if m.asciiArt == "" {
			m.updateAscii()
		}
	}
	return m, nil
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
		maxW := m.termW - 45 // Leave room for left pane
		if m.Width > maxW && maxW > 10 {
			m.Width = maxW
		}
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
	controls.WriteString(titleStyle.Render("CATS AS A SERVICE") + "\n\n")

	renderItem := func(idx int, label string, val string) {
		cursor := "  "
		style := itemStyle
		if m.cursor == idx {
			cursor = "▶ "
			style = selectedItemStyle
		}
		controls.WriteString(style.Render(fmt.Sprintf("%s%-10s %s\n", cursor, label, val)))
	}

	imgName := "None"
	if len(m.Images) > 0 {
		imgName = filepath.Base(m.Images[m.ImageIdx])
		if len(imgName) > 15 {
			imgName = imgName[:12] + "..."
		}
	}

	renderItem(0, "Image:", imgName)
	renderItem(1, "Width:", fmt.Sprintf("%d", m.Width))
	renderItem(2, "Color:", fmt.Sprintf("%v", m.Colorize))
	renderItem(3, "Style:", rampNames[m.RampIdx])
	renderItem(4, "Invert:", fmt.Sprintf("%v", m.Invert))

	leftPane := activePaneStyle.
		Width(35).
		Height(m.termH - 6).
		Render(controls.String())

	// --- Right Pane (Preview) ---
	rightPane := paneStyle.
		Width(m.termW - 42).
		Height(m.termH - 6).
		Render(m.asciiArt)

	// Join panes
	layout := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, "  ", rightPane)

	// Footer
	footer := footerStyle.Render("↑/↓: Navigate • ←/→: Adjust • q: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, layout, footer)
}
