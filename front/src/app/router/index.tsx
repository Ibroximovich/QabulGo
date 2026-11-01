import { lazy, Suspense } from 'react';
import { createBrowserRouter, Navigate } from 'react-router-dom';
import { PublicRoute } from './PublicRoute';
import { ProtectedRoute } from './ProtectedRoute';
import { LoadingSpinner } from '@/components/common/LoadingSpinner';
import { ClinicLayout } from '@/components/layout/clinic/ClinicLayout';
import { AdminLayout } from '@/components/layout/admin/AdminLayout';

// Patient Auth & Pages
const LoginPage = lazy(() =>
  import('@/features/auth/pages/LoginPage').then((m) => ({ default: m.LoginPage }))
);
const RegisterPage = lazy(() =>
  import('@/features/auth/pages/RegisterPage').then((m) => ({ default: m.RegisterPage }))
);
const VerifyEmailPage = lazy(() =>
  import('@/features/auth/pages/VerifyEmailPage').then((m) => ({ default: m.VerifyEmailPage }))
);
const ForgotPasswordPage = lazy(() =>
  import('@/features/auth/pages/ForgotPasswordPage').then((m) => ({
    default: m.ForgotPasswordPage,
  }))
);
const ResetPasswordPage = lazy(() =>
  import('@/features/auth/pages/ResetPasswordPage').then((m) => ({
    default: m.ResetPasswordPage,
  }))
);
const DashboardPage = lazy(() =>
  import('@/features/patient/pages/DashboardPage').then((m) => ({ default: m.DashboardPage }))
);
const ProfilePage = lazy(() =>
  import('@/features/patient/pages/ProfilePage').then((m) => ({ default: m.ProfilePage }))
);

// Clinic Auth & Pages
const ClinicRegisterPage = lazy(() =>
  import('@/features/clinic/pages/ClinicRegisterPage').then((m) => ({
    default: m.ClinicRegisterPage,
  }))
);
const ClinicVerifyEmailPage = lazy(() =>
  import('@/features/clinic/pages/ClinicVerifyEmailPage').then((m) => ({
    default: m.ClinicVerifyEmailPage,
  }))
);
const ClinicLoginPage = lazy(() =>
  import('@/features/clinic/pages/ClinicLoginPage').then((m) => ({
    default: m.ClinicLoginPage,
  }))
);
const ClinicOnboardingPage = lazy(() =>
  import('@/features/clinic/pages/ClinicOnboardingPage').then((m) => ({
    default: m.ClinicOnboardingPage,
  }))
);
const ClinicDashboardPage = lazy(() =>
  import('@/features/clinic/pages/ClinicDashboardPage').then((m) => ({
    default: m.ClinicDashboardPage,
  }))
);
const ClinicProfilePage = lazy(() =>
  import('@/features/clinic/pages/ClinicProfilePage').then((m) => ({
    default: m.ClinicProfilePage,
  }))
);
const DoctorListPage = lazy(() =>
  import('@/features/clinic/pages/DoctorListPage').then((m) => ({
    default: m.DoctorListPage,
  }))
);
const DoctorCreatePage = lazy(() =>
  import('@/features/clinic/pages/DoctorCreatePage').then((m) => ({
    default: m.DoctorCreatePage,
  }))
);
const DoctorEditPage = lazy(() =>
  import('@/features/clinic/pages/DoctorEditPage').then((m) => ({
    default: m.DoctorEditPage,
  }))
);

// Admin Auth & Pages
const AdminLoginPage = lazy(() =>
  import('@/features/admin/pages/AdminLoginPage').then((m) => ({
    default: m.AdminLoginPage,
  }))
);
const AdminClinicsPage = lazy(() =>
  import('@/features/admin/pages/AdminClinicsPage').then((m) => ({
    default: m.AdminClinicsPage,
  }))
);

// Common
const NotFoundPage = lazy(() =>
  import('@/features/common/pages/NotFoundPage').then((m) => ({ default: m.NotFoundPage }))
);

function SuspenseWrapper({ children }: { children: React.ReactNode }) {
  return <Suspense fallback={<LoadingSpinner size="full" />}>{children}</Suspense>;
}

export const router = createBrowserRouter([
  // Root Redirect
  {
    path: '/',
    element: <Navigate to="/dashboard" replace />,
  },

  // 1. Public Guest Routes (Patient Auth)
  {
    element: <PublicRoute />,
    children: [
      {
        path: '/login',
        element: (
          <SuspenseWrapper>
            <LoginPage />
          </SuspenseWrapper>
        ),
      },
      {
        path: '/register',
        element: (
          <SuspenseWrapper>
            <RegisterPage />
          </SuspenseWrapper>
        ),
      },
      {
        path: '/verify-email',
        element: (
          <SuspenseWrapper>
            <VerifyEmailPage />
          </SuspenseWrapper>
        ),
      },
      {
        path: '/forgot-password',
        element: (
          <SuspenseWrapper>
            <ForgotPasswordPage />
          </SuspenseWrapper>
        ),
      },
      {
        path: '/reset-password',
        element: (
          <SuspenseWrapper>
            <ResetPasswordPage />
          </SuspenseWrapper>
        ),
      },
    ],
  },

  // 2. Public Guest Routes (Clinic Auth)
  {
    element: <PublicRoute />,
    children: [
      {
        path: '/clinic/register',
        element: (
          <SuspenseWrapper>
            <ClinicRegisterPage />
          </SuspenseWrapper>
        ),
      },
      {
        path: '/clinic/login',
        element: (
          <SuspenseWrapper>
            <ClinicLoginPage />
          </SuspenseWrapper>
        ),
      },
      {
        path: '/clinic/verify-email',
        element: (
          <SuspenseWrapper>
            <ClinicVerifyEmailPage />
          </SuspenseWrapper>
        ),
      },
    ],
  },

  // 3. Public Guest Routes (Superadmin Auth)
  {
    element: <PublicRoute />,
    children: [
      {
        path: '/admin/login',
        element: (
          <SuspenseWrapper>
            <AdminLoginPage />
          </SuspenseWrapper>
        ),
      },
    ],
  },

  // 4. Protected Patient Routes
  {
    element: <ProtectedRoute allowedRoles={['patient']} />,
    children: [
      {
        path: '/dashboard',
        element: (
          <SuspenseWrapper>
            <DashboardPage />
          </SuspenseWrapper>
        ),
      },
      {
        path: '/profile',
        element: (
          <SuspenseWrapper>
            <ProfilePage />
          </SuspenseWrapper>
        ),
      },
    ],
  },

  // 5. Protected Clinic Routes (Clinic Owner / Admin)
  {
    element: <ProtectedRoute allowedRoles={['clinic_owner', 'clinic_admin']} />,
    children: [
      {
        path: '/clinic/onboarding',
        element: (
          <SuspenseWrapper>
            <ClinicOnboardingPage />
          </SuspenseWrapper>
        ),
      },
      {
        element: <ClinicLayout />,
        children: [
          {
            path: '/clinic/dashboard',
            element: (
              <SuspenseWrapper>
                <ClinicDashboardPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/clinic/profile',
            element: (
              <SuspenseWrapper>
                <ClinicProfilePage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/clinic/doctors',
            element: (
              <SuspenseWrapper>
                <DoctorListPage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/clinic/doctors/new',
            element: (
              <SuspenseWrapper>
                <DoctorCreatePage />
              </SuspenseWrapper>
            ),
          },
          {
            path: '/clinic/doctors/:doctorId/edit',
            element: (
              <SuspenseWrapper>
                <DoctorEditPage />
              </SuspenseWrapper>
            ),
          },
        ],
      },
    ],
  },

  // 6. Protected Superadmin Routes
  {
    element: <ProtectedRoute allowedRoles={['super_admin']} />,
    children: [
      {
        element: <AdminLayout />,
        children: [
          {
            path: '/admin/clinics',
            element: (
              <SuspenseWrapper>
                <AdminClinicsPage />
              </SuspenseWrapper>
            ),
          },
        ],
      },
    ],
  },

  // 7. 404 Catch All
  {
    path: '*',
    element: (
      <SuspenseWrapper>
        <NotFoundPage />
      </SuspenseWrapper>
    ),
  },
]);
