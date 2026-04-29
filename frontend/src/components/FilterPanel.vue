<script setup lang="ts">
import { ref, watch } from "vue";
import { usePlacesStore } from "../stores/places";

const places = usePlacesStore();

const search = ref(places.filter.q ?? "");
const categoryId = ref(places.filter.category_id ?? 0);
const noiseMax = ref(places.filter.noise_max ?? 0);
const lightMax = ref(places.filter.light_max ?? 0);
const crowdMax = ref(places.filter.crowd_max ?? 0);

let debounceTimer: number | undefined;

function applyDebounced(): void {
  window.clearTimeout(debounceTimer);
  debounceTimer = window.setTimeout(apply, 250);
}

async function apply(): Promise<void> {
  places.setFilter({
    q: search.value || undefined,
    category_id: categoryId.value || undefined,
    noise_max: noiseMax.value || undefined,
    light_max: lightMax.value || undefined,
    crowd_max: crowdMax.value || undefined,
  });
  await places.loadAll();
}

function reset(): void {
  search.value = "";
  categoryId.value = 0;
  noiseMax.value = 0;
  lightMax.value = 0;
  crowdMax.value = 0;
  apply();
}

watch(search, applyDebounced);
watch([categoryId, noiseMax, lightMax, crowdMax], apply);
</script>

<template>
  <aside class="panel card">
    <h3>Поиск и фильтры</h3>

    <div class="field">
      <label>Поиск по названию или адресу</label>
      <input
        v-model="search"
        type="text"
        placeholder="например, парк или библиотека"
        aria-label="строка поиска"
      />
    </div>

    <div class="field">
      <label>Категория</label>
      <select v-model.number="categoryId">
        <option :value="0">Все категории</option>
        <option v-for="c in places.categories" :key="c.id" :value="c.id">
          {{ c.name }}
        </option>
      </select>
    </div>

    <div class="muted small">
      Сенсорная нагрузка не выше (1 — низкая, 5 — высокая):
    </div>

    <div class="field">
      <label>Шум — не выше {{ noiseMax || "любой" }}</label>
      <input v-model.number="noiseMax" type="range" min="0" max="5" />
    </div>
    <div class="field">
      <label>Свет — не выше {{ lightMax || "любой" }}</label>
      <input v-model.number="lightMax" type="range" min="0" max="5" />
    </div>
    <div class="field">
      <label>Заполненность — не выше {{ crowdMax || "любой" }}</label>
      <input v-model.number="crowdMax" type="range" min="0" max="5" />
    </div>

    <div class="row">
      <button class="btn-secondary" @click="reset">Сбросить</button>
      <span class="spacer"></span>
      <span class="muted">{{ places.items.length }} мест</span>
    </div>
  </aside>
</template>

<style scoped>
.panel {
  position: absolute;
  top: 16px;
  left: 16px;
  width: 320px;
  max-height: calc(100% - 32px);
  overflow-y: auto;
  z-index: 800;
}
h3 { margin-top: 0; }
.small { font-size: 12px; margin-bottom: 4px; }
input[type="range"] { width: 100%; }
</style>
