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
- **Export**: plain-text (`.txt`) and Discord-ready 34-column markdown codeblocks.
- **Headless CLI**: print ASCII straight to stdout.

---

## TUI Studio

```bash
cats
# or
go run .
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
| `a` | Toggle fit-info in the footer |
| `e` | Export via modal (editable path, overwrite confirm) |
| `d` | Instant Discord codeblock export |
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
# Windows executable (embeds favicon.ico as the app icon via rsrc_windows_amd64.syso)
go build -trimpath -ldflags "-s -w" -o catgen.exe .

# Regenerate the icon resource after changing favicon.ico
go install github.com/akavel/rsrc@latest
go generate ./...

# Stamp a version
go build -trimpath -ldflags "-s -w -X main.version=v2.0.0" -o catgen.exe .
```

Linux and macOS builds are cross-compiled on demand
(`GOOS=linux`/`darwin go build ...`) when cutting a release.

---

## Configuration

| Variable | Default | Description |
| :--- | :--- | :--- |
| `IMAGES_DIR` | `images` | Path to the image asset directory (overridden by `--dir`) |
