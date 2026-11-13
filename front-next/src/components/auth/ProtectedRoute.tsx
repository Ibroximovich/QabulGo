'use client';

import React, { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/providers/AuthProvider';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';

export interface ProtectedRouteProps {
  allowedRoles?: string[];
  children: React.ReactNode;
}

export function ProtectedRoute({ allowedRoles, children }: ProtectedRouteProps) {
  const { user, isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (isLoading) return;

    if (!isAuthenticated) {
      if (allowedRoles?.includes('super_admin')) {
        router.replace('/admin/login');
        return;
      }
      if (allowedRoles?.includes('clinic_owner') || allowedRoles?.includes('clinic_admin')) {
        router.replace('/clinic/login');
        return;
      }
      router.replace('/login');
      return;
    }

    // Wait until user object and its role are fully resolved!
    if (!user || !user.role) {
      return;
    }

    if (allowedRoles && allowedRoles.length > 0) {
      const currentRole = user.role;
      if (!allowedRoles.includes(currentRole)) {
        if (currentRole === 'super_admin') {
          router.replace('/admin/clinics');
          return;
        }
        if (currentRole === 'clinic_owner' || currentRole === 'clinic_admin') {
          router.replace('/clinic/dashboard');
          return;
        }
        if (currentRole === 'doctor') {
          router.replace('/doctor/dashboard');
          return;
        }
        router.replace('/dashboard');
        return;
      }
    }
  }, [isLoading, isAuthenticated, user, allowedRoles, router]);

  if (isLoading || !isAuthenticated || !user || !user.role) {
    return <LoadingSpinner size="full" label="Yuklanmoqda..." />;
  }

  if (allowedRoles && allowedRoles.length > 0) {
    const currentRole = user.role;
    if (!allowedRoles.includes(currentRole)) {
      return <LoadingSpinner size="full" label="Yuklanmoqda..." />;
    }
  }

  return <>{children}</>;
}
