# CATGEN: Project Roadmap

**CATGEN** is a terminal-only ASCII cat studio: a Go image-to-ASCII engine, an
interactive Bubble Tea TUI, and a headless CLI. It is deliberately scoped as a
TUI app in the spirit of BANGEN — there is no web app, no HTTP server, and no
plan to add other surfaces.

The current tree ships as **v1** plus Phases 1.7–1.9 (image export, chrome
themes, layout polish, structural cleanup, exports browser). Phase 2 cuts the
**v2** release with cross-platform executables.

---

## Phase 1: Core Terminal Engine (CLI & TUI) — v1.0.0 [COMPLETED]

- [x] **Core ASCII converter** — image decoding, font-aspect correction, and
  luminance mapping; density ramps (Blocks, Standard, Braille, Detailed, Binary,
  Minimal); 24-bit TrueColor ANSI plus 6 color themes; Auto-Fit constraint
  solver; live brightness / contrast / inversion.
- [x] **Interactive TUI studio (Bubble Tea & Lipgloss)** — BANGEN-style 2-pane
  UI with embedded centered border titles, glowing CATGEN logo banner, full-width
  teal selection highlights, radio-button effect lists, live centered preview,
  color-coded keybind footer.
- [x] **Discord export** — 34-column codeblock formatter with 16-color ANSI.
- [x] **CLI mode** — `--cli`, `--discord`, `--width`, `--color`, `--invert`,
  `--file`.

## Phase 1.5: Arbitrary Image Input [COMPLETED]

- [x] **`imgsrc` loader** — `LoadImage` resolves a local path, `~` path, or
  `http(s)` URL; `png` / `jpeg` / `gif` / `webp` decoders; quote/whitespace
  trimming and `~` expansion; 15s / 25 MiB remote fetch caps; downscale sources
  over 4000 px on the long edge.
- [x] **TUI integration** — `o` opens any image by path or URL; external images
  join the carousel; decoded-image cache; load errors surface in the status line.
- [x] **CLI** — `--file` accepts paths / `~` paths / URLs; `--dir` overrides the
  image directory.

## Phase 1.6: TUI Studio Polish [COMPLETED]

- [x] **Look presets** — JSON presets under `~/.catgen/presets/`, `s` to save,
  `p` to pick; ships `matrix`, `cyberpunk`, `amber-crt`, `noir`. Theme and ramp
  stored by name so presets survive catalog reordering.
- [x] **Export modal** — `e` opens a modal with a format toggle, editable output
  path, and overwrite confirmation.
- [x] **Fit-info toggle** — `a` shows the resolved output size in the footer.
- [x] **Active-choice glyph** — selected theme/ramp shows a green check.
- [x] **Shared catalog & CLI discovery** — theme and ramp catalogs live in the
  `ascii` package; `--list-themes` / `--list-ramps` print them and exit.
- [x] **Clipboard paste** — the `o` overlay accepts pasted paths/URLs.

---

## Phase 1.7: Image Export [COMPLETED]

**Goal**: produce a shareable colour image of the ASCII art that a `.txt` file
cannot carry.

- [x] **Grid intermediate representation** — `ConvertGrid(img, opts) [][]Cell`
  where each `Cell` is `{Ch rune, R, G, B uint8}`. `Convert` is now
  `renderANSI(ConvertGrid(...))`; a golden test in `converter_test.go` plus an
  offline 7,824-case sweep confirm byte-identical output.
- [x] **Collapse `ConvertToDiscord`** onto `ConvertGrid` ("grid → 16-colour
  ANSI"), deleting its parallel sampling loop, luminance math, and dimension
  solver; the palette matcher is now the package-level `closestDiscordANSI`.
- [x] **`RenderPNG(grid, scale)`** — draws each cell's glyph in its colour on a
  near-black background. **DejaVu Sans Mono** is bundled via `//go:embed`
  (`internal/ascii/fonts/`, covers the `░▒▓█` block and `⠿` braille glyphs);
  rendered with `golang.org/x/image/font/opentype` + `font.Drawer`. Cell metrics
  come from the face; `scale` multiplies the font size and output resolution.
- [x] **Export modal → two options** — the format toggle is now **Plain text
  (`.txt`)** / **Image (`.png`)**; toggling swaps the path extension while
  keeping the user's stem (`swapExportExt`).

**Cleanup carried in this pass** (entangled with the converter refactor):

- [x] Hoist `fontAspect = 0.46` to one package-level `const` in `ascii`.
- [x] Drop the standalone `d` Discord quick-export shortcut, its `exportToDiscord`
  handler, and its footer entry (`--discord` flag stays).
- [x] Remove `--server` HTTP mode, `CatServer`, all handlers, `safeImgDir`, and
  the Dockerfile — out of scope for a TUI-only tool. `main.go` is now a ~135-line
  dispatcher. The former `allowedExts` map is replaced by `imgsrc.SupportedExts`.

## Phase 1.8: TUI Chrome Themes, Layout Polish & Cleanup [COMPLETED]

**Goal**: let the studio chrome match the art, tighten the layout, and pay down
structural redundancy before cutting v2.

- [x] **TUI chrome schemes** — five built-in palettes (`teal`, `amber`,
  `magenta`, `green`, `mono`) setting border / accent / section / text colours;
  `c` cycles them, the choice persists to `~/.catgen/config.json`
  (`internal/config`) and is restored on next launch. Separate from the ASCII
  art themes; no custom palette editing.
- [x] **Split `internal/tui/tui.go`** into `tui.go` (model, `Update`, adjust
  handlers), `tui_overlays.go` (overlay input handlers + overlay actions +
  modal bodies), `tui_view.go` (`View`, `buildFramedBox`, framing helpers).
- [x] **Full-size layout** — the controls and Live Preview frames stay
  edge-to-edge for the whole terminal, BANGEN-style; only the art inside the
  preview resizes. (An earlier "hug the content" experiment was reverted.)
- [x] **Footer never wraps** — the keybind bar is width-capped and sheds its
  least-essential hints (`nav`, `adjust`, `toggle`, `info`, …) on narrow
  terminals instead of spilling onto a second line.
- [x] **Width control fixed** — adjusting `Width` now switches to a `Custom`
  fit mode that actually renders at `CustomW` (it previously fell through to the
  hard-coded Compact width and did nothing).
- [x] **"Auto Fit" → "Auto"** label in the Fit Mode control (all modes intact).
- [x] **`exports/` folder** — TUI exports default to `exports/cat_ascii.{png,txt}`,
  created on write and git-ignored, so user output is not mixed into the tree.
- [x] **Repo tidy** — `main` package and its Windows icon assets moved to
  `cmd/catgen/`; repo root is now `go.{mod,sum}`, `README.md`, `ROADMAP.md`,
  `images/`, `cmd/`, `internal/`.
- [x] **General trim** — dropped the unused `headerStyle`; collapsed the colour
  block around the chrome system.

---

## Phase 1.9: Exports Tab [COMPLETED]

**Goal**: let the studio manage what it has produced.

- [x] `x` opens an exports browser (`overlayExports`) listing `exports/`
  newest-first with each file's size and age: `↑↓` move, `Enter` opens the file
  in the OS viewer (`start` / `open` / `xdg-open`), `c` copies its absolute path
  to the clipboard, `d` / `Backspace` deletes it after a one-key confirm. The
  `atotto/clipboard` dependency is promoted from indirect to direct.

---

## Phase 2: v2 Release

**Goal**: ship CATGEN v2 as ready-to-run executables.

- [ ] **Curate the bundled image set** — the user is auditing `images/` by hand,
  dropping frames that convert poorly to ASCII (text / logo overlays, busy
  backgrounds, low-contrast or tiny subjects).
- [x] **`scripts/release.ps1`** — cross-compiles the `windows/amd64`,
  `darwin/{amd64,arm64}` and `linux/{amd64,arm64}` matrix with
  `-trimpath -ldflags "-s -w -X main.version=..."` and `CGO_ENABLED=0`, then
  assembles one clean folder + zip per target under `dist/`: `catgen(.exe)` +
  `images/` + an empty `exports/` + `README.md`, no source.
- [ ] Run `scripts/release.ps1 -Version v2.0.0` and cut a GitHub Release with
  the five zips attached and notes covering image export, chrome themes, and
  the exports browser.
- [ ] Final `README.md` pass: screenshots / cast of the studio, install steps per
  platform.
- [ ] Tag **v2.0.0**.

*This is the end of the roadmap. CATGEN is a finished TUI tool at v2 — no web
app, no server, no further phases planned.*
