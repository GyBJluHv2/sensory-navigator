<script setup lang="ts">
import { computed } from "vue";
import type { Place } from "../types/models";
import SensoryRating from "./SensoryRating.vue";

const props = defineProps<{ place: Place }>();
defineEmits<{
  (e: "close"): void;
  (e: "open", id: number): void;
}>();

const overall = computed(() => props.place.overall_avg || 0);
const overallDesc = computed(() => {
  if (overall.value === 0) return "нет данных";
  if (overall.value <= 2) return "комфортно";
  if (overall.value <= 3.5) return "умеренная нагрузка";
  return "высокая нагрузка";
});
</script>

<template>
  <div class="place-card card">
    <div class="row">
      <div>
        <h3 class="title">{{ place.name }}</h3>
        <div class="muted">{{ place.address }}</div>
      </div>
      <span class="spacer"></span>
      <button class="btn-secondary" @click="$emit('close')">×</button>
    </div>

    <div class="row mt-2">
      <span class="tag">{{ place.category?.name || "—" }}</span>
      <span class="tag">{{ place.reviews_count }} отзыв(ов)</span>
      <span :class="['tag', overall <= 2 ? 'bg-comfort-low' : overall <= 3.5 ? 'bg-comfort-mid' : 'bg-comfort-high']">
        Сенсорно — {{ overallDesc }}
      </span>
    </div>

    <div class="divider"></div>

    <SensoryRating label="Шум" :value="place.avg_noise" />
    <SensoryRating label="Свет" :value="place.avg_light" />
    <SensoryRating label="Заполненность" :value="place.avg_crowd" />
    <SensoryRating label="Запахи" :value="place.avg_smell" />
    <SensoryRating label="Визуальная нагрузка" :value="place.avg_visual" />

    <div class="row mt-2">
      <button @click="$emit('open', place.id)">Подробнее и отзывы</button>
    </div>
  </div>
</template>

<style scoped>
.place-card { width: 320px; }
.title { margin: 0 0 2px; font-size: 16px; }
.mt-2 { margin-top: 8px; }
</style>
