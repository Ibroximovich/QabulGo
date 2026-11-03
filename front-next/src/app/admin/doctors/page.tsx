import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { AdminLayout } from '@/components/layout/admin/AdminLayout';
import { AdminDoctorsPage } from '@/features/admin/pages/AdminDoctorsPage';

export const metadata = {
  title: 'Shifokorlar — QabulGo Superadmin',
  description: 'QabulGo platformasidagi barcha shifokorlar va ularning faoliyati',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['super_admin']}>
      <AdminLayout>
        <AdminDoctorsPage />
      </AdminLayout>
    </ProtectedRoute>
  );
}

