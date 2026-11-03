import { VerifyEmailPage } from '@/features/auth/pages/VerifyEmailPage';

export const metadata = {
  title: 'Emailni tasdiqlash — QabulGo',
  description: 'Emailingizga yuborilgan 6 xonali tasdiqlash kodini kiriting',
  robots: {
    index: false,
    follow: false,
  },
};

export default function Page() {
  return <VerifyEmailPage />;
}

