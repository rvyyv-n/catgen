# 🐾 ASCII Cat Generator: Project Roadmap

This document outlines the phased development plan for transforming the `cats` repository into an interactive ASCII Cat Generator and Web App.

## Phase 1: The Core Terminal Engine (CLI & TUI)
**Goal**: Build a robust Go-based image-to-ASCII converter and an interactive terminal user interface.

- [ ] **Core ASCII Converter**
  - [ ] Implement image resizing and grayscale mapping in Go.
  - [ ] Add multiple character density ramps (Standard, Block, Braille).
  - [ ] Add 24-bit ANSI color support mapping.
- [ ] **Interactive TUI (using Bubble Tea)**
  - [ ] Implement an interactive browser (`Space`/`Enter` for the next cat).
  - [ ] Add keyboard controls for resolution (`+`/`-`), color toggles (`c`), and inversion (`i`).
  - [ ] Add export functionality (save to `.txt` or copy to clipboard).
- [ ] **CLI Mode**
  - [ ] Add CLI flags for headless execution (e.g., `cats --cli --color=true --width=80`).

## Phase 2: Server Integrations (`curl` Mode)
**Goal**: Make the cats easily accessible via HTTP requests directly in the terminal.

- [ ] **`curl`-Friendly Endpoint**
  - [ ] Add User-Agent detection in `main.go`.
  - [ ] Serve raw ANSI-colored strings when requested via terminal clients (like `curl localhost:8090/ascii`).
- [ ] **Extended JSON API**
  - [ ] Expand the existing `/api/cat` to optionally return pre-generated ASCII string representations alongside URLs.

## Phase 3: The Retro Web App
**Goal**: Deploy a modern web application (e.g., Next.js, Astro, or Go WASM) with a retro CRT aesthetic.

- [ ] **Web UI Setup**
  - [ ] Initialize frontend framework and setup hosting (Vercel, Cloudflare, etc.).
  - [ ] Design a retro CRT terminal aesthetic (CSS shaders, green phosphor scanlines).
- [ ] **Interactive Web Tools**
  - [ ] Implement client-side/server-side live ASCII conversion.
  - [ ] Add drag-and-drop support for users to convert their own cat images.
  - [ ] Implement real-time UI sliders for density, contrast, and resolution.
- [ ] **Export Options**
  - [ ] One-click export to `.txt`, `.svg`, or `.png`.
  - [ ] Copy-to-clipboard button formatted specifically for Discord codeblocks.

## Phase 4: Extra Fun Mechanics (Stretch Goals)
**Goal**: Add delightful, highly shareable mechanics to the project.

- [ ] **Cat MOTD (`cats motd`)**
  - [ ] Combine ASCII cats with `fortune`-style daily cat facts or Unix wisdom.
- [ ] **Animated Cats**
  - [ ] Support converting and playing animated `.gif` cat files in the terminal.
- [ ] **Nyan Cat Live Stream**
  - [ ] Create an infinite scrolling ASCII rainbow cat mode (`cats stream`).
