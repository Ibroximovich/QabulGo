import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ClinicLayout } from '@/components/layout/clinic/ClinicLayout';
import { ClinicDiagnosticsPage } from '@/features/diagnostics/pages/ClinicDiagnosticsPage';

export const metadata = {
  title: 'Diagnostika boshqaruvi — QabulGo',
  description: 'Klinika diagnostika uskunalari, xizmatlari va bronlarini boshqarish',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['clinic_owner', 'clinic_admin']}>
      <ClinicLayout>
        <ClinicDiagnosticsPage />
      </ClinicLayout>
    </ProtectedRoute>
  );
}

