/** UI strings. English is the default; French is used when it is the
 * user's first preferred language. */

const en = {
    "app.title": "EXIF Photos Helper",
    "toolbar.all": "All",
    "toolbar.noDate": "No date",
    "toolbar.noLocation": "No location",
    "toolbar.selectAll": "Select all",
    "toolbar.selectAllTitle": "Add all visible photos to the basket (⌘A)",
    "toolbar.openFolder": "Open folder…",
    "start.title": "Drop a photo folder here",
    "start.text": "JPG and HEIC files directly inside the folder are loaded (subfolders are ignored).",
    "start.choose": "Choose folder…",
    "fatal.title": "Can't start",
    "fatal.exiftool": "exiftool was not found. Install it with: sudo port install p5-image-exiftool",
    "grid.empty": "No photos match this filter.",
    "grid.noDate": "no date",
    "grid.noDateTitle": "No date",
    "grid.noLocationTitle": "No location",
    "basket.title": "Basket",
    "basket.clear": "Clear",
    "basket.hint": "Click photos to add them. Shift-click selects a range.",
    "basket.remove": "click to remove",
    "date.title": "Date",
    "date.keep": "Don't change",
    "date.placeholder": "Pick a decade — stop at any level",
    "date.startOver": "Start over",
    "date.decadeChip": "{d}s",
    "date.decadeCrumb": "{d}s",
    "date.refine": "Refine {level} (optional)",
    "date.level.year": "year",
    "date.level.month": "month",
    "date.level.day": "day",
    "date.level.hour": "hour",
    "date.unchanged": "Dates won't be changed.",
    "date.willWrite": "Will write {date}  ({exif})",
    "loc.title": "Location",
    "loc.keep": "Don't change",
    "loc.search": "Search a place…",
    "loc.unchanged": "Locations won't be changed.",
    "loc.unchangedClick": "Locations won't be changed. Click the map to set one.",
    "loc.willWrite": "Will write {lat}, {lon}",
    "loc.noToken": "No Mapbox token. Add MAPBOX_TOKEN to .env and rebuild.",
    "loc.noResults": "No results",
    "loc.searchFailed": "Search failed ({status})",
    "apply.emptyBasket": "Add photos to the basket",
    "apply.pickSomething": "Pick a date and/or a location",
    "apply.date": "date",
    "apply.location": "location",
    "apply.and": " + ",
    "apply.button": "Write {what} to {n} {photos}",
    "apply.writing": "Writing…",
    "apply.confirmTitle": "Overwrite {n} {photos}?",
    "apply.confirmDate": "Date: {date}",
    "apply.confirmLocation": "Location: {lat}, {lon}",
    "apply.confirmWarning": "Files are modified in place, without backup.",
    "apply.confirmWrite": "Write",
    "apply.confirmCancel": "Cancel",
    "apply.doneOne": "Updated 1 photo.",
    "apply.doneMany": "Updated {n} photos.",
    "apply.failed": "Some photos could not be updated: {error}",
    "photo.one": "photo",
    "photo.many": "photos",
    "folder.loading": "Loading {dir}…",
    "folder.empty": "No JPG or HEIC photos directly in this folder.",
    "lightbox.noDate": "no date",
    "lightbox.noLocation": "no location",
};

type Key = keyof typeof en;

const fr: Record<Key, string> = {
    "app.title": "EXIF Photos Helper",
    "toolbar.all": "Toutes",
    "toolbar.noDate": "Sans date",
    "toolbar.noLocation": "Sans lieu",
    "toolbar.selectAll": "Tout sélectionner",
    "toolbar.selectAllTitle": "Ajouter toutes les photos visibles au panier (⌘A)",
    "toolbar.openFolder": "Ouvrir un dossier…",
    "start.title": "Déposez un dossier de photos ici",
    "start.text": "Les fichiers JPG et HEIC situés directement dans le dossier sont chargés (les sous-dossiers sont ignorés).",
    "start.choose": "Choisir un dossier…",
    "fatal.title": "Démarrage impossible",
    "fatal.exiftool": "exiftool est introuvable. Installez-le avec : sudo port install p5-image-exiftool",
    "grid.empty": "Aucune photo ne correspond à ce filtre.",
    "grid.noDate": "sans date",
    "grid.noDateTitle": "Sans date",
    "grid.noLocationTitle": "Sans lieu",
    "basket.title": "Panier",
    "basket.clear": "Vider",
    "basket.hint": "Cliquez sur des photos pour les ajouter. Maj-clic sélectionne une plage.",
    "basket.remove": "cliquer pour retirer",
    "date.title": "Date",
    "date.keep": "Ne pas modifier",
    "date.placeholder": "Choisissez une décennie — arrêtez-vous où vous voulez",
    "date.startOver": "Recommencer",
    "date.decadeChip": "{d}",
    "date.decadeCrumb": "années {d}",
    "date.refine": "Préciser {level} (facultatif)",
    "date.level.year": "l'année",
    "date.level.month": "le mois",
    "date.level.day": "le jour",
    "date.level.hour": "l'heure",
    "date.unchanged": "Les dates ne seront pas modifiées.",
    "date.willWrite": "Écrira {date}  ({exif})",
    "loc.title": "Lieu",
    "loc.keep": "Ne pas modifier",
    "loc.search": "Rechercher un lieu…",
    "loc.unchanged": "Les lieux ne seront pas modifiés.",
    "loc.unchangedClick": "Les lieux ne seront pas modifiés. Cliquez sur la carte pour en choisir un.",
    "loc.willWrite": "Écrira {lat}, {lon}",
    "loc.noToken": "Aucune clé Mapbox. Ajoutez MAPBOX_TOKEN dans .env et recompilez.",
    "loc.noResults": "Aucun résultat",
    "loc.searchFailed": "La recherche a échoué ({status})",
    "apply.emptyBasket": "Ajoutez des photos au panier",
    "apply.pickSomething": "Choisissez une date et/ou un lieu",
    "apply.date": "la date",
    "apply.location": "le lieu",
    "apply.and": " et ",
    "apply.button": "Écrire {what} sur {n} {photos}",
    "apply.writing": "Écriture…",
    "apply.confirmTitle": "Modifier {n} {photos} ?",
    "apply.confirmDate": "Date : {date}",
    "apply.confirmLocation": "Lieu : {lat}, {lon}",
    "apply.confirmWarning": "Les fichiers sont modifiés directement, sans copie de sauvegarde.",
    "apply.confirmWrite": "Écrire",
    "apply.confirmCancel": "Annuler",
    "apply.doneOne": "1 photo mise à jour.",
    "apply.doneMany": "{n} photos mises à jour.",
    "apply.failed": "Certaines photos n'ont pas pu être modifiées : {error}",
    "photo.one": "photo",
    "photo.many": "photos",
    "folder.loading": "Chargement de {dir}…",
    "folder.empty": "Aucune photo JPG ou HEIC directement dans ce dossier.",
    "lightbox.noDate": "sans date",
    "lightbox.noLocation": "sans lieu",
};

export let lang: "en" | "fr" = "en";

/** Picks the UI language from a BCP 47 tag such as "fr-CH". */
export function setLanguage(tag: string) {
    lang = tag.toLowerCase().startsWith("fr") ? "fr" : "en";
    document.documentElement.lang = lang;
}

export function t(key: Key, params: Record<string, string | number> = {}): string {
    const s = (lang === "fr" ? fr : en)[key] ?? en[key];
    return s.replace(/\{(\w+)\}/g, (_, k) => String(params[k] ?? `{${k}}`));
}

/** "photo" or "photos" for n. */
export function photos(n: number): string {
    return t(n > 1 ? "photo.many" : "photo.one");
}

/** Translates elements marked with data-i18n (text), data-i18n-title and
 * data-i18n-placeholder. */
export function translatePage() {
    for (const el of Array.from(document.querySelectorAll<HTMLElement>("[data-i18n]"))) {
        el.textContent = t(el.dataset.i18n as Key);
    }
    for (const el of Array.from(document.querySelectorAll<HTMLElement>("[data-i18n-title]"))) {
        el.title = t(el.dataset.i18nTitle as Key);
    }
    for (const el of Array.from(document.querySelectorAll<HTMLInputElement>("[data-i18n-placeholder]"))) {
        el.placeholder = t(el.dataset.i18nPlaceholder as Key);
    }
    document.title = t("app.title");
}

/** Short month names in the UI language ("Jul" / "juil."). */
export function monthNames(): string[] {
    const fmt = new Intl.DateTimeFormat(lang, {month: "short", timeZone: "UTC"});
    return Array.from({length: 12}, (_, i) => fmt.format(new Date(Date.UTC(2000, i, 1))));
}
