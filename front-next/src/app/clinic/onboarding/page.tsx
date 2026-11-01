import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ClinicOnboardingPage } from '@/features/clinic/pages/ClinicOnboardingPage';

export const metadata = {
  title: 'Klinika ma’lumotlarini to‘ldirish — QabulGo',
  description: 'Klinikangiz ma’lumotlarini to‘ldirib tasdiqlashga yuboring',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['clinic_owner', 'clinic_admin']}>
      <ClinicOnboardingPage />
    </ProtectedRoute>
  );
}

