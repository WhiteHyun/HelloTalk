import {apiClient} from './client';

export interface UserResponse {
  id: string;
  email: string;
  name: string;
  avatar_url: string | null;
  created_at: string;
}

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface AuthResponse {
  user: UserResponse;
  token: TokenResponse;
}

export function signup(email: string, password: string, name: string) {
  return apiClient.request<AuthResponse>({
    method: 'POST',
    path: '/api/v1/auth/signup',
    body: {email, password, name},
    auth: false,
  });
}

export function login(email: string, password: string) {
  return apiClient.request<AuthResponse>({
    method: 'POST',
    path: '/api/v1/auth/login',
    body: {email, password},
    auth: false,
  });
}

export function logout(refreshToken: string) {
  return apiClient.request<void>({
    method: 'POST',
    path: '/api/v1/auth/logout',
    body: {refresh_token: refreshToken},
  });
}
