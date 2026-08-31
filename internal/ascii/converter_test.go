package ascii

import (
	"image"
	"strings"
	"testing"
)

func solid(w, h int) image.Image {
	return image.NewRGBA(image.Rect(0, 0, w, h))
}

func TestMeasureMatchesConvertRowCount(t *testing.T) {
	img := solid(200, 100)
	opts := Options{Width: 60, Theme: ThemeGrayscale, DensityRamp: RampBlocks, Contrast: 1.0}

	w, h := Measure(img, opts)
	if w <= 0 || h <= 0 {
		t.Fatalf("Measure returned non-positive dims: %dx%d", w, h)
	}

	// Grayscale output is plain text, so row/column counts are exact.
	rows := strings.Split(Convert(img, opts), "\n")
	if len(rows) != h {
		t.Errorf("Convert produced %d rows, Measure said %d", len(rows), h)
	}
	if got := len([]rune(rows[0])); got != w {
		t.Errorf("Convert row width = %d, Measure said %d", got, w)
	}
}

func TestMeasureAutoFitRespectsBounds(t *testing.T) {
	img := solid(400, 400)
	w, h := Measure(img, Options{AutoFit: true, MaxWidth: 50, MaxHeight: 20})
	if w > 50 || h > 20 {
		t.Errorf("Auto-Fit grid %dx%d exceeds 50x20 bounds", w, h)
	}
}

func TestThemeAndRampByName(t *testing.T) {
	if ThemeByName("matrix") != 2 {
		t.Errorf("ThemeByName(matrix) = %d, want 2", ThemeByName("matrix"))
	}
	if ThemeByName("nope") != -1 {
		t.Errorf("ThemeByName(nope) = %d, want -1", ThemeByName("nope"))
	}
	if RampByName("binary") != 4 {
		t.Errorf("RampByName(binary) = %d, want 4", RampByName("binary"))
	}
}
