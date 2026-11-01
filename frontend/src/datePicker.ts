import type {PartialDate} from "../bindings/exif-photos-helper/internal/datesel/models";
import {PhotoService} from "../bindings/exif-photos-helper";
import {formatExifDate} from "./state";
import {t, monthNames} from "./i18n";

const crumbsEl = document.getElementById("date-crumbs")!;
const chipsEl = document.getElementById("date-chips")!;
const previewEl = document.getElementById("date-preview")!;

const FIRST_DECADE = 1880;
const LAST_DECADE = Math.floor(new Date().getFullYear() / 10) * 10;

/** Picked values, in order: decade, year, month, day, hour. */
let picked: number[] = [];
let onChange: () => void = () => {};

export function initDatePicker(changed: () => void) {
    onChange = changed;
    document.getElementById("date-clear")!.addEventListener("click", () => setLevel(0));
    crumbsEl.addEventListener("click", e => {
        const btn = (e.target as HTMLElement).closest<HTMLElement>("[data-level]");
        if (btn) setLevel(Number(btn.dataset.level));
    });
    chipsEl.addEventListener("click", e => {
        const btn = (e.target as HTMLElement).closest<HTMLElement>("[data-value]");
        if (!btn) return;
        picked.push(Number(btn.dataset.value));
        render();
        onChange();
    });
    render();
}

/** Keeps only the first n picked values. */
function setLevel(n: number) {
    picked = picked.slice(0, n);
    render();
    onChange();
}

/** The selected partial date, or null if the date shouldn't change. */
export function selectedDate(): PartialDate | null {
    if (picked.length === 0) return null;
    const [decade, year, month, day, hour] = picked;
    return {decade, year: year ?? null, month: month ?? null, day: day ?? null, hour: hour ?? null};
}

function crumbLabel(level: number, v: number): string {
    switch (level) {
        case 0: return t("date.decadeCrumb", {d: v});
        case 2: return monthNames()[v - 1];
        case 4: return `${String(v).padStart(2, "0")}h`;
        default: return String(v);
    }
}

function options(): {value: number, label: string}[] {
    const [decade, year, month] = picked;
    const range = (a: number, b: number) => Array.from({length: b - a + 1}, (_, i) => a + i);
    switch (picked.length) {
        case 0:
            return range(0, (LAST_DECADE - FIRST_DECADE) / 10).map(i => {
                const d = LAST_DECADE - i * 10;
                return {value: d, label: t("date.decadeChip", {d})};
            });
        case 1:
            return range(decade, decade + 9).map(y => ({value: y, label: String(y)}));
        case 2:
            return range(1, 12).map(m => ({value: m, label: monthNames()[m - 1]}));
        case 3:
            return range(1, new Date(year, month, 0).getDate()).map(d => ({value: d, label: String(d)}));
        case 4:
            return range(0, 23).map(h => ({value: h, label: String(h).padStart(2, "0")}));
        default:
            return [];
    }
}

const LEVEL_KEYS = ["", "date.level.year", "date.level.month", "date.level.day", "date.level.hour"] as const;

function render() {
    crumbsEl.innerHTML = picked.length === 0
        ? `<span class="crumb-placeholder">${t("date.placeholder")}</span>`
        : `<button class="crumb" data-level="0" title="${t("date.startOver")}">✕</button>` +
          picked.map((v, i) => `<button class="crumb ${i === picked.length - 1 ? "current" : ""}" data-level="${i + 1}">${crumbLabel(i, v)}</button>`).join(`<span class="sep">›</span>`);

    const opts = options();
    chipsEl.className = `chips level-${picked.length}`;
    chipsEl.innerHTML = opts.length
        ? (picked.length > 0 ? `<div class="chips-label">${t("date.refine", {level: t(LEVEL_KEYS[picked.length] as Exclude<typeof LEVEL_KEYS[number], "">)})}</div>` : "") +
          opts.map(o => `<button class="chip" data-value="${o.value}">${o.label}</button>`).join("")
        : "";

    const d = selectedDate();
    if (!d) {
        previewEl.textContent = t("date.unchanged");
        previewEl.classList.add("muted");
        return;
    }
    previewEl.classList.remove("muted");
    PhotoService.DatePreview(d)
        .then(s => { previewEl.textContent = t("date.willWrite", {date: formatExifDate(s), exif: s}); })
        .catch(err => { previewEl.textContent = String(err); });
}
