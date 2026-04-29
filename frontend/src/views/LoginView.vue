<script setup lang="ts">
import { ref } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useAuthStore } from "../stores/auth";

const auth = useAuthStore();
const router = useRouter();
const route = useRoute();

const email = ref("");
const password = ref("");
const showPassword = ref(false);
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
        <div class="password-wrap">
          <input
            id="password"
            v-model="password"
            :type="showPassword ? 'text' : 'password'"
            autocomplete="current-password"
            required
          />
          <button
            type="button"
            class="toggle-pass"
            @click="showPassword = !showPassword"
            :aria-label="showPassword ? 'Скрыть пароль' : 'Показать пароль'"
            :title="showPassword ? 'Скрыть пароль' : 'Показать пароль'"
          >
            {{ showPassword ? '🙈' : '👁' }}
          </button>
        </div>
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
.password-wrap {
  position: relative;
  display: flex;
  align-items: center;
}
.password-wrap input {
  width: 100%;
  padding-right: 44px;
}
.toggle-pass {
  position: absolute;
  right: 6px;
  top: 50%;
  transform: translateY(-50%);
  background: transparent;
  border: 0;
  padding: 4px 8px;
  cursor: pointer;
  font-size: 18px;
  line-height: 1;
  color: var(--fg-muted);
}
.toggle-pass:hover { color: var(--fg); }
</style>