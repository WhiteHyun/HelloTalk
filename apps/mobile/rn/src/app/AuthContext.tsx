import React, {createContext, useContext, useEffect, useReducer} from 'react';
import {UserResponse} from '../infrastructure/api/auth';
import * as authApi from '../infrastructure/api/auth';
import * as userApi from '../infrastructure/api/user';
import {saveTokens, getTokens, clearTokens} from '../infrastructure/storage/token';
import {apiClient} from '../infrastructure/api/client';

interface AuthState {
  isLoading: boolean;
  user: UserResponse | null;
}

type AuthAction =
  | {type: 'LOADING'}
  | {type: 'SIGNED_IN'; user: UserResponse}
  | {type: 'SIGNED_OUT'};

interface AuthContextType extends AuthState {
  signUp: (email: string, password: string, name: string) => Promise<void>;
  signIn: (email: string, password: string) => Promise<void>;
  signOut: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'LOADING':
      return {...state, isLoading: true};
    case 'SIGNED_IN':
      return {isLoading: false, user: action.user};
    case 'SIGNED_OUT':
      return {isLoading: false, user: null};
  }
}

export function AuthProvider({children}: {children: React.ReactNode}) {
  const [state, dispatch] = useReducer(authReducer, {
    isLoading: true,
    user: null,
  });

  useEffect(() => {
    // 앱 시작 시 저장된 토큰으로 유저 정보 복원
    async function restore() {
      try {
        const tokens = await getTokens();
        if (tokens) {
          const user = await userApi.getMe();
          dispatch({type: 'SIGNED_IN', user});
          return;
        }
      } catch {
        await clearTokens();
      }
      dispatch({type: 'SIGNED_OUT'});
    }
    restore();

    // 토큰 만료 시 로그아웃
    apiClient.setOnUnauthorized(() => {
      dispatch({type: 'SIGNED_OUT'});
    });
  }, []);

  const authContext: AuthContextType = {
    ...state,
    signUp: async (email, password, name) => {
      const result = await authApi.signup(email, password, name);
      await saveTokens({
        accessToken: result.token.access_token,
        refreshToken: result.token.refresh_token,
      });
      dispatch({type: 'SIGNED_IN', user: result.user});
    },
    signIn: async (email, password) => {
      const result = await authApi.login(email, password);
      await saveTokens({
        accessToken: result.token.access_token,
        refreshToken: result.token.refresh_token,
      });
      dispatch({type: 'SIGNED_IN', user: result.user});
    },
    signOut: async () => {
      try {
        const tokens = await getTokens();
        if (tokens) {
          await authApi.logout(tokens.refreshToken);
        }
      } catch {
        // 로그아웃 API 실패해도 로컬은 정리
      }
      await clearTokens();
      dispatch({type: 'SIGNED_OUT'});
    },
  };

  return (
    <AuthContext.Provider value={authContext}>{children}</AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
