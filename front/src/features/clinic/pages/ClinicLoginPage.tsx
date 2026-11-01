import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Link, useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { Mail, Lock, Eye, EyeOff } from 'lucide-react';
import { AuthLayout } from '@/components/layout/AuthLayout';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Alert } from '@/components/ui/Alert';
import { clinicLoginSchema, ClinicLoginFormData } from '../schemas/clinicSchemas';
import { clinicApi } from '../api/clinicApi';
import { useAuth } from '@/app/providers/AuthProvider';

export function ClinicLoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [showPassword, setShowPassword] = useState(false);
  const [serverError, setServerError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ClinicLoginFormData>({
    resolver: zodResolver(clinicLoginSchema),
    defaultValues: {
      email: '',
      password: '',
    },
    mode: 'onTouched',
  });

  const onSubmit = async (data: ClinicLoginFormData) => {
    setServerError(null);
    try {
      const res = await clinicApi.login({
        email: data.email,
        password: data.password,
      });

      const loginData = res.data?.data;
      if (!loginData?.access_token || !loginData?.user) {
        throw new Error('Kirish ma’lumotlari to‘liq emas');
      }

      login(loginData.access_token, loginData.user);
      toast.success('Klinika portaliga xush kelibsiz!');

      // Check if clinic profile exists
      try {
        const clinicRes = await clinicApi.getMyClinic();
        if (clinicRes.data?.data) {
          navigate('/clinic/dashboard', { replace: true });
          return;
        }
      } catch (clinicErr: unknown) {
        const axiosErr = clinicErr as { response?: { status?: number } };
        if (axiosErr.response?.status === 404) {
          navigate('/clinic/onboarding', { replace: true });
          return;
        }
      }

      navigate('/clinic/dashboard', { replace: true });
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: { status?: number; data?: { message?: string } };
      };
      const status = axiosErr.response?.status;
      const msg = axiosErr.response?.data?.message;

      if (status === 403 && msg?.includes('tasdiqlanmagan')) {
        setServerError(msg);
        navigate(`/clinic/verify-email?email=${encodeURIComponent(data.email)}`);
        return;
      }

      const finalMsg = msg || 'Email yoki parol noto‘g‘ri';
      setServerError(finalMsg);
      toast.error(finalMsg);
    }
  };

  return (
    <AuthLayout
      title="Klinika portaliga kirish"
      subtitle="Klinikani boshqarish, shifokorlar va qabullarni sozlash uchun tizimga kiring"
      badgeText="Klinika portali"
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {serverError && (
          <Alert variant="error" title="Kirishda xatolik" message={serverError} />
        )}

        <Input
          type="email"
          label="Email manzili"
          placeholder="jamshid@clinic.uz"
          leftIcon={<Mail className="w-4 h-4 text-slate-400" />}
          error={errors.email?.message}
          {...register('email')}
        />

        <div>
          <div className="flex items-center justify-between mb-1.5">
            <label className="text-xs font-semibold text-slate-700">Parol</label>
            <Link
              to="/forgot-password"
              className="text-xs text-[#0866F5] hover:underline"
            >
              Parolni unutdingizmi?
            </Link>
          </div>
          <Input
            type={showPassword ? 'text' : 'password'}
            placeholder="Parolingizni kiriting"
            leftIcon={<Lock className="w-4 h-4 text-slate-400" />}
            rightIcon={
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="text-slate-400 hover:text-slate-600 focus:outline-hidden"
              >
                {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
              </button>
            }
            error={errors.password?.message}
            {...register('password')}
          />
        </div>

        <Button
          type="submit"
          variant="primary"
          size="lg"
          fullWidth
          isLoading={isSubmitting}
          className="bg-[#0866F5] hover:bg-blue-600 shadow-md"
        >
          Kirish
        </Button>

        <div className="pt-4 text-center space-y-2 border-t border-slate-100">
          <p className="text-xs text-slate-600">
            Hali hisobingiz yo‘qmi?{' '}
            <Link
              to="/clinic/register"
              className="font-semibold text-[#0866F5] hover:underline"
            >
              Klinikani ro‘yxatdan o‘tkazish
            </Link>
          </p>

          <p className="text-xs text-slate-400">
            Bemor kabinetiga kirmoqchimisiz?{' '}
            <Link to="/login" className="text-slate-600 hover:underline">
              Bemor sifatida kirish
            </Link>
          </p>
        </div>
      </form>
    </AuthLayout>
  );
}
