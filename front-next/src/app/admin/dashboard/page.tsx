import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { AdminLayout } from '@/components/layout/admin/AdminLayout';
import { AdminDashboardPage } from '@/features/admin/pages/AdminDashboardPage';

export const metadata = {
  title: 'Dashboard — QabulGo Superadmin',
  description: 'QabulGo platformasi asosiy ko‘rsatkichlari, tahlillari va statistikasi',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['super_admin']}>
      <AdminLayout>
        <AdminDashboardPage />
      </AdminLayout>
    </ProtectedRoute>
  );
}

