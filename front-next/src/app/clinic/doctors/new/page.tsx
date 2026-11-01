import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ClinicLayout } from '@/components/layout/clinic/ClinicLayout';
import { DoctorCreatePage } from '@/features/clinic/pages/DoctorCreatePage';

export const metadata = {
  title: 'Yangi shifokor qo‘shish — QabulGo',
  description: 'Klinikaga yangi shifokor biriktirish',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['clinic_owner', 'clinic_admin']}>
      <ClinicLayout>
        <DoctorCreatePage />
      </ClinicLayout>
    </ProtectedRoute>
  );
}

