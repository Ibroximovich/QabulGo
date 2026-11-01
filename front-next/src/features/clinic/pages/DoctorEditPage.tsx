'use client';

import { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { useClinicContext } from '@/components/layout/clinic/ClinicLayout';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { ArrowLeft, User, UploadCloud, Loader2, Eye } from 'lucide-react';
import { ClinicDetail } from '../types';
import { clinicApi } from '../api/clinicApi';
import { doctorFormSchema, DoctorFormData } from '../schemas/clinicSchemas';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Alert } from '@/components/ui/Alert';
import { getMediaUrl } from '@/lib/utils/media';

interface ClinicContextType {
  clinic?: ClinicDetail;
  isLoadingClinic: boolean;
}

export function DoctorEditPage() {
  const { clinic, isLoadingClinic } = useClinicContext();
  const params = useParams();
  const doctorId = (params?.doctorId as string) || '';
  const router = useRouter();
  const queryClient = useQueryClient();
  const [photoUrl, setPhotoUrl] = useState('');
  const [isUploadingPhoto, setIsUploadingPhoto] = useState(false);
  const [serverError, setServerError] = useState<string | null>(null);

  const { data: specialtiesRes } = useQuery({
    queryKey: ['specialties'],
    queryFn: async () => {
      const res = await clinicApi.getSpecialties();
      return res.data?.data || [];
    },
  });
  const specialties = specialtiesRes || [];

  const {
    data: doctorData,
    isLoading: isLoadingDoctor,
    isError,
  } = useQuery({
    queryKey: ['doctor', clinic?.id, doctorId],
    queryFn: async () => {
      if (!clinic?.id || !doctorId) return null;
      const res = await clinicApi.getDoctor(clinic.id, doctorId);
      return res.data?.data || null;
    },
    enabled: Boolean(clinic?.id && doctorId),
  });

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<DoctorFormData>({
    resolver: zodResolver(doctorFormSchema),
    mode: 'onTouched',
  });

  useEffect(() => {
    if (doctorData) {
      reset({
        first_name: doctorData.first_name,
        last_name: doctorData.last_name,
        specialty_id: doctorData.specialty?.id || '',
        experience_years: doctorData.experience_years,
        consultation_price: doctorData.consultation_price,
        appointment_duration_minutes: doctorData.appointment_duration_minutes,
        bio: doctorData.bio || '',
        photo_url: doctorData.photo_url || '',
        is_active: doctorData.is_active,
      });
      setPhotoUrl(doctorData.photo_url || '');
    }
  }, [doctorData, reset]);

  const handlePhotoUpload = async (file: File) => {
    if (file.size > 5 * 1024 * 1024) {
      toast.error('Rasm hajmi 5MB dan oshmasligi kerak');
      return;
    }
    setIsUploadingPhoto(true);
    try {
      const res = await clinicApi.uploadImage(file, 'doctors');
      if (res.data?.url) {
        setPhotoUrl(res.data.url);
        setValue('photo_url', res.data.url);
        toast.success('Rasm yuklandi');
      }
    } catch {
      toast.error('Rasmni yuklashda xatolik yuz berdi');
    } finally {
      setIsUploadingPhoto(false);
    }
  };

  const onSubmit = async (data: DoctorFormData) => {
    if (!clinic?.id || !doctorId) return;
    setServerError(null);

    try {
      await clinicApi.updateDoctor(clinic.id, doctorId, {
        first_name: data.first_name,
        last_name: data.last_name,
        specialty_id: data.specialty_id,
        experience_years: Number(data.experience_years),
        consultation_price: Number(data.consultation_price),
        appointment_duration_minutes: Number(data.appointment_duration_minutes),
        bio: data.bio || '',
        photo_url: photoUrl || undefined,
        is_active: data.is_active,
      });

      toast.success('Shifokor ma’lumotlari muvaffaqiyatli yangilandi');
      queryClient.invalidateQueries({ queryKey: ['clinicDoctors'] });
      queryClient.invalidateQueries({ queryKey: ['doctor', clinic.id, doctorId] });
      router.push('/clinic/doctors');
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      const msg =
        axiosErr.response?.data?.message || 'Shifokorni tahrirlashda xatolik yuz berdi';
      setServerError(msg);
      toast.error(msg);
    }
  };

  if (isLoadingClinic || (Boolean(clinic?.id) && isLoadingDoctor) || (Boolean(clinic?.id) && !doctorData && !isError)) {
    return (
      <div className="flex flex-col items-center justify-center py-16 gap-3 text-slate-500">
        <Loader2 className="w-8 h-8 animate-spin text-[#0866F5]" />
        <p className="text-sm">Shifokor ma’lumotlari yuklanmoqda...</p>
      </div>
    );
  }

  if ((!isLoadingDoctor && !doctorData) || isError) {
    return (
      <div className="bg-white rounded-2xl border border-slate-200 p-8 text-center max-w-lg mx-auto shadow-xs my-8">
        <h3 className="text-base font-bold text-slate-900 mb-2">Shifokor topilmadi</h3>
        <p className="text-xs text-slate-500 mb-4">
          Ushbu shifokor mavjud emas yoki boshqa klinikaga tegishli bo‘lishi mumkin.
        </p>
        <Button variant="primary" onClick={() => router.push('/clinic/doctors')} className="bg-[#0866F5]">
          Shifokorlar ro‘yxatiga qaytish
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-3xl">
      <div className="flex items-center gap-3">
        <Button
          variant="outline"
          size="sm"
          onClick={() => router.push('/clinic/doctors')}
          className="gap-1.5"
        >
          <ArrowLeft className="w-4 h-4" />
          Orqaga
        </Button>
        <div>
          <h2 className="text-xl font-extrabold text-slate-900">
            Shifokor ma’lumotlarini tahrirlash
          </h2>
          <p className="text-xs text-slate-500">
            {doctorData?.first_name} {doctorData?.last_name} ma’lumotlarini yangilash
          </p>
        </div>
      </div>

      <div className="bg-white rounded-2xl border border-slate-200 p-6 sm:p-8 shadow-xs">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5" noValidate>
          {serverError && (
            <Alert variant="error" title="Xatolik" message={serverError} />
          )}

          {/* Photo upload section */}
          <div className="flex flex-col sm:flex-row sm:items-center gap-5 pb-5 border-b border-slate-100">
            <div className="relative group shrink-0">
              {photoUrl ? (
                <>
                  <img
                    src={getMediaUrl(photoUrl)}
                    alt="Shifokor rasmi"
                    className="w-24 h-24 sm:w-28 sm:h-28 rounded-2xl object-cover object-top border-2 border-slate-200 shadow-xs"
                    style={{ imageRendering: '-webkit-optimize-contrast' }}
                  />
                  <a
                    href={getMediaUrl(photoUrl)}
                    target="_blank"
                    rel="noreferrer"
                    className="absolute inset-0 bg-slate-900/40 text-white rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity flex flex-col items-center justify-center text-[11px] font-semibold gap-1"
                    title="To‘liq o‘lchamda ko‘rish"
                  >
                    <Eye className="w-5 h-5" />
                    <span>Kattalashtirish</span>
                  </a>
                </>
              ) : (
                <div className="w-24 h-24 sm:w-28 sm:h-28 rounded-2xl bg-slate-100 border-2 border-dashed border-slate-300 flex items-center justify-center text-slate-400">
                  <User className="w-10 h-10 opacity-40" />
                </div>
              )}
            </div>

            <div className="space-y-2">
              <div>
                <span className="text-xs font-bold text-slate-800 block">
                  Shifokor fotosurati
                </span>
                <p className="text-[11px] text-slate-500 mt-0.5">
                  Shifokorning yuzi yaqindan (portret formatida) ko‘ringan 1:1 kvadrat yoki 3:4 sifatli fotosurat yuklang.
                </p>
              </div>

              <div className="flex items-center gap-2 pt-1">
                <label className="cursor-pointer inline-flex items-center gap-1.5 px-3.5 py-2 bg-white border border-slate-300 rounded-xl text-xs font-semibold text-slate-700 hover:bg-slate-50 shadow-2xs transition-colors">
                  {isUploadingPhoto ? (
                    <Loader2 className="w-3.5 h-3.5 animate-spin text-[#0866F5]" />
                  ) : (
                    <UploadCloud className="w-3.5 h-3.5 text-[#0866F5]" />
                  )}
                  <span>{photoUrl ? 'Rasmni almashtirish' : 'Yangi rasm yuklash'}</span>
                  <input
                    type="file"
                    accept="image/png,image/jpeg,image/webp"
                    className="hidden"
                    onChange={(e) => {
                      const file = e.target.files?.[0];
                      if (file) handlePhotoUpload(file);
                    }}
                  />
                </label>

                {photoUrl && (
                  <a
                    href={getMediaUrl(photoUrl)}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center gap-1 px-3 py-2 text-xs font-medium text-slate-600 hover:text-[#0866F5] hover:bg-blue-50/50 rounded-xl transition-colors border border-transparent hover:border-blue-200"
                  >
                    <Eye className="w-3.5 h-3.5" />
                    Asl o‘lchamda
                  </a>
                )}
              </div>
              <p className="text-[10px] text-slate-400">Formatlar: PNG, JPG, WebP (maksimal hajm 5MB)</p>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <Input
              label="Ism *"
              error={errors.first_name?.message}
              {...register('first_name')}
            />
            <Input
              label="Familiya *"
              error={errors.last_name?.message}
              {...register('last_name')}
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="text-xs font-semibold text-slate-700 block mb-1.5">
                Mutaxassislik *
              </label>
              <select
                className="w-full h-11 px-3.5 rounded-xl border border-slate-300 text-sm focus:outline-hidden focus:border-[#0866F5] focus:ring-2 focus:ring-blue-100 bg-white"
                {...register('specialty_id')}
              >
                <option value="">Mutaxassislikni tanlang</option>
                {specialties.map((spec) => (
                  <option key={spec.id} value={spec.id}>
                    {spec.name}
                  </option>
                ))}
              </select>
              {errors.specialty_id && (
                <p className="mt-1 text-xs text-rose-500">
                  {errors.specialty_id.message}
                </p>
              )}
            </div>

            <Input
              type="number"
              label="Ish tajribasi (yillarda) *"
              error={errors.experience_years?.message}
              {...register('experience_years')}
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <Input
              type="number"
              label="Konsultatsiya narxi (so‘mda) *"
              error={errors.consultation_price?.message}
              {...register('consultation_price')}
            />

            <div>
              <label className="text-xs font-semibold text-slate-700 block mb-1.5">
                Qabul davomiyligi (daqiqa) *
              </label>
              <select
                className="w-full h-11 px-3.5 rounded-xl border border-slate-300 text-sm focus:outline-hidden focus:border-[#0866F5] focus:ring-2 focus:ring-blue-100 bg-white font-mono"
                {...register('appointment_duration_minutes')}
              >
                <option value={15}>15 daqiqa</option>
                <option value={20}>20 daqiqa</option>
                <option value={30}>30 daqiqa</option>
                <option value={45}>45 daqiqa</option>
                <option value={60}>60 daqiqa (1 soat)</option>
              </select>
              {errors.appointment_duration_minutes && (
                <p className="mt-1 text-xs text-rose-500">
                  {errors.appointment_duration_minutes.message}
                </p>
              )}
            </div>
          </div>

          <div>
            <label className="text-xs font-semibold text-slate-700 block mb-1.5">
              Shifokor haqida ma’lumot (Bio)
            </label>
            <textarea
              rows={3}
              className="w-full px-3.5 py-2.5 rounded-xl border border-slate-300 text-sm focus:outline-hidden focus:border-[#0866F5] focus:ring-2 focus:ring-blue-100 resize-none"
              {...register('bio')}
            />
            {errors.bio && (
              <p className="mt-1 text-xs text-rose-500">{errors.bio.message}</p>
            )}
          </div>

          <div className="pt-2">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                className="rounded border-slate-300 text-[#0866F5] focus:ring-[#0866F5] h-4 w-4"
                {...register('is_active')}
              />
              <span className="text-xs font-medium text-slate-700">
                Klinikada faol (bemorlar qabuliga yozilishi mumkin)
              </span>
            </label>
          </div>

          <div className="flex items-center justify-end gap-3 pt-4 border-t border-slate-100">
            <Button
              type="button"
              variant="outline"
              onClick={() => router.push('/clinic/doctors')}
            >
              Bekor qilish
            </Button>

            <Button
              type="submit"
              variant="primary"
              isLoading={isSubmitting}
              className="bg-[#0866F5] hover:bg-blue-600 shadow-xs"
            >
              O‘zgarishlarni saqlash
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
