<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import * as api from "../services/api";
import { useAuthStore } from "../stores/auth";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();

const email = ref<string>((route.query.email as string) || "");
const code = ref("");
const error = ref<string | null>(null);
const success = ref<string | null>(null);
const loading = ref(false);
const resending = ref(false);

onMounted(() => {
  if (!email.value) {
    router.replace({ name: "register" });
  }
});

async function submit(): Promise<void> {
  error.value = null;
  success.value = null;
  if (!/^\d{6}$/.test(code.value)) {
    error.value = "введите 6-значный код из письма";
    return;
  }
  loading.value = true;
  try {
    const res = await api.auth.confirmRegister({
      email: email.value,
      code: code.value,
    });
    auth.applyToken(res.token, res.user);
    router.push("/");
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    loading.value = false;
  }
}

async function resend(): Promise<void> {
  error.value = null;
  success.value = null;
  resending.value = true;
  try {
    await api.auth.resendCode({ email: email.value });
    success.value = "новый код отправлен";
  } catch (e) {
    error.value = (e as Error).message;
  } finally {
    resending.value = false;
  }
}
</script>

<template>
  <div class="center-form">
    <form class="card" @submit.prevent="submit">
      <h1>Подтверждение email</h1>
      <div v-if="error" class="error">{{ error }}</div>
      <div v-if="success" class="success">{{ success }}</div>

      <p class="muted">
        Мы отправили 6-значный код на адрес
        <strong>{{ email }}</strong>.
        Введите его ниже, чтобы завершить регистрацию.
      </p>

      <div class="field">
        <label>Код подтверждения</label>
        <input
          v-model="code"
          type="text"
          inputmode="numeric"
          pattern="\d{6}"
          maxlength="6"
          minlength="6"
          required
          autocomplete="one-time-code"
          placeholder="123456"
          class="code-input"
        />
      </div>

      <button type="submit" :disabled="loading">
        {{ loading ? 'Проверка…' : 'Подтвердить' }}
      </button>

      <div class="muted mt">
        Не пришёл код?
        <button
          type="button"
          class="link-btn"
          @click="resend"
          :disabled="resending"
        >
          {{ resending ? 'Отправка…' : 'Отправить ещё раз' }}
        </button>
      </div>

      <div class="muted mt">
        <RouterLink :to="{ name: 'register' }">← к регистрации</RouterLink>
      </div>
    </form>
  </div>
</template>

<style scoped>
.mt { margin-top: 8px; }
.code-input {
  font-size: 22px;
  letter-spacing: 8px;
  text-align: center;
  font-family: ui-monospace, "JetBrains Mono", Consolas, monospace;
}
.success {
  color: #2e7d32;
  background: #e8f5e9;
  border: 1px solid #a5d6a7;
  padding: 6px 10px;
  border-radius: 6px;
  margin-bottom: 8px;
}
.link-btn {
  background: transparent;
  border: 0;
  padding: 0;
  margin: 0;
  color: var(--primary, #1565c0);
  cursor: pointer;
  text-decoration: underline;
  font: inherit;
}
.link-btn:disabled {
  color: var(--fg-muted);
  cursor: default;
  text-decoration: none;
}
</style>