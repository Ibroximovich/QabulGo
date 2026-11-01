import { useSearchParams, Link } from 'react-router-dom';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Mail, Lock, ArrowLeft } from 'lucide-react';
import { AuthLayout } from '@/components/layout/AuthLayout';
import { Input } from '@/components/ui/Input';
import { OtpInput } from '@/components/ui/OtpInput';
import { Button } from '@/components/ui/Button';
import { resetPasswordSchema, ResetPasswordFormValues } from '../schemas/authSchemas';
import { useResetPassword } from '../hooks/useResetPassword';

export function ResetPasswordPage() {
  const [searchParams] = useSearchParams();
  const emailParam = searchParams.get('email') || '';

  const { mutate: resetPassword, isPending } = useResetPassword();

  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<ResetPasswordFormValues>({
    resolver: zodResolver(resetPasswordSchema),
    defaultValues: {
      email: emailParam,
      code: '',
      new_password: '',
      confirm_password: '',
    },
    mode: 'onBlur',
  });

  const onSubmit = (values: ResetPasswordFormValues) => {
    resetPassword({
      email: values.email.trim().toLowerCase(),
      code: values.code.trim(),
      new_password: values.new_password,
      confirm_password: values.confirm_password,
    });
  };

  return (
    <AuthLayout
      title="Yangi parol o‘rnatish"
      subtitle="Emailingizga yuborilgan 6 xonali kodni va yangi xavfsiz parolingizni kiriting"
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {/* Email */}
        <Input
          label="Email manzili"
          type="email"
          placeholder="ali@example.com"
          leftIcon={<Mail className="h-4 w-4" />}
          error={errors.email?.message}
          {...register('email')}
        />

        {/* 6-digit OTP */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium text-brand-navy">
            6 xonali tasdiqlash kodi
          </label>
          <Controller
            control={control}
            name="code"
            render={({ field }) => (
              <OtpInput
                value={field.value}
                onChange={field.onChange}
                hasError={Boolean(errors.code)}
              />
            )}
          />
          {errors.code && (
            <p className="text-xs text-brand-error font-medium">{errors.code.message}</p>
          )}
        </div>

        {/* Yangi parol */}
        <Input
          label="Yangi parol"
          isPassword
          placeholder="••••••••"
          leftIcon={<Lock className="h-4 w-4" />}
          error={errors.new_password?.message}
          helperText="Kamida 8 ta belgi, harf va raqam"
          {...register('new_password')}
        />

        {/* Parolni tasdiqlash */}
        <Input
          label="Yangi parolni tasdiqlang"
          isPassword
          placeholder="••••••••"
          leftIcon={<Lock className="h-4 w-4" />}
          error={errors.confirm_password?.message}
          {...register('confirm_password')}
        />

        <div className="pt-2">
          <Button
            type="submit"
            variant="primary"
            size="lg"
            fullWidth
            isLoading={isPending}
            loadingText="Parol yangilanmoqda..."
          >
            Parolni yangilash
          </Button>
        </div>

        <div className="text-center pt-2">
          <Link
            to="/login"
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
