// Package samples embeds the bundled cat images directly into the catgen
// binary so a single downloaded executable has a working image pool with no
// accompanying images/ folder. --dir and IMAGES_DIR still take priority when
// the caller points catgen at a real directory.
package samples

import (
	"embed"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed images
var fsys embed.FS

// root is the embedded directory the images live under.
const root = "images"

// Prefix marks an image reference as a name inside the embedded pool rather
// than a path on disk. Callers that build refs from Names should keep it.
const Prefix = "embedded:"

// Names returns the embedded sample images as "embedded:"-prefixed refs,
// sorted for a stable, deterministic pool order.
func Names() []string {
	var names []string
	_ = fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(p, root+"/")
		names = append(names, Prefix+rel)
		return nil
	})
	sort.Strings(names)
	return names
}

// Open opens an embedded image by ref (with or without the Prefix).
func Open(ref string) (fs.File, error) {
	name := strings.TrimPrefix(ref, Prefix)
	return fsys.Open(path.Join(root, name))
}
