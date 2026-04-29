export interface Category {
  id: number;
  name: string;
  slug: string;
  description?: string;
  icon?: string;
}

export interface User {
  id: number;
  email: string;
  username: string;
  display_name?: string;
  avatar_url?: string;
  noise_pref: number;
  light_pref: number;
  crowd_pref: number;
}

export interface Place {
  id: number;
  name: string;
  address?: string;
  description?: string;
  category_id: number;
  category?: Category;
  latitude: number;
  longitude: number;
  avg_noise: number;
  avg_light: number;
  avg_crowd: number;
  avg_smell: number;
  avg_visual: number;
  overall_avg: number;
  reviews_count: number;
  created_at?: string;
}

export interface Review {
  id: number;
  place_id: number;
  user_id: number;
  user?: User;
  text: string;
  noise: number;
  light: number;
  crowd: number;
  smell: number;
  visual: number;
  created_at: string;
}

export interface PlaceFilter {
  q?: string;
  category_id?: number;
  noise_max?: number;
  light_max?: number;
  crowd_max?: number;
  smell_max?: number;
  visual_max?: number;
}

export interface AuthResponse {
  token: string;
  expires_at: string;
  user: User;
}
