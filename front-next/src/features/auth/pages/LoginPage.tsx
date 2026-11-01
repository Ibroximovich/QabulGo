'use client';

import { useEffect, Suspense } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Mail, Lock } from 'lucide-react';
import { toast } from 'sonner';
import { AuthLayout } from '@/components/layout/AuthLayout';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Alert } from '@/components/ui/Alert';
import { loginSchema, LoginFormValues } from '../schemas/authSchemas';
import { useLogin } from '../hooks/useLogin';

function LoginForm() {
  const searchParams = useSearchParams();
  const isVerified = searchParams?.get('verified') === 'true';

  const { mutate: loginUser, isPending } = useLogin();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: '',
      password: '',
    },
    mode: 'onBlur',
  });

  useEffect(() => {
    if (isVerified) {
      toast.success('Emailingiz tasdiqlandi! Endi tizimga kirishingiz mumkin.');
    }
  }, [isVerified]);

  const onSubmit = (values: LoginFormValues) => {
    loginUser({
      email: values.email.trim().toLowerCase(),
      password: values.password,
    });
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
      {isVerified && (
        <Alert variant="success" title="Email muvaffaqiyatli tasdiqlandi!">
          Profilingiz faollashtirildi. Iltimos, parolingizni kiritib kabinetga kiring.
        </Alert>
      )}

      {/* Email */}
      <Input
        id="login-email"
        label="Email manzili"
        type="email"
        placeholder="ali@example.com"
        leftIcon={<Mail className="h-4 w-4" />}
        error={errors.email?.message}
        autoComplete="email"
        {...register('email')}
      />

      {/* Parol */}
      <div>
        <div className="flex items-center justify-between mb-1.5">
          <label htmlFor="login-password" className="text-sm font-medium text-brand-navy">
            Parol
          </label>
          <Link
            href="/forgot-password"
            className="text-xs font-semibold text-brand-blue hover:underline"
          >
            Parolni unutdingizmi?
          </Link>
        </div>
        <Input
          id="login-password"
          isPassword
          placeholder="••••••••"
          leftIcon={<Lock className="h-4 w-4" />}
          error={errors.password?.message}
          autoComplete="current-password"
          {...register('password')}
        />
      </div>

      {/* Submit tugmasi */}
      <div className="pt-2">
        <Button
          type="submit"
          variant="primary"
          size="lg"
          fullWidth
          isLoading={isPending}
          loadingText="Kirilmoqda..."
        >
          Kirish
        </Button>
      </div>

      {/* Ro‘yxatdan o‘tish havolasi */}
      <p className="text-center text-sm text-brand-muted pt-2">
        Hali hisobingiz yo‘qmi?{' '}
        <Link
          href="/register"
          className="text-brand-blue font-bold hover:underline inline-flex items-center gap-1"
        >
          Ro‘yxatdan o‘tish
        </Link>
      </p>

      {/* Klinika portali havolasi */}
      <div className="pt-3 border-t border-slate-100 text-center">
        <p className="text-xs text-slate-500">
          Klinika egasimisiz?{' '}
          <Link href="/clinic/login" className="text-brand-blue font-semibold hover:underline">
            Klinika portaliga kirish
          </Link>
        </p>
        <p className="text-xs text-slate-400 mt-1.5">
          <Link href="/admin/login" className="hover:text-slate-600 hover:underline transition-colors">
            Admin panelga o'tish
          </Link>
        </p>
      </div>
    </form>
  );
}

export function LoginPage() {
  return (
    <AuthLayout
      title="Tizimga kirish"
      subtitle="Shaxsiy kabinetingizga kirish uchun login ma’lumotlaringizni kiriting"
    >
      <Suspense fallback={<div className="h-48 animate-pulse bg-slate-50 rounded-xl" />}>
        <LoginForm />
      </Suspense>
    </AuthLayout>
  );
}
