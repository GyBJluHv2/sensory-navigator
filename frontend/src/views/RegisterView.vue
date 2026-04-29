<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import * as api from "../services/api";

const router = useRouter();

const email = ref("");
const username = ref("");
const displayName = ref("");
const password = ref("");
const confirm = ref("");
const showPassword = ref(false);
const showConfirm = ref(false);
const error = ref<string | null>(null);
const loading = ref(false);

const emailValid = ref(true);

function checkEmail(): void {
  // Базовая валидация email на стороне клиента
  emailValid.value = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value);
}

async function submit(): Promise<void> {
  error.value = null;
  checkEmail();
  if (!emailValid.value) {
    error.value = "введите корректный email";
    return;
  }
  if (password.value !== confirm.value) {
    error.value = "пароли не совпадают";
    return;
  }
  loading.value = true;
  try {
    await api.auth.requestRegister({
      email: email.value,
      username: username.value,
      password: password.value,
      display_name: displayName.value,
    });
    // Переходим на экран ввода кода. Сами учётные данные передаём через query,
    // чтобы пользователь не вводил их повторно (на втором шаге уже идёт verify).
    router.push({
      name: "verifyEmail",
      query: { email: email.value },
    });
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
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
        <input
          v-model="email"
          type="email"
          required
          autocomplete="email"
          @blur="checkEmail"
        />
        <div v-if="!emailValid" class="muted small error-hint">
          формат email некорректен
        </div>
      </div>

      <div class="field">
        <label>Имя пользователя</label>
        <input v-model="username" type="text" minlength="3" maxlength="64" required />
      </div>

      <div class="field">
        <label>Отображаемое имя (необязательно)</label>
        <input v-model="displayName" type="text" />
      </div>

      <div class="field">
        <label>Пароль (мин. 6 символов)</label>
        <div class="password-wrap">
          <input
            v-model="password"
            :type="showPassword ? 'text' : 'password'"
            minlength="6"
            required
            autocomplete="new-password"
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

      <div class="field">
        <label>Подтверждение пароля</label>
        <div class="password-wrap">
          <input
            v-model="confirm"
            :type="showConfirm ? 'text' : 'password'"
            minlength="6"
            required
            autocomplete="new-password"
          />
          <button
            type="button"
            class="toggle-pass"
            @click="showConfirm = !showConfirm"
            :aria-label="showConfirm ? 'Скрыть пароль' : 'Показать пароль'"
            :title="showConfirm ? 'Скрыть пароль' : 'Показать пароль'"
          >
            {{ showConfirm ? '🙈' : '👁' }}
          </button>
        </div>
      </div>

      <button type="submit" :disabled="loading">
        {{ loading ? 'Отправка кода…' : 'Отправить код на email' }}
      </button>
      <div class="muted mt">
        Уже зарегистрированы?
        <RouterLink to="/login">Войти</RouterLink>
      </div>
    </form>
  </div>
</template>

<style scoped>
.mt { margin-top: 8px; }
.small { font-size: 12px; }
.error-hint { color: #c0392b; margin-top: 4px; }
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