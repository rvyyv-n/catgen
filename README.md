```text
 █▀▀ ▄▀█ ▀█▀ █▀▀ █▀▀ █▄ █
 █▄▄ █▀█  █  █▄█ ██▄ █ ▀█
  ── ASCII Art Studio ──

        _                        
       \`*-.                    
        )  _`-.                 
       .  : `. .                
       : _   '  \               
       ; *` _.   `*-._          
       `-.-'          `-.       
         ;       `       `.     
         :.       .        \    
         . \  .   :   .-'   .   
         '  `+.;  ;  '      :   
         :  '  |    ;       ;-. 
         ; '   : :`-:     _.`* ;
[bug] .*' /  .*' ; .*`- +'  `*' 
      `*-*   `*-*  `*-*'
```

**CATGEN** turns any image into ASCII art in your terminal. It is a Go
image-to-ASCII engine wrapped in an interactive TUI studio and a headless CLI —
no web app, no server, just the terminal. It ships with a pool of sample images
(cats, naturally) so it does something the moment you run it.

---

## Install

Download the archive for your platform from the
[latest release](https://github.com/rvyyv-n/catgen/releases/latest), unzip it,
and run `catgen` from a terminal:

| Platform | Archive |
| :--- | :--- |
| Windows (x64) | `catgen-<ver>-windows-x64.zip` |
| Windows (ARM64) | `catgen-<ver>-windows-arm64.zip` |
| Linux (x64) | `catgen-<ver>-linux-x64.zip` |
| Linux (ARM64) | `catgen-<ver>-linux-arm64.zip` |
| macOS (Apple Silicon) | `catgen-<ver>-macos-arm64.zip` |

Each archive contains the `catgen` binary, the bundled `images/`, and an empty
`exports/` folder for your output. Nothing is installed system-wide; delete the
folder to uninstall.

Or build from source (Go 1.25+):

```bash
go install github.com/rvyyv-n/catgen/cmd/catgen@latest
```

---

## TUI Studio

Run `catgen` with no arguments:

```bash
catgen
```

<!-- screenshot: hero -->

Left pane is the controls, right pane is a live preview that re-renders on every
change.

### Controls

| Key | Action |
| :--- | :--- |
| `↑` / `↓` (`k` / `j`) | Navigate options |
| `←` / `→` (`h` / `l`) | Adjust values and properties |
| `Enter` / `Space` | Select / toggle |
| `o` | Open any image by path or URL |
| `r` | Load a random bundled image |
| `s` / `p` | Save the current look as a preset / pick a preset |
| `e` | Export — plain text (`.txt`) or colour image (`.png`) to `exports/` |
| `x` | Browse `exports/` — `Enter` open · `c` copy path · `d` delete |
| `t` | Cycle the UI colour theme (persisted) |
| `a` | Toggle fit info in the footer |
| `q` / `Ctrl+C` | Quit |

### Rendering options

- **Fit Mode / Width** — `Auto` fits the preview pane; `Compact` / `Wide` / `Max`
  are quick presets; adjusting `Width` switches to `Custom` and renders at that
  column count.
- **Palette** — TrueColor (RGB), Grayscale, Matrix Glow, Cyberpunk,
  Amber Phosphor, Ice Blue.
- **Character Ramp** — Blocks (`░▒▓█`), Standard (`.:-=+*#%@`), Braille,
  Detailed ASCII, Binary (`01`), Minimal.
- **Tuning** — live brightness, contrast, and inversion.

### Themes

Seven UI colour schemes — `teal`, `amber`, `magenta`, `green`, `ocean`,
`violet`, `mono` — cycled with `t` and saved to `~/.catgen/config.json`. These
are the studio chrome, separate from the art palettes.

<!-- screenshot: theme -->

### Image export

`e` opens the export modal. **Plain text** writes the ANSI-free characters to a
`.txt`. **Image (PNG)** rasterises the art — every glyph drawn in its colour on a
dark background, using an embedded DejaVu Sans Mono so the block and braille
ramps render correctly — to a `.png` you can share anywhere a text file can't go.

<!-- screenshot: png-export -->

---

## CLI

```bash
# Print a random sample image in full colour
catgen --cli

# A specific image — local path, ~ path, or URL
catgen --cli --file="images/blep.png"
catgen --cli --file="~/Pictures/skyline.jpg"
catgen --cli --file="https://example.com/photo.png"

# Width, colour, inversion
catgen --cli --file="photo.png" --width=100 --color=false --invert

# Discord-ready markdown codeblock (<1,500 chars, 34 cols, 16-colour ANSI)
catgen --discord --file="photo.png"

# Point the random pool at another folder
catgen --dir="~/Pictures" --cli

# Discovery
catgen --list-themes
catgen --list-ramps
catgen --version
```

<!-- screenshot: cli (optional) -->

---

## Configuration

| Path / Var | Purpose |
| :--- | :--- |
| `~/.catgen/config.json` | Persisted UI theme |
| `~/.catgen/presets/` | Saved look presets (`s` to save, `p` to load) |
| `IMAGES_DIR` env var | Overrides the `images/` pool location (also `--dir`) |
| `./exports/` | Where the TUI writes `.txt` / `.png` exports |

---

## Building from source

```bash
# Local build
go build -trimpath -ldflags "-s -w" -o catgen ./cmd/catgen

# Versioned build
go build -trimpath -ldflags "-s -w -X main.version=v2.0.0" -o catgen ./cmd/catgen

# Regenerate the Windows icon after changing cmd/catgen/favicon.ico
go install github.com/akavel/rsrc@latest
go generate ./...

# All release archives (Windows x64/ARM64, Linux x64/ARM64, macOS ARM64)
pwsh scripts/release.ps1 -Version v2.0.0
```
