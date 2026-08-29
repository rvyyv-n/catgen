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

Terminal-first ASCII cat generator, interactive TUI studio, and HTTP server written in Go.

---

## Features

- **Interactive TUI Studio**: 2-pane terminal interface built with Bubble Tea and Lipgloss.
- **Dynamic Auto-Fit**: Automatic scaling to fit any terminal size without overflow or clipping.
- **Discord-Ready Export**: Formats ASCII into 34-column codeblocks with 16-color ANSI support.
- **Color Palettes**: TrueColor (RGB), Grayscale, Matrix Glow, Cyberpunk, Amber Phosphor, Ice Blue.
- **Character Ramps**: Blocks (`░▒▓█`), Standard (`.:-=+*#%@`), Braille (`⠋⠙⠹`), Detailed, Binary (`01`), Minimal.
- **Image Tuning**: Live brightness, contrast, inversion, and custom width scaling.
- **Export Options**: Plain-text (`.txt`) and Discord markdown snippets.
- **Headless CLI**: Print ASCII directly to terminal stdout.
- **HTTP Server**: Image delivery and JSON metadata API.

---

## TUI Studio

Launch the interactive terminal studio:

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
| `r` | Load random cat |
| `e` | Export plain ASCII to `cat_ascii.txt` |
| `d` | Export Discord codeblock to `cat_discord.txt` |
| `q` / `Ctrl+C` | Quit |

---

## CLI Usage

Generate ASCII directly to stdout:

```bash
# Print a random cat in full color
cats --cli

# Generate Discord-ready markdown codeblock (< 1,500 chars, max width 34)
cats --discord

# Custom width, monochrome, and inverted
cats --cli --width=60 --color=false --invert

# Process a specific image file
cats --cli --file="images/blep.png"
```

---

## HTTP Server & API

Run the HTTP server:

```bash
cats --server
```

### Routes

| Route | Method | Description |
| :--- | :--- | :--- |
| `/` | `GET` | Redirects (302) to random cat image URL |
| `/cat` | `GET` | Directly streams a random cat image |
| `/cats/<filename>` | `GET` | Serves specific cat image file from `images/` |
| `/api/cat` | `GET` | Returns JSON metadata for random cat image |
| `/api/cats` | `GET` | Returns JSON list of all available cat images |
| `/favicon.ico` | `GET` | Serves favicon.svg |
| `/healthz` | `GET` | Health check endpoint |

---

## Docker

```bash
# Build
docker build -t catgen .

# Run
docker run -p 8090:8090 catgen
```

---

## Configuration

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8090` | Server listen port |
| `IMAGES_DIR` | `images` | Path to image asset directory |
