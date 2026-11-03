import { ForgotPasswordPage } from '@/features/auth/pages/ForgotPasswordPage';

export const metadata = {
  title: 'Parolni tiklash — QabulGo',
  description: 'Parolni tiklash uchun emailingizga kod yuborish',
  robots: {
    index: false,
    follow: false,
  },
};

export default function Page() {
  return <ForgotPasswordPage />;
}

