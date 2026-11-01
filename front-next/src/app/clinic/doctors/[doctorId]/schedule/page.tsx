import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ClinicLayout } from '@/components/layout/clinic/ClinicLayout';
import { DoctorSchedulePage } from '@/features/clinic/pages/DoctorSchedulePage';

export const metadata = {
  title: 'Shifokor qabul jadvali — QabulGo',
  description: 'Shifokor qabul vaqtlarini avtomatik va qo‘lda boshqarish',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['clinic_owner', 'clinic_admin', 'super_admin']}>
      <ClinicLayout>
        <DoctorSchedulePage />
      </ClinicLayout>
    </ProtectedRoute>
  );
}

