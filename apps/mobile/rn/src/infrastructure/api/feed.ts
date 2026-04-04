import {apiClient} from './client';

export interface PostResponse {
  id: string;
  user_id: string;
  user_name: string;
  content: string;
  like_count: number;
  liked: boolean;
  created_at: string;
}

export interface FeedResponse {
  posts: PostResponse[];
  next_cursor: string | null;
}

export function getFeed(cursor?: string, limit = 20) {
  const params = new URLSearchParams();
  if (cursor) {
    params.set('cursor', cursor);
  }
  params.set('limit', String(limit));

  return apiClient.request<FeedResponse>({
    method: 'GET',
    path: `/api/v1/posts?${params.toString()}`,
  });
}

export function getPost(id: string) {
  return apiClient.request<PostResponse>({
    method: 'GET',
    path: `/api/v1/posts/${id}`,
  });
}

export function createPost(content: string) {
  return apiClient.request<PostResponse>({
    method: 'POST',
    path: '/api/v1/posts',
    body: {content},
  });
}

export function updatePost(id: string, content: string) {
  return apiClient.request<PostResponse>({
    method: 'PUT',
    path: `/api/v1/posts/${id}`,
    body: {content},
  });
}

export function deletePost(id: string) {
  return apiClient.request<void>({
    method: 'DELETE',
    path: `/api/v1/posts/${id}`,
  });
}

export function likePost(id: string) {
  return apiClient.request<void>({
    method: 'POST',
    path: `/api/v1/posts/${id}/like`,
  });
}

export function unlikePost(id: string) {
  return apiClient.request<void>({
    method: 'DELETE',
    path: `/api/v1/posts/${id}/like`,
  });
}
