# Plan — EXIF Photos Helper

## Context
You want a fast, easy-to-launch desktop tool for batch-fixing the **date** and **GPS location** of old scanned or undated photos (mostly JPG, some HEIC). The workflow is:
1. Drop a folder onto the app.
2. Browse thumbnails and collect photos into a "basket".
3. Choose a date at whatever precision you know (decade → year → month → day → hour), and click a map for the location.
4. Write the EXIF.

The project folder is empty except for `readme.md`. Step 0 of implementation copies this plan to `/Users/mfh/dev/exif-photos-helper/plan.md`, as you asked.

## Decisions (confirmed)
- **Shell:** native window with **Wails** (Go backend + system WebKit webview). There is no Electron and no Node runtime when the app runs.
- **EXIF engine:** **exiftool**, driven from Go. You will install it through MacPorts (`sudo port install p5-image-exiftool`, which provides `/opt/local/bin/exiftool`).
- **Safety:** files are edited in place with **no backup** (`-overwrite_original`).
- **Date fields:** `DateTimeOriginal`, `CreateDate`, `ModifyDate` and the filesystem modification time.

## Stack
- Go (go.mod requires 1.26; the toolchain auto-downloads) + **Wails v3.0.0** (`go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0`).
- Frontend: **plain TypeScript + Vite, with no UI framework**, to keep it as low-level as possible. It is bundled into the Go binary.
- Map: **Mapbox GL JS**, using your API key.
- No database. The app is stateless apart from a small config file.

## Architecture
```
exif-photos-helper/
├── main.go                 # Wails bootstrap, window options, file-drop enabled
├── app.go                  # Methods bound to the frontend (OpenFolder, Scan, ApplyEdit, GetMapboxToken…)
├── internal/
│   ├── exif/exiftool.go    # Locates the exiftool binary; one persistent `-stay_open` process for reads/writes
│   ├── scan/scan.go        # Lists the folder (top level only, no subfolders), filters .jpg/.jpeg/.heic, batch-reads metadata as JSON
│   ├── thumbs/thumbs.go    # Extracts the embedded ThumbnailImage via exiftool, falls back to `sips` resize; memory + disk cache
│   ├── datesel/date.go     # Partial-date model {decade, year?, month?, day?, hour?} → "YYYY:MM:DD HH:00:00"
│   └── config/config.go    # Mapbox token: $MAPBOX_TOKEN env → .env (dev) → value embedded at build time
└── frontend/
    ├── index.html
    └── src/{main.ts, grid.ts, basket.ts, datePicker.ts, mapPicker.ts, style.css}
```

### Key details
- **Finding exiftool:** apps launched from Finder don't inherit the shell `PATH`. The app checks `/opt/local/bin`, `/usr/local/bin`, `/opt/homebrew/bin` and then `PATH`. If exiftool isn't found, it shows a clear error screen.
- **Speed:** one `exiftool -stay_open True -@ -` process stays alive for the whole session. Scanning uses `-json -fast2 -DateTimeOriginal -GPSLatitude# -GPSLongitude# -Orientation` over the folder's files (no `-r`) in a single call. Thumbnails are served through a Wails asset handler (`/thumb?path=…`) and loaded lazily (`loading="lazy"` plus IntersectionObserver).
- **Folder drop:** Wails `OnFileDrop` receives the native path. A "Choose folder…" button (native dialog) is the fallback.
- **Partial date rule:** fields you don't pick default to their lowest value. Picking only 1996 gives `1996:01:01 00:00:00`. Picking 1996-07 gives `1996:07:01 00:00:00`, and so on. Choosing only a decade (e.g. "1990s") writes `1990:01:01`.
- **Write command, one call per basket:**
  `-overwrite_original -DateTimeOriginal=… -CreateDate=… -ModifyDate=… -FileModifyDate=… -GPSLatitude=… -GPSLatitudeRef=N|S -GPSLongitude=… -GPSLongitudeRef=E|W <files>`
  Date and location are applied independently, so you can set only one of them.
- **After writing:** the app re-reads the edited files, refreshes their badges and empties the basket.

## UI
- **Start screen:** a drop zone plus a "Choose folder" button.
- **Main view:**
  - Left (large): a thumbnail grid. Each tile shows small badges for its current date and GPS (📅/📍, greyed when missing).
  - Selection: click to toggle; shift-click selects a range; `⌘A` selects all; a filter shows only photos with no date or no GPS.
  - **Basket panel** (right sidebar): selected photos, a count, and "clear".
- **Edit panel** (below the basket):
  - **Date picker:** a row of decade chips (1900s…2020s). Picking one shows its 10 year chips, then 12 month chips, then a day grid, then 24 hour chips. A breadcrumb (`1990s › 1996 › Jul`) lets you stop or step back at any level.
  - **Map picker:** Mapbox map with a search box (Mapbox geocoding). Clicking drops a pin, and the coordinates are shown.
  - An **Apply to N photos** button with a confirmation, since there is no backup.
- **Mapbox token:** read from `.env` (`MAPBOX_TOKEN`, git-ignored). `wails build` embeds it via `-ldflags "-X …config.embeddedToken=…"` so the `.app` works when launched from Finder. No settings screen.

## Implementation steps
0. Copy this plan to `./plan.md`, run `git init`, and add a `.gitignore`.
1. Run `wails init` with the vanilla-ts template and trim the boilerplate.
2. `internal/exif`: binary lookup plus the stay_open wrapper, with unit tests against fixture JPGs.
3. `internal/scan` and `internal/thumbs`, with folder drop and the grid render.
4. Selection and basket UI.
5. `internal/datesel` with table-driven tests, then the drill-down date picker UI.
6. Mapbox token loading (`.env` / build-time embed) and the map picker.
7. `ApplyEdit` end to end, then refresh.
8. `wails build`, producing a `.app` bundle in `build/bin/`.

## Verification
- `go test ./...`: date-precision mapping, GPS sign→Ref conversion, and exiftool argument building, with round-trip tests on a copied fixture JPG and HEIC.
- `wails dev`: drop a test folder, select photos, set "1996" plus a map click, and apply.
- Check the result independently: `exiftool -a -G1 -time:all -gps:all file.jpg` shows `1996:01:01 00:00:00` and the correct lat/lon. Open the photo in Apple Photos/Preview to confirm the location is on the map.
- Always test on **copies** of photos, because writes are destructive.

## Open questions / assumptions
- Mapbox: public token (Mapbox token name `go-app-exif-geo`) stored in `.env`.
- Timezone: never written (no `OffsetTime*` tags); dates are plain wall-clock time.
- Subfolders: **not** scanned; only files directly inside the dropped folder.

## Status & usage
Implemented (steps 0–8). Requires `~/go/bin` on `PATH` for `wails3`.

| Command | What it does |
|---|---|
| `wails3 dev` | Dev mode with hot reload |
| `wails3 task test` | Go tests (unit + real exiftool round-trips on JPG/HEIC) |
| `wails3 package` | Builds `bin/exif-photos-helper.app` (Mapbox token from `.env` embedded) |
| `bin/exif-photos-helper.app/Contents/MacOS/exif-photos-helper <folder>` | Opens a folder directly |

Extras added during implementation: double-click a photo for a large preview; filters "No date" / "No location";
basket photos that already have GPS show as grey dots on the map; place search (Mapbox geocoding);
the pin is draggable; Esc clears the basket, ⌘A selects all visible.
