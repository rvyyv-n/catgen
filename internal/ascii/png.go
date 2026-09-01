package ascii

import (
	_ "embed"
	"errors"
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// DejaVu Sans Mono covers the block (░▒▓█) and braille (⠿) ramp glyphs that the
// Go stdlib bundled faces do not. Licence: internal/ascii/fonts/LICENSE.
//
//go:embed fonts/DejaVuSansMono.ttf
var dejavuMonoTTF []byte

// pngBaseSize is the font size in points at scale 1. Cell metrics are derived
// from the face so glyphs never clip.
const pngBaseSize = 16

// PNGBackground is the fill behind the glyphs — a near-black that keeps dim
// colours legible.
var PNGBackground = color.RGBA{0x0d, 0x0d, 0x0d, 0xff}

// RenderPNG rasterises a Cell grid to an RGBA image: every glyph drawn in its
// cell colour on PNGBackground. scale multiplies the font size (and therefore
// the output resolution); values below 1 are treated as 1.
func RenderPNG(grid [][]Cell, scale int) (*image.RGBA, error) {
	if len(grid) == 0 || len(grid[0]) == 0 {
		return nil, errors.New("ascii: empty grid")
	}
	if scale < 1 {
		scale = 1
	}

	fnt, err := opentype.Parse(dejavuMonoTTF)
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    float64(pngBaseSize * scale),
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}
	defer face.Close()

	metrics := face.Metrics()
	cellH := (metrics.Ascent + metrics.Descent).Ceil()
	adv, ok := face.GlyphAdvance('M')
	if !ok || adv == 0 {
		return nil, errors.New("ascii: font has no advance width")
	}
	cellW := adv.Ceil()
	baseline := metrics.Ascent.Ceil()

	rows := len(grid)
	cols := len(grid[0])
	dst := image.NewRGBA(image.Rect(0, 0, cols*cellW, rows*cellH))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(PNGBackground), image.Point{}, draw.Src)

	drawer := &font.Drawer{Dst: dst, Face: face}
	for y, row := range grid {
		for x, c := range row {
			if c.Ch == ' ' || c.Ch == 0 {
				continue
			}
			drawer.Src = image.NewUniform(color.RGBA{c.R, c.G, c.B, 0xff})
			drawer.Dot = fixed.P(x*cellW, y*cellH+baseline)
			drawer.DrawString(string(c.Ch))
		}
	}
	return dst, nil
}
