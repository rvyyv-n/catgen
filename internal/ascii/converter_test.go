package ascii

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

func solid(w, h int) image.Image {
	return image.NewRGBA(image.Rect(0, 0, w, h))
}

// ramp8x6 is a small deterministic gradient used to lock converter output.
func ramp8x6() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 8, 6))
	for y := 0; y < 6; y++ {
		for x := 0; x < 8; x++ {
			v := uint8((x + y) * 20)
			img.Set(x, y, color.RGBA{v, uint8(x * 30), uint8(y * 40), 255})
		}
	}
	return img
}

// TestConvertGolden pins the ANSI byte stream for representative theme/ramp
// combos so the grid-IR refactor (and any future change) stays byte-identical
// to the reference implementation.
func TestConvertGolden(t *testing.T) {
	img := ramp8x6()
	cases := []struct {
		name string
		opts Options
		want string
	}{
		{
			"grayscale",
			Options{Width: 8, Theme: ThemeGrayscale, DensityRamp: RampStandard, Contrast: 1},
			" .:-=+*#\n .:-=+*#",
		},
		{
			"truecolor",
			Options{Width: 8, Theme: ThemeTrueColor, DensityRamp: RampBlocks, Contrast: 1},
			"\x1b[38;2;0;0;0m \x1b[38;2;20;30;0m \x1b[38;2;40;60;0m░\x1b[38;2;60;90;0m░\x1b[38;2;80;120;0m▒\x1b[38;2;100;150;0m▒\x1b[38;2;120;180;0m▓\x1b[38;2;140;210;0m▓\x1b[0m\n\x1b[38;2;60;0;120m \x1b[38;2;80;30;120m \x1b[38;2;100;60;120m░\x1b[38;2;120;90;120m░\x1b[38;2;140;120;120m▒\x1b[38;2;160;150;120m▒\x1b[38;2;180;180;120m▓\x1b[38;2;200;210;120m▓\x1b[0m",
		},
		{
			"matrix-inverted",
			Options{Width: 8, Theme: ThemeMatrix, DensityRamp: RampBraille, Invert: true, Contrast: 1},
			"\x1b[38;2;30;255;50m⠿\x1b[38;2;26;229;44m⠿\x1b[38;2;23;203;39m⠟\x1b[38;2;20;177;34m⠏\x1b[38;2;17;152;29m⠏\x1b[38;2;14;126;24m⠇\x1b[38;2;11;100;19m⠃\x1b[38;2;8;75;14m⠃\x1b[0m\n\x1b[38;2;27;233;45m⠿\x1b[38;2;24;207;40m⠟\x1b[38;2;21;182;35m⠟\x1b[38;2;18;156;30m⠏\x1b[38;2;15;130;25m⠇\x1b[38;2;12;105;20m⠃\x1b[38;2;9;79;15m⠃\x1b[38;2;6;53;10m⠁\x1b[0m",
		},
	}
	for _, c := range cases {
		if got := Convert(img, c.opts); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
}

// TestConvertToDiscordGolden pins the collapsed Discord path (now built on
// ConvertGrid) to its reference byte stream.
func TestConvertToDiscordGolden(t *testing.T) {
	want := "```ansi\n\x1b[30m     ....::::----=====++++\x1b[32m****\x1b[33m####\x1b[0m\n\x1b[30m     ....::::----=====++++\x1b[32m****\x1b[33m####\x1b[0m\n\x1b[30m     ....::::----=====++++\x1b[32m****\x1b[33m####\x1b[0m\n\x1b[30m     ....::::----=====++++\x1b[32m****\x1b[33m####\x1b[0m\n\x1b[30m     ....::::----=====\x1b[32m++++\x1b[33m****####\x1b[0m\n\x1b[30m     ....::::----=====\x1b[32m++++\x1b[33m****####\x1b[0m\n\x1b[30m     ....::::----\x1b[35m=====++++\x1b[33m****####\x1b[0m\n\x1b[30m     ....::::----\x1b[35m=====++++\x1b[33m****####\x1b[0m\n\x1b[30m.....\x1b[34m::::----\x1b[35m====+++++****\x1b[33m####%%%%\x1b[0m\n\x1b[30m.....\x1b[34m::::----\x1b[35m====+++++****\x1b[33m####%%%%\x1b[0m\n\x1b[34m.....::::----====\x1b[35m+++++****\x1b[37m####%%%%\x1b[0m\n```"
	if got := ConvertToDiscord(ramp8x6(), true, RampStandard); got != want {
		t.Errorf("Discord output drifted:\n got %q\nwant %q", got, want)
	}
}

// TestConvertUsesGrid asserts Convert is exactly renderANSI over ConvertGrid.
func TestConvertUsesGrid(t *testing.T) {
	img := ramp8x6()
	opts := Options{Width: 8, Theme: ThemeCyberpunk, DensityRamp: RampDetailed, Contrast: 1.2, Brightness: 10}
	if Convert(img, opts) != renderANSI(ConvertGrid(img, opts), opts.Theme) {
		t.Error("Convert diverged from renderANSI(ConvertGrid(...))")
	}
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
