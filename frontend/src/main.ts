import "./style.css";
import {Dialogs, Events} from "@wailsio/runtime";
import {PhotoService} from "../bindings/exif-photos-helper";
import {state, visiblePhotos, basketPhotos, thumbURL, orientationTransform, formatExifDate, escapeHTML, type Filter, type Photo} from "./state";
import {initGrid, renderGrid} from "./grid";
import {initDatePicker, selectedDate} from "./datePicker";
import {initMapPicker, selectedPosition, showBasketPositions} from "./mapPicker";

const $ = <T extends HTMLElement = HTMLElement>(id: string) => document.getElementById(id) as T;

const startEl = $("start");
const workspaceEl = $("workspace");
const toolbarEl = $("toolbar");
const folderEl = $("folder");
const basketEl = $("basket");
const applyBtn = $<HTMLButtonElement>("apply");

async function main() {
    const problem = await PhotoService.Status();
    if (problem) {
        startEl.hidden = true;
        $("fatal").hidden = false;
        $("fatal-msg").textContent = problem;
        return;
    }

    initGrid(render, openLightbox);
    initDatePicker(renderApply);
    initMapPicker(await PhotoService.MapboxToken(), renderApply);

    $("open-folder").addEventListener("click", chooseFolder);
    $("open-folder-2").addEventListener("click", chooseFolder);
    Events.On("folder-dropped", e => loadFolder(e.data));

    $("filter").addEventListener("click", e => {
        const btn = (e.target as HTMLElement).closest<HTMLElement>("[data-filter]");
        if (!btn) return;
        state.filter = btn.dataset.filter as Filter;
        render();
    });
    $("select-all").addEventListener("click", selectAllVisible);
    $("basket-clear").addEventListener("click", () => { state.basket.clear(); render(); });
    basketEl.addEventListener("click", e => {
        const item = (e.target as HTMLElement).closest<HTMLElement>("[data-path]");
        if (item) { state.basket.delete(item.dataset.path!); render(); }
    });
    applyBtn.addEventListener("click", apply);

    const initial = await PhotoService.InitialFolder();
    if (initial) await loadFolder(initial);

    $("lightbox").addEventListener("click", () => { $("lightbox").hidden = true; });
    document.addEventListener("keydown", e => {
        const typing = (e.target as HTMLElement).matches("input, textarea");
        if (e.key === "Escape") {
            if (!$("lightbox").hidden) $("lightbox").hidden = true;
            else if (!typing) { state.basket.clear(); render(); }
        }
        if (e.key === "a" && e.metaKey && !typing && state.photos.length) {
            e.preventDefault();
            selectAllVisible();
        }
    });
}

async function chooseFolder() {
    try {
        const dir = await PhotoService.ChooseFolder();
        if (dir) await loadFolder(dir);
    } catch (err) {
        toast(String(err), true);
    }
}

async function loadFolder(dir: string) {
    folderEl.textContent = `Loading ${dir}…`;
    try {
        const photos = await PhotoService.Scan(dir);
        state.folder = dir;
        state.photos = photos ?? [];
        state.basket.clear();
        state.filter = "all";
        startEl.hidden = true;
        workspaceEl.hidden = false;
        toolbarEl.hidden = false;
        folderEl.textContent = dir;
        folderEl.title = dir;
        $("grid-wrap").scrollTop = 0;
        render();
        if (state.photos.length === 0) toast("No JPG or HEIC photos directly in this folder.", true);
    } catch (err) {
        folderEl.textContent = state.folder;
        toast(String(err), true);
    }
}

function selectAllVisible() {
    for (const p of visiblePhotos()) state.basket.add(p.path);
    render();
}

function render() {
    $("count-all").textContent = String(state.photos.length);
    $("count-nodate").textContent = String(state.photos.filter(p => !p.date).length);
    $("count-nogps").textContent = String(state.photos.filter(p => p.lat == null).length);
    for (const b of Array.from($("filter").children) as HTMLElement[]) {
        b.classList.toggle("active", b.dataset.filter === state.filter);
    }
    renderGrid();
    renderBasket();
    renderApply();
}

function renderBasket() {
    const photos = basketPhotos();
    $("basket-count").textContent = String(photos.length);
    $("basket-hint").hidden = photos.length > 0;
    basketEl.innerHTML = photos.map(p =>
        `<div class="basket-item" data-path="${escapeHTML(p.path)}" title="${escapeHTML(p.name)} — click to remove">
            <img src="${escapeHTML(thumbURL(p))}" style="transform:${orientationTransform(p.orientation)}" alt=""/>
        </div>`).join("");
    showBasketPositions(photos);
}

function renderApply() {
    const n = state.basket.size;
    const date = selectedDate();
    const pos = selectedPosition();
    const what = [date && "date", pos && "location"].filter(Boolean).join(" + ");
    applyBtn.disabled = n === 0 || !what;
    applyBtn.textContent = n === 0
        ? "Add photos to the basket"
        : !what
            ? "Pick a date and/or a location"
            : `Write ${what} to ${n} photo${n > 1 ? "s" : ""}`;
}

async function apply() {
    const files = basketPhotos().map(p => p.path);
    const date = selectedDate();
    const position = selectedPosition();
    if (!files.length || (!date && !position)) return;

    const lines: string[] = [];
    if (date) lines.push(`Date: ${formatExifDate(await PhotoService.DatePreview(date))}`);
    if (position) lines.push(`Location: ${position.lat.toFixed(6)}, ${position.lon.toFixed(6)}`);
    const answer = await Dialogs.Question({
        Title: `Overwrite ${files.length} photo${files.length > 1 ? "s" : ""}?`,
        Message: `${lines.join("\n")}\n\nFiles are modified in place, without backup.`,
        Buttons: [{Label: "Write", IsDefault: true}, {Label: "Cancel", IsCancel: true}],
    });
    if (answer !== "Write") return;

    applyBtn.disabled = true;
    applyBtn.textContent = "Writing…";
    let updated: Photo[] | null = null;
    try {
        updated = await PhotoService.Apply({files, date, position});
        toast(`Updated ${files.length} photo${files.length > 1 ? "s" : ""}.`);
        state.basket.clear();
    } catch (err) {
        toast(String(err), true);
        // Apply may still have updated some files; reload them all.
        updated = await PhotoService.Scan(state.folder).catch(() => null);
    }
    if (updated) mergePhotos(updated);
    render();
}

function mergePhotos(updated: Photo[]) {
    const byPath = new Map(updated.map(p => [p.path, p]));
    state.photos = state.photos.map(p => byPath.get(p.path) ?? p);
}

function openLightbox(p: Photo) {
    const img = $<HTMLImageElement>("lightbox-img");
    img.src = thumbURL(p, true);
    img.style.transform = orientationTransform(p.orientation);
    img.classList.toggle("sideways", p.orientation >= 5);
    const where = p.lat != null ? `📍 ${p.lat.toFixed(5)}, ${p.lon!.toFixed(5)}` : "📍 no location";
    $("lightbox-caption").textContent = `${p.name} · 📅 ${p.date ? formatExifDate(p.date) : "no date"} · ${where}`;
    $("lightbox").hidden = false;
}

let toastTimer: ReturnType<typeof setTimeout> | undefined;
function toast(msg: string, error = false) {
    const el = $("toast");
    el.textContent = msg;
    el.classList.toggle("error", error);
    el.hidden = false;
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => { el.hidden = true; }, error ? 8000 : 3500);
}

main();
