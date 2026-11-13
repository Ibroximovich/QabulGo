'use client';

import React, { createContext, useContext, useEffect, useState, useCallback } from 'react';
import axios from 'axios';
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
      // 1. Check sessionStorage for super admin active in current session
      const sessionCached = sessionStorage.getItem('qabulgo_user');
      if (sessionCached) {
        const parsed = JSON.parse(sessionCached);
        if (parsed.role === 'super_admin') return parsed;
      }
      // 2. Check localStorage for persistent 30-day users
      const localCached = localStorage.getItem('qabulgo_user');
      if (localCached) {
        const parsed = JSON.parse(localCached);
        if (parsed.role !== 'super_admin') return parsed;
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
                if (isMounted) {
                  setUser(localUser);
                  setIsAuthenticated(true);
                  setIsLoading(false);
                }

                // Silently refresh in background to keep tokens fresh
                authApi.refreshToken().then((res) => {
                  const refreshedToken = res.data?.access_token;
                  if (refreshedToken && isMounted) {
                    tokenStorage.setAccessToken(refreshedToken, localUser.role);
                  }
                }).catch(() => {
                  // If refresh fails due to network/offline, keep session active since token is valid for 30 days
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
          const role = newClaims?.role;

          // If role is super_admin and this is a fresh browser opening without session token, do NOT auto-login!
          if (role === 'super_admin' && !sessionStorage.getItem('qabulgo_admin_access_token')) {
            tokenStorage.clear();
            setUser(null);
            setIsAuthenticated(false);
            return;
          }

          tokenStorage.setAccessToken(newToken, role);

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

          // For clinic owners, clinic admins, doctors, or superadmin
          const storage = role === 'super_admin' ? sessionStorage : localStorage;
          const cachedStr = typeof window !== 'undefined' ? storage.getItem('qabulgo_user') : null;
          if (cachedStr) {
            try {
              const cached = JSON.parse(cachedStr);
              if (newClaims?.user_id && cached.id === newClaims.user_id) {
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
              role: newClaims.role || 'user',
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
