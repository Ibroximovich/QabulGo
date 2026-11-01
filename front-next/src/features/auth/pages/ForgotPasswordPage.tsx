'use client';

import Link from 'next/link';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Mail, ArrowLeft, KeyRound } from 'lucide-react';
import { AuthLayout } from '@/components/layout/AuthLayout';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { forgotPasswordSchema, ForgotPasswordFormValues } from '../schemas/authSchemas';
import { useForgotPassword } from '../hooks/useForgotPassword';

export function ForgotPasswordPage() {
  const { mutate: sendCode, isPending } = useForgotPassword();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ForgotPasswordFormValues>({
    resolver: zodResolver(forgotPasswordSchema),
    defaultValues: { email: '' },
    mode: 'onBlur',
  });

  const onSubmit = (values: ForgotPasswordFormValues) => {
    sendCode({ email: values.email.trim().toLowerCase() });
  };

  return (
    <AuthLayout
      title="Parolni tiklash"
      subtitle="Email manzilingizni kiriting. Sizga parolni tiklash uchun 6 xonali tasdiqlash kodi yuboriladi"
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        <Input
          label="Email manzili"
          type="email"
          placeholder="ali@example.com"
          leftIcon={<Mail className="h-4 w-4" />}
          error={errors.email?.message}
          helperText="Hisobingiz ro‘yxatdan o‘tgan email manzilini kiriting"
          {...register('email')}
        />

        <div className="pt-2">
          <Button
            type="submit"
            variant="primary"
            size="lg"
            fullWidth
            isLoading={isPending}
            loadingText="Yuborilmoqda..."
            rightIcon={<KeyRound className="h-4 w-4" />}
          >
            Tiklash kodini yuborish
          </Button>
        </div>

        <div className="text-center pt-2">
          <Link
            href="/login"
            className="inline-flex items-center gap-1.5 text-xs text-brand-muted hover:text-brand-navy transition-colors font-medium"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            <span>Login sahifasiga qaytish</span>
          </Link>
        </div>
      </form>
    </AuthLayout>
  );
}
