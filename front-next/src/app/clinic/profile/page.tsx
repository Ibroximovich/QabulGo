import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ClinicLayout } from '@/components/layout/clinic/ClinicLayout';
import { ClinicProfilePage } from '@/features/clinic/pages/ClinicProfilePage';

export const metadata = {
  title: 'Klinika profili — QabulGo',
  description: 'Klinika asosiy ma’lumotlari, ish vaqtlari va rasmlarini tahrirlash',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['clinic_owner', 'clinic_admin']}>
      <ClinicLayout>
        <ClinicProfilePage />
      </ClinicLayout>
    </ProtectedRoute>
  );
}

