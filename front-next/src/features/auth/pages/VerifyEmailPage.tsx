'use client';

import { useState, useEffect, Suspense } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { Mail, RefreshCw, ArrowLeft } from 'lucide-react';
import { AuthLayout } from '@/components/layout/AuthLayout';
import { OtpInput } from '@/components/ui/OtpInput';
import { Button } from '@/components/ui/Button';
import { Alert } from '@/components/ui/Alert';
import { useVerifyEmail } from '../hooks/useVerifyEmail';

function VerifyEmailForm() {
  const searchParams = useSearchParams();
  const email = searchParams?.get('email') || '';

  const [code, setCode] = useState('');
  const [resendCooldown, setResendCooldown] = useState(60);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const { verifyEmail, isVerifying, resendVerification, isResending } = useVerifyEmail();

  // Countdown timer for resend
  useEffect(() => {
    if (resendCooldown <= 0) return;
    const interval = setInterval(() => {
      setResendCooldown((prev) => prev - 1);
    }, 1000);
    return () => clearInterval(interval);
  }, [resendCooldown]);

  const handleVerify = async (otpCode: string) => {
    if (otpCode.length !== 6) {
      setErrorMessage('Iltimos, 6 xonali tasdiqlash kodini to‘liq kiriting');
      return;
    }
    setErrorMessage(null);
    try {
      await verifyEmail({ email, code: otpCode });
    } catch {
      // Error handled by hook toast
    }
  };

  const handleOtpChange = (newVal: string) => {
    setCode(newVal);
    setErrorMessage(null);
    // Auto submit when 6 digits are reached
    if (newVal.length === 6) {
      handleVerify(newVal);
    }
  };

  const handleResend = async () => {
    if (resendCooldown > 0 || isResending) return;
    try {
      await resendVerification({ email });
      setResendCooldown(60);
    } catch {
      // Handled by hook
    }
  };

  return (
    <div className="space-y-6">
      {/* Email Badge */}
      <div className="flex items-center justify-center p-3 rounded-xl bg-brand-blue-50/70 border border-brand-blue-100 text-center">
        <div className="flex items-center gap-2 text-xs sm:text-sm text-brand-navy">
          <Mail className="h-4 w-4 text-brand-blue shrink-0" />
          <span className="text-brand-muted">Kod yuborildi:</span>
          <span className="font-bold text-brand-navy truncate max-w-[200px] sm:max-w-xs">
            {email || 'Sizning emailingiz'}
          </span>
        </div>
      </div>

      {errorMessage && (
        <Alert variant="error" onClose={() => setErrorMessage(null)}>
          {errorMessage}
        </Alert>
      )}

      {/* OTP Input Fields */}
      <div className="space-y-2">
        <label className="block text-center text-xs font-semibold uppercase tracking-wider text-brand-muted">
          6 xonali tasdiqlash kodi
        </label>
        <OtpInput
          value={code}
          onChange={handleOtpChange}
          disabled={isVerifying}
          hasError={Boolean(errorMessage)}
          autoFocus
        />
        <p className="text-center text-xs text-brand-muted mt-2">
          Kod 10 daqiqa davomida amal qiladi
        </p>
      </div>

      {/* Submit Button */}
      <Button
        type="button"
        variant="primary"
        size="lg"
        fullWidth
        disabled={code.length !== 6}
        isLoading={isVerifying}
        loadingText="Tasdiqlanmoqda..."
        onClick={() => handleVerify(code)}
      >
        Emailni tasdiqlash
      </Button>

      {/* Resend Code Section */}
      <div className="text-center space-y-2 pt-2 border-t border-slate-100">
        <p className="text-xs text-brand-muted">Kod kelmadimi?</p>
        <button
          type="button"
          disabled={resendCooldown > 0 || isResending}
          onClick={handleResend}
          className="inline-flex items-center gap-1.5 text-xs font-bold text-brand-blue hover:text-brand-blue-700 disabled:text-slate-400 disabled:cursor-not-allowed transition-colors"
        >
          <RefreshCw className={`h-3.5 w-3.5 ${isResending ? 'animate-spin' : ''}`} />
          {resendCooldown > 0
            ? `Kodni qayta yuborish (${resendCooldown}s)`
            : 'Yangi kod yuborish'}
        </button>
      </div>

      {/* Back link */}
      <div className="text-center pt-2">
        <Link
          href="/login"
          className="inline-flex items-center gap-1.5 text-xs text-brand-muted hover:text-brand-navy transition-colors font-medium"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          <span>Login sahifasiga qaytish</span>
        </Link>
      </div>
    </div>
  );
}

export function VerifyEmailPage() {
  return (
    <AuthLayout
      title="Emailni tasdiqlang"
      subtitle="Ro‘yxatdan o‘tishni yakunlash uchun pochtangizga yuborilgan 6 xonali kodni kiriting"
    >
      <Suspense fallback={<div className="h-48 animate-pulse bg-slate-50 rounded-xl" />}>
        <VerifyEmailForm />
      </Suspense>
    </AuthLayout>
  );
}
