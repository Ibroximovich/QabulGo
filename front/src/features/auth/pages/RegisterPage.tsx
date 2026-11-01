import { useState, ChangeEvent } from 'react';
import { Link } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { User, Mail, Phone, Lock } from 'lucide-react';
import { AuthLayout } from '@/components/layout/AuthLayout';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';
import { registerSchema, RegisterFormValues } from '../schemas/authSchemas';
import { useRegister } from '../hooks/useRegister';
import { formatUzbekPhoneDisplay, cleanPhoneNumber } from '@/lib/utils/phone';

export function RegisterPage() {
  const { mutate: registerUser, isPending } = useRegister();
  const [isTermsModalOpen, setIsTermsModalOpen] = useState(false);

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      first_name: '',
      last_name: '',
      email: '',
      phone: '+998 ',
      password: '',
      confirm_password: '',
      accept_terms: false,
    },
    mode: 'onBlur',
  });

  const phoneValue = watch('phone');

  const handlePhoneChange = (e: ChangeEvent<HTMLInputElement>) => {
    const formatted = formatUzbekPhoneDisplay(e.target.value);
    setValue('phone', formatted, { shouldValidate: true });
  };

  const onSubmit = (values: RegisterFormValues) => {
    registerUser({
      first_name: values.first_name.trim(),
      last_name: values.last_name.trim(),
      email: values.email.trim().toLowerCase(),
      phone: cleanPhoneNumber(values.phone),
      password: values.password,
      confirm_password: values.confirm_password,
      accept_terms: values.accept_terms,
    });
  };

  return (
    <AuthLayout
      title="Ro‘yxatdan o‘tish"
      subtitle="Shifokor qabuliga qulay va onlayn yozilish uchun o‘z shaxsiy profilingizni yarating"
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {/* Ism va Familiya (2 ustun) */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <Input
            label="Ismingiz"
            placeholder="Ali"
            leftIcon={<User className="h-4 w-4" />}
            error={errors.first_name?.message}
            {...register('first_name')}
          />

          <Input
            label="Familiyangiz"
            placeholder="Valiyev"
            leftIcon={<User className="h-4 w-4" />}
            error={errors.last_name?.message}
            {...register('last_name')}
          />
        </div>

        {/* Email */}
        <Input
          label="Email manzili"
          type="email"
          placeholder="ali@example.com"
          leftIcon={<Mail className="h-4 w-4" />}
          error={errors.email?.message}
          helperText="Tasdiqlash kodi ushbu emailga yuboriladi"
          {...register('email')}
        />

        {/* Telefon raqam */}
        <Input
          label="Telefon raqamingiz"
          type="tel"
          placeholder="+998 (90) 123-45-67"
          leftIcon={<Phone className="h-4 w-4" />}
          value={phoneValue}
          onChange={handlePhoneChange}
          error={errors.phone?.message}
        />

        {/* Parol */}
        <Input
          label="Parol"
          isPassword
          placeholder="••••••••"
          leftIcon={<Lock className="h-4 w-4" />}
          error={errors.password?.message}
          helperText="Kamida 8 ta belgi, harf va raqamdan iborat bo‘lsin"
          {...register('password')}
        />

        {/* Parolni tasdiqlash */}
        <Input
          label="Parolni tasdiqlang"
          isPassword
          placeholder="••••••••"
          leftIcon={<Lock className="h-4 w-4" />}
          error={errors.confirm_password?.message}
          {...register('confirm_password')}
        />

        {/* Foydalanish shartlari */}
        <div className="pt-1">
          <label className="flex items-start gap-3 cursor-pointer select-none">
            <input
              type="checkbox"
              className="mt-1 h-4 w-4 rounded border-slate-300 text-brand-blue focus:ring-brand-blue/30 cursor-pointer"
              {...register('accept_terms')}
            />
            <span className="text-xs text-brand-muted leading-relaxed">
              Men{' '}
              <button
                type="button"
                onClick={() => setIsTermsModalOpen(true)}
                className="text-brand-blue font-semibold hover:underline"
              >
                foydalanish shartlari
              </button>{' '}
              va maxfiylik siyosatiga to‘liq roziman.
            </span>
          </label>
          {errors.accept_terms && (
            <p className="text-xs text-brand-error font-medium mt-1">
              {errors.accept_terms.message}
            </p>
          )}
        </div>

        {/* Submit tugmasi */}
        <div className="pt-2">
          <Button
            type="submit"
            variant="primary"
            size="lg"
            fullWidth
            isLoading={isPending}
            loadingText="Ro‘yxatdan o‘tilmoqda..."
          >
            Ro‘yxatdan o‘tish
          </Button>
        </div>

        {/* Login havolasi */}
        <p className="text-center text-sm text-brand-muted pt-2">
          Profilingiz bormi?{' '}
          <Link
            to="/login"
            className="text-brand-blue font-bold hover:underline inline-flex items-center gap-1"
          >
            Tizimga kirish
          </Link>
        </p>

        {/* Klinika portali havolasi */}
        <div className="pt-3 border-t border-slate-100 text-center">
          <p className="text-xs text-slate-500">
            Klinikani ro‘yxatdan o‘tkazmoqchimisiz?{' '}
            <Link to="/clinic/register" className="text-brand-blue font-semibold hover:underline">
              Klinika ro‘yxatidan o‘tish
            </Link>
          </p>
        </div>
      </form>

      {/* Foydalanish shartlari modali */}
      <Modal
        isOpen={isTermsModalOpen}
        onClose={() => setIsTermsModalOpen(false)}
        title="QabulGo foydalanish shartlari"
        description="Versiya: v1.0 — 2026-yil"
      >
        <div className="space-y-3 text-xs text-slate-600 max-h-80 overflow-y-auto pr-2">
          <p>
            1. <strong>Umumiy qoidalar:</strong> QabulGo platformasi shifokorlar va bemorlar o‘rtasida onlayn navbat olish xizmatini taqdim etadi.
          </p>
          <p>
            2. <strong>Bemor majburiyatlari:</strong> Ro‘yxatdan o‘tish vaqtida haqiqiy va aniq shaxsiy ma’lumotlarni kiritish talab etiladi.
          </p>
          <p>
            3. <strong>Xavfsizlik:</strong> Sizning shaxsiy va tibbiy ma’lumotlaringiz shifrlangan holatda saqlanadi va uchinchi shaxslarga berilmaydi.
          </p>
          <p>
            4. <strong>Hisob ma’lumotlari:</strong> Parolingizni begona shaxslarga bermang. Akkaunt xavfsizligi foydalanuvchi javobgarligidadir.
          </p>
        </div>
        <div className="mt-5 flex justify-end">
          <Button variant="primary" size="sm" onClick={() => setIsTermsModalOpen(false)}>
            Tushundim
          </Button>
        </div>
      </Modal>
    </AuthLayout>
  );
}
