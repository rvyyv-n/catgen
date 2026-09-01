package tui

import (
	"fmt"
	"strings"
	"time"

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
	// The frame stays edge-to-edge; only the art inside it resizes. Centre the
	// render in the full pane.
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
	case overlayExports:
		mainLayout = centeredModal("Exports", m.exportsModalBody(), totalW, contentH)
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
	case overlayExports:
		footer = hintFooter("↑↓ move · Enter open · c copy path · d delete · Esc close", m.statusMsg, totalW)
	default:
		pairs := []struct{ key, desc string }{
			{"↑↓", "nav"}, {"↔", "adjust"}, {"⏎", "toggle"},
			{"o", "open"}, {"r", "random"}, {"e", "export"}, {"s", "save"},
			{"p", "presets"}, {"x", "exports"}, {"t", "themes"}, {"a", "info"},
			{"q", "quit"},
		}
		var lead string
		if m.statusMsg != "" {
			lead = msgStyle.Render(m.statusMsg)
		} else if m.showFitInfo && m.curImg != nil {
			gw, gh := ascii.Measure(m.curImg, m.renderOpts())
			b := m.curImg.Bounds()
			lead = footerDescStyle.Render(fmt.Sprintf("fit:%s · src %dx%d · grid %dx%d",
				fitModes[m.FitModeIdx], b.Dx(), b.Dy(), gw, gh))
		}
		leadW := 0
		if lead != "" {
			leadW = lipgloss.Width(lead) + 3 // + the "   " separator
		}

		// Drop the least-essential hints, in this order, until the whole line
		// (lead + keybinds) fits without wrapping.
		drop := []string{"nav", "adjust", "toggle", "info", "random", "save"}
		footerItems := renderFooterKeys(pairs)
		for _, d := range drop {
			if footerFits(footerItems, totalW-leadW) {
				break
			}
			for i := range pairs {
				if pairs[i].desc == d {
					pairs = append(pairs[:i], pairs[i+1:]...)
					break
				}
			}
			footerItems = renderFooterKeys(pairs)
		}

		if lead != "" {
			footerItems = lipgloss.JoinHorizontal(lipgloss.Top, lead, "   ", footerItems)
		}

		// Hard cap the content width before aligning so it can never wrap onto
		// a second line and break the layout.
		footerItems = lipgloss.NewStyle().MaxWidth(totalW).Render(footerItems)

		footer = lipgloss.NewStyle().
			Width(totalW).
			Align(lipgloss.Right).
			Render(footerItems)
	}

	return lipgloss.JoinVertical(lipgloss.Left, mainLayout, "", footer)
}

// renderFooterKeys joins key/description pairs into the main keybind bar.
func renderFooterKeys(pairs []struct{ key, desc string }) string {
	parts := make([]string, 0, len(pairs)*2)
	for i, p := range pairs {
		sep := "  "
		if i == len(pairs)-1 {
			sep = ""
		}
		parts = append(parts, footerKeyStyle.Render(p.key), footerDescStyle.Render(" "+p.desc+sep))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// footerFits leaves a small margin for width-ambiguous glyphs (↑↓ ↔ ⏎) that
// some terminals render wider than lipgloss measures.
func footerFits(s string, totalW int) bool {
	return lipgloss.Width(s) <= totalW-2
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

// exportsModalBody renders the exports browser: one row per file with its size
// and age, newest first.
func (m Model) exportsModalBody() string {
	if len(m.exportList) == 0 {
		return "  (no exports yet — press e to make one)"
	}
	rows := make([]string, len(m.exportList))
	for i, e := range m.exportList {
		meta := fmt.Sprintf("%9s · %s", humanSize(e.Size), humanAge(time.Since(e.ModTime)))
		line := fmt.Sprintf("%-28s %s", truncName(e.Name, 28), meta)
		if i == m.exportListCursor {
			rows[i] = lipgloss.NewStyle().Foreground(colorTeal).Bold(true).Render("► " + line)
		} else {
			rows[i] = "  " + line
		}
	}
	return strings.Join(rows, "\n")
}

func humanSize(n int64) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.0f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func humanAge(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func truncName(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
