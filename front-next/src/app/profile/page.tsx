import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { ProfilePage } from '@/features/patient/pages/ProfilePage';

export const metadata = {
  title: 'Profil ma’lumotlari — QabulGo',
  description: 'Bemor profil ma’lumotlarini ko‘rish va yangilash',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['patient']}>
      <ProfilePage />
    </ProtectedRoute>
  );
}

