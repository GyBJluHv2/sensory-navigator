<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import L from "leaflet";
import * as api from "../services/api";
import { usePlacesStore } from "../stores/places";

const router = useRouter();
const places = usePlacesStore();

const name = ref("");
const address = ref("");
const description = ref("");
const categoryId = ref<number>(0);
const latitude = ref<number>(55.751244);
const longitude = ref<number>(37.618423);
const error = ref<string | null>(null);

const mapEl = ref<HTMLDivElement | null>(null);
let map: L.Map | null = null;
let marker: L.Marker | null = null;

function placeMarker(lat: number, lon: number): void {
  if (!map) return;
  if (!marker) {
    marker = L.marker([lat, lon], { draggable: true }).addTo(map);
    marker.on("dragend", () => {
      const ll = marker!.getLatLng();
      latitude.value = ll.lat;
      longitude.value = ll.lng;
    });
  } else {
    marker.setLatLng([lat, lon]);
  }
}

onMounted(async () => {
  await places.loadCategories();
  if (places.categories.length > 0) categoryId.value = places.categories[0].id;
  if (!mapEl.value) return;
  map = L.map(mapEl.value).setView([latitude.value, longitude.value], 13);
  L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
    attribution: "© OpenStreetMap",
    maxZoom: 19,
  }).addTo(map);
  placeMarker(latitude.value, longitude.value);
  map.on("click", (e: L.LeafletMouseEvent) => {
    latitude.value = e.latlng.lat;
    longitude.value = e.latlng.lng;
    placeMarker(latitude.value, longitude.value);
  });
});

watch([latitude, longitude], ([lat, lon]) => {
  if (Number.isFinite(lat) && Number.isFinite(lon)) {
    placeMarker(lat, lon);
  }
});

async function submit(): Promise<void> {
  error.value = null;
  try {
    const created = await api.places.create({
      name: name.value,
      address: address.value,
      description: description.value,
      category_id: categoryId.value,
      latitude: latitude.value,
      longitude: longitude.value,
    });
    router.push({ name: "place", params: { id: created.id } });
  } catch (e) {
    error.value = (e as Error).message;
  }
}
</script>

<template>
  <div class="page">
    <div class="row">
      <button class="btn-secondary" @click="router.back()">← Назад</button>
      <h1 class="title">Новое место</h1>
    </div>
    <div v-if="error" class="error">{{ error }}</div>

    <div class="grid">
      <form class="card" @submit.prevent="submit">
        <div class="field">
          <label>Название</label>
          <input v-model="name" type="text" required />
        </div>
        <div class="field">
          <label>Адрес</label>
          <input v-model="address" type="text" />
        </div>
        <div class="field">
          <label>Категория</label>
          <select v-model.number="categoryId">
            <option v-for="c in places.categories" :key="c.id" :value="c.id">
              {{ c.name }}
            </option>
          </select>
        </div>
        <div class="field">
          <label>Описание</label>
          <textarea v-model="description" rows="4"></textarea>
        </div>
        <div class="field row">
          <div style="flex:1">
            <label>Широта</label>
            <input v-model.number="latitude" type="number" step="0.000001" />
          </div>
          <div style="flex:1">
            <label>Долгота</label>
            <input v-model.number="longitude" type="number" step="0.000001" />
          </div>
        </div>
        <button type="submit">Сохранить место</button>
      </form>

      <div class="card">
        <p class="muted small">Кликните по карте, чтобы поставить маркер. Маркер можно перетаскивать.</p>
        <div ref="mapEl" class="picker"></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page {
  padding: 16px 24px;
  max-width: 1100px;
  margin: 0 auto;
  height: 100%;
  overflow-y: auto;
}
.title { margin-left: 12px; }
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-top: 16px;
}
.picker { width: 100%; height: 420px; border-radius: 8px; }
@media (max-width: 800px) {
  .grid { grid-template-columns: 1fr; }
}
</style>
