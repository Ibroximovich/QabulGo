import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { AppointmentDetailPage } from '@/features/appointments/pages/AppointmentDetailPage';

export const metadata = {
  title: 'Qabul tafsilotlari — QabulGo',
  description: 'Shifokor qabuli tafsilotlari va holati',
};

export default function AppointmentPage() {
  return (
    <ProtectedRoute allowedRoles={['patient']}>
      <AppointmentDetailPage />
    </ProtectedRoute>
  );
}
