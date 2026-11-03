import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { AdminLayout } from '@/components/layout/admin/AdminLayout';
import { AdminAppointmentsPage } from '@/features/admin/pages/AdminAppointmentsPage';

export const metadata = {
  title: 'Barcha bronlar — QabulGo Superadmin',
  description: 'QabulGo platformasidagi barcha bronlar monitoringi va nazorati',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['super_admin']}>
      <AdminLayout>
        <AdminAppointmentsPage />
      </AdminLayout>
    </ProtectedRoute>
  );
}

