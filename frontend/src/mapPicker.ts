import mapboxgl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import type {LatLon} from "../bindings/exif-photos-helper/internal/exif/models";
import type {Photo} from "./state";
import {escapeHTML} from "./state";

const previewEl = document.getElementById("loc-preview")!;
const searchForm = document.getElementById("search") as HTMLFormElement;
const searchInput = document.getElementById("search-input") as HTMLInputElement;
const resultsEl = document.getElementById("search-results")!;

let map: mapboxgl.Map | null = null;
let marker: mapboxgl.Marker | null = null;
let picked: LatLon | null = null;
let token = "";
let onChange: () => void = () => {};

const EXISTING_SOURCE = "basket-positions";

export function initMapPicker(accessToken: string, changed: () => void) {
    token = accessToken;
    onChange = changed;
    document.getElementById("loc-clear")!.addEventListener("click", () => setPosition(null));

    if (!token) {
        document.getElementById("map")!.innerHTML =
            `<div class="map-missing">No Mapbox token. Add MAPBOX_TOKEN to .env and rebuild.</div>`;
        renderPreview();
        return;
    }

    mapboxgl.accessToken = token;
    map = new mapboxgl.Map({
        container: "map",
        style: "mapbox://styles/mapbox/outdoors-v12",
        center: [8.2, 46.8],
        zoom: 5,
        attributionControl: false,
    });
    map.addControl(new mapboxgl.NavigationControl({showCompass: false}), "top-right");
    map.addControl(new mapboxgl.AttributionControl({compact: true}));
    map.getCanvas().style.cursor = "crosshair";
    map.on("click", e => setPosition({lat: e.lngLat.lat, lon: e.lngLat.lng}));
    map.on("load", () => {
        map!.addSource(EXISTING_SOURCE, {type: "geojson", data: {type: "FeatureCollection", features: []}});
        map!.addLayer({
            id: EXISTING_SOURCE,
            type: "circle",
            source: EXISTING_SOURCE,
            paint: {"circle-radius": 5, "circle-color": "#94a3b8", "circle-stroke-color": "#fff", "circle-stroke-width": 1.5},
        });
    });
    new ResizeObserver(() => map?.resize()).observe(document.getElementById("map")!);

    initSearch();
    renderPreview();
}

export function selectedPosition(): LatLon | null {
    return picked;
}

function setPosition(p: LatLon | null) {
    picked = p;
    marker?.remove();
    marker = null;
    if (p && map) {
        marker = new mapboxgl.Marker({color: "#f97316", draggable: true}).setLngLat([p.lon, p.lat]).addTo(map);
        marker.on("dragend", () => {
            const ll = marker!.getLngLat();
            picked = {lat: ll.lat, lon: ll.lng};
            renderPreview();
        });
    }
    renderPreview();
    onChange();
}

function renderPreview() {
    if (!picked) {
        previewEl.textContent = token ? "Locations won't be changed. Click the map to set one." : "Locations won't be changed.";
        previewEl.classList.add("muted");
        return;
    }
    previewEl.classList.remove("muted");
    previewEl.textContent = `Will write ${picked.lat.toFixed(6)}, ${picked.lon.toFixed(6)}`;
}

/** Shows the current positions of basket photos as grey dots. */
export function showBasketPositions(photos: Photo[]) {
    const src = map?.getSource(EXISTING_SOURCE) as mapboxgl.GeoJSONSource | undefined;
    if (!src) return;
    src.setData({
        type: "FeatureCollection",
        features: photos.filter(p => p.lat != null).map(p => ({
            type: "Feature",
            properties: {},
            geometry: {type: "Point", coordinates: [p.lon!, p.lat!]},
        })),
    });
}

interface GeoResult {
    name: string;
    lat: number;
    lon: number;
    bbox?: [number, number, number, number];
}

function initSearch() {
    let results: GeoResult[] = [];

    searchForm.addEventListener("submit", async e => {
        e.preventDefault();
        const q = searchInput.value.trim();
        if (!q) return;
        try {
            results = await geocode(q);
        } catch (err) {
            results = [];
            resultsEl.innerHTML = `<li class="muted">${escapeHTML(String(err))}</li>`;
            resultsEl.hidden = false;
            return;
        }
        resultsEl.innerHTML = results.length
            ? results.map((r, i) => `<li data-i="${i}">${escapeHTML(r.name)}</li>`).join("")
            : `<li class="muted">No results</li>`;
        resultsEl.hidden = false;
    });

    resultsEl.addEventListener("click", e => {
        const li = (e.target as HTMLElement).closest<HTMLElement>("[data-i]");
        if (!li) return;
        const r = results[Number(li.dataset.i)];
        resultsEl.hidden = true;
        if (r.bbox) {
            map?.fitBounds(r.bbox, {padding: 30, maxZoom: 15, duration: 600});
        } else {
            map?.flyTo({center: [r.lon, r.lat], zoom: 13, duration: 600});
        }
        setPosition({lat: r.lat, lon: r.lon});
    });

    searchInput.addEventListener("keydown", e => {
        if (e.key === "Escape") resultsEl.hidden = true;
    });
}

async function geocode(q: string): Promise<GeoResult[]> {
    const url = `https://api.mapbox.com/search/geocode/v6/forward?q=${encodeURIComponent(q)}&limit=6&access_token=${token}`;
    const res = await fetch(url);
    if (!res.ok) throw new Error(`Search failed (${res.status})`);
    const json = await res.json();
    return (json.features ?? []).map((f: any) => ({
        name: f.properties?.full_address || f.properties?.name || "?",
        lat: f.geometry.coordinates[1],
        lon: f.geometry.coordinates[0],
        bbox: f.properties?.bbox,
    }));
}
