import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { MyAppointmentsPage } from '@/features/appointments/pages/MyAppointmentsPage';

export const metadata = {
  title: 'Mening bronlarim — QabulGo',
  description: 'Bemorning barcha shifokor qabullari va bronlari',
};

export default function AppointmentsPage() {
  return (
    <ProtectedRoute allowedRoles={['patient']}>
      <MyAppointmentsPage />
    </ProtectedRoute>
  );
}
