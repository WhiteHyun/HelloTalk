import {getTokens, saveTokens, clearTokens} from '../storage/token';

const BASE_URL = 'http://localhost:8080';

type Method = 'GET' | 'POST' | 'PUT' | 'DELETE';

interface RequestOptions {
  method: Method;
  path: string;
  body?: unknown;
  auth?: boolean;
  multipart?: FormData;
}

interface ApiError {
  code: string;
  message: string;
}

export class ApiClient {
  private onUnauthorized?: () => void;

  setOnUnauthorized(callback: () => void) {
    this.onUnauthorized = callback;
  }

  async request<T>(options: RequestOptions): Promise<T> {
    const {method, path, body, auth = true, multipart} = options;

    const headers: Record<string, string> = {};

    if (auth) {
      const tokens = await getTokens();
      if (tokens) {
        headers['Authorization'] = `Bearer ${tokens.accessToken}`;
      }
    }

    if (multipart) {
      // Content-Type은 fetch가 자동으로 multipart/form-data로 설정
    } else if (body) {
      headers['Content-Type'] = 'application/json';
    }

    const response = await fetch(`${BASE_URL}${path}`, {
      method,
      headers,
      body: multipart ? multipart : body ? JSON.stringify(body) : undefined,
    });

    // 204 No Content
    if (response.status === 204) {
      return undefined as T;
    }

    // 401 → 토큰 갱신 시도
    if (response.status === 401 && auth) {
      const refreshed = await this.tryRefresh();
      if (refreshed) {
        return this.request({...options});
      }
      this.onUnauthorized?.();
      throw new ApiRequestError(401, 'UNAUTHORIZED', '인증이 만료되었습니다');
    }

    const data = await response.json();

    if (!response.ok) {
      const error = data as ApiError;
      throw new ApiRequestError(response.status, error.code, error.message);
    }

    return data as T;
  }

  private async tryRefresh(): Promise<boolean> {
    try {
      const tokens = await getTokens();
      if (!tokens?.refreshToken) {
        return false;
      }

      const response = await fetch(`${BASE_URL}/api/v1/auth/refresh`, {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({refresh_token: tokens.refreshToken}),
      });

      if (!response.ok) {
        await clearTokens();
        return false;
      }

      const data = await response.json();
      await saveTokens({
        accessToken: data.token.access_token,
        refreshToken: data.token.refresh_token,
      });
      return true;
    } catch {
      await clearTokens();
      return false;
    }
  }
}

export class ApiRequestError extends Error {
  constructor(
    public status: number,
    public code: string,
    public userMessage: string,
  ) {
    super(userMessage);
  }
}

export const apiClient = new ApiClient();
