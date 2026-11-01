import {state, visiblePhotos, thumbURL, orientationTransform, formatExifDate, escapeHTML, type Photo} from "./state";

const grid = document.getElementById("grid")!;
const empty = document.getElementById("grid-empty")!;

let lastClicked: string | null = null;
let onChange: () => void = () => {};
let onOpen: (p: Photo) => void = () => {};

export function initGrid(changed: () => void, open: (p: Photo) => void) {
    onChange = changed;
    onOpen = open;

    grid.addEventListener("click", e => {
        const tile = (e.target as HTMLElement).closest<HTMLElement>(".tile");
        if (!tile) return;
        const path = tile.dataset.path!;
        if (e.shiftKey && lastClicked) {
            selectRange(lastClicked, path);
        } else if (state.basket.has(path)) {
            state.basket.delete(path);
        } else {
            state.basket.add(path);
        }
        lastClicked = path;
        onChange();
    });

    grid.addEventListener("dblclick", e => {
        const tile = (e.target as HTMLElement).closest<HTMLElement>(".tile");
        const photo = tile && state.photos.find(p => p.path === tile.dataset.path);
        if (photo) onOpen(photo);
    });
}

function selectRange(from: string, to: string) {
    const paths = visiblePhotos().map(p => p.path);
    let a = paths.indexOf(from), b = paths.indexOf(to);
    if (a < 0 || b < 0) return;
    if (a > b) [a, b] = [b, a];
    for (const p of paths.slice(a, b + 1)) state.basket.add(p);
}

/** Rebuilds the grid. Tiles are keyed by path so images aren't reloaded. */
export function renderGrid() {
    const photos = visiblePhotos();
    empty.hidden = photos.length > 0;

    const existing = new Map<string, HTMLElement>();
    for (const el of Array.from(grid.children) as HTMLElement[]) existing.set(el.dataset.key!, el);

    const frag = document.createDocumentFragment();
    for (const p of photos) {
        const key = p.path + "|" + p.date + "|" + p.lat + "|" + p.lon;
        let tile = existing.get(key);
        if (!tile) {
            tile = createTile(p);
            tile.dataset.key = key;
        }
        tile.classList.toggle("selected", state.basket.has(p.path));
        frag.appendChild(tile);
    }
    grid.replaceChildren(frag);
}

function createTile(p: Photo): HTMLElement {
    const tile = document.createElement("div");
    tile.className = "tile";
    tile.dataset.path = p.path;
    tile.title = p.name;

    const hasGPS = p.lat != null;
    const year = p.date ? p.date.slice(0, 4) : "no date";
    tile.innerHTML = `
        <div class="thumb"><img loading="lazy" decoding="async" alt="" style="transform:${orientationTransform(p.orientation)}" src="${escapeHTML(thumbURL(p))}"/></div>
        <div class="check">✓</div>
        <div class="meta">
            <span class="badge ${p.date ? "" : "missing"}" title="${escapeHTML(p.date ? formatExifDate(p.date) : "No date")}">📅 ${escapeHTML(year)}</span>
            <span class="badge ${hasGPS ? "" : "missing"}" title="${hasGPS ? `${p.lat!.toFixed(5)}, ${p.lon!.toFixed(5)}` : "No location"}">📍</span>
        </div>
        <div class="name">${escapeHTML(p.name)}</div>`;
    return tile;
}
