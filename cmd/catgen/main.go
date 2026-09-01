// Regenerate the Windows executable icon from favicon.ico after changing it:
//
//	go install github.com/akavel/rsrc@latest
//	go generate ./...
//
//go:generate rsrc -ico favicon.ico -arch amd64 -o rsrc_windows_amd64.syso
package main

import (
	"flag"
	"fmt"
	"image"
	"io/fs"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"cats/internal/ascii"
	"cats/internal/imgsrc"
	"cats/internal/tui"
)

// version is stamped at build time via -ldflags "-X main.version=...".
var version = "dev"

// scanImages returns every supported image file under dir as a slash-separated
// path relative to dir.
func scanImages(dir string) []string {
	var list []string
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if imgsrc.SupportedExts[strings.ToLower(filepath.Ext(d.Name()))] {
			if rel, rerr := filepath.Rel(dir, p); rerr == nil {
				list = append(list, filepath.ToSlash(rel))
			}
		}
		return nil
	})
	return list
}

// resolveImageDir picks the image directory: the --dir override, then
// IMAGES_DIR, then ./images, falling back to the working directory.
func resolveImageDir(override string) string {
	dir := override
	if dir == "" {
		dir = os.Getenv("IMAGES_DIR")
	}
	if dir == "" {
		dir = "images"
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		dir = "."
	}
	return dir
}

// resolveImage loads ref, or a random image from images when ref is empty. The
// reference may be a local path, a ~ path, or an http(s) URL.
func resolveImage(dir string, images []string, ref string) (image.Image, error) {
	if ref == "" {
		if len(images) == 0 {
			return nil, fmt.Errorf("no images found in %q", dir)
		}
		ref = filepath.Join(dir, filepath.FromSlash(images[rand.Intn(len(images))]))
	}
	img, _, err := imgsrc.LoadImage(ref)
	return img, err
}

func main() {
	cliMode := flag.Bool("cli", false, "Run in one-shot CLI mode")
	discordMode := flag.Bool("discord", false, "Output a Discord-optimized markdown codeblock snippet")
	cliWidth := flag.Int("width", 80, "Width of the ASCII output in CLI mode")
	cliColor := flag.Bool("color", true, "Enable ANSI color in CLI mode")
	cliInvert := flag.Bool("invert", false, "Invert luminance in CLI mode")
	cliFile := flag.String("file", "", "Image to render: a local path, ~ path, or http(s) URL")
	cliDir := flag.String("dir", "", "Image directory to browse (overrides IMAGES_DIR)")
	showVersion := flag.Bool("version", false, "Print the version and exit")
	listThemes := flag.Bool("list-themes", false, "List available color themes and exit")
	listRamps := flag.Bool("list-ramps", false, "List available character ramps and exit")
	flag.Parse()

	switch {
	case *showVersion:
		fmt.Printf("catgen %s\n", version)
		return
	case *listThemes:
		for _, t := range ascii.Themes {
			fmt.Printf("%-12s %s\n", t.Name, t.Label)
		}
		return
	case *listRamps:
		for _, r := range ascii.Ramps {
			fmt.Printf("%-12s %s\n", r.Name, r.Label)
		}
		return
	}

	imageDir := resolveImageDir(*cliDir)
	images := scanImages(imageDir)

	switch {
	case *discordMode:
		runDiscordCLI(imageDir, images, *cliColor, *cliFile)
	case *cliMode:
		runCLI(imageDir, images, *cliWidth, *cliColor, *cliInvert, *cliFile)
	default:
		m := tui.NewModel(imageDir, images)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal("Error running TUI:", err)
		}
	}
}

func runCLI(dir string, images []string, width int, colorize, invert bool, file string) {
	img, err := resolveImage(dir, images, file)
	if err != nil {
		log.Fatal(err)
	}
	theme := ascii.ThemeTrueColor
	if !colorize {
		theme = ascii.ThemeGrayscale
	}
	fmt.Print(ascii.Convert(img, ascii.Options{
		Width:       width,
		Theme:       theme,
		Invert:      invert,
		DensityRamp: ascii.RampBlocks,
	}))
}

func runDiscordCLI(dir string, images []string, colorize bool, file string) {
	img, err := resolveImage(dir, images, file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ascii.ConvertToDiscord(img, colorize, ascii.RampStandard))
}
