import { defineStore } from "pinia";
import { ref, computed } from "vue";
import * as api from "../services/api";
import type { Category, Place, PlaceFilter } from "../types/models";

export const usePlacesStore = defineStore("places", () => {
  const items = ref<Place[]>([]);
  const categories = ref<Category[]>([]);
  const selected = ref<Place | null>(null);
  const filter = ref<PlaceFilter>({});
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function loadCategories(): Promise<void> {
    if (categories.value.length > 0) return;
    const res = await api.categories.list();
    categories.value = res.items;
  }

  async function loadAll(): Promise<void> {
    loading.value = true;
    error.value = null;
    try {
      const res = await api.places.list(filter.value);
      items.value = res.items;
    } catch (e) {
      error.value = (e as Error).message;
    } finally {
      loading.value = false;
    }
  }

  async function loadPlace(id: number): Promise<Place | null> {
    try {
      selected.value = await api.places.get(id);
      return selected.value;
    } catch (e) {
      error.value = (e as Error).message;
      return null;
    }
  }

  async function createPlace(input: {
    name: string;
    address?: string;
    description?: string;
    category_id: number;
    latitude: number;
    longitude: number;
  }): Promise<Place> {
    const created = await api.places.create(input);
    items.value.push(created);
    return created;
  }

  function setFilter(next: PlaceFilter): void {
    filter.value = { ...filter.value, ...next };
  }

  const visibleItems = computed(() => items.value);

  return {
    items,
    categories,
    selected,
    filter,
    loading,
    error,
    visibleItems,
    loadAll,
    loadCategories,
    loadPlace,
    createPlace,
    setFilter,
  };
});
