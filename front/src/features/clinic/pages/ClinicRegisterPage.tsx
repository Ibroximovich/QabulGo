import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Link, useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { Eye, EyeOff, User, Mail, Phone, Lock } from 'lucide-react';
import { AuthLayout } from '@/components/layout/AuthLayout';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Alert } from '@/components/ui/Alert';
import {
  clinicRegisterSchema,
  ClinicRegisterFormData,
} from '../schemas/clinicSchemas';
import { clinicApi } from '../api/clinicApi';
import { formatPhone, normalizePhone } from '@/lib/utils/phone';

export function ClinicRegisterPage() {
  const navigate = useNavigate();
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [serverError, setServerError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<ClinicRegisterFormData>({
    resolver: zodResolver(clinicRegisterSchema),
    defaultValues: {
      first_name: '',
      last_name: '',
      email: '',
      phone: '+998 ',
      password: '',
      confirm_password: '',
      accept_terms: true,
    },
    mode: 'onTouched',
  });

  const phoneValue = watch('phone');

  const handlePhoneChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatPhone(e.target.value);
    setValue('phone', formatted, { shouldValidate: true });
  };

  const onSubmit = async (data: ClinicRegisterFormData) => {
    setServerError(null);
    try {
      const res = await clinicApi.register({
        first_name: data.first_name,
        last_name: data.last_name,
        email: data.email,
        phone: normalizePhone(data.phone),
        password: data.password,
        confirm_password: data.confirm_password,
        accept_terms: Boolean(data.accept_terms),
      });

      toast.success(res.data?.message || 'Tasdiqlash kodi emailingizga yuborildi');
      navigate(`/clinic/verify-email?email=${encodeURIComponent(data.email)}`);
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      const msg =
        axiosErr.response?.data?.message ||
        'Ro‘yxatdan o‘tishda xatolik yuz berdi. Iltimos, ma’lumotlarni tekshiring.';
      setServerError(msg);
      toast.error(msg);
    }
  };

  return (
    <AuthLayout
      title="Klinikani ro‘yxatdan o‘tkazish"
      subtitle="Klinikangizni QabulGo platformasiga qo‘shish uchun hisob yarating"
      badgeText="Klinika portali"
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {serverError && (
          <Alert variant="error" title="Xatolik" message={serverError} />
        )}

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
          <Input
            label="Ism"
            placeholder="Jamshid"
            leftIcon={<User className="w-4 h-4 text-slate-400" />}
            error={errors.first_name?.message}
            {...register('first_name')}
          />
          <Input
            label="Familiya"
            placeholder="Karimov"
            leftIcon={<User className="w-4 h-4 text-slate-400" />}
            error={errors.last_name?.message}
            {...register('last_name')}
          />
        </div>

        <Input
          type="email"
          label="Email manzili"
          placeholder="jamshid@clinic.uz"
          leftIcon={<Mail className="w-4 h-4 text-slate-400" />}
          error={errors.email?.message}
          {...register('email')}
        />

        <Input
          type="tel"
          label="Telefon raqami"
          placeholder="+998 90 123 45 67"
          value={phoneValue}
          leftIcon={<Phone className="w-4 h-4 text-slate-400" />}
          error={errors.phone?.message}
          onChange={handlePhoneChange}
        />

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
          <Input
            type={showPassword ? 'text' : 'password'}
            label="Parol"
            placeholder="Kamida 8 ta belgi"
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
            helperText="Katta va kichik harf, raqam"
            {...register('password')}
          />

          <Input
            type={showConfirmPassword ? 'text' : 'password'}
            label="Parolni tasdiqlash"
            placeholder="Parolni qayta kiriting"
            leftIcon={<Lock className="w-4 h-4 text-slate-400" />}
            rightIcon={
              <button
                type="button"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                className="text-slate-400 hover:text-slate-600 focus:outline-hidden"
              >
                {showConfirmPassword ? (
                  <EyeOff className="w-4 h-4" />
                ) : (
                  <Eye className="w-4 h-4" />
                )}
              </button>
            }
            error={errors.confirm_password?.message}
            {...register('confirm_password')}
          />
        </div>

        <div className="pt-1">
          <label className="flex items-start gap-2.5 cursor-pointer">
            <input
              type="checkbox"
              className="mt-1 h-4 w-4 rounded border-slate-300 text-[#0866F5] focus:ring-[#0866F5]"
              {...register('accept_terms')}
            />
            <span className="text-xs text-slate-600 leading-relaxed">
              Men QabulGo platformasining{' '}
              <a href="#" className="text-[#0866F5] hover:underline">
                Foydalanish shartlari
              </a>{' '}
              va{' '}
              <a href="#" className="text-[#0866F5] hover:underline">
                Maxfiylik siyosati
              </a>
              ga roziman.
            </span>
          </label>
          {errors.accept_terms && (
            <p className="mt-1 text-xs text-rose-500">{errors.accept_terms.message}</p>
          )}
        </div>

        <Button
          type="submit"
          variant="primary"
          size="lg"
          fullWidth
          isLoading={isSubmitting}
          className="mt-2 bg-[#0866F5] hover:bg-blue-600 shadow-md"
        >
          Klinika hisobini yaratish
        </Button>

        <div className="pt-3 text-center space-y-2 border-t border-slate-100">
          <p className="text-xs text-slate-600">
            Allaqachon hisobingiz bormi?{' '}
            <Link
              to="/clinic/login"
              className="font-semibold text-[#0866F5] hover:underline"
            >
              Klinika portaliga kirish
            </Link>
          </p>

          <p className="text-xs text-slate-400">
            Bemor sifatida kirmoqchimisiz?{' '}
            <Link to="/register" className="text-slate-600 hover:underline">
              Bemor ro‘yxatidan o‘tish
            </Link>
          </p>
        </div>
      </form>
    </AuthLayout>
  );
}
