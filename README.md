# CATGEN 🐱

```text
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

**CATGEN** is a terminal-first ASCII cat generator, interactive TUI studio, and HTTP cat server built with Go.

---

## ✨ Features

* **Interactive TUI Studio**: Full-screen 2-pane terminal interface powered by Bubble Tea and Lipgloss.
* **Auto-Fit Engine**: Dynamic boundary math scales any image to fit your terminal window perfectly with zero overflow or cutoffs.
* **6 Color Palettes**: TrueColor (24-bit RGB), Grayscale, Matrix Glow, Cyberpunk Neon, Amber Phosphor, and Ice Blue.
* **6 Character Ramps**: Blocks (`░▒▓█`), Standard (`.:-=+*#%@`), Braille (`⠋⠙⠹`), Detailed ASCII, Binary Matrix (`01`), and Minimal.
* **Image Tuning**: Live brightness, contrast, inversion, and custom resolution adjustments.
* **Export**: One-key export of clean ASCII art to `.txt`.
* **One-Shot CLI**: Print random or specific ASCII cats directly to stdout.
* **Web & API Server**: HTTP image & JSON metadata endpoints.

---

## 🎮 Interactive TUI

Launch the full TUI studio:

```bash
cats
# or
go run .
```

### Keybindings

| Key | Action |
| :--- | :--- |
| `↑` / `↓` (`k` / `j`) | Navigate menu items |
| `←` / `→` (`h` / `l`) | Adjust values and properties |
| `Enter` / `Space` | Select / toggle effect |
| `r` | Load random cat |
| `e` | Export clean ASCII to `cat_ascii.txt` |
| `q` / `Ctrl+C` | Quit |

---

## 💻 CLI Mode

Generate ASCII directly to your terminal prompt:

```bash
# Print a random cat in full color
cats --cli

# Specify width, disable color, and invert
cats --cli --width=60 --color=false --invert

# Render a specific cat image
cats --cli --file="images/blep.png"
```

---

## 🌐 Web Server & API

Run the background HTTP server:

```bash
cats --server
```

### Endpoints

| Route | Method | Description |
| :--- | :--- | :--- |
| `/` | `GET` | Redirects (`302`) to a random cat image URL |
| `/cat` | `GET` | Directly streams a random cat image |
| `/cats/<filename>` | `GET` | Serves a specific cat image file from `images/` |
| `/api/cat` | `GET` | Returns JSON metadata for a random cat image |
| `/api/cats` | `GET` | Returns a JSON list of all available cat images |
| `/favicon.ico` | `GET` | Serves [`favicon.svg`](./favicon.svg) |
| `/healthz` | `GET` | Health check endpoint |

---

## 🐳 Docker

```bash
# Build
docker build -t catgen .

# Run Web Server on port 8090
docker run -p 8090:8090 catgen
```

---

## ⚙️ Environment Variables

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8090` | HTTP server port (when running in `--server` mode) |
| `IMAGES_DIR` | `images` | Directory containing source cat images |
