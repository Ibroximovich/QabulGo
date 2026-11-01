import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { AdminLayout } from '@/components/layout/admin/AdminLayout';
import { AdminClinicsPage } from '@/features/admin/pages/AdminClinicsPage';

export const metadata = {
  title: 'Klinikalar nazorati — QabulGo Superadmin',
  description: 'Klinikalar arizalarini ko‘rib chiqish, tasdiqlash va rad etish',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['super_admin']}>
      <AdminLayout>
        <AdminClinicsPage />
      </AdminLayout>
    </ProtectedRoute>
  );
}

