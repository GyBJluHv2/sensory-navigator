import { defineStore } from "pinia";
import { ref, computed } from "vue";
import * as api from "../services/api";
import type { User } from "../types/models";

export const useAuthStore = defineStore("auth", () => {
  const token = ref<string | null>(api.getToken());
  const user = ref<User | null>(null);
  const loading = ref(false);

  const isAuthenticated = computed(() => !!token.value);

  async function login(email: string, password: string): Promise<void> {
    loading.value = true;
    try {
      const res = await api.auth.login({ email, password });
      token.value = res.token;
      user.value = res.user;
      api.setToken(res.token);
    } finally {
      loading.value = false;
    }
  }

  async function register(input: {
    email: string;
    username: string;
    password: string;
    display_name?: string;
  }): Promise<void> {
    loading.value = true;
    try {
      const res = await api.auth.register(input);
      token.value = res.token;
      user.value = res.user;
      api.setToken(res.token);
    } finally {
      loading.value = false;
    }
  }

  function logout(): void {
    token.value = null;
    user.value = null;
    api.setToken(null);
  }

  // applyToken используется на экране подтверждения email после
  // успешной верификации кода: сервер уже выдал JWT, клиенту достаточно
  // сохранить его и не делать повторный логин.
  function applyToken(t: string, u: User): void {
    token.value = t;
    user.value = u;
    api.setToken(t);
  }

  async function fetchMe(): Promise<void> {
    if (!token.value) return;
    try {
      user.value = await api.auth.me();
    } catch {
      logout();
    }
  }

  async function updateProfile(input: Partial<User>): Promise<void> {
    user.value = await api.auth.updateMe(input);
  }

  return {
    token,
    user,
    loading,
    isAuthenticated,
    login,
    register,
    logout,
    applyToken,
    fetchMe,
    updateProfile,
  };
});
