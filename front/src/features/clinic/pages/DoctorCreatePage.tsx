import { useState } from 'react';
import { useOutletContext, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { ArrowLeft, User, UploadCloud, Loader2 } from 'lucide-react';
import { ClinicDetail } from '../types';
import { clinicApi } from '../api/clinicApi';
import { doctorFormSchema, DoctorFormData } from '../schemas/clinicSchemas';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Alert } from '@/components/ui/Alert';

interface ClinicContextType {
  clinic?: ClinicDetail;
  isLoadingClinic: boolean;
}

export function DoctorCreatePage() {
  const { clinic } = useOutletContext<ClinicContextType>();
  const navigate = useNavigate();
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
    register,
    handleSubmit,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<DoctorFormData>({
    resolver: zodResolver(doctorFormSchema),
    defaultValues: {
      first_name: '',
      last_name: '',
      specialty_id: '',
      experience_years: 5,
      consultation_price: 150000,
      appointment_duration_minutes: 30,
      bio: '',
      photo_url: '',
      is_active: true,
    },
    mode: 'onTouched',
  });

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
        setValue('photo_url', res.data.url, { shouldValidate: true });
        toast.success('Rasm yuklandi');
      }
    } catch {
      toast.error('Rasmni yuklashda xatolik yuz berdi');
    } finally {
      setIsUploadingPhoto(false);
    }
  };

  const onSubmit = async (data: DoctorFormData) => {
    if (!clinic?.id) return;
    setServerError(null);

    try {
      await clinicApi.createDoctor(clinic.id, {
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

      toast.success('Shifokor muvaffaqiyatli qo‘shildi');
      queryClient.invalidateQueries({ queryKey: ['clinicDoctors'] });
      queryClient.invalidateQueries({ queryKey: ['myClinic'] });
      navigate('/clinic/doctors');
    } catch (err: unknown) {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      const msg =
        axiosErr.response?.data?.message || 'Shifokorni qo‘shishda xatolik yuz berdi';
      setServerError(msg);
      toast.error(msg);
    }
  };

  return (
    <div className="space-y-6 max-w-3xl">
      <div className="flex items-center gap-3">
        <Button
          variant="outline"
          size="sm"
          onClick={() => navigate('/clinic/doctors')}
          className="gap-1.5"
        >
          <ArrowLeft className="w-4 h-4" />
          Orqaga
        </Button>
        <div>
          <h2 className="text-xl font-extrabold text-slate-900">Yangi shifokor qo‘shish</h2>
          <p className="text-xs text-slate-500">
            Shifokor ma’lumotlari, mutaxassisligi va konsultatsiya shartlarini kiriting
          </p>
        </div>
      </div>

      <div className="bg-white rounded-2xl border border-slate-200 p-6 sm:p-8 shadow-xs">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5" noValidate>
          {serverError && (
            <Alert variant="error" title="Xatolik" message={serverError} />
          )}

          {/* Photo upload section */}
          <div className="flex items-center gap-4 pb-4 border-b border-slate-100">
            {photoUrl ? (
              <img
                src={photoUrl}
                alt="Shifokor rasmi"
                className="w-20 h-20 rounded-2xl object-cover border border-slate-200"
              />
            ) : (
              <div className="w-20 h-20 rounded-2xl bg-slate-100 border border-slate-200 flex items-center justify-center text-slate-400">
                <User className="w-8 h-8 opacity-40" />
              </div>
            )}

            <div>
              <span className="text-xs font-semibold text-slate-700 block mb-1">
                Shifokor fotosurati
              </span>
              <label className="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 bg-white border border-slate-300 rounded-lg text-xs font-medium text-slate-700 hover:bg-slate-50 shadow-2xs">
                {isUploadingPhoto ? (
                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                ) : (
                  <UploadCloud className="w-3.5 h-3.5" />
                )}
                <span>Rasm yuklash</span>
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
              <p className="text-[11px] text-slate-400 mt-1">PNG, JPG, WebP (maks. 5MB)</p>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <Input
              label="Ism *"
              placeholder="Masalan: Anvar"
              error={errors.first_name?.message}
              {...register('first_name')}
            />
            <Input
              label="Familiya *"
              placeholder="Masalan: Saidov"
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
              placeholder="10"
              error={errors.experience_years?.message}
              {...register('experience_years')}
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <Input
              type="number"
              label="Konsultatsiya narxi (so‘mda) *"
              placeholder="150000"
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
              placeholder="Ma’lumoti, toifasi, ilmiy unvoni va yutuqlari..."
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
              onClick={() => navigate('/clinic/doctors')}
            >
              Bekor qilish
            </Button>

            <Button
              type="submit"
              variant="primary"
              isLoading={isSubmitting}
              className="bg-[#0866F5] hover:bg-blue-600 shadow-xs"
            >
              Shifokorni saqlash
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
