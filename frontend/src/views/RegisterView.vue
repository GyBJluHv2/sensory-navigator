<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "../stores/auth";

const auth = useAuthStore();
const router = useRouter();

const email = ref("");
const username = ref("");
const displayName = ref("");
const password = ref("");
const confirm = ref("");
const error = ref<string | null>(null);

async function submit(): Promise<void> {
  error.value = null;
  if (password.value !== confirm.value) {
    error.value = "пароли не совпадают";
    return;
  }
  try {
    await auth.register({
      email: email.value,
      username: username.value,
      password: password.value,
      display_name: displayName.value,
    });
    router.push("/");
  } catch (e) {
    error.value = (e as Error).message;
  }
}
</script>

<template>
  <div class="center-form">
    <form class="card" @submit.prevent="submit">
      <h1>Регистрация</h1>
      <div v-if="error" class="error">{{ error }}</div>

      <div class="field">
        <label>Email</label>
        <input v-model="email" type="email" required />
      </div>

      <div class="field">
        <label>Имя пользователя</label>
        <input v-model="username" type="text" minlength="3" required />
      </div>

      <div class="field">
        <label>Отображаемое имя</label>
        <input v-model="displayName" type="text" />
      </div>

      <div class="field">
        <label>Пароль (мин. 6 символов)</label>
        <input v-model="password" type="password" minlength="6" required />
      </div>

      <div class="field">
        <label>Подтверждение пароля</label>
        <input v-model="confirm" type="password" minlength="6" required />
      </div>

      <button type="submit" :disabled="auth.loading">Создать аккаунт</button>
      <div class="muted mt">
        Уже зарегистрированы?
        <RouterLink to="/login">Войти</RouterLink>
      </div>
    </form>
  </div>
</template>

<style scoped>
.mt { margin-top: 8px; }
</style>
