import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ClinicLayout } from '@/components/layout/clinic/ClinicLayout';
import { ClinicAppointmentDetailPage } from '@/features/clinic/pages/ClinicAppointmentDetailPage';

export const metadata = {
  title: 'Bron tafsilotlari — QabulGo Klinika',
  description: 'Bemor bronining to‘liq tafsilotlari',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['clinic_owner', 'clinic_admin', 'super_admin']}>
      <ClinicLayout>
        <ClinicAppointmentDetailPage />
      </ClinicLayout>
    </ProtectedRoute>
  );
}
