import { LoginPage } from '@/features/auth/pages/LoginPage';

export const metadata = {
  title: 'Tizimga kirish — QabulGo',
  description: 'Shaxsiy kabinetingizga kirish uchun login ma’lumotlaringizni kiriting',
  robots: {
    index: false,
    follow: false,
  },
};

export default function Page() {
  return <LoginPage />;
}

