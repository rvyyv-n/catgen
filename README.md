```text
 █▀▀ ▄▀█ ▀█▀ █▀▀ █▀▀ █▄ █
 █▄▄ █▀█  █  █▄█ ██▄ █ ▀█
  ── ASCII Cat Studio ──

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

Terminal-only ASCII cat generator: an interactive TUI studio and a headless CLI,
written in Go. No web app, no server — just the terminal.

---

## Features

- **Interactive TUI Studio**: 2-pane terminal interface built with Bubble Tea and Lipgloss.
- **Dynamic Auto-Fit**: automatic scaling to fit any terminal size without overflow or clipping.
- **Color Palettes**: TrueColor (RGB), Grayscale, Matrix Glow, Cyberpunk, Amber Phosphor, Ice Blue.
- **Character Ramps**: Blocks (`░▒▓█`), Standard (`.:-=+*#%@`), Braille (`⠋⠙⠹`), Detailed, Binary (`01`), Minimal.
- **Any Image Source**: convert any local file (`png`, `jpg`, `gif`, `webp`) or `http(s)` URL, not just the bundled cats.
- **Image Tuning**: live brightness, contrast, inversion, and custom width scaling.
- **Look Presets**: save and recall named looks from `~/.catgen/presets/`.
- **Chrome Themes**: five built-in UI colour schemes (`teal`, `amber`, `magenta`, `green`, `mono`); `c` cycles, the choice is saved to `~/.catgen/config.json`.
- **Export**: plain-text (`.txt`) or a colour image (`.png`) of the art from the TUI, written to `exports/`; Discord-ready 34-column markdown codeblocks from the CLI.
- **Headless CLI**: print ASCII straight to stdout.

---

## TUI Studio

```bash
cats
# or
go run ./cmd/catgen
```

### Controls

| Key | Action |
| :--- | :--- |
| `↑` / `↓` (`k` / `j`) | Navigate options |
| `←` / `→` (`h` / `l`) | Adjust values and properties |
| `Enter` / `Space` | Select / toggle effect |
| `o` | Open any image by path or URL |
| `r` | Load random cat |
| `s` / `p` | Save look preset / pick a preset |
| `x` | Browse `exports/` — `Enter` open · `c` copy path · `d` delete |
| `c` | Cycle the chrome colour scheme (persisted) |
| `a` | Toggle fit-info in the footer |
| `e` | Export via modal — plain text (`.txt`) or image (`.png`) to `exports/`, editable path, overwrite confirm |
| `q` / `Ctrl+C` | Quit |

---

## CLI Usage

```bash
# Print a random cat in full color
cats --cli

# Discord-ready markdown codeblock (< 1,500 chars, max width 34)
cats --discord

# Custom width, monochrome, inverted
cats --cli --width=60 --color=false --invert

# Any local image, ~ path, or URL
cats --cli --file="images/blep.png"
cats --cli --file="~/Pictures/dog.jpg"
cats --cli --file="https://example.com/photo.png"

# Point the random pool at a different folder
cats --dir="~/Pictures" --cli

# Discovery
cats --list-themes
cats --list-ramps
cats --version
```

---

## Building

```bash
# Windows executable (embeds cmd/catgen/favicon.ico via rsrc_windows_amd64.syso)
go build -trimpath -ldflags "-s -w" -o catgen.exe ./cmd/catgen

# Regenerate the icon resource after changing cmd/catgen/favicon.ico
go install github.com/akavel/rsrc@latest
go generate ./...

# Stamp a version
go build -trimpath -ldflags "-s -w -X main.version=v2.0.0" -o catgen.exe ./cmd/catgen
```

Linux and macOS builds are cross-compiled on demand
(`GOOS=linux`/`darwin go build ... ./cmd/catgen`) when cutting a release.

---

## Configuration

| Variable | Default | Description |
| :--- | :--- | :--- |
| `IMAGES_DIR` | `images` | Path to the image asset directory (overridden by `--dir`) |
