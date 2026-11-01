'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import { toast } from 'sonner';
import { UserSummary } from '@/types/api';
import { authApi } from '@/features/auth/api/authApi';
import { patientApi } from '@/features/patient/api/patientApi';
import { tokenStorage } from '@/lib/api/tokenStorage';
import { queryClient } from './QueryProvider';

interface AuthContextType {
  user: UserSummary | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (token: string, user: UserSummary) => void;
  logout: () => Promise<void>;
  updateUser: (user: Partial<UserSummary>) => void;
  refetchUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

function parseJwtRoleAndId(token: string): { user_id?: string; role?: string } | null {
  try {
    const base64Url = token.split('.')[1];
    if (!base64Url) return null;
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );
    return JSON.parse(jsonPayload);
  } catch {
    return null;
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<UserSummary | null>(() => {
    if (typeof window === 'undefined') return null;
    try {
      const cached = localStorage.getItem('qabulgo_user');
      return cached ? JSON.parse(cached) : null;
    } catch {
      return null;
    }
  });
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  const login = useCallback((token: string, userData: UserSummary) => {
    tokenStorage.setAccessToken(token);
    setUser(userData);
    setIsAuthenticated(true);
    if (typeof window !== 'undefined') {
      try {
        localStorage.setItem('qabulgo_user', JSON.stringify(userData));
      } catch {
        // Ignore storage errors
      }
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } catch {
      // Ignore network errors on logout
    } finally {
      tokenStorage.clear();
      setUser(null);
      setIsAuthenticated(false);
      if (typeof window !== 'undefined') {
        try {
          localStorage.removeItem('qabulgo_user');
        } catch {
          // Ignore storage errors
        }
      }
      queryClient.clear();
      toast.success('Kabinetdan muvaffaqiyatli chiqildi');
    }
  }, []);

  const updateUser = useCallback((updatedFields: Partial<UserSummary>) => {
    setUser((prev) => {
      if (!prev) return null;
      const updated = { ...prev, ...updatedFields };
      if (typeof window !== 'undefined') {
        try {
          localStorage.setItem('qabulgo_user', JSON.stringify(updated));
        } catch {
          // Ignore
        }
      }
      return updated;
    });
  }, []);

  const refetchUser = useCallback(async () => {
    if (!user || user.role === 'patient') {
      try {
        const res = await patientApi.getProfile();
        if (res.data) {
          const freshUser: UserSummary = {
            id: res.data.id,
            first_name: res.data.first_name,
            last_name: res.data.last_name,
            email: res.data.email,
            phone: res.data.phone,
            role: res.data.role,
          };
          setUser(freshUser);
          if (typeof window !== 'undefined') {
            localStorage.setItem('qabulgo_user', JSON.stringify(freshUser));
          }
        }
      } catch (err) {
        console.error('Failed to refetch patient profile', err);
      }
    }
  }, [user]);

  // Initial Silent Refresh
  useEffect(() => {
    let isMounted = true;

    async function initAuth() {
      try {
        const res = await authApi.refreshToken();
        const newToken = res.data?.access_token;

        if (newToken && isMounted) {
          tokenStorage.setAccessToken(newToken);
          const claims = parseJwtRoleAndId(newToken);
          const role = claims?.role;

          if (role === 'patient') {
            try {
              const profileRes = await patientApi.getProfile();
              if (profileRes.data && isMounted) {
                const freshUser: UserSummary = {
                  id: profileRes.data.id,
                  first_name: profileRes.data.first_name,
                  last_name: profileRes.data.last_name,
                  email: profileRes.data.email,
                  phone: profileRes.data.phone,
                  role: profileRes.data.role,
                };
                setUser(freshUser);
                setIsAuthenticated(true);
                localStorage.setItem('qabulgo_user', JSON.stringify(freshUser));
                return;
              }
            } catch {
              // Fallback to cached
            }
          }

          // For clinic owners, clinic admins, super admins, or if profile fetch failed
          const cachedStr = typeof window !== 'undefined' ? localStorage.getItem('qabulgo_user') : null;
          if (cachedStr) {
            try {
              const cached = JSON.parse(cachedStr);
              if (claims?.user_id && cached.id === claims.user_id) {
                setUser(cached);
                setIsAuthenticated(true);
                return;
              }
            } catch {
              // Ignore
            }
          }

          // Minimal fallback user from JWT
          if (claims?.user_id) {
            const fallbackUser: UserSummary = {
              id: claims.user_id,
              first_name: '',
              last_name: '',
              email: '',
              phone: '',
              role: claims.role || 'user',
            };
            setUser(fallbackUser);
            setIsAuthenticated(true);
          }
        }
      } catch {
        // Not authenticated or refresh token expired
        if (isMounted) {
          tokenStorage.clear();
          setUser(null);
          setIsAuthenticated(false);
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }

    initAuth();

    const handleSessionExpired = () => {
      setUser(null);
      setIsAuthenticated(false);
      tokenStorage.clear();
      try {
        localStorage.removeItem('qabulgo_user');
      } catch {
        // Ignore
      }
      queryClient.clear();
      toast.error('Sessiyangiz muddati tugadi. Iltimos, qayta kiring.');
    };

    window.addEventListener('auth:session-expired', handleSessionExpired);

    return () => {
      isMounted = false;
      window.removeEventListener('auth:session-expired', handleSessionExpired);
    };
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated,
        isLoading,
        login,
        logout,
        updateUser,
        refetchUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

