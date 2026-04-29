<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import * as api from "../services/api";
import type { Place, Review } from "../types/models";
import { useAuthStore } from "../stores/auth";
import SensoryRating from "../components/SensoryRating.vue";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const place = ref<Place | null>(null);
const reviews = ref<Review[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

const review = ref({
  text: "",
  noise: 3,
  light: 3,
  crowd: 3,
  smell: 1,
  visual: 3,
});
const isFavorite = ref(false);

async function load(): Promise<void> {
  loading.value = true;
  try {
    const id = Number(route.params.id);
    place.value = await api.places.get(id);
    const r = await api.reviews.byPlace(id);
    reviews.value = r.items;

    if (auth.isAuthenticated) {
      const favs = await api.favorites.list();
      isFavorite.value = favs.items.some((p) => p.id === id);
    }
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}

async function submitReview(): Promise<void> {
  if (!place.value) return;
  if (!auth.isAuthenticated) {
    router.push({ name: "login", query: { redirect: route.fullPath } });
    return;
  }
  try {
    await api.reviews.create(place.value.id, review.value);
    await load();
  } catch (e) {
    error.value = (e as Error).message;
  }
}

async function deleteReview(id: number): Promise<void> {
  if (!confirm("удалить отзыв?")) return;
  try {
    await api.reviews.remove(id);
    await load();
  } catch (e) {
    error.value = (e as Error).message;
  }
}

async function toggleFavorite(): Promise<void> {
  if (!place.value) return;
  if (!auth.isAuthenticated) {
    router.push({ name: "login" });
    return;
  }
  try {
    if (isFavorite.value) {
      await api.favorites.remove(place.value.id);
    } else {
      await api.favorites.add(place.value.id);
    }
    isFavorite.value = !isFavorite.value;
  } catch (e) {
    error.value = (e as Error).message;
  }
}

onMounted(load);
</script>

<template>
  <div class="page">
    <div v-if="loading">загрузка…</div>
    <div v-if="error" class="error">{{ error }}</div>

    <template v-if="place">
      <div class="row">
        <button class="btn-secondary" @click="router.back()">← Назад</button>
        <span class="spacer"></span>
        <button :class="isFavorite ? 'btn-danger' : 'btn-secondary'" @click="toggleFavorite">
          {{ isFavorite ? "★ В избранном" : "☆ В избранное" }}
        </button>
      </div>

      <div class="grid">
        <div class="card">
          <h2 class="title">{{ place.name }}</h2>
          <div class="muted">{{ place.address }}</div>
          <div class="row tags">
            <span class="tag">{{ place.category?.name || "—" }}</span>
            <span class="tag">{{ place.reviews_count }} отзыв(ов)</span>
          </div>

          <div class="divider"></div>

          <p>{{ place.description }}</p>

          <div class="divider"></div>
          <h3>Сенсорные оценки</h3>
          <SensoryRating label="Шум" :value="place.avg_noise" />
          <SensoryRating label="Свет" :value="place.avg_light" />
          <SensoryRating label="Заполненность" :value="place.avg_crowd" />
          <SensoryRating label="Запахи" :value="place.avg_smell" />
          <SensoryRating label="Визуальная нагрузка" :value="place.avg_visual" />
        </div>

        <div class="card">
          <h3>Оставить отзыв</h3>
          <div v-if="!auth.isAuthenticated" class="muted">
            Чтобы оставить отзыв, <RouterLink to="/login">войдите</RouterLink>.
          </div>
          <form v-else @submit.prevent="submitReview">
            <div class="field">
              <label>Шум: {{ review.noise }} (1 — тихо, 5 — шумно)</label>
              <input v-model.number="review.noise" type="range" min="1" max="5" />
            </div>
            <div class="field">
              <label>Свет: {{ review.light }}</label>
              <input v-model.number="review.light" type="range" min="1" max="5" />
            </div>
            <div class="field">
              <label>Заполненность: {{ review.crowd }}</label>
              <input v-model.number="review.crowd" type="range" min="1" max="5" />
            </div>
            <div class="field">
              <label>Резкие запахи: {{ review.smell }}</label>
              <input v-model.number="review.smell" type="range" min="1" max="5" />
            </div>
            <div class="field">
              <label>Визуальная нагрузка: {{ review.visual }}</label>
              <input v-model.number="review.visual" type="range" min="1" max="5" />
            </div>
            <div class="field">
              <label>Комментарий (необязательно)</label>
              <textarea v-model="review.text" rows="3" maxlength="1000"></textarea>
            </div>
            <button type="submit">Отправить отзыв</button>
          </form>
        </div>
      </div>

      <div class="card mt">
        <h3>Отзывы</h3>
        <div v-if="reviews.length === 0" class="muted">Пока нет отзывов</div>
        <div v-for="r in reviews" :key="r.id" class="review">
          <div class="row">
            <strong>{{ r.user?.display_name || r.user?.username || "Пользователь" }}</strong>
            <span class="spacer"></span>
            <span class="muted">{{ new Date(r.created_at).toLocaleDateString("ru") }}</span>
            <button v-if="auth.user?.id === r.user_id" class="btn-secondary" @click="deleteReview(r.id)">
              Удалить
            </button>
          </div>
          <div class="row tags">
            <span class="tag">Шум: {{ r.noise }}</span>
            <span class="tag">Свет: {{ r.light }}</span>
            <span class="tag">Людность: {{ r.crowd }}</span>
            <span class="tag">Запахи: {{ r.smell }}</span>
            <span class="tag">Визуально: {{ r.visual }}</span>
          </div>
          <p v-if="r.text">{{ r.text }}</p>
          <div class="divider"></div>
        </div>
      </div>
    </template>
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
.title { margin: 0 0 4px; font-size: 22px; }
.tags { margin-top: 8px; }
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-top: 16px;
}
.review { padding: 8px 0; }
.mt { margin-top: 16px; }
@media (max-width: 800px) {
  .grid { grid-template-columns: 1fr; }
}
</style>
