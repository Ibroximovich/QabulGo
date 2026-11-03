import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { AdminLayout } from '@/components/layout/admin/AdminLayout';
import { AdminPatientsPage } from '@/features/admin/pages/AdminPatientsPage';

export const metadata = {
  title: 'Bemorlar — QabulGo Superadmin',
  description: 'QabulGo platformasidagi barcha bemorlar va ularning faolligi',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['super_admin']}>
      <AdminLayout>
        <AdminPatientsPage />
      </AdminLayout>
    </ProtectedRoute>
  );
}

