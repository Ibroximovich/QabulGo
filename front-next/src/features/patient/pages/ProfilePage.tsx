'use client';

import { useEffect, ChangeEvent } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import {
  User,
  Phone,
  Mail,
  ShieldCheck,
  CheckCircle2,
  Lock,
  Save,
  RotateCcw,
  Calendar,
  AlertCircle,
  RefreshCw,
} from 'lucide-react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { Skeleton } from '@/components/ui/Skeleton';
import { usePatientProfile } from '../hooks/usePatientProfile';
import { useUpdateProfile } from '../hooks/useUpdateProfile';
import { updateProfileSchema, UpdateProfileFormValues } from '../schemas/patientSchemas';
import { formatUzbekPhoneDisplay, cleanPhoneNumber } from '@/lib/utils/phone';
import { formatDateUz, formatRoleUz } from '@/lib/utils/formatters';

export function ProfilePage() {
  const { data: profile, isLoading, isError, refetch } = usePatientProfile();
  const { mutate: updateProfile, isPending } = useUpdateProfile();

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    reset,
    formState: { errors, isDirty },
  } = useForm<UpdateProfileFormValues>({
    resolver: zodResolver(updateProfileSchema),
    defaultValues: {
      first_name: '',
      last_name: '',
      phone: '',
    },
    mode: 'onBlur',
  });

  // Populate form once profile loads
  useEffect(() => {
    if (profile) {
      reset({
        first_name: profile.first_name,
        last_name: profile.last_name,
        phone: formatUzbekPhoneDisplay(profile.phone),
      });
    }
  }, [profile, reset]);

  const phoneValue = watch('phone');

  const handlePhoneChange = (e: ChangeEvent<HTMLInputElement>) => {
    const formatted = formatUzbekPhoneDisplay(e.target.value);
    setValue('phone', formatted, { shouldDirty: true, shouldValidate: true });
  };

  const handleReset = () => {
    if (profile) {
      reset({
        first_name: profile.first_name,
        last_name: profile.last_name,
        phone: formatUzbekPhoneDisplay(profile.phone),
      });
    }
  };

  const onSubmit = (values: UpdateProfileFormValues) => {
    updateProfile({
      first_name: values.first_name.trim(),
      last_name: values.last_name.trim(),
      phone: cleanPhoneNumber(values.phone),
    });
  };

  return (
    <AppLayout>
      <div className="max-w-4xl mx-auto space-y-6 animate-fade-in">
        {/* Page Header */}
        <div>
          <h1 className="text-2xl sm:text-3xl font-extrabold text-brand-navy tracking-tight">
            Profil ma’lumotlari
          </h1>
          <p className="text-sm text-brand-muted mt-1">
            Shaxsiy va aloqa ma’lumotlaringizni ko‘ring hamda tahrirlang
          </p>
        </div>

        {isLoading ? (
          <Card>
            <CardContent className="p-8 space-y-6">
              <Skeleton className="h-6 w-48" />
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <Skeleton className="h-12 w-full" />
                <Skeleton className="h-12 w-full" />
              </div>
              <Skeleton className="h-12 w-full" />
            </CardContent>
          </Card>
        ) : isError ? (
          <Card className="border-red-200 bg-red-50/20">
            <CardContent className="p-8 text-center space-y-4">
              <AlertCircle className="h-10 w-10 text-brand-error mx-auto" />
              <h3 className="font-bold text-brand-navy">Profilni yuklab bo‘lmadi</h3>
              <p className="text-sm text-brand-muted">
                Server bilan ulanishda muammo yuz berdi. Iltimos, qayta urinib ko‘ring.
              </p>
              <Button variant="outline" size="sm" onClick={() => refetch()} leftIcon={<RefreshCw className="h-4 w-4" />}>
                Qayta yuklash
              </Button>
            </CardContent>
          </Card>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Sidebar / Profile overview */}
            <div className="lg:col-span-1 space-y-6">
              <Card>
                <CardContent className="p-6 text-center space-y-4">
                  <div className="h-20 w-20 rounded-full gradient-primary text-white font-extrabold text-2xl flex items-center justify-center mx-auto shadow-md">
                    {profile?.first_name?.[0]}
                    {profile?.last_name?.[0]}
                  </div>

                  <div>
                    <h3 className="text-lg font-bold text-brand-navy">
                      {profile?.first_name} {profile?.last_name}
                    </h3>
                    <p className="text-xs text-brand-muted mt-0.5">{profile?.email}</p>
                  </div>

                  <div className="flex flex-wrap items-center justify-center gap-2 pt-2">
                    <Badge variant="teal" size="sm">
                      {formatRoleUz(profile?.role)}
                    </Badge>
                    {profile?.email_verified && (
                      <Badge variant="success" size="sm" className="gap-1">
                        <CheckCircle2 className="h-3 w-3" />
                        <span>Tasdiqlangan</span>
                      </Badge>
                    )}
                  </div>

                  <div className="border-t border-slate-100 pt-4 text-left text-xs text-brand-muted space-y-2">
                    <div className="flex items-center gap-2">
                      <Calendar className="h-4 w-4 text-slate-400" />
                      <span>A’zolik: {formatDateUz(profile?.created_at)}</span>
                    </div>
                    <div className="flex items-center gap-2 text-emerald-700">
                      <ShieldCheck className="h-4 w-4" />
                      <span>Shaxsiy ma’lumotlar himoyalangan</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Editable Form */}
            <div className="lg:col-span-2">
              <Card elevated>
                <CardHeader>
                  <CardTitle>Ma’lumotlarni tahrirlash</CardTitle>
                  <CardDescription>
                    Ism, familiya va telefon raqamingizni o‘zgartirishingiz mumkin
                  </CardDescription>
                </CardHeader>

                <CardContent>
                  <form onSubmit={handleSubmit(onSubmit)} className="space-y-5" noValidate>
                    {/* Read-only Email Field */}
                    <Input
                      label="Email manzili (O‘zgartirib bo‘lmaydi)"
                      value={profile?.email || ''}
                      disabled
                      leftIcon={<Mail className="h-4 w-4" />}
                      rightIcon={<Lock className="h-4 w-4 text-slate-400" />}
                      helperText="Xavfsizlik maqsadida ro‘yxatdan o‘tgan email manzilini o‘zgartirish mumkin emas"
                    />

                    {/* First & Last Name */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <Input
                        label="Ism"
                        placeholder="Ali"
                        leftIcon={<User className="h-4 w-4" />}
                        error={errors.first_name?.message}
                        {...register('first_name')}
                      />

                      <Input
                        label="Familiya"
                        placeholder="Valiyev"
                        leftIcon={<User className="h-4 w-4" />}
                        error={errors.last_name?.message}
                        {...register('last_name')}
                      />
                    </div>

                    {/* Phone */}
                    <Input
                      label="Telefon raqami"
                      type="tel"
                      placeholder="+998 (90) 123-45-67"
                      leftIcon={<Phone className="h-4 w-4" />}
                      value={phoneValue}
                      onChange={handlePhoneChange}
                      error={errors.phone?.message}
                      helperText="Qabul eslatmalari va tasdiqlash uchun"
                    />

                    {/* Action buttons */}
                    <div className="pt-4 border-t border-slate-100 flex flex-col-reverse sm:flex-row items-center justify-end gap-3">
                      {isDirty && (
                        <Button
                          type="button"
                          variant="ghost"
                          size="md"
                          disabled={isPending}
                          onClick={handleReset}
                          leftIcon={<RotateCcw className="h-4 w-4" />}
                        >
                          Bekor qilish
                        </Button>
                      )}

                      <Button
                        type="submit"
                        variant="primary"
                        size="md"
                        disabled={!isDirty || isPending}
                        isLoading={isPending}
                        loadingText="Saqlanmoqda..."
                        leftIcon={<Save className="h-4 w-4" />}
                      >
                        O‘zgarishlarni saqlash
                      </Button>
                    </div>
                  </form>
                </CardContent>
              </Card>
            </div>
          </div>
        )}
      </div>
    </AppLayout>
  );
}
