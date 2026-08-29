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

## Phase 2: Server Integrations (`curl` Mode & API) [UP NEXT]
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
