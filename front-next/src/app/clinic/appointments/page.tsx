import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ClinicLayout } from '@/components/layout/clinic/ClinicLayout';
import { ClinicAppointmentsPage } from '@/features/clinic/pages/ClinicAppointmentsPage';

export const metadata = {
  title: 'Qabul bronlari — QabulGo Klinika',
  description: 'Klinikadagi bemorlar qabuli va bronlar ro‘yxati',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['clinic_owner', 'clinic_admin', 'super_admin']}>
      <ClinicLayout>
        <ClinicAppointmentsPage />
      </ClinicLayout>
    </ProtectedRoute>
  );
}
