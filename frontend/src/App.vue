<script setup lang="ts">
import { onMounted } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";
import { useAuthStore } from "./stores/auth";

const auth = useAuthStore();
const route = useRoute();

onMounted(async () => {
  if (auth.token) {
    await auth.fetchMe();
  }
});
</script>

<template>
  <div class="app-shell">
    <header class="topbar">
      <RouterLink to="/" class="brand">
        <span class="brand-icon" aria-hidden="true"></span>
        Сенсорный навигатор
      </RouterLink>
      <nav class="nav">
        <RouterLink to="/" :class="{ active: route.name === 'map' }">Карта</RouterLink>
        <RouterLink to="/favorites" v-if="auth.isAuthenticated">Избранное</RouterLink>
        <RouterLink to="/profile" v-if="auth.isAuthenticated">Профиль</RouterLink>
        <template v-else>
          <RouterLink to="/login">Вход</RouterLink>
          <RouterLink to="/register">Регистрация</RouterLink>
        </template>
      </nav>
    </header>

    <main class="main">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.app-shell {
  display: grid;
  grid-template-rows: 56px 1fr;
  height: 100vh;
  background: var(--bg);
  color: var(--fg);
}
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  border-bottom: 1px solid var(--border);
  background: var(--surface);
}
.brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  font-size: 16px;
  color: var(--fg);
  text-decoration: none;
}
.brand-icon {
  width: 22px;
  height: 22px;
  border-radius: 6px;
  background: linear-gradient(135deg, #4caf50, #009688);
}
.nav {
  display: inline-flex;
  gap: 6px;
  align-items: center;
}
.nav a {
  color: var(--fg-muted);
  text-decoration: none;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 14px;
}
.nav a:hover {
  color: var(--fg);
  background: var(--surface-2);
}
.nav a.active,
.nav a.router-link-exact-active {
  color: var(--fg);
  background: var(--surface-2);
}
.main {
  overflow: hidden;
}
</style>
