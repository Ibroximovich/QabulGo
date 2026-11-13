'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import axios from 'axios';
import { toast } from 'sonner';
import { UserSummary } from '@/types/api';
import { authApi } from '@/features/auth/api/authApi';
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
      // 1. Check active super admin in sessionStorage
      const sessionCached = sessionStorage.getItem('qabulgo_user');
      if (sessionCached) {
        const parsed = JSON.parse(sessionCached);
        if (parsed.role === 'super_admin') return parsed;
      }

      // 2. Check localStorage for persistent users (patient, clinic, doctor)
      const localCached = localStorage.getItem('qabulgo_user');
      if (localCached) {
        const parsed = JSON.parse(localCached);
        if (parsed.role !== 'super_admin') {
          // If token exists, make sure role strictly reflects the signed JWT claims
          const token = tokenStorage.getAccessToken();
          if (token) {
            const claims = parseJwtRoleAndId(token);
            if (claims?.role) {
              parsed.role = claims.role;
            }
          }
          return parsed;
        }
      }
      return null;
    } catch {
      return null;
    }
  });

  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(() => {
    if (typeof window === 'undefined') return false;
    return tokenStorage.hasToken();
  });

  const [isLoading, setIsLoading] = useState<boolean>(true);

  const login = useCallback((token: string, userData: UserSummary) => {
    tokenStorage.setAccessToken(token, userData.role);
    setUser(userData);
    setIsAuthenticated(true);
    if (typeof window !== 'undefined') {
      try {
        if (userData.role === 'super_admin') {
          sessionStorage.setItem('qabulgo_user', JSON.stringify(userData));
          localStorage.removeItem('qabulgo_user');
        } else {
          localStorage.setItem('qabulgo_user', JSON.stringify(userData));
          sessionStorage.removeItem('qabulgo_user');
        }
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
          sessionStorage.removeItem('qabulgo_user');
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
          if (updated.role === 'super_admin') {
            sessionStorage.setItem('qabulgo_user', JSON.stringify(updated));
          } else {
            localStorage.setItem('qabulgo_user', JSON.stringify(updated));
          }
        } catch {
          // Ignore
        }
      }
      return updated;
    });
  }, []);

  const refetchUser = useCallback(async () => {
    try {
      const res = await authApi.getMe();
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
          if (freshUser.role === 'super_admin') {
            sessionStorage.setItem('qabulgo_user', JSON.stringify(freshUser));
          } else {
            localStorage.setItem('qabulgo_user', JSON.stringify(freshUser));
          }
        }
      }
    } catch (err) {
      console.error('Failed to refetch user profile', err);
    }
  }, []);

  // Initial Auth Check / Silent Refresh
  useEffect(() => {
    let isMounted = true;

    async function initAuth() {
      // 1. Check existing token from tokenStorage
      const existingToken = tokenStorage.getAccessToken();
      const claims = existingToken ? parseJwtRoleAndId(existingToken) : null;
      const isSuperAdmin = claims?.role === 'super_admin';

      if (existingToken && claims?.user_id) {
        if (isSuperAdmin) {
          const sessionUserStr = typeof window !== 'undefined' ? sessionStorage.getItem('qabulgo_user') : null;
          if (sessionUserStr) {
            try {
              const sessionUser = JSON.parse(sessionUserStr);
              if (sessionUser.id === claims.user_id) {
                if (claims.role) {
                  sessionUser.role = claims.role;
                }
                if (isMounted) {
                  setUser(sessionUser);
                  setIsAuthenticated(true);
                  setIsLoading(false);
                }
                return;
              }
            } catch {}
          }
        } else {
          // Patient, Clinic, Doctor: restore session immediately from localStorage!
          const localUserStr = typeof window !== 'undefined' ? localStorage.getItem('qabulgo_user') : null;
          if (localUserStr) {
            try {
              const localUser = JSON.parse(localUserStr);
              if (localUser.id === claims.user_id) {
                // Ground truth: strictly preserve role from signed token
                if (claims.role) {
                  localUser.role = claims.role;
                }
                if (isMounted) {
                  setUser(localUser);
                  setIsAuthenticated(true);
                  setIsLoading(false);
                }

                // Verify fresh profile with backend in background via GET /api/v1/auth/me
                authApi.getMe().then((res) => {
                  if (res.data && isMounted) {
                    const freshUser = res.data;
                    setUser(freshUser);
                    localStorage.setItem('qabulgo_user', JSON.stringify(freshUser));
                  }
                }).catch(() => {
                  // Silent refresh fallback if needed
                  authApi.refreshToken().then((refreshRes) => {
                    const refreshedToken = refreshRes.data?.access_token;
                    if (refreshedToken && isMounted) {
                      const refreshedUser = refreshRes.data?.user || localUser;
                      tokenStorage.setAccessToken(refreshedToken, refreshedUser.role);
                      setUser(refreshedUser);
                      localStorage.setItem('qabulgo_user', JSON.stringify(refreshedUser));
                    }
                  }).catch(() => {});
                });
                return;
              }
            } catch {}
          }
        }
      }

      // 2. If no valid stored session or needing silent refresh from cookie:
      try {
        const res = await authApi.refreshToken();
        const newToken = res.data?.access_token;

        if (newToken && isMounted) {
          const newClaims = parseJwtRoleAndId(newToken);
          const role = res.data?.user?.role || newClaims?.role;

          // If role is super_admin and this is a fresh browser opening without session token, do NOT auto-login!
          if (role === 'super_admin' && !sessionStorage.getItem('qabulgo_admin_access_token')) {
            tokenStorage.clear();
            setUser(null);
            setIsAuthenticated(false);
            return;
          }

          tokenStorage.setAccessToken(newToken, role);

          // If refreshToken returned user directly:
          if (res.data?.user) {
            const freshUser = res.data.user;
            setUser(freshUser);
            setIsAuthenticated(true);
            if (role === 'super_admin') {
              sessionStorage.setItem('qabulgo_user', JSON.stringify(freshUser));
            } else {
              localStorage.setItem('qabulgo_user', JSON.stringify(freshUser));
            }
            return;
          }

          // Otherwise fetch via authApi.getMe()
          try {
            const meRes = await authApi.getMe();
            if (meRes.data && isMounted) {
              const freshUser = meRes.data;
              setUser(freshUser);
              setIsAuthenticated(true);
              if (role === 'super_admin') {
                sessionStorage.setItem('qabulgo_user', JSON.stringify(freshUser));
              } else {
                localStorage.setItem('qabulgo_user', JSON.stringify(freshUser));
              }
              return;
            }
          } catch {}

          // Fallback to cached storage
          const storage = role === 'super_admin' ? sessionStorage : localStorage;
          const cachedStr = typeof window !== 'undefined' ? storage.getItem('qabulgo_user') : null;
          if (cachedStr) {
            try {
              const cached = JSON.parse(cachedStr);
              if (newClaims?.user_id && cached.id === newClaims.user_id) {
                if (role) cached.role = role;
                setUser(cached);
                setIsAuthenticated(true);
                return;
              }
            } catch {
              // Ignore
            }
          }

          // Minimal fallback user from JWT
          if (newClaims?.user_id) {
            const fallbackUser: UserSummary = {
              id: newClaims.user_id,
              first_name: '',
              last_name: '',
              email: '',
              phone: '',
              role: role || 'patient',
            };
            setUser(fallbackUser);
            setIsAuthenticated(true);
            if (role !== 'super_admin') {
              localStorage.setItem('qabulgo_user', JSON.stringify(fallbackUser));
            } else {
              sessionStorage.setItem('qabulgo_user', JSON.stringify(fallbackUser));
            }
          }
        }
      } catch (err: unknown) {
        const isAxiosErr = axios.isAxiosError(err);
        const status = isAxiosErr ? err.response?.status : undefined;

        // Only clear session if server explicitly returns 401 Unauthorized
        if (status === 401) {
          if (isMounted) {
            tokenStorage.clear();
            setUser(null);
            setIsAuthenticated(false);
          }
        } else {
          // Temporary network failure: keep cached non-superadmin user
          if (isMounted) {
            const cachedStr = typeof window !== 'undefined' ? localStorage.getItem('qabulgo_user') : null;
            if (cachedStr) {
              try {
                const cached = JSON.parse(cachedStr);
                if (cached.role !== 'super_admin') {
                  setUser(cached);
                  setIsAuthenticated(true);
                }
              } catch {
                tokenStorage.clear();
                setUser(null);
                setIsAuthenticated(false);
              }
            }
          }
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
