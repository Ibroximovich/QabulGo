import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ClinicLayout } from '@/components/layout/clinic/ClinicLayout';
import { DoctorListPage } from '@/features/clinic/pages/DoctorListPage';

export const metadata = {
  title: 'Shifokorlar ro‘yxati — QabulGo',
  description: 'Klinika shifokorlarini ko‘rish va boshqarish',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['clinic_owner', 'clinic_admin']}>
      <ClinicLayout>
        <DoctorListPage />
      </ClinicLayout>
    </ProtectedRoute>
  );
}

