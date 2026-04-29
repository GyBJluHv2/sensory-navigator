<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";
import * as api from "../services/api";
import type { Review } from "../types/models";

const auth = useAuthStore();
const router = useRouter();

const displayName = ref(auth.user?.display_name ?? "");
const noise = ref(auth.user?.noise_pref ?? 3);
const light = ref(auth.user?.light_pref ?? 3);
const crowd = ref(auth.user?.crowd_pref ?? 3);

const reviews = ref<Review[]>([]);
const message = ref<string | null>(null);
const error = ref<string | null>(null);

const oldPw = ref("");
const newPw = ref("");
const newPwConfirm = ref("");

async function load(): Promise<void> {
  await auth.fetchMe();
  if (auth.user) {
    displayName.value = auth.user.display_name ?? "";
    noise.value = auth.user.noise_pref;
    light.value = auth.user.light_pref;
    crowd.value = auth.user.crowd_pref;
  }
  try {
    const r = await api.reviews.myReviews();
    reviews.value = r.items;
  } catch (e) {
    error.value = (e as Error).message;
  }
}

async function saveProfile(): Promise<void> {
  message.value = null;
  error.value = null;
  try {
    await auth.updateProfile({
      display_name: displayName.value,
      noise_pref: noise.value,
      light_pref: light.value,
      crowd_pref: crowd.value,
    });
    message.value = "профиль сохранён";
  } catch (e) {
    error.value = (e as Error).message;
  }
}

async function changePassword(): Promise<void> {
  message.value = null;
  error.value = null;
  if (newPw.value !== newPwConfirm.value) {
    error.value = "новые пароли не совпадают";
    return;
  }
  try {
    await api.auth.changePassword(oldPw.value, newPw.value);
    oldPw.value = newPw.value = newPwConfirm.value = "";
    message.value = "пароль изменён";
  } catch (e) {
    error.value = (e as Error).message;
  }
}

function logout(): void {
  auth.logout();
  router.push("/");
}

onMounted(load);
</script>

<template>
  <div class="page">
    <div class="row">
      <h1>Профиль</h1>
      <span class="spacer"></span>
      <button class="btn-secondary" @click="logout">Выйти</button>
    </div>
    <div v-if="message" class="success">{{ message }}</div>
    <div v-if="error" class="error">{{ error }}</div>

    <div v-if="auth.user" class="grid">
      <form class="card" @submit.prevent="saveProfile">
        <h3>Личные данные</h3>
        <div class="muted">{{ auth.user.email }} · @{{ auth.user.username }}</div>
        <div class="field">
          <label>Отображаемое имя</label>
          <input v-model="displayName" type="text" />
        </div>
        <div class="divider"></div>

        <h3>Сенсорные предпочтения</h3>
        <div class="muted small">
          Места с нагрузкой не выше указанной будут выделяться при поиске.
        </div>

        <div class="field">
          <label>Допустимый шум: {{ noise }}</label>
          <input v-model.number="noise" type="range" min="1" max="5" />
        </div>
        <div class="field">
          <label>Допустимый свет: {{ light }}</label>
          <input v-model.number="light" type="range" min="1" max="5" />
        </div>
        <div class="field">
          <label>Допустимая заполненность: {{ crowd }}</label>
          <input v-model.number="crowd" type="range" min="1" max="5" />
        </div>

        <button type="submit">Сохранить</button>
      </form>

      <form class="card" @submit.prevent="changePassword">
        <h3>Смена пароля</h3>
        <div class="field">
          <label>Текущий пароль</label>
          <input v-model="oldPw" type="password" autocomplete="current-password" required />
        </div>
        <div class="field">
          <label>Новый пароль</label>
          <input v-model="newPw" type="password" minlength="6" autocomplete="new-password" required />
        </div>
        <div class="field">
          <label>Подтверждение</label>
          <input v-model="newPwConfirm" type="password" minlength="6" required />
        </div>
        <button type="submit">Изменить пароль</button>
      </form>
    </div>

    <div class="card mt">
      <h3>Мои отзывы ({{ reviews.length }})</h3>
      <div v-for="r in reviews" :key="r.id" class="review">
        <div class="row">
          <RouterLink :to="`/place/${r.place_id}`">Место #{{ r.place_id }}</RouterLink>
          <span class="spacer"></span>
          <span class="muted">{{ new Date(r.created_at).toLocaleDateString("ru") }}</span>
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
      <div v-if="reviews.length === 0" class="muted">Пока нет отзывов</div>
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
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-top: 12px;
}
.tags { margin-top: 4px; }
.review { padding: 8px 0; }
.mt { margin-top: 16px; }
.success {
  background: #e8f5e9;
  color: var(--primary-dark);
  padding: 8px 10px;
  border-radius: 6px;
  margin-bottom: 8px;
}
@media (max-width: 800px) {
  .grid { grid-template-columns: 1fr; }
}
</style>
