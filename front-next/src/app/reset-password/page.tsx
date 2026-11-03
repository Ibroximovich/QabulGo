import { ResetPasswordPage } from '@/features/auth/pages/ResetPasswordPage';

export const metadata = {
  title: 'Yangi parol o‘rnatish — QabulGo',
  description: 'Emailingizga yuborilgan tasdiqlash kodi bilan yangi parol o‘rnating',
  robots: {
    index: false,
    follow: false,
  },
};

export default function Page() {
  return <ResetPasswordPage />;
}

