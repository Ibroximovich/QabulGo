'use client';

import { useLayoutEffect, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/providers/AuthProvider';

const useIsomorphicLayoutEffect = typeof window !== 'undefined' ? useLayoutEffect : useEffect;

/**
 * Automatically redirects authenticated staff/admin users from the public home page
 * directly to their respective portals (Clinic Dashboard, Doctor Dashboard, Admin).
 */
export function HomeRoleRedirect() {
  const { user, isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  useIsomorphicLayoutEffect(() => {
    // 1. Instant check from localStorage to avoid any render delay
    if (typeof window !== 'undefined') {
      try {
        const rawUser = localStorage.getItem('qabulgo_user');
        const token = localStorage.getItem('qabulgo_access_token');
        if (rawUser && token) {
          const parsed = JSON.parse(rawUser);
          const role = parsed?.role;
          if (role === 'clinic_owner' || role === 'clinic_admin') {
            router.replace('/clinic/dashboard');
            return;
          }
          if (role === 'doctor') {
            router.replace('/doctor/dashboard');
            return;
          }
          if (role === 'super_admin') {
            router.replace('/admin/clinics');
            return;
          }
        }
      } catch {
        // ignore JSON parse errors
      }
    }

    // 2. React state check when auth provider finishes initializing
    if (!isLoading && isAuthenticated && user?.role) {
      if (user.role === 'clinic_owner' || user.role === 'clinic_admin') {
        router.replace('/clinic/dashboard');
      } else if (user.role === 'doctor') {
        router.replace('/doctor/dashboard');
      } else if (user.role === 'super_admin') {
        router.replace('/admin/clinics');
      }
    }
  }, [user, isAuthenticated, isLoading, router]);

  return null;
}

