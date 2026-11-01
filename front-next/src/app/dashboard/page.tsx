import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { DashboardPage } from '@/features/patient/pages/DashboardPage';

export const metadata = {
  title: 'Bemor kabineti — QabulGo',
  description: 'Bemorning shaxsiy boshqaruv kabineti',
};

export default function Page() {
  return (
    <ProtectedRoute allowedRoles={['patient']}>
      <DashboardPage />
    </ProtectedRoute>
  );
}

