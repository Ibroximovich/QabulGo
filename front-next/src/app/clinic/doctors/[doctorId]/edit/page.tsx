import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ClinicLayout } from '@/components/layout/clinic/ClinicLayout';
import { DoctorEditPage } from '@/features/clinic/pages/DoctorEditPage';

export const metadata = {
  title: 'Shifokor ma’lumotlarini tahrirlash — QabulGo',
  description: 'Shifokor profili va narxlarini yangilash',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['clinic_owner', 'clinic_admin']}>
      <ClinicLayout>
        <DoctorEditPage />
      </ClinicLayout>
    </ProtectedRoute>
  );
}

