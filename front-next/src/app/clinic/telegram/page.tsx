import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ClinicLayout } from '@/components/layout/clinic/ClinicLayout';
import { ClinicTelegramPage } from '@/features/clinic/pages/ClinicTelegramPage';

export const metadata = {
  title: 'Telegram bot — QabulGo Klinika',
  description: 'Telegram bot orqali yangi bronlar va xabarnomalarni sozlash',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['clinic_owner', 'clinic_admin']}>
      <ClinicLayout>
        <ClinicTelegramPage />
      </ClinicLayout>
    </ProtectedRoute>
  );
}

