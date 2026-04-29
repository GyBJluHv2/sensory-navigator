import { createRouter, createWebHashHistory } from "vue-router";
import MapView from "../views/MapView.vue";
import LoginView from "../views/LoginView.vue";
import RegisterView from "../views/RegisterView.vue";
import VerifyEmailView from "../views/VerifyEmailView.vue";
import PlaceView from "../views/PlaceView.vue";
import ProfileView from "../views/ProfileView.vue";
import FavoritesView from "../views/FavoritesView.vue";
import AddPlaceView from "../views/AddPlaceView.vue";
import { useAuthStore } from "../stores/auth";

export const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/", name: "map", component: MapView },
    { path: "/login", name: "login", component: LoginView },
    { path: "/register", name: "register", component: RegisterView },
    { path: "/verify-email", name: "verifyEmail", component: VerifyEmailView },
    { path: "/place/:id", name: "place", component: PlaceView, props: true },
    {
      path: "/profile",
      name: "profile",
      component: ProfileView,
      meta: { auth: true },
    },
    {
      path: "/favorites",
      name: "favorites",
      component: FavoritesView,
      meta: { auth: true },
    },
    {
      path: "/places/new",
      name: "addPlace",
      component: AddPlaceView,
      meta: { auth: true },
    },
  ],
});

router.beforeEach((to) => {
  if (to.meta.auth) {
    const auth = useAuthStore();
    if (!auth.isAuthenticated) {
      return { name: "login", query: { redirect: to.fullPath } };
    }
  }
  return true;
});
