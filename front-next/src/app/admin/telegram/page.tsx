import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { AdminLayout } from '@/components/layout/admin/AdminLayout';
import { AdminTelegramPage } from '@/features/admin/pages/AdminTelegramPage';

export const metadata = {
  title: 'Telegram Bot — QabulGo Superadmin',
  description: 'Superadmin Telegram boti sozlamalari va arizalar monitoringi',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['super_admin']}>
      <AdminLayout>
        <AdminTelegramPage />
      </AdminLayout>
    </ProtectedRoute>
  );
}

