<script setup lang="ts">
import { onMounted, ref } from "vue";
import * as api from "../services/api";
import type { Place } from "../types/models";

const items = ref<Place[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

async function load(): Promise<void> {
  loading.value = true;
  try {
    const r = await api.favorites.list();
    items.value = r.items;
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}

async function remove(placeId: number): Promise<void> {
  await api.favorites.remove(placeId);
  items.value = items.value.filter((p) => p.id !== placeId);
}

onMounted(load);
</script>

<template>
  <div class="page">
    <h1>Избранные места</h1>
    <div v-if="loading">загрузка…</div>
    <div v-if="error" class="error">{{ error }}</div>

    <div class="grid">
      <div v-for="p in items" :key="p.id" class="card">
        <RouterLink :to="`/place/${p.id}`" class="title">{{ p.name }}</RouterLink>
        <div class="muted">{{ p.address }}</div>
        <div class="row tags">
          <span class="tag">{{ p.category?.name || "—" }}</span>
          <span class="tag">{{ p.reviews_count }} отзыв(ов)</span>
        </div>
        <div class="divider"></div>
        <button class="btn-secondary" @click="remove(p.id)">Убрать из избранного</button>
      </div>
      <div v-if="items.length === 0 && !loading" class="muted">
        Пока ничего не добавлено в избранное.
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
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 12px;
  margin-top: 12px;
}
.title {
  font-weight: 600;
  text-decoration: none;
  color: var(--fg);
}
.tags { margin-top: 6px; }
</style>
