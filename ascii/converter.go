package ascii

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

// Density ramps from dark to light
const (
	RampStandard = " .:-=+*#%@"
	RampBlock    = " ░▒▓█"
	RampBraille  = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
)

// Options holds configuration for the ASCII generation
type Options struct {
	Width       int
	Colorize    bool
	Invert      bool
	DensityRamp string
}

// Convert takes an image and converts it to an ASCII string
func Convert(img image.Image, opts Options) string {
	if opts.Width <= 0 {
		opts.Width = 80 // default terminal width
	}
	if opts.DensityRamp == "" {
		opts.DensityRamp = RampStandard
	}

	bounds := img.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()

	// Terminal characters are typically taller than they are wide.
	// We multiply by ~0.45 to compensate for font aspect ratios.
	aspectRatio := float64(origH) / float64(origW)
	targetH := int(float64(opts.Width) * aspectRatio * 0.45)

	if targetH <= 0 {
		targetH = 1
	}

	rampRunes := []rune(opts.DensityRamp)
	rampLen := len(rampRunes)
	var builder strings.Builder

	for y := 0; y < targetH; y++ {
		for x := 0; x < opts.Width; x++ {
			// Nearest neighbor sampling for speed and simplicity
			srcX := bounds.Min.X + (x * origW / opts.Width)
			srcY := bounds.Min.Y + (y * origH / targetH)

			r, g, b, _ := img.At(srcX, srcY).RGBA()

			// Convert to 8-bit color depth (0-255)
			r8, g8, b8 := r>>8, g>>8, b>>8

			// Calculate standard luminance
			lum := 0.2126*float64(r8) + 0.7152*float64(g8) + 0.0722*float64(b8)
			lumNorm := lum / 255.0
			
			if opts.Invert {
				lumNorm = 1.0 - lumNorm
			}

			// Map to character
			charIdx := int(lumNorm * float64(rampLen-1))
			if charIdx < 0 {
				charIdx = 0
			}
			if charIdx >= rampLen {
				charIdx = rampLen - 1
			}

			char := string(rampRunes[charIdx])

			if opts.Colorize {
				// ANSI 24-bit TrueColor foreground
				builder.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s", r8, g8, b8, char))
			} else {
				builder.WriteString(char)
			}
		}
		if opts.Colorize {
			// Reset colors at the end of each line
			builder.WriteString("\x1b[0m")
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
