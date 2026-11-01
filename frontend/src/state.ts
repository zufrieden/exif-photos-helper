import type {Photo} from "../bindings/exif-photos-helper/internal/exif/models";
import {lang} from "./i18n";

export type {Photo};
export type Filter = "all" | "nodate" | "nogps";

/** Shared app state. Mutate through the helpers, then call render(). */
export const state = {
    folder: "",
    photos: [] as Photo[],
    basket: new Set<string>(),
    filter: "all" as Filter,
};

export function visiblePhotos(): Photo[] {
    switch (state.filter) {
        case "nodate":
            return state.photos.filter(p => !p.date);
        case "nogps":
            return state.photos.filter(p => p.lat == null);
        default:
            return state.photos;
    }
}

export function basketPhotos(): Photo[] {
    return state.photos.filter(p => state.basket.has(p.path));
}

export function thumbURL(p: Photo, large = false): string {
    return `/thumb?path=${encodeURIComponent(p.path)}${large ? "&size=large" : ""}&d=${encodeURIComponent(p.date)}`;
}

/** CSS transform that displays an unrotated image per its EXIF orientation. */
export function orientationTransform(o: number): string {
    return [
        "",
        "",                              // 1 normal
        "scaleX(-1)",                    // 2 mirror horizontal
        "rotate(180deg)",                // 3
        "scaleY(-1)",                    // 4 mirror vertical
        "scaleX(-1) rotate(90deg)",      // 5 transpose
        "rotate(90deg)",                 // 6
        "scaleX(-1) rotate(-90deg)",     // 7 transverse
        "rotate(-90deg)",                // 8
    ][o] ?? "";
}

/** "1996:07:14 18:00:00" → "14 Jul 1996, 18:00" (or "14 juil. 1996, 18:00"). */
export function formatExifDate(d: string): string {
    const m = d.match(/^(\d{4}):(\d{2}):(\d{2}) (\d{2}):(\d{2})/);
    if (!m) return d;
    const [y, mo, da, h, mi] = m.slice(1).map(Number);
    const date = new Date(Date.UTC(2000, mo - 1, da, h, mi));
    date.setUTCFullYear(y); // Date.UTC maps years 0–99 to 1900–1999
    return new Intl.DateTimeFormat(lang, {
        day: "numeric", month: "short", year: "numeric",
        hour: "2-digit", minute: "2-digit", timeZone: "UTC",
    }).format(date);
}

export function escapeHTML(s: string): string {
    return s.replace(/[&<>"']/g, c => ({"&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;"}[c]!));
}
