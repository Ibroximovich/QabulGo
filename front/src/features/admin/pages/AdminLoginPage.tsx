import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { Mail, Lock, Eye, EyeOff } from 'lucide-react';
import { BrandLogo } from '@/components/common/BrandLogo';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Alert } from '@/components/ui/Alert';
import { adminLoginSchema, AdminLoginFormData } from '@/features/clinic/schemas/clinicSchemas';
import { adminApi } from '../api/adminApi';
import { useAuth } from '@/app/providers/AuthProvider';

export function AdminLoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [showPassword, setShowPassword] = useState(false);
  const [serverError, setServerError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<AdminLoginFormData>({
    resolver: zodResolver(adminLoginSchema),
    defaultValues: {
      email: '',
      password: '',
    },
    mode: 'onTouched',
  });

  const onSubmit = async (data: AdminLoginFormData) => {
    setServerError(null);
    try {
      const res = await adminApi.login({
        email: data.email,
        password: data.password,
      });

      const loginData = res.data?.data;
      if (!loginData?.access_token || !loginData?.user) {
        throw new Error('Kirish ma’lumotlari to‘liq emas');
      }

      login(loginData.access_token, loginData.user);
      toast.success('Superadmin paneliga xush kelibsiz');
      navigate('/admin/clinics', { replace: true });
    } catch (err: unknown) {
      const axiosErr = err as {
        response?: { data?: { message?: string } };
      };
      const msg = axiosErr.response?.data?.message || 'Email yoki parol noto‘g‘ri';
      setServerError(msg);
      toast.error(msg);
    }
  };

  return (
    <div className="min-h-screen bg-[#08234A] flex flex-col justify-center py-12 sm:px-6 lg:px-8 font-sans">
      <div className="sm:mx-auto sm:w-full sm:max-w-md text-center">
        <div className="inline-flex items-center justify-center mb-4">
          <BrandLogo size="lg" subtitle="Superadmin" isDark />
        </div>
        <h2 className="text-xl font-extrabold text-white">QabulGo Nazorat Markazi</h2>
        <p className="mt-1 text-xs text-blue-200/70">
          Platforma boshqaruvi va klinikalarni tasdiqlash portali
        </p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md px-4 sm:px-0">
        <div className="bg-white py-8 px-6 shadow-2xl rounded-3xl sm:px-10 border border-blue-900/40">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
            {serverError && (
              <Alert variant="error" title="Xatolik" message={serverError} />
            )}

            <Input
              type="email"
              label="Superadmin emaili"
              placeholder="admin@qabulgo.uz"
              leftIcon={<Mail className="w-4 h-4 text-slate-400" />}
              error={errors.email?.message}
              {...register('email')}
            />

            <div>
              <Input
                type={showPassword ? 'text' : 'password'}
                label="Maxfiy parol"
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
              className="bg-[#0866F5] hover:bg-blue-600 shadow-md mt-2"
            >
              Tizimga kirish
            </Button>
          </form>
        </div>
      </div>
    </div>
  );
}
