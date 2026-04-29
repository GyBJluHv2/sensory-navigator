<script setup lang="ts">
import { onMounted, ref, watch, onBeforeUnmount } from "vue";
import { useRouter } from "vue-router";
import L from "leaflet";
import "leaflet.markercluster";
import { usePlacesStore } from "../stores/places";
import { useAuthStore } from "../stores/auth";
import type { Place } from "../types/models";
import FilterPanel from "../components/FilterPanel.vue";
import PlaceCard from "../components/PlaceCard.vue";

const router = useRouter();
const places = usePlacesStore();
const auth = useAuthStore();

const mapEl = ref<HTMLDivElement | null>(null);
let map: L.Map | null = null;
let cluster: L.MarkerClusterGroup | null = null;
const markers = new Map<number, L.Marker>();
const popupCard = ref<Place | null>(null);
const popupPos = ref<{ x: number; y: number } | null>(null);

// Центр Москвы по умолчанию
const DEFAULT_CENTER: L.LatLngTuple = [55.751244, 37.618423];

function comfortClass(value: number): string {
  if (value === 0) return "bg-comfort-mid";
  if (value <= 2) return "bg-comfort-low";
  if (value <= 3.5) return "bg-comfort-mid";
  return "bg-comfort-high";
}

function makeMarker(place: Place): L.Marker {
  const overall = place.overall_avg || 0;
  const cls = comfortClass(overall);
  const label = overall === 0 ? "?" : overall.toFixed(1);
  const html = `<div class="sn-marker ${cls}">${label}</div>`;
  const icon = L.divIcon({
    className: "sn-divicon",
    html,
    iconSize: [28, 28],
    iconAnchor: [14, 14],
  });
  const m = L.marker([place.latitude, place.longitude], { icon });
  m.on("click", () => selectPlace(place));
  return m;
}

function selectPlace(place: Place): void {
  popupCard.value = place;
  if (map) {
    const point = map.latLngToContainerPoint([place.latitude, place.longitude]);
    popupPos.value = { x: point.x, y: point.y };
    map.flyTo([place.latitude, place.longitude], Math.max(map.getZoom(), 15));
  }
}

function rebuildMarkers(items: Place[]): void {
  if (!cluster) return;
  cluster.clearLayers();
  markers.clear();
  for (const p of items) {
    const m = makeMarker(p);
    markers.set(p.id, m);
    cluster.addLayer(m);
  }
}

function locateMe(): void {
  if (!map) return;
  if (!navigator.geolocation) {
    alert("геолокация недоступна в этом браузере");
    return;
  }
  navigator.geolocation.getCurrentPosition(
    (pos) => {
      const ll: L.LatLngTuple = [pos.coords.latitude, pos.coords.longitude];
      map?.flyTo(ll, 15);
      L.circleMarker(ll, {
        radius: 8,
        color: "#1565c0",
        fillColor: "#1976d2",
        fillOpacity: 0.7,
      }).addTo(map!);
    },
    () => alert("не удалось определить местоположение"),
  );
}

async function searchNearby(): Promise<void> {
  if (!map) return;
  const center = map.getCenter();
  try {
    const res = await import("../services/api").then((m) =>
      m.places.nearby(center.lat, center.lng, 2000),
    );
    rebuildMarkers(res.items);
  } catch (e) {
    alert((e as Error).message);
  }
}

function onAddPlace(): void {
  if (!auth.isAuthenticated) {
    router.push({ name: "login", query: { redirect: "/places/new" } });
    return;
  }
  router.push({ name: "addPlace" });
}

onMounted(async () => {
  if (!mapEl.value) return;

  map = L.map(mapEl.value, { zoomControl: true }).setView(DEFAULT_CENTER, 12);
  L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
    attribution:
      '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
    maxZoom: 19,
  }).addTo(map);

  cluster = L.markerClusterGroup({ chunkedLoading: true });
  map.addLayer(cluster);

  map.on("move", () => {
    if (popupCard.value && map) {
      const p = map.latLngToContainerPoint([
        popupCard.value.latitude,
        popupCard.value.longitude,
      ]);
      popupPos.value = { x: p.x, y: p.y };
    }
  });
  map.on("zoom", () => {
    if (popupCard.value && map) {
      const p = map.latLngToContainerPoint([
        popupCard.value.latitude,
        popupCard.value.longitude,
      ]);
      popupPos.value = { x: p.x, y: p.y };
    }
  });

  await places.loadCategories();
  await places.loadAll();
});

watch(
  () => places.items,
  (items) => rebuildMarkers(items),
  { immediate: true },
);

onBeforeUnmount(() => {
  map?.remove();
  map = null;
});
</script>

<template>
  <div class="map-page">
    <div ref="mapEl" class="map-canvas" aria-label="карта мест"></div>

    <FilterPanel />

    <div class="actions">
      <button @click="locateMe" title="Определить местоположение">📍 Я здесь</button>
      <button @click="searchNearby" class="btn-secondary">Места рядом</button>
      <button @click="onAddPlace" class="btn-secondary">＋ Добавить место</button>
    </div>

    <div
      v-if="popupCard && popupPos"
      class="floating-popup"
      :style="{ left: popupPos.x + 'px', top: popupPos.y + 'px' }"
    >
      <PlaceCard
        :place="popupCard"
        @close="popupCard = null"
        @open="(id: number) => router.push({ name: 'place', params: { id } })"
      />
    </div>

    <div v-if="places.loading" class="loading">загрузка…</div>
    <div v-if="places.error" class="error floating">{{ places.error }}</div>
  </div>
</template>

<style scoped>
.map-page {
  position: relative;
  width: 100%;
  height: 100%;
}
.map-canvas {
  position: absolute;
  inset: 0;
  z-index: 1;
}
.actions {
  position: absolute;
  bottom: 16px;
  right: 16px;
  z-index: 800;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.actions button {
  box-shadow: var(--shadow);
}
.floating-popup {
  position: absolute;
  z-index: 900;
  transform: translate(-50%, calc(-100% - 18px));
  pointer-events: auto;
}
.loading {
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 900;
  background: var(--surface);
  padding: 6px 12px;
  border-radius: 8px;
  border: 1px solid var(--border);
}
.error.floating {
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 900;
  max-width: 360px;
}
</style>
