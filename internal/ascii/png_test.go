package ascii

import "testing"

func TestRenderPNGDimensionsAndInk(t *testing.T) {
	grid := ConvertGrid(ramp8x6(), Options{Width: 8, Theme: ThemeTrueColor, DensityRamp: RampBlocks, Contrast: 1})
	if len(grid) == 0 {
		t.Fatal("empty grid from ConvertGrid")
	}

	img, err := RenderPNG(grid, 2)
	if err != nil {
		t.Fatalf("RenderPNG: %v", err)
	}

	rows, cols := len(grid), len(grid[0])
	b := img.Bounds()
	if b.Dx()%cols != 0 || b.Dy()%rows != 0 {
		t.Errorf("image %dx%d is not a whole multiple of the %dx%d grid", b.Dx(), b.Dy(), cols, rows)
	}
	if b.Dx() <= cols || b.Dy() <= rows {
		t.Errorf("image %dx%d too small for a %dx%d grid", b.Dx(), b.Dy(), cols, rows)
	}

	// At least one pixel must differ from the background: glyphs were drawn.
	inked := false
	for y := b.Min.Y; y < b.Max.Y && !inked; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			br, bg, bb, _ := PNGBackground.RGBA()
			if r != br || g != bg || bl != bb {
				inked = true
				break
			}
		}
	}
	if !inked {
		t.Error("rendered image is entirely background — no glyphs drawn")
	}
}

func TestRenderPNGEmptyGrid(t *testing.T) {
	if _, err := RenderPNG(nil, 1); err == nil {
		t.Error("expected error for nil grid")
	}
	if _, err := RenderPNG([][]Cell{}, 1); err == nil {
		t.Error("expected error for empty grid")
	}
}
