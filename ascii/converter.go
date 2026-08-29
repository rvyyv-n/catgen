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

// Convert takes an image and converts it to an ASCII string strictly within bounds
func Convert(img image.Image, opts Options) string {
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
		return ""
	}

	// Font aspect ratio correction (terminal characters are roughly twice as tall as wide)
	const fontAspect = 0.46
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

	rampRunes := []rune(opts.DensityRamp)
	rampLen := len(rampRunes)
	var builder strings.Builder

	for y := 0; y < targetH; y++ {
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
			lum = (lum/255.0 - 0.5) * opts.Contrast + 0.5 + (float64(opts.Brightness) / 100.0)
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
			char := string(rampRunes[charIdx])

			// Apply color themes
			switch opts.Theme {
			case ThemeTrueColor:
				// Boost color with brightness/contrast
				cr := math.Min(255, math.Max(0, (r8-128)*opts.Contrast+128+float64(opts.Brightness)*2.55))
				cg := math.Min(255, math.Max(0, (g8-128)*opts.Contrast+128+float64(opts.Brightness)*2.55))
				cb := math.Min(255, math.Max(0, (b8-128)*opts.Contrast+128+float64(opts.Brightness)*2.55))
				builder.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s", int(cr), int(cg), int(cb), char))

			case ThemeMatrix:
				// Hacker Green glow based on luminance
				mg := int(lum * 255)
				mr := int(lum * 30)
				mb := int(lum * 50)
				builder.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s", mr, mg, mb, char))

			case ThemeCyberpunk:
				// Gradient between Neon Cyan (#00f0ff) and Neon Magenta (#ff007f)
				blend := (float64(x)/float64(targetW) + lum) / 2.0
				cr := int((1.0-blend)*0 + blend*255)
				cg := int((1.0-blend)*240 + blend*0)
				cb := int((1.0-blend)*255 + blend*130)
				builder.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s", cr, cg, cb, char))

			case ThemeAmber:
				// Vintage CRT Amber (#ffb000)
				ar := int(lum * 255)
				ag := int(lum * 176)
				ab := int(lum * 20)
				builder.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s", ar, ag, ab, char))

			case ThemeIceBlue:
				// Frost Cyan (#00ffff)
				ir := int(lum * 50)
				ig := int(lum * 230)
				ib := int(lum * 255)
				builder.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s", ir, ig, ib, char))

			case ThemeGrayscale:
				fallthrough
			default:
				// Monochrome / Terminal Default
				builder.WriteString(char)
			}
		}
		if opts.Theme != ThemeGrayscale {
			builder.WriteString("\x1b[0m")
		}
		if y < targetH-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}
