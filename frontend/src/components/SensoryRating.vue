<script setup lang="ts">
defineProps<{
  label: string;
  value: number;
  /** "1 — комфортно, 5 — некомфортно" */
  helper?: string;
}>();

function colorClass(v: number): string {
  if (v <= 2) return "comfort-low";
  if (v <= 3.5) return "comfort-mid";
  return "comfort-high";
}

function description(label: string, v: number): string {
  if (v === 0) return "нет данных";
  switch (label) {
    case "Шум":
      return v <= 2 ? "тихо" : v <= 3.5 ? "умеренно" : "шумно";
    case "Свет":
      return v <= 2 ? "приглушённый" : v <= 3.5 ? "средний" : "яркий";
    case "Заполненность":
      return v <= 2 ? "малолюдно" : v <= 3.5 ? "средне" : "очень многолюдно";
    case "Запахи":
      return v <= 2 ? "почти нет" : v <= 3.5 ? "ощутимы" : "сильные";
    case "Визуальная нагрузка":
      return v <= 2 ? "спокойно" : v <= 3.5 ? "средне" : "пёстро";
  }
  return v.toFixed(1);
}
</script>

<template>
  <div class="rating">
    <div class="row">
      <span class="label">{{ label }}</span>
      <span class="spacer"></span>
      <span :class="['value', colorClass(value)]">
        {{ value > 0 ? value.toFixed(1) : "—" }}
      </span>
    </div>
    <div class="bar">
      <div
        class="fill"
        :class="colorClass(value)"
        :style="{ width: (value / 5) * 100 + '%' }"
      ></div>
    </div>
    <div class="muted desc">{{ description(label, value) }}</div>
  </div>
</template>

<style scoped>
.rating { margin-bottom: 8px; }
.label { font-weight: 500; }
.value { font-weight: 600; }
.bar {
  height: 6px;
  background: var(--surface-2);
  border-radius: 3px;
  overflow: hidden;
  margin-top: 4px;
}
.fill { height: 100%; transition: width 0.2s ease; border-radius: 3px; }
.fill.comfort-low { background: var(--comfort-low); }
.fill.comfort-mid { background: var(--comfort-mid); }
.fill.comfort-high { background: var(--comfort-high); }
.desc { margin-top: 2px; font-size: 12px; }
</style>
