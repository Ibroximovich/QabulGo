import { RegisterPage } from '@/features/auth/pages/RegisterPage';

export const metadata = {
  title: 'Ro‘yxatdan o‘tish — QabulGo',
  description: 'Shifokor qabuliga qulay va onlayn yozilish uchun shaxsiy profilingizni yarating',
  robots: {
    index: false,
    follow: false,
  },
};

export default function Page() {
  return <RegisterPage />;
}

