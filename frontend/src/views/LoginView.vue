<script setup lang="ts">
import { ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useAuthStore } from "../stores/auth";

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const email = ref("");
const password = ref("");
const error = ref<string | null>(null);

async function submit(): Promise<void> {
  error.value = null;
  try {
    await auth.login(email.value, password.value);
    const dst = (route.query.redirect as string) || "/";
    router.push(dst);
  } catch (e) {
    error.value = (e as Error).message;
  }
}
</script>

<template>
  <div class="center-form">
    <form class="card" @submit.prevent="submit">
      <h1>Вход в аккаунт</h1>
      <div v-if="error" class="error">{{ error }}</div>

      <div class="field">
        <label for="email">Email</label>
        <input id="email" v-model="email" type="email" autocomplete="email" required />
      </div>

      <div class="field">
        <label for="password">Пароль</label>
        <input id="password" v-model="password" type="password" autocomplete="current-password" required />
      </div>

      <button type="submit" :disabled="auth.loading">Войти</button>
      <div class="muted mt">
        Нет аккаунта?
        <RouterLink to="/register">Зарегистрироваться</RouterLink>
      </div>
    </form>
  </div>
</template>

<style scoped>
.mt { margin-top: 8px; }
</style>
