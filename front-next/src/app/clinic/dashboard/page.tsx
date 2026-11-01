import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ClinicLayout } from '@/components/layout/clinic/ClinicLayout';
import { ClinicDashboardPage } from '@/features/clinic/pages/ClinicDashboardPage';

export const metadata = {
  title: 'Klinika boshqaruv paneli — QabulGo',
  description: 'Klinika boshqaruv paneli va statistika',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['clinic_owner', 'clinic_admin']}>
      <ClinicLayout>
        <ClinicDashboardPage />
      </ClinicLayout>
    </ProtectedRoute>
  );
}

