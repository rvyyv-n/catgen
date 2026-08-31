package imgsrc

import (
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// writeImage encodes a solid-color image of the given size to a temp file using
// the encoder matching ext, and returns its path.
func writeImage(t *testing.T, dir, name string, w, h int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	p := filepath.Join(dir, name)
	f, err := os.Create(p)
	if err != nil {
		t.Fatalf("create %s: %v", p, err)
	}
	defer f.Close()

	switch strings.ToLower(filepath.Ext(name)) {
	case ".png":
		err = png.Encode(f, img)
	case ".jpg", ".jpeg":
		err = jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
	case ".gif":
		err = gif.Encode(f, img, nil)
	default:
		t.Fatalf("unhandled ext for %s", name)
	}
	if err != nil {
		t.Fatalf("encode %s: %v", p, err)
	}
	return p
}

func TestRef(t *testing.T) {
	home, _ := os.UserHomeDir()
	cases := []struct {
		in   string
		want string
	}{
		{"  spaced.png  ", "spaced.png"},
		{`"quoted path.png"`, "quoted path.png"},
		{"'single.png'", "single.png"},
		{"  \"  trimmed.png  \"  ", "trimmed.png"},
		{"https://example.com/cat.png", "https://example.com/cat.png"},
		{"~", home},
		{"~/pics/cat.png", filepath.Join(home, "pics", "cat.png")},
	}
	for _, c := range cases {
		if got := Ref(c.in); got != c.want {
			t.Errorf("Ref(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsURL(t *testing.T) {
	for in, want := range map[string]bool{
		"http://x/y.png":  true,
		"https://x/y.png": true,
		"/local/y.png":    false,
		"y.png":           false,
		"ftp://x/y.png":   false,
	} {
		if got := IsURL(in); got != want {
			t.Errorf("IsURL(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestLoadImageLocalFormats(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name       string
		wantFormat string
	}{
		{"pic.png", "png"},
		{"pic.jpg", "jpeg"},
		{"pic.gif", "gif"},
	}
	for _, c := range cases {
		p := writeImage(t, dir, c.name, 40, 30)
		img, format, err := LoadImage(p)
		if err != nil {
			t.Fatalf("LoadImage(%s): %v", c.name, err)
		}
		if format != c.wantFormat {
			t.Errorf("LoadImage(%s) format = %q, want %q", c.name, format, c.wantFormat)
		}
		if img.Bounds().Dx() != 40 || img.Bounds().Dy() != 30 {
			t.Errorf("LoadImage(%s) bounds = %v, want 40x30", c.name, img.Bounds())
		}
	}
}

func TestLoadImageQuotedPath(t *testing.T) {
	dir := t.TempDir()
	p := writeImage(t, dir, "cat.png", 10, 10)
	if _, _, err := LoadImage(`"` + p + `"`); err != nil {
		t.Fatalf("LoadImage with quoted path: %v", err)
	}
}

func TestLoadImageMissing(t *testing.T) {
	_, _, err := LoadImage(filepath.Join(t.TempDir(), "nope.png"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadImageNotAnImage(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "fake.png")
	if err := os.WriteFile(p, []byte("this is not a PNG"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := LoadImage(p); err == nil {
		t.Fatal("expected decode error for non-image content, got nil")
	}
}

func TestLoadImageEmptyRef(t *testing.T) {
	if _, _, err := LoadImage("   "); err == nil {
		t.Fatal("expected error for empty reference, got nil")
	}
}

func TestLoadImageDownscalesLargeSource(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large-image encode in -short mode")
	}
	dir := t.TempDir()
	// Long edge well over MaxPixelDim; short edge tiny to keep the file small.
	p := writeImage(t, dir, "wide.png", MaxPixelDim+1200, 40)

	img, _, err := LoadImage(p)
	if err != nil {
		t.Fatalf("LoadImage: %v", err)
	}
	if got := img.Bounds().Dx(); got != MaxPixelDim {
		t.Errorf("downscaled width = %d, want %d", got, MaxPixelDim)
	}
	if got := img.Bounds().Dy(); got > MaxPixelDim {
		t.Errorf("downscaled height = %d, want <= %d", got, MaxPixelDim)
	}
}

func TestDownscaleKeepsSmallImageUntouched(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 100, 50))
	got := downscale(src, MaxPixelDim)
	if got != image.Image(src) {
		t.Error("downscale returned a new image for an already-small source")
	}
}

func TestRefTildeNeedsHome(t *testing.T) {
	// Guard the ~ branch on platforms without a resolvable home dir.
	if _, err := os.UserHomeDir(); err != nil {
		t.Skipf("no home dir on %s: %v", runtime.GOOS, err)
	}
}
