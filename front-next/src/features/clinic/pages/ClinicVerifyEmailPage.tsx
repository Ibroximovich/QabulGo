'use client';

import { useState, useEffect, Suspense } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { toast } from 'sonner';
import { Mail, RefreshCw } from 'lucide-react';
import { AuthLayout } from '@/components/layout/AuthLayout';
import { OtpInput } from '@/components/ui/OtpInput';
import { Button } from '@/components/ui/Button';
import { Alert } from '@/components/ui/Alert';
import { authApi } from '@/features/auth/api/authApi';
import { useAuth } from '@/providers/AuthProvider';

function ClinicVerifyEmailForm() {
  const searchParams = useSearchParams();
  const emailParam = searchParams?.get('email') || '';
  const router = useRouter();
  const { login } = useAuth();

  const [otp, setOtp] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isResending, setIsResending] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [countdown, setCountdown] = useState(60);

  useEffect(() => {
    if (countdown <= 0) return;
    const timer = setInterval(() => {
      setCountdown((prev) => prev - 1);
    }, 1000);
    return () => clearInterval(timer);
  }, [countdown]);

  const handleVerify = async (codeToVerify?: string) => {
    const code = codeToVerify || otp;
    if (code.length < 6) {
      setErrorMessage('6 xonali tasdiqlash kodini to‘liq kiriting');
      return;
    }

    if (!emailParam) {
      setErrorMessage('Email manzili topilmadi. Iltimos, qayta ro‘yxatdan o‘ting.');
      return;
    }

    setIsSubmitting(true);
    setErrorMessage(null);

    try {
      const res = await authApi.verifyEmail({
        email: emailParam,
        code,
      });

      const data = res.data;
      if (data?.access_token && data?.user) {
        // Authenticate immediately
        login(data.access_token, data.user);
        toast.success('Email tasdiqlandi! Klinika ma’lumotlarini kiritishga xush kelibsiz.');
        router.replace('/clinic/onboarding');
      } else {
        toast.success('Emailingiz tasdiqlandi. Iltimos, tizimga kiring.');
        router.replace('/clinic/login');
      }
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      const msg =
        axiosErr.response?.data?.message ||
        'Kiritilgan tasdiqlash kodi noto‘g‘ri yoki muddati tugagan.';
      setErrorMessage(msg);
      toast.error(msg);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleOtpChange = (newOtp: string) => {
    setOtp(newOtp);
    if (errorMessage) setErrorMessage(null);
    if (newOtp.length === 6) {
      handleVerify(newOtp);
    }
  };

  const handleResend = async () => {
    if (countdown > 0 || !emailParam) return;
    setIsResending(true);
    setErrorMessage(null);

    try {
      await authApi.resendVerification({ email: emailParam });
      toast.success('Yangi tasdiqlash kodi emailingizga yuborildi');
      setCountdown(60);
      setOtp('');
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      const msg =
        axiosErr.response?.data?.message || 'Kodni qayta yuborishda xatolik yuz berdi.';
      setErrorMessage(msg);
      toast.error(msg);
    } finally {
      setIsResending(false);
    }
  };

  return (
    <div className="space-y-6">
      {errorMessage && (
        <Alert variant="error" title="Xatolik" message={errorMessage} />
      )}

      <div className="p-3.5 bg-blue-50/70 border border-blue-100 rounded-xl flex items-center gap-3 text-sm text-blue-900">
        <Mail className="w-5 h-5 text-[#0866F5] shrink-0" />
        <div className="min-w-0">
          <span className="text-xs text-blue-600 block">Kod yuborilgan email:</span>
          <span className="font-semibold truncate block">
            {emailParam || 'Belgilanmagan'}
          </span>
        </div>
      </div>

      <div className="flex flex-col items-center justify-center space-y-4">
        <label className="text-xs font-semibold text-slate-700">
          6 xonali tasdiqlash kodi
        </label>
        <OtpInput
          value={otp}
          onChange={handleOtpChange}
          disabled={isSubmitting}
          hasError={Boolean(errorMessage)}
        />
      </div>

      <Button
        type="button"
        variant="primary"
        size="lg"
        fullWidth
        disabled={otp.length < 6}
        isLoading={isSubmitting}
        onClick={() => handleVerify()}
        className="bg-[#0866F5] hover:bg-blue-600 shadow-md"
      >
        Tasdiqlash va davom etish
      </Button>

      <div className="flex flex-col items-center gap-2 pt-2 text-center text-xs text-slate-600">
        {countdown > 0 ? (
          <p>
            Yangi kod olish mumkin:{' '}
            <span className="font-semibold text-slate-900">{countdown} soniya</span>
          </p>
        ) : (
          <button
            type="button"
            onClick={handleResend}
            disabled={isResending}
            className="inline-flex items-center gap-1.5 font-semibold text-[#0866F5] hover:underline"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isResending ? 'animate-spin' : ''}`} />
            Kodni qayta yuborish
          </button>
        )}

        <Link
          href="/clinic/register"
          className="text-slate-400 hover:text-slate-600 hover:underline pt-2"
        >
          Boshqa email bilan ro‘yxatdan o‘tish
        </Link>
      </div>
    </div>
  );
}

export function ClinicVerifyEmailPage() {
  return (
    <AuthLayout
      title="Emailni tasdiqlash"
      subtitle="Klinikani faollashtirish uchun emailingizga yuborilgan 6 xonali kodni kiriting"
      badgeText="Klinika tasdiqlash"
    >
      <Suspense fallback={<div className="h-48 animate-pulse bg-slate-50 rounded-xl" />}>
        <ClinicVerifyEmailForm />
      </Suspense>
    </AuthLayout>
  );
}
