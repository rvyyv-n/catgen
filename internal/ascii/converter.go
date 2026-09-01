package ascii

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"strings"
)

// Density ramps from dark to light
const (
	RampBlocks   = " ░▒▓█"
	RampStandard = " .:-=+*#%@"
	RampBraille  = " ⠁⠃⠇⠏⠟⠿"
	RampMinimal  = " .oO@"
	RampDetailed = " .'`^\",:;Il!i~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"
	RampBinary   = " 01"
)

// Theme enum
type Theme int

const (
	ThemeTrueColor Theme = iota
	ThemeGrayscale
	ThemeMatrix
	ThemeCyberpunk
	ThemeAmber
	ThemeIceBlue
)

// ThemeInfo pairs a short, stable identifier (used by CLI flags and presets)
// and a human-facing label with a Theme value.
type ThemeInfo struct {
	Name  string
	Label string
	Theme Theme
}

// Themes is the ordered theme catalog shared by the TUI and the --list-themes
// flag. The order matches the historical TUI ordering so stored selection
// indices stay valid.
var Themes = []ThemeInfo{
	{"truecolor", "TrueColor (RGB)", ThemeTrueColor},
	{"grayscale", "Grayscale", ThemeGrayscale},
	{"matrix", "Matrix Glow", ThemeMatrix},
	{"cyberpunk", "Cyberpunk", ThemeCyberpunk},
	{"amber", "Amber Phosphor", ThemeAmber},
	{"iceblue", "Ice Blue", ThemeIceBlue},
}

// RampInfo pairs a short, stable identifier and a label with a density ramp.
type RampInfo struct {
	Name  string
	Label string
	Chars string
}

// Ramps is the ordered character-ramp catalog shared by the TUI and the
// --list-ramps flag.
var Ramps = []RampInfo{
	{"blocks", "Blocks (░▒▓█)", RampBlocks},
	{"standard", "Standard (.-=+*)", RampStandard},
	{"braille", "Braille (⠋⠙⠹)", RampBraille},
	{"detailed", "Detailed ASCII", RampDetailed},
	{"binary", "Binary Matrix (01)", RampBinary},
	{"minimal", "Minimal ( .oO@)", RampMinimal},
}

// ThemeByName returns the catalog index of the named theme, or -1 if unknown.
func ThemeByName(name string) int {
	for i, t := range Themes {
		if t.Name == name {
			return i
		}
	}
	return -1
}

// RampByName returns the catalog index of the named ramp, or -1 if unknown.
func RampByName(name string) int {
	for i, r := range Ramps {
		if r.Name == name {
			return i
		}
	}
	return -1
}

// fontAspect corrects for terminal cells being roughly twice as tall as wide,
// so square source images map to square-looking character grids.
const fontAspect = 0.46

// Cell is one character of the ASCII grid: the glyph plus the colour it renders
// in. The colour is always populated (grayscale stores the luminance grey), so
// RenderPNG can draw every theme without re-sampling.
type Cell struct {
	Ch      rune
	R, G, B uint8
}

// Options holds configuration for the ASCII generation
type Options struct {
	Width       int
	MaxHeight   int
	MaxWidth    int
	AutoFit     bool
	Theme       Theme
	Invert      bool
	Brightness  int     // -50 to +50
	Contrast    float64 // 0.5 to 2.0 (1.0 = normal)
	DensityRamp string
}

// resolveDims computes the character-grid width and height for a source of
// origW x origH pixels under opts, applying font-aspect correction and the
// Auto-Fit / max-bound constraints.
func resolveDims(origW, origH int, opts Options) (int, int) {
	imgAspect := (float64(origH) / float64(origW)) * fontAspect

	targetW := opts.Width
	if targetW <= 0 {
		targetW = 45
	}
	targetH := int(float64(targetW) * imgAspect)

	// If AutoFit is enabled or bounds are given, strictly fit inside max bounds
	if opts.AutoFit && opts.MaxWidth > 0 && opts.MaxHeight > 0 {
		targetW = opts.MaxWidth
		targetH = int(float64(targetW) * imgAspect)
		if targetH > opts.MaxHeight {
			targetH = opts.MaxHeight
			targetW = int(float64(targetH) / imgAspect)
		}
		if targetW > opts.MaxWidth {
			targetW = opts.MaxWidth
			targetH = int(float64(targetW) * imgAspect)
		}
	} else {
		// Enforce MaxHeight limit if specified
		if opts.MaxHeight > 0 && targetH > opts.MaxHeight {
			targetH = opts.MaxHeight
			targetW = int(float64(targetH) / imgAspect)
		}
		if opts.MaxWidth > 0 && targetW > opts.MaxWidth {
			targetW = opts.MaxWidth
			targetH = int(float64(targetW) * imgAspect)
		}
	}

	if targetW <= 0 {
		targetW = 1
	}
	if targetH <= 0 {
		targetH = 1
	}
	return targetW, targetH
}

// Measure reports the character-grid dimensions Convert will produce for img
// under opts, without rendering. The TUI's fit-info toggle uses it to show the
// resolved output size.
func Measure(img image.Image, opts Options) (w, h int) {
	b := img.Bounds()
	if b.Dx() <= 0 || b.Dy() <= 0 {
		return 0, 0
	}
	return resolveDims(b.Dx(), b.Dy(), opts)
}

// Convert takes an image and converts it to an ANSI ASCII string strictly
// within bounds. It is the grid produced by ConvertGrid serialised to ANSI.
func Convert(img image.Image, opts Options) string {
	return renderANSI(ConvertGrid(img, opts), opts.Theme)
}

// ConvertGrid samples img into a character grid under opts. Each Cell carries
// its glyph and the RGB it should render in for the active theme. Returns nil
// for a zero-area image.
func ConvertGrid(img image.Image, opts Options) [][]Cell {
	if opts.DensityRamp == "" {
		opts.DensityRamp = RampBlocks
	}
	if opts.Contrast <= 0 {
		opts.Contrast = 1.0
	}

	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()
	if origW <= 0 || origH <= 0 {
		return nil
	}

	targetW, targetH := resolveDims(origW, origH, opts)

	rampRunes := []rune(opts.DensityRamp)
	rampLen := len(rampRunes)

	grid := make([][]Cell, targetH)
	for y := 0; y < targetH; y++ {
		row := make([]Cell, targetW)
		for x := 0; x < targetW; x++ {
			// Nearest neighbor coordinate mapping
			srcX := bounds.Min.X + (x * origW / targetW)
			srcY := bounds.Min.Y + (y * origH / targetH)

			r, g, b, _ := img.At(srcX, srcY).RGBA()
			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := float64(b >> 8)

			// Calculate standard luminance
			lum := 0.2126*r8 + 0.7152*g8 + 0.0722*b8

			// Apply Brightness & Contrast
			lum = (lum/255.0-0.5)*opts.Contrast + 0.5 + (float64(opts.Brightness) / 100.0)
			if lum < 0 {
				lum = 0
			}
			if lum > 1 {
				lum = 1
			}

			if opts.Invert {
				lum = 1.0 - lum
			}

			// Map luminance to character ramp
			charIdx := int(math.Floor(lum * float64(rampLen)))
			if charIdx < 0 {
				charIdx = 0
			}
			if charIdx >= rampLen {
				charIdx = rampLen - 1
			}

			cr, cg, cb := cellColor(opts, r8, g8, b8, lum, x, targetW)
			row[x] = Cell{Ch: rampRunes[charIdx], R: cr, G: cg, B: cb}
		}
		grid[y] = row
	}
	return grid
}

// cellColor computes the render colour for one cell under opts.Theme. The math
// mirrors the historical per-theme branches in Convert exactly.
func cellColor(opts Options, r8, g8, b8, lum float64, x, targetW int) (uint8, uint8, uint8) {
	switch opts.Theme {
	case ThemeTrueColor:
		// Boost color with brightness/contrast
		cr := math.Min(255, math.Max(0, (r8-128)*opts.Contrast+128+float64(opts.Brightness)*2.55))
		cg := math.Min(255, math.Max(0, (g8-128)*opts.Contrast+128+float64(opts.Brightness)*2.55))
		cb := math.Min(255, math.Max(0, (b8-128)*opts.Contrast+128+float64(opts.Brightness)*2.55))
		return uint8(cr), uint8(cg), uint8(cb)

	case ThemeMatrix:
		// Hacker Green glow based on luminance
		return uint8(lum * 30), uint8(lum * 255), uint8(lum * 50)

	case ThemeCyberpunk:
		// Gradient between Neon Cyan (#00f0ff) and Neon Magenta (#ff007f)
		blend := (float64(x)/float64(targetW) + lum) / 2.0
		cr := int((1.0-blend)*0 + blend*255)
		cg := int((1.0-blend)*240 + blend*0)
		cb := int((1.0-blend)*255 + blend*130)
		return uint8(cr), uint8(cg), uint8(cb)

	case ThemeAmber:
		// Vintage CRT Amber (#ffb000)
		return uint8(lum * 255), uint8(lum * 176), uint8(lum * 20)

	case ThemeIceBlue:
		// Frost Cyan (#00ffff)
		return uint8(lum * 50), uint8(lum * 230), uint8(lum * 255)

	default:
		// Grayscale / terminal default: store the luminance grey for RenderPNG.
		g := uint8(lum * 255)
		return g, g, g
	}
}

// renderANSI serialises a grid to the same ANSI byte stream the old Convert
// produced: TrueColor SGR per cell for the colour themes, bare glyphs for
// grayscale, a reset at each row end for every theme except grayscale, and no
// trailing newline.
func renderANSI(grid [][]Cell, theme Theme) string {
	if len(grid) == 0 {
		return ""
	}
	var builder strings.Builder
	for y, row := range grid {
		for _, c := range row {
			switch theme {
			case ThemeTrueColor, ThemeMatrix, ThemeCyberpunk, ThemeAmber, ThemeIceBlue:
				builder.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s", c.R, c.G, c.B, string(c.Ch)))
			default:
				builder.WriteString(string(c.Ch))
			}
		}
		if theme != ThemeGrayscale {
			builder.WriteString("\x1b[0m")
		}
		if y < len(grid)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

// discordPalette is Discord's renderable 16-colour ANSI subset (8 foregrounds).
var discordPalette = []struct {
	r, g, b int
	ansi    string
}{
	{50, 54, 60, "\x1b[30m"},    // Dark Gray
	{237, 66, 69, "\x1b[31m"},   // Red
	{87, 242, 135, "\x1b[32m"},  // Green
	{254, 231, 92, "\x1b[33m"},  // Yellow
	{88, 101, 242, "\x1b[34m"},  // Blurple / Blue
	{235, 69, 158, "\x1b[35m"},  // Pink / Magenta
	{0, 176, 244, "\x1b[36m"},   // Cyan
	{255, 255, 255, "\x1b[37m"}, // White
}

// closestDiscordANSI returns the nearest discordPalette SGR code to an RGB
// triple by squared Euclidean distance.
func closestDiscordANSI(r, g, b int) string {
	bestDist := math.MaxFloat64
	bestCode := "\x1b[37m"
	for _, c := range discordPalette {
		dr := float64(r - c.r)
		dg := float64(g - c.g)
		db := float64(b - c.b)
		dist := dr*dr + dg*dg + db*db
		if dist < bestDist {
			bestDist = dist
			bestCode = c.ansi
		}
	}
	return bestCode
}

// ConvertToDiscord generates a Discord-optimized ASCII snippet (< 1,500 chars, max width 34)
// wrapped in a ready-to-paste markdown codeblock (with Discord 16-color ANSI support).
func ConvertToDiscord(img image.Image, colorize bool, ramp string) string {
	if ramp == "" {
		ramp = RampStandard
	}
	grid := ConvertGrid(img, Options{
		Width:       34,
		MaxHeight:   22,
		Theme:       ThemeTrueColor,
		Contrast:    1.0,
		DensityRamp: ramp,
	})
	if len(grid) == 0 {
		return ""
	}

	var builder strings.Builder
	if colorize {
		builder.WriteString("```ansi\n")
	} else {
		builder.WriteString("```\n")
	}

	lastColor := ""
	for _, row := range grid {
		for _, c := range row {
			if colorize {
				colorCode := closestDiscordANSI(int(c.R), int(c.G), int(c.B))
				if colorCode != lastColor {
					builder.WriteString(colorCode)
					lastColor = colorCode
				}
			}
			builder.WriteString(string(c.Ch))
		}
		if colorize {
			builder.WriteString("\x1b[0m")
			lastColor = ""
		}
		builder.WriteString("\n")
	}
	builder.WriteString("```")
	return builder.String()
}
