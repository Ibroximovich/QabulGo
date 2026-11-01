import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '@/app/providers/AuthProvider';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';

interface ProtectedRouteProps {
  allowedRoles?: string[];
}

export function ProtectedRoute({ allowedRoles }: ProtectedRouteProps) {
  const { user, isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return <LoadingSpinner size="full" label="Yuklanmoqda..." />;
  }

  if (!isAuthenticated) {
    if (allowedRoles?.includes('super_admin')) {
      return <Navigate to="/admin/login" state={{ from: location }} replace />;
    }
    if (allowedRoles?.includes('clinic_owner') || allowedRoles?.includes('clinic_admin')) {
      return <Navigate to="/clinic/login" state={{ from: location }} replace />;
    }
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (allowedRoles && allowedRoles.length > 0) {
    const currentRole = user?.role || 'patient';
    if (!allowedRoles.includes(currentRole)) {
      // Role not authorized — redirect to user's primary portal
      if (currentRole === 'super_admin') {
        return <Navigate to="/admin/clinics" replace />;
      }
      if (currentRole === 'clinic_owner' || currentRole === 'clinic_admin') {
        return <Navigate to="/clinic/dashboard" replace />;
      }
      return <Navigate to="/dashboard" replace />;
    }
  }

  return <Outlet />;
}
