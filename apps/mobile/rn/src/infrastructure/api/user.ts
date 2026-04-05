import {apiClient} from './client';
import {UserResponse} from './auth';

export function getMe() {
  return apiClient.request<UserResponse>({
    method: 'GET',
    path: '/api/v1/users/me',
  });
}

export function updateMe(name: string) {
  return apiClient.request<UserResponse>({
    method: 'PUT',
    path: '/api/v1/users/me',
    body: {name},
  });
}

export function uploadAvatar(uri: string) {
  const formData = new FormData();
  formData.append('avatar', {
    uri,
    type: 'image/jpeg',
    name: 'avatar.jpg',
  } as unknown as Blob);

  return apiClient.request<UserResponse>({
    method: 'POST',
    path: '/api/v1/users/me/avatar',
    multipart: formData,
  });
}

export function getUser(id: string) {
  return apiClient.request<UserResponse>({
    method: 'GET',
    path: `/api/v1/users/${id}`,
  });
}
