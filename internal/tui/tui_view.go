package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"cats/internal/ascii"
)

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

	logoLine1 := logoCatStyle.Render("█▀▀ ▄▀█ ▀█▀") + " " + logoGenStyle.Render("█▀▀ █▀▀ █▄ █")
	logoLine2 := logoCatStyle.Render("█▄▄ █▀█  █ ") + " " + logoGenStyle.Render("█▄█ ██▄ █ ▀█")
	allRows = append(allRows, RenderedRow{
		isSection: true,
		text:      lipgloss.PlaceHorizontal(innerLeftW, lipgloss.Center, logoLine1),
	})
	allRows = append(allRows, RenderedRow{
		isSection: true,
		text:      lipgloss.PlaceHorizontal(innerLeftW, lipgloss.Center, logoLine2),
	})
	allRows = append(allRows, RenderedRow{
		isSection: true,
		text:      lipgloss.PlaceHorizontal(innerLeftW, lipgloss.Center, subLogoStyle.Render("── ASCII Cat Studio ──")),
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

	// Size the controls frame to hug its content, capped at the terminal.
	leftBoxH := len(allRows) + 2
	if leftBoxH > contentH {
		leftBoxH = contentH
	}
	if leftBoxH < 8 {
		leftBoxH = 8
	}
	innerHeight := leftBoxH - 2

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

	leftBox := buildFramedBox("", leftBody, leftPaneW, leftBoxH, colorBorder, colorTeal)

	// --- Right Pane (Live Preview) ---
	// Size the frame to hug the art instead of stretching to the whole pane,
	// so a small render doesn't sit in a large empty box.
	artW := lipgloss.Width(m.asciiArt)
	artH := lipgloss.Height(m.asciiArt)

	previewW := artW + 4
	if previewW > rightPaneW {
		previewW = rightPaneW
	}
	if previewW < 24 {
		previewW = 24
	}
	previewH := artH + 2
	if previewH > contentH {
		previewH = contentH
	}
	if previewH < 8 {
		previewH = 8
	}

	centeredArt := lipgloss.Place(
		previewW-2,
		previewH-2,
		lipgloss.Center,
		lipgloss.Center,
		m.asciiArt,
	)

	rightBox := buildFramedBox("Live Preview", centeredArt, previewW, previewH, colorBorder, colorTeal)

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
			footerKeyStyle.Render("s"), footerDescStyle.Render(" save  "),
			footerKeyStyle.Render("p"), footerDescStyle.Render(" presets  "),
			footerKeyStyle.Render("c"), footerDescStyle.Render(" chrome  "),
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
	plain, png := "○ Plain text", "○ Image (PNG)"
	if m.exportPNG {
		png = "◉ Image (PNG)"
	} else {
		plain = "◉ Plain text"
	}
	fmtLabel, pathLabel := "  Format:  ", "  Output:  "
	if m.exportField == 0 {
		fmtLabel = "► Format:  "
	} else {
		pathLabel = "► Output:  "
	}
	return fmtLabel + plain + "    " + png + "\n" + pathLabel + m.input.View()
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
