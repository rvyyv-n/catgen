# CATGEN: Project Roadmap

This document outlines the phased development plan for transforming the `cats` repository into **CATGEN**, an interactive ASCII Cat Studio and Web App.

---

## Phase 1: The Core Terminal Engine (CLI & TUI) — v1.0.0 [COMPLETED]
**Goal**: Build a robust Go-based image-to-ASCII converter and an interactive BANGEN-style terminal user interface.

- [x] **Core ASCII Converter**
  - [x] Implement image decoding, aspect-ratio correction, and luminance mapping in Go.
  - [x] Add multiple character density ramps (Blocks, Standard, Braille, Detailed, Binary, Minimal).
  - [x] Add 24-bit TrueColor ANSI output and 6 color themes (Matrix, Cyberpunk, Amber, Ice Blue, Grayscale).
  - [x] Implement Auto-Fit dynamic constraint solver to prevent screen overflow.
  - [x] Live brightness, contrast, and inversion controls.
  - [x] Discord-optimized 34-column codeblock formatter with 16-color ANSI support.
- [x] **Interactive TUI Studio (Bubble Tea & Lipgloss)**
  - [x] Build BANGEN-style 2-pane UI with embedded centered border titles and thin framing.
  - [x] Add glowing dual-tone block ASCII CATGEN logo banner to sidebar.
  - [x] Left pane property controls with full-width solid teal selection highlights.
  - [x] Radio-button effect lists with smooth scrolling viewport.
  - [x] Live preview right pane with horizontal & vertical image centering.
  - [x] Action shortcuts: `r` for instant random cat, `e` for plain text export, `d` for Discord export.
  - [x] Right-aligned color-coded keybind footer.
- [x] **CLI Mode**
  - [x] Add `--cli`, `--discord`, `--width`, `--color`, `--invert`, and `--file` flags for headless terminal generation.

---

## Phase 1.5: Arbitrary Image Input [COMPLETED]
**Goal**: Convert any image, not only the bundled cat assets, from every entry point.

- [x] **Shared loader (`imgsrc` package)**
  - [x] `LoadImage` resolves a local path, `~` path, or `http(s)` URL into a decoded image.
  - [x] Register `gif` and `webp` decoders alongside `png`/`jpeg`; drop the unsupported `svg` claim.
  - [x] Path cleaning: trim whitespace and surrounding quotes (terminal drag-and-drop), expand `~`.
  - [x] Remote fetch with a 15s timeout and a 25 MiB size cap.
  - [x] Downscale sources larger than 4000 px on the long edge before sampling.
- [x] **TUI integration**
  - [x] `o` opens an image by path or URL via a Bubbles text-input overlay.
  - [x] External images join the carousel — cycle, random, and export like bundled cats.
  - [x] Decoded-image cache so slider tweaks (and remote URLs) do not reload every keystroke.
  - [x] Load errors surface in the status line without disturbing the current selection.
- [x] **CLI**
  - [x] `--file` accepts local paths, `~` paths, and URLs.
  - [x] `--dir` overrides the image directory (previously only `IMAGES_DIR`).

---

## Phase 1.6: TUI Studio Polish [COMPLETED]
**Goal**: Bring the terminal studio closer to feature parity with BANGEN's editor.

- [x] **Look presets** — JSON presets under `~/.catgen/presets/`, `s` to save the
  current look, `p` to pick one from a list; ships `matrix`, `cyberpunk`,
  `amber-crt`, and `noir` built-ins. Theme and ramp are stored by name so presets
  survive catalog reordering.
- [x] **Export modal** — `e` opens a modal with a plain-text / Discord-snippet
  toggle, an editable output path, and an overwrite confirmation, replacing the
  silent fixed-filename dump. `d` still does an instant Discord export.
- [x] **Fit-info toggle** — `a` shows the resolved output size
  (`fit:<mode> · src WxH · grid WxH`) in the footer, surfacing what the Auto-Fit
  solver actually produced.
- [x] **Active-choice glyph** — the selected theme/ramp shows a green check
  instead of a filled bullet, matching BANGEN's radio styling.
- [x] **Shared catalog & CLI discovery** — theme and ramp catalogs moved into the
  `ascii` package; `--list-themes` and `--list-ramps` print them and exit.
- [x] **Clipboard paste** — the `o` open-image overlay accepts pasted paths/URLs.

---

## Phase 1.7: Image Export & Codebase Cleanup [UP NEXT]
**Goal**: Produce a universal, shareable image output for colour ASCII art that a
plain `.txt` file cannot carry, and pay down structural redundancy before the
server work in Phase 2.

### Image export

- [ ] **Grid intermediate representation**
  - [ ] Split the converter: `ConvertGrid(img, opts) [][]Cell` where each `Cell`
    is `{Ch rune, R, G, B uint8}`. `Convert` becomes "grid → ANSI string"; its
    output is unchanged.
  - [ ] Re-express `ConvertToDiscord` as "grid → 16-colour ANSI" on top of
    `ConvertGrid`, deleting its parallel sampling loop, luminance math, and
    palette matcher (~100 lines collapse).
- [ ] **PNG renderer (`RenderPNG`)**
  - [ ] Draw each cell's glyph in its colour on a dark background using a bundled
    monospace font embedded with `//go:embed`.
  - [ ] Use **DejaVu Sans Mono** (permissive licence) — it covers the block
    `░▒▓█` and braille `⠿` ramp glyphs; `golang.org/x/image/font/opentype` is
    already in the dependency tree, so no new module.
  - [ ] Image size = `cols × cellW` by `rows × cellH`; expose a cell size / scale
    knob.
- [ ] **Export modal: two options only**
  - [ ] Replace the plain-text / Discord toggle with **Plain text (`.txt`)** and
    **Image (`.png`)**; auto-swap the path extension with the format.
  - [ ] Decide whether the standalone `d` Discord quick-export stays or is
    removed (leaning: keep it as a separate shortcut).

### Codebase cleanup

- [ ] **Extract `internal/server/`** — move `CatServer`, its methods, every HTTP
  handler, and `safeImgDir` out of `main.go`. `main.go` shrinks to a ~70-line
  mode dispatcher (flags → CLI / TUI / server). Done now so Phase 2 grows the
  server package, not `main.go`.
- [ ] **Unify the extension allowlists** — `internal/server` imports
  `imgsrc.SupportedExts` and adds `.svg` locally (the file server ships SVG bytes
  without decoding); delete `main.go`'s separate `allowedExts` map.
- [ ] **Hoist `fontAspect = 0.46`** to a single package-level `const` in `ascii`
  (currently redeclared in `resolveDims` and `ConvertToDiscord`).
- [ ] *(optional)* Split `internal/tui/tui.go` (~1200 lines) along its natural
  seams: `tui_overlays.go` (overlay handlers + modal rendering) and
  `tui_view.go` (`View` + `buildFramedBox` + helpers).

---

## Phase 2: Server Integrations (`curl` Mode & API)
**Goal**: Connect the ASCII engine to the HTTP server for instant terminal streaming.

- [ ] **`curl`-Friendly Endpoint**
  - [ ] Detect `curl` / terminal User-Agents on `GET /cat` and stream colored ANSI ASCII directly to stdout.
  - [ ] Add query parameters for terminal streaming (e.g. `curl localhost:8090/cat?theme=matrix&width=60`).
- [ ] **Extended JSON API**
  - [ ] Add `ascii` string field to `/api/cat` and `/api/cats` responses.

---

## Phase 3: The Retro Web App
**Goal**: Deploy a modern web application (e.g., Next.js, Astro, or Go WASM) with a retro CRT aesthetic.

- [ ] **Web UI Setup**
  - [ ] Initialize frontend framework and setup hosting (Vercel, Cloudflare Pages, etc.).
  - [ ] Design a retro CRT terminal aesthetic (CSS shaders, green phosphor scanlines).
- [ ] **Interactive Web Tools**
  - [ ] Implement client-side live ASCII conversion (via Canvas / Web Worker / WASM).
  - [ ] Add drag-and-drop support for users to convert their own cat images.
  - [ ] Real-time UI sliders for density, contrast, brightness, and resolution.
- [ ] **Export Options**
  - [ ] One-click export to `.txt`, `.svg`, or `.png`.
  - [ ] Copy-to-clipboard button formatted specifically for Discord codeblocks.

---

## Phase 4: Extra Fun Mechanics (Stretch Goals)
**Goal**: Add delightful, highly shareable mechanics to the project.

- [ ] **Cat MOTD (`cats motd`)**
  - [ ] Combine ASCII cats with `fortune`-style daily cat facts or Unix wisdom.
- [ ] **Animated Cats**
  - [ ] Support converting and playing animated `.gif` cat files in the terminal.
- [ ] **Nyan Cat Live Stream**
  - [ ] Create an infinite scrolling ASCII rainbow cat mode (`cats stream`).
