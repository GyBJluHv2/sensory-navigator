import type {
  AuthResponse,
  Category,
  Place,
  PlaceFilter,
  Review,
  User,
} from "../types/models";

const API_BASE =
  (import.meta.env.VITE_API_BASE as string) || "http://localhost:8080";

class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
    this.name = "ApiError";
  }
}

let currentToken: string | null = localStorage.getItem("sn_token");

export function setToken(token: string | null): void {
  currentToken = token;
  if (token) localStorage.setItem("sn_token", token);
  else localStorage.removeItem("sn_token");
}

export function getToken(): string | null {
  return currentToken;
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  authRequired = false,
): Promise<T> {
  const headers: Record<string, string> = {
    Accept: "application/json",
    ...((options.headers as Record<string, string>) || {}),
  };
  if (options.body && !(options.body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }
  if (authRequired || currentToken) {
    if (!currentToken) {
      throw new ApiError("требуется авторизация", 401);
    }
    headers["Authorization"] = `Bearer ${currentToken}`;
  }
  const res = await fetch(API_BASE + path, { ...options, headers });
  if (res.status === 204) {
    return undefined as unknown as T;
  }
  const text = await res.text();
  const data = text ? JSON.parse(text) : ({} as Record<string, unknown>);
  if (!res.ok) {
    const msg = (data as { error?: string }).error || res.statusText;
    throw new ApiError(msg, res.status);
  }
  return data as T;
}

// === Auth ===

export const auth = {
  // Старый одношаговый эндпоинт регистрации сохранён для обратной совместимости
  // (используется CI/seed). Новый flow — два шага: requestRegister + confirmRegister.
  register: (input: {
    email: string;
    username: string;
    password: string;
    display_name?: string;
  }) =>
    request<AuthResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  // Шаг 1: проверка email-формата + отправка 6-значного кода на почту.
  requestRegister: (input: {
    email: string;
    username: string;
    password: string;
    display_name?: string;
  }) =>
    request<{ status: string; expires_at: string }>("/api/auth/register-request", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  // Шаг 2: подтверждение кода → создание пользователя → выдача JWT.
  confirmRegister: (input: { email: string; code: string }) =>
    request<AuthResponse>("/api/auth/register-confirm", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  // Повторная отправка кода на тот же email.
  resendCode: (input: { email: string }) =>
    request<{ status: string; expires_at: string }>("/api/auth/resend-code", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  login: (input: { email: string; password: string }) =>
    request<AuthResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify(input),
    }),
  me: () => request<User>("/api/me", {}, true),
  updateMe: (input: Partial<User>) =>
    request<User>("/api/me", {
      method: "PUT",
      body: JSON.stringify(input),
    }, true),
  changePassword: (old_password: string, new_password: string) =>
    request<void>("/api/me/password", {
      method: "PUT",
      body: JSON.stringify({ old_password, new_password }),
    }, true),
};

// === Places ===

export const places = {
  list: (filter: PlaceFilter = {}) => {
    const params = new URLSearchParams();
    if (filter.q) params.set("q", filter.q);
    if (filter.category_id) params.set("category_id", String(filter.category_id));
    if (filter.noise_max) params.set("noise_max", String(filter.noise_max));
    if (filter.light_max) params.set("light_max", String(filter.light_max));
    if (filter.crowd_max) params.set("crowd_max", String(filter.crowd_max));
    if (filter.smell_max) params.set("smell_max", String(filter.smell_max));
    if (filter.visual_max) params.set("visual_max", String(filter.visual_max));
    const qs = params.toString();
    return request<{ items: Place[]; count: number }>(
      "/api/places" + (qs ? "?" + qs : ""),
    );
  },
  search: (q: string) =>
    request<{ items: Place[] }>(
      "/api/places/search?q=" + encodeURIComponent(q),
    ),
  nearby: (lat: number, lon: number, radius = 1000) =>
    request<{ items: Place[] }>(
      `/api/places/nearby?lat=${lat}&lon=${lon}&radius=${radius}`,
    ),
  get: (id: number) => request<Place>(`/api/places/${id}`),
  create: (input: {
    name: string;
    address?: string;
    description?: string;
    category_id: number;
    latitude: number;
    longitude: number;
  }) =>
    request<Place>("/api/places", {
      method: "POST",
      body: JSON.stringify(input),
    }, true),
  update: (id: number, input: Partial<Place>) =>
    request<Place>(`/api/places/${id}`, {
      method: "PUT",
      body: JSON.stringify(input),
    }, true),
  remove: (id: number) =>
    request<void>(`/api/places/${id}`, { method: "DELETE" }, true),
};

// === Categories ===

export const categories = {
  list: () => request<{ items: Category[] }>("/api/categories"),
};

// === Reviews ===

export const reviews = {
  byPlace: (placeId: number) =>
    request<{ items: Review[] }>(`/api/places/${placeId}/reviews`),
  create: (
    placeId: number,
    input: {
      text?: string;
      noise: number;
      light: number;
      crowd: number;
      smell: number;
      visual: number;
    },
  ) =>
    request<Review>(`/api/places/${placeId}/reviews`, {
      method: "POST",
      body: JSON.stringify(input),
    }, true),
  update: (
    id: number,
    input: {
      text?: string;
      noise: number;
      light: number;
      crowd: number;
      smell: number;
      visual: number;
    },
  ) =>
    request<Review>(`/api/reviews/${id}`, {
      method: "PUT",
      body: JSON.stringify(input),
    }, true),
  remove: (id: number) =>
    request<void>(`/api/reviews/${id}`, { method: "DELETE" }, true),
  myReviews: () => request<{ items: Review[] }>(`/api/reviews/me`, {}, true),
};

// === Favorites ===

export const favorites = {
  list: () => request<{ items: Place[] }>("/api/favorites", {}, true),
  add: (placeId: number) =>
    request<void>(`/api/places/${placeId}/favorite`, {
      method: "POST",
    }, true),
  remove: (placeId: number) =>
    request<void>(`/api/places/${placeId}/favorite`, {
      method: "DELETE",
    }, true),
};

export { ApiError };
