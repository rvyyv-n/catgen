// Regenerate the Windows executable icon from favicon.ico after changing it:
//
//	go install github.com/akavel/rsrc@latest
//	go generate ./...
//
//go:generate rsrc -ico favicon.ico -arch amd64 -o rsrc_windows_amd64.syso
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"cats/internal/ascii"
	"cats/internal/imgsrc"
	"cats/internal/tui"
)

// version is stamped at build time via -ldflags "-X main.version=...".
var version = "dev"

var allowedExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
	".svg":  true,
}

type CatServer struct {
	imageDir string
	mu       sync.RWMutex
	images   []string
	rng      *rand.Rand
}

func NewCatServer(dir string) *CatServer {
	s := &CatServer{
		imageDir: dir,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	s.reloadImages()
	return s
}

func (s *CatServer) reloadImages() {
	s.mu.Lock()
	defer s.mu.Unlock()

	var list []string
	_ = filepath.WalkDir(s.imageDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if allowedExts[ext] {
			rel, err := filepath.Rel(s.imageDir, p)
			if err == nil {
				list = append(list, filepath.ToSlash(rel))
			}
		}
		return nil
	})

	s.images = list
}

func (s *CatServer) getRandomCat() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.images) == 0 {
		return "", errors.New("no cat images found")
	}
	idx := s.rng.Intn(len(s.images))
	return s.images[idx], nil
}

func (s *CatServer) getAllCats() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]string, len(s.images))
	copy(out, s.images)
	return out
}

func main() {
	serverMode := flag.Bool("server", false, "Run the HTTP web server")
	cliMode := flag.Bool("cli", false, "Run in one-shot CLI mode")
	discordMode := flag.Bool("discord", false, "Output a Discord-optimized markdown codeblock snippet")
	cliWidth := flag.Int("width", 80, "Width of the ASCII output in CLI mode")
	cliColor := flag.Bool("color", true, "Enable ANSI color in CLI mode")
	cliInvert := flag.Bool("invert", false, "Invert luminance in CLI mode")
	cliFile := flag.String("file", "", "Image to render: a local path, ~ path, or http(s) URL")
	cliDir := flag.String("dir", "", "Image directory to browse (overrides IMAGES_DIR)")
	showVersion := flag.Bool("version", false, "Print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("catgen %s\n", version)
		return
	}

	imageDir := *cliDir
	if imageDir == "" {
		imageDir = os.Getenv("IMAGES_DIR")
	}
	if imageDir == "" {
		imageDir = "images"
	}
	if _, err := os.Stat(imageDir); os.IsNotExist(err) {
		imageDir = "."
	}

	server := NewCatServer(imageDir)

	if *discordMode {
		runDiscordCLI(server, *cliColor, *cliFile)
		return
	}

	if *cliMode {
		runCLI(server, *cliWidth, *cliColor, *cliInvert, *cliFile)
		return
	}

	if !*serverMode {
		// Default to interactive TUI
		m := tui.NewModel(server.imageDir, server.getAllCats())
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal("Error running TUI:", err)
		}
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	mux := http.NewServeMux()

	// Redirect root to random cat image URL
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		cat, err := server.getRandomCat()
		if err != nil {
			http.Error(w, "No cats available :(", http.StatusNotFound)
			return
		}
		http.Redirect(w, req, "/cats/"+cat, http.StatusFound)
	})

	// Direct stream of a random cat image
	mux.HandleFunc("/cat", func(w http.ResponseWriter, req *http.Request) {
		cat, err := server.getRandomCat()
		if err != nil {
			http.Error(w, "No cats available :(", http.StatusNotFound)
			return
		}
		fullPath := filepath.Join(server.imageDir, filepath.FromSlash(cat))
		http.ServeFile(w, req, fullPath)
	})

	// Static cat file server
	fileServer := http.FileServer(safeImgDir(server.imageDir))
	mux.Handle("/cats/", http.StripPrefix("/cats/", fileServer))

	// JSON API: Get single random cat metadata
	mux.HandleFunc("/api/cat", func(w http.ResponseWriter, req *http.Request) {
		cat, err := server.getRandomCat()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "No cats available"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"name": filepath.Base(cat),
			"url":  "/cats/" + cat,
		})
	})

	// JSON API: List all cat images
	mux.HandleFunc("/api/cats", func(w http.ResponseWriter, req *http.Request) {
		cats := server.getAllCats()
		urls := make([]string, len(cats))
		for i, c := range cats {
			urls[i] = "/cats/" + c
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count": len(urls),
			"cats":  urls,
		})
	})

	// Favicon
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, req *http.Request) {
		if _, err := os.Stat("favicon.svg"); err == nil {
			http.ServeFile(w, req, "favicon.svg")
			return
		}
		http.NotFound(w, req)
	})

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "OK")
	})

	log.Printf("Cats server running on http://localhost%s (serving from '%s')", port, imageDir)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal(err)
	}
}

type safeImgDir string

func (d safeImgDir) Open(name string) (http.File, error) {
	ext := strings.ToLower(filepath.Ext(name))
	if !allowedExts[ext] {
		return nil, errors.New("file type not allowed")
	}
	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
		return nil, errors.New("invalid character in file path")
	}
	dir := string(d)
	fullName := filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name)))
	f, err := os.Open(fullName)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// resolveImage loads the requested image reference, or a random cat from the
// image directory when no reference is given. The reference may be a local path,
// a ~ path, or an http(s) URL.
func resolveImage(server *CatServer, ref string) (image.Image, error) {
	if ref == "" {
		cat, err := server.getRandomCat()
		if err != nil {
			return nil, err
		}
		ref = filepath.Join(server.imageDir, filepath.FromSlash(cat))
	}
	img, _, err := imgsrc.LoadImage(ref)
	return img, err
}

func runCLI(server *CatServer, width int, colorize bool, invert bool, file string) {
	img, err := resolveImage(server, file)
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

func runDiscordCLI(server *CatServer, colorize bool, file string) {
	img, err := resolveImage(server, file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ascii.ConvertToDiscord(img, colorize, ascii.RampStandard))
}
