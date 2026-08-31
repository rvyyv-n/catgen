// Package imgsrc resolves an arbitrary image reference — a local file path or an
// http(s) URL — into a decoded image ready for ASCII conversion. It is the shared
// loading layer used by the CLI, the TUI, and (later) the server, so every entry
// point accepts the same set of formats and the same path conveniences.
package imgsrc

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"  // register GIF decoder
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // register WebP decoder (decode-only)
)

const (
	// MaxDownloadBytes caps how much of a remote image we will read.
	MaxDownloadBytes = 25 << 20 // 25 MiB
	// MaxPixelDim is the largest long-edge size we keep; bigger sources are
	// downscaled so ASCII sampling stays fast on full-resolution photos.
	MaxPixelDim = 4000
	// httpTimeout bounds a single remote fetch.
	httpTimeout = 15 * time.Second
)

// SupportedExts is the set of file extensions LoadImage can decode. SVG is
// intentionally absent: the standard library cannot rasterize it.
var SupportedExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
}

// Ref normalizes a raw, user-supplied image reference. It trims surrounding
// whitespace, strips one pair of wrapping quotes (terminals paste drag-and-dropped
// paths quoted), and expands a leading "~" to the user's home directory for local
// paths. URLs are returned unchanged apart from trimming.
func Ref(raw string) string {
	s := strings.TrimSpace(raw)
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			s = strings.TrimSpace(s[1 : len(s)-1])
		}
	}
	if s == "~" || strings.HasPrefix(s, "~/") || strings.HasPrefix(s, `~\`) {
		if home, err := os.UserHomeDir(); err == nil {
			s = filepath.Join(home, s[1:])
		}
	}
	return s
}

// IsURL reports whether ref should be fetched over HTTP rather than read from disk.
func IsURL(ref string) bool {
	return strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://")
}

// LoadImage resolves ref, decodes it, and downscales very large images. It returns
// the decoded image and the detected format name ("png", "jpeg", "gif", "webp").
func LoadImage(ref string) (image.Image, string, error) {
	ref = Ref(ref)
	if ref == "" {
		return nil, "", errors.New("no image reference given")
	}

	var src io.Reader
	if IsURL(ref) {
		rc, err := fetch(ref)
		if err != nil {
			return nil, "", err
		}
		defer rc.Close()
		src = rc
	} else {
		f, err := os.Open(ref)
		if err != nil {
			return nil, "", fmt.Errorf("cannot open %q: %w", ref, err)
		}
		defer f.Close()
		src = f
	}

	img, format, err := image.Decode(src)
	if err != nil {
		return nil, "", fmt.Errorf("could not decode image %q: %w", ref, err)
	}
	return downscale(img, MaxPixelDim), format, nil
}

func fetch(url string) (io.ReadCloser, error) {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(url) //nolint:noctx // timeout is set on the client
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("fetch failed: HTTP %s", resp.Status)
	}
	return &limitReadCloser{
		r: io.LimitReader(resp.Body, MaxDownloadBytes),
		c: resp.Body,
	}, nil
}

type limitReadCloser struct {
	r io.Reader
	c io.Closer
}

func (l *limitReadCloser) Read(p []byte) (int, error) { return l.r.Read(p) }
func (l *limitReadCloser) Close() error               { return l.c.Close() }

// downscale shrinks img so its longest edge is at most maxDim pixels, preserving
// aspect ratio. Images already within the limit are returned unchanged.
func downscale(img image.Image, maxDim int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxDim && h <= maxDim {
		return img
	}

	scale := float64(maxDim) / float64(w)
	if h > w {
		scale = float64(maxDim) / float64(h)
	}
	nw := max(1, int(float64(w)*scale))
	nh := max(1, int(float64(h)*scale))

	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)
	return dst
}
