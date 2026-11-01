'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  Building2,
  MapPin,
  Clock,
  Image as ImageIcon,
  CheckCircle,
  ArrowRight,
  ArrowLeft,
  UploadCloud,
  Check,
  Send,
  Loader2,
} from 'lucide-react';
import { BrandLogo } from '@/components/common/BrandLogo';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Alert } from '@/components/ui/Alert';
import { clinicApi } from '../api/clinicApi';
import { WorkingHourItem, CreateClinicPayload } from '../types';
import { formatPhone, isValidUzbekPhone } from '@/lib/utils/phone';
import { UZBEKISTAN_REGIONS } from '@/lib/constants/regions';
import { isValidMapLocationUrl } from '@/lib/utils/location';
import { getMediaUrl } from '@/lib/utils/media';

const WEEKDAYS = [
  { id: 1, name: 'Dushanba' },
  { id: 2, name: 'Seshanba' },
  { id: 3, name: 'Chorshanba' },
  { id: 4, name: 'Payshanba' },
  { id: 5, name: 'Juma' },
  { id: 6, name: 'Shanba' },
  { id: 7, name: 'Yakshanba' },
];

const DEFAULT_WORKING_HOURS: WorkingHourItem[] = WEEKDAYS.map((day) => ({
  weekday: day.id,
  opening_time: '08:00',
  closing_time: '18:00',
  is_closed: day.id === 7,
}));

export function ClinicOnboardingPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const [currentStep, setCurrentStep] = useState(1);
  const [formError, setFormError] = useState<string | null>(null);

  // Form State
  const [clinicId, setClinicId] = useState<string | null>(null);
  const [name, setName] = useState('');
  const [phone, setPhone] = useState('+998 ');
  const [email, setEmail] = useState('');
  const [description, setDescription] = useState('');

  const [region, setRegion] = useState('');
  const [district, setDistrict] = useState('');
  const [address, setAddress] = useState('');
  const [locationUrl, setLocationUrl] = useState('');
  const [locationUrlError, setLocationUrlError] = useState<string | null>(null);

  const [workingHours, setWorkingHours] = useState<WorkingHourItem[]>(DEFAULT_WORKING_HOURS);

  const [logoUrl, setLogoUrl] = useState('');
  const [coverImageUrl, setCoverImageUrl] = useState('');
  const [isUploadingLogo, setIsUploadingLogo] = useState(false);
  const [isUploadingCover, setIsUploadingCover] = useState(false);

  // Load existing clinic draft if available
  const { data: existingClinicRes } = useQuery({
    queryKey: ['myClinicOnboarding'],
    queryFn: async () => {
      try {
        const res = await clinicApi.getMyClinic();
        return res.data?.data || null;
      } catch {
        return null;
      }
    },
    retry: 0,
  });

  useEffect(() => {
    if (existingClinicRes) {
      setClinicId(existingClinicRes.id);
      setName(existingClinicRes.name || '');
      setPhone(existingClinicRes.phone || '+998 ');
      setEmail(existingClinicRes.email || '');
      setDescription(existingClinicRes.description || '');
      setRegion(existingClinicRes.region || '');
      setDistrict(existingClinicRes.district || '');
      setAddress(existingClinicRes.address || '');
      setLocationUrl(existingClinicRes.location_url || '');
      setLogoUrl(existingClinicRes.logo_url || '');
      setCoverImageUrl(existingClinicRes.cover_image_url || '');
      if (existingClinicRes.working_hours && existingClinicRes.working_hours.length > 0) {
        setWorkingHours(existingClinicRes.working_hours);
      }
      if (existingClinicRes.status !== 'draft') {
        router.replace('/clinic/dashboard');
      }
    }
  }, [existingClinicRes, router]);

  // Save draft mutation
  const saveDraftMutation = useMutation({
    mutationFn: async () => {
      const payload: CreateClinicPayload = {
        name,
        phone,
        email: email || undefined,
        description: description || undefined,
        region: region || 'Toshkent shahri',
        district: district || 'Yunusobod',
        address: address || 'Kiritilmagan',
        location_url: locationUrl.trim() || undefined,
        logo_url: logoUrl || undefined,
        cover_image_url: coverImageUrl || undefined,
        working_hours: workingHours,
      };

      if (clinicId) {
        const res = await clinicApi.updateClinic(clinicId, payload);
        return res.data?.data;
      } else {
        const res = await clinicApi.createClinic(payload);
        return res.data?.data;
      }
    },
    onSuccess: (data) => {
      if (data?.id) {
        setClinicId(data.id);
      }
      queryClient.invalidateQueries({ queryKey: ['myClinic'] });
    },
  });

  // Submit mutation
  const submitMutation = useMutation({
    mutationFn: async () => {
      const saved = await saveDraftMutation.mutateAsync();
      const targetId = saved?.id || clinicId;
      if (!targetId) throw new Error('Klinika ID topilmadi');
      const res = await clinicApi.submitClinic(targetId);
      return res.data;
    },
    onSuccess: () => {
      toast.success('Klinikangiz tasdiqlash uchun superadminga yuborildi!');
      queryClient.invalidateQueries({ queryKey: ['myClinic'] });
      router.replace('/clinic/dashboard');
    },
    onError: (err: unknown) => {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      toast.error(
        axiosErr.response?.data?.message || 'Arizani yuborishda xatolik yuz berdi'
      );
    },
  });

  const handleFileUpload = async (file: File, type: 'logo' | 'cover') => {
    if (file.size > 5 * 1024 * 1024) {
      toast.error('Rasm hajmi 5MB dan oshmasligi kerak');
      return;
    }

    if (type === 'logo') setIsUploadingLogo(true);
    else setIsUploadingCover(true);

    try {
      const res = await clinicApi.uploadImage(file, 'clinics');
      if (res.data?.url) {
        if (type === 'logo') setLogoUrl(res.data.url);
        else setCoverImageUrl(res.data.url);
        toast.success('Rasm muvaffaqiyatli yuklandi');
      }
    } catch {
      toast.error('Rasmni yuklashda xatolik yuz berdi');
    } finally {
      if (type === 'logo') setIsUploadingLogo(false);
      else setIsUploadingCover(false);
    }
  };

  const validateStep = (step: number): boolean => {
    setFormError(null);
    if (step === 1) {
      if (!name.trim() || name.length < 2) {
        setFormError('Klinika nomi kamida 2 ta belgidan iborat bo‘lishi kerak');
        return false;
      }
      if (!isValidUzbekPhone(phone)) {
        setFormError('Telefon raqami noto‘g‘ri. Masalan: +998 71 200 11 22');
        return false;
      }
    } else if (step === 2) {
      if (!region) {
        setFormError('Viloyatni tanlang');
        return false;
      }
      if (!district) {
        setFormError('Tumanni tanlang');
        return false;
      }
      if (!address.trim() || address.length < 5) {
        setFormError('Aniq manzilni kiriting (kamida 5 ta belgi)');
        return false;
      }
      if (locationUrl.trim() && !isValidMapLocationUrl(locationUrl.trim())) {
        setLocationUrlError('Faqat Google Maps yoki Yandex Maps lokatsiya linki kiritilishi shart');
        setFormError('Xaritadagi lokatsiya havolasi noto‘g‘ri formatda');
        return false;
      }
      setLocationUrlError(null);
    }
    return true;
  };

  const handleNext = async () => {
    if (!validateStep(currentStep)) return;
    try {
      await saveDraftMutation.mutateAsync();
      setCurrentStep((prev) => Math.min(prev + 1, 5));
    } catch (err) {
      console.error('Draft autosave error', err);
      setCurrentStep((prev) => Math.min(prev + 1, 5));
    }
  };

  const handlePrev = () => {
    setFormError(null);
    setCurrentStep((prev) => Math.max(prev - 1, 1));
  };

  const steps = [
    { num: 1, title: 'Asosiy', icon: Building2 },
    { num: 2, title: 'Joylashuv', icon: MapPin },
    { num: 3, title: 'Ish vaqti', icon: Clock },
    { num: 4, title: 'Rasmlar', icon: ImageIcon },
    { num: 5, title: 'Tekshirish', icon: CheckCircle },
  ];

  return (
    <div className="min-h-screen bg-[#F7FAFC] flex flex-col font-sans">
      {/* Header */}
      <header className="bg-white border-b border-slate-200 py-4 px-6 sticky top-0 z-30 shadow-xs">
        <div className="max-w-4xl mx-auto flex items-center justify-between">
          <BrandLogo size="md" subtitle="Onboarding" />
          <div className="flex items-center gap-2 text-xs text-slate-500 font-medium">
            <span>Bosqich {currentStep} / 5</span>
          </div>
        </div>
      </header>

      {/* Main Container */}
      <div className="flex-1 max-w-4xl w-full mx-auto p-4 sm:p-6 lg:p-8">
        {/* Stepper Card */}
        <div className="bg-white rounded-2xl border border-slate-200 p-4 sm:p-6 shadow-xs mb-8">
          <div className="grid grid-cols-5 gap-2 sm:gap-4">
            {steps.map((s) => {
              const Icon = s.icon;
              const isPassed = currentStep > s.num;
              const isCurrent = currentStep === s.num;
              return (
                <div key={s.num} className="flex flex-col items-center text-center">
                  <div
                    className={`w-9 h-9 sm:w-11 sm:h-11 rounded-full flex items-center justify-center transition-all ${
                      isPassed
                        ? 'bg-emerald-500 text-white shadow-xs'
                        : isCurrent
                        ? 'bg-[#0866F5] text-white ring-4 ring-blue-100 shadow-sm'
                        : 'bg-slate-100 text-slate-400'
                    }`}
                  >
                    {isPassed ? (
                      <Check className="w-5 h-5" />
                    ) : (
                      <Icon className="w-4 h-4 sm:w-5 sm:h-5" />
                    )}
                  </div>
                  <span
                    className={`text-[11px] sm:text-xs font-semibold mt-2 hidden sm:block ${
                      isCurrent ? 'text-slate-900' : 'text-slate-400'
                    }`}
                  >
                    {s.title}
                  </span>
                </div>
              );
            })}
          </div>
        </div>

        {/* Step Content */}
        <div className="bg-white rounded-2xl border border-slate-200 p-6 sm:p-8 shadow-sm">
          {formError && (
            <div className="mb-6">
              <Alert variant="error" title="Diqqat" message={formError} />
            </div>
          )}

          {/* 1-BOSQICH: ASOSIY MA'LUMOTLAR */}
          {currentStep === 1 && (
            <div className="space-y-5">
              <div>
                <h3 className="text-lg font-bold text-slate-900">1. Asosiy ma’lumotlar</h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  Klinikangizning rasmiy nomi va bog‘lanish vositalari
                </p>
              </div>

              <Input
                label="Klinika rasmiy nomi *"
                placeholder="Masalan: Shifo Nur Med Clinic"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Input
                  label="Telefon raqami *"
                  placeholder="+998 71 200 11 22"
                  value={phone}
                  onChange={(e) => setPhone(formatPhone(e.target.value))}
                />
                <Input
                  type="email"
                  label="Klinika emaili (ixtiyoriy)"
                  placeholder="info@shifonur.uz"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>

              <div>
                <label className="text-xs font-semibold text-slate-700 block mb-1.5">
                  Klinika haqida qisqacha tavsif
                </label>
                <textarea
                  rows={4}
                  className="w-full px-3.5 py-2.5 rounded-xl border border-slate-300 text-sm focus:outline-hidden focus:border-[#0866F5] focus:ring-2 focus:ring-blue-100 transition-all resize-none"
                  placeholder="Klinika yo‘nalishlari, zamonaviy laboratoriya va afzalliklari haqida yozing..."
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                />
              </div>
            </div>
          )}

          {/* 2-BOSQICH: JOYLASHUV */}
          {currentStep === 2 && (
            <div className="space-y-5">
              <div>
                <h3 className="text-lg font-bold text-slate-900">2. Joylashuv ma’lumotlari</h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  Bemorlar sizni oson topishi uchun aniq manzilni belgilang
                </p>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="text-xs font-semibold text-slate-700 block mb-1.5">
                    Viloyat *
                  </label>
                  <select
                    className="w-full h-11 px-3.5 rounded-xl border border-slate-300 text-sm focus:outline-hidden focus:border-[#0866F5] focus:ring-2 focus:ring-blue-100 bg-white"
                    value={region}
                    onChange={(e) => {
                      setRegion(e.target.value);
                      setDistrict('');
                    }}
                  >
                    <option value="">Viloyatni tanlang</option>
                    {Object.keys(UZBEKISTAN_REGIONS).map((reg) => (
                      <option key={reg} value={reg}>
                        {reg}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="text-xs font-semibold text-slate-700 block mb-1.5">
                    Tuman / Shahar *
                  </label>
                  <select
                    className="w-full h-11 px-3.5 rounded-xl border border-slate-300 text-sm focus:outline-hidden focus:border-[#0866F5] focus:ring-2 focus:ring-blue-100 bg-white"
                    value={district}
                    disabled={!region}
                    onChange={(e) => setDistrict(e.target.value)}
                  >
                    <option value="">Tumanni tanlang</option>
                    {region &&
                      UZBEKISTAN_REGIONS[region]?.map((dist) => (
                        <option key={dist} value={dist}>
                          {dist}
                        </option>
                      ))}
                  </select>
                </div>
              </div>

              <Input
                label="Aniq manzil (ko‘cha, uy, mo‘ljal) *"
                placeholder="Masalan: Amir Temur ko‘chasi, 45-uy. Mo‘ljal: Minor metro bekati"
                value={address}
                onChange={(e) => setAddress(e.target.value)}
              />

              <div>
                <label className="text-xs font-semibold text-slate-700 block mb-1.5">
                  Xarita lokatsiya havolasi (Google Maps yoki Yandex Maps)
                </label>
                <input
                  type="url"
                  placeholder="https://maps.app.goo.gl/... yoki https://yandex.uz/maps/..."
                  className={`w-full h-11 px-3.5 rounded-xl border text-sm transition-all focus:outline-hidden ${
                    locationUrlError
                      ? 'border-red-400 bg-red-50/30 text-red-900 focus:border-red-500 focus:ring-2 focus:ring-red-100'
                      : 'border-slate-300 focus:border-[#0866F5] focus:ring-2 focus:ring-blue-100 bg-white'
                  }`}
                  value={locationUrl}
                  onChange={(e) => {
                    const val = e.target.value;
                    setLocationUrl(val);
                    if (val.trim() && !isValidMapLocationUrl(val)) {
                      setLocationUrlError('Faqat Google Maps yoki Yandex Maps linki qabul qilinadi');
                    } else {
                      setLocationUrlError(null);
                    }
                  }}
                />
                {locationUrlError ? (
                  <p className="text-xs text-red-600 mt-1 font-medium">{locationUrlError}</p>
                ) : (
                  <p className="text-[11px] text-slate-500 mt-1">
                    Google Maps yoki Yandex Maps ilovasida klinikangiz nuqtasini belgilab, &quot;Ulashish (Поделиться)&quot; tugmasi orqali havolani nusxalab qo‘ying.
                  </p>
                )}
              </div>
            </div>
          )}

          {/* 3-BOSQICH: ISH VAQTI */}
          {currentStep === 3 && (
            <div className="space-y-5">
              <div>
                <h3 className="text-lg font-bold text-slate-900">3. Haftalik ish vaqti</h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  Har bir hafta kuni bo‘yicha qabul soatlarini kiriting
                </p>
              </div>

              <div className="space-y-3 divide-y divide-slate-100">
                {workingHours.map((wh) => {
                  const dayName = WEEKDAYS.find((d) => d.id === wh.weekday)?.name || '';
                  return (
                    <div
                      key={wh.weekday}
                      className="pt-3 first:pt-0 flex flex-col sm:flex-row sm:items-center justify-between gap-3"
                    >
                      <div className="w-32 font-semibold text-sm text-slate-800">
                        {dayName}
                      </div>

                      <div className="flex items-center gap-4">
                        <label className="flex items-center gap-2 cursor-pointer text-xs text-slate-600">
                          <input
                            type="checkbox"
                            checked={wh.is_closed}
                            onChange={(e) => {
                              const updated = workingHours.map((item) =>
                                item.weekday === wh.weekday
                                  ? { ...item, is_closed: e.target.checked }
                                  : item
                              );
                              setWorkingHours(updated);
                            }}
                            className="rounded border-slate-300 text-rose-500 focus:ring-rose-400 h-4 w-4"
                          />
                          Dam olish kuni
                        </label>

                        {!wh.is_closed ? (
                          <div className="flex items-center gap-2">
                            <input
                              type="time"
                              value={wh.opening_time}
                              onChange={(e) => {
                                const updated = workingHours.map((item) =>
                                  item.weekday === wh.weekday
                                    ? { ...item, opening_time: e.target.value }
                                    : item
                                );
                                setWorkingHours(updated);
                              }}
                              className="px-2.5 py-1.5 border border-slate-300 rounded-lg text-xs font-mono"
                            />
                            <span className="text-xs text-slate-400">—</span>
                            <input
                              type="time"
                              value={wh.closing_time}
                              onChange={(e) => {
                                const updated = workingHours.map((item) =>
                                  item.weekday === wh.weekday
                                    ? { ...item, closing_time: e.target.value }
                                    : item
                                );
                                setWorkingHours(updated);
                              }}
                              className="px-2.5 py-1.5 border border-slate-300 rounded-lg text-xs font-mono"
                            />
                          </div>
                        ) : (
                          <span className="text-xs text-rose-500 font-medium px-2 py-1 bg-rose-50 rounded-md">
                            Yopiq
                          </span>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* 4-BOSQICH: RASMLAR */}
          {currentStep === 4 && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-bold text-slate-900">4. Rasmlar va logotip</h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  Klinika brendingizni namoyish qilish uchun logotip va muqova rasmini yuklang
                </p>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div className="border border-slate-200 rounded-2xl p-4 flex flex-col items-center justify-center text-center bg-slate-50/50">
                  <span className="text-xs font-semibold text-slate-700 mb-3">
                    Klinika logotipi (PNG, JPG, max 5MB)
                  </span>

                  {logoUrl ? (
                    <div className="relative mb-3 group">
                      <img
                        src={getMediaUrl(logoUrl)}
                        alt="Klinika logotipi"
                        className="w-24 h-24 object-contain rounded-xl border border-slate-200 bg-white p-1"
                      />
                    </div>
                  ) : (
                    <div className="w-24 h-24 rounded-xl border-2 border-dashed border-slate-300 flex items-center justify-center text-slate-400 mb-3 bg-white">
                      <ImageIcon className="w-8 h-8 opacity-40" />
                    </div>
                  )}

                  <label className="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 bg-white border border-slate-300 rounded-lg text-xs font-medium text-slate-700 hover:bg-slate-50 shadow-2xs">
                    {isUploadingLogo ? (
                      <Loader2 className="w-3.5 h-3.5 animate-spin" />
                    ) : (
                      <UploadCloud className="w-3.5 h-3.5" />
                    )}
                    <span>{logoUrl ? 'Almashtirish' : 'Logotip yuklash'}</span>
                    <input
                      type="file"
                      accept="image/png,image/jpeg,image/webp"
                      className="hidden"
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file) handleFileUpload(file, 'logo');
                      }}
                    />
                  </label>
                </div>

                <div className="border border-slate-200 rounded-2xl p-4 flex flex-col items-center justify-center text-center bg-slate-50/50">
                  <span className="text-xs font-semibold text-slate-700 mb-3">
                    Muqova rasmi (bino ko‘rinishi)
                  </span>

                  {coverImageUrl ? (
                    <div className="relative mb-3">
                      <img
                        src={getMediaUrl(coverImageUrl)}
                        alt="Muqova rasmi"
                        className="w-40 h-24 object-cover rounded-xl border border-slate-200 bg-white"
                      />
                    </div>
                  ) : (
                    <div className="w-40 h-24 rounded-xl border-2 border-dashed border-slate-300 flex items-center justify-center text-slate-400 mb-3 bg-white">
                      <Building2 className="w-8 h-8 opacity-40" />
                    </div>
                  )}

                  <label className="cursor-pointer inline-flex items-center gap-1.5 px-3 py-1.5 bg-white border border-slate-300 rounded-lg text-xs font-medium text-slate-700 hover:bg-slate-50 shadow-2xs">
                    {isUploadingCover ? (
                      <Loader2 className="w-3.5 h-3.5 animate-spin" />
                    ) : (
                      <UploadCloud className="w-3.5 h-3.5" />
                    )}
                    <span>{coverImageUrl ? 'Almashtirish' : 'Rasm yuklash'}</span>
                    <input
                      type="file"
                      accept="image/png,image/jpeg,image/webp"
                      className="hidden"
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file) handleFileUpload(file, 'cover');
                      }}
                    />
                  </label>
                </div>
              </div>
            </div>
          )}

          {/* 5-BOSQICH: TEKSHIRISH VA YUBORISH */}
          {currentStep === 5 && (
            <div className="space-y-6">
              <div>
                <h3 className="text-lg font-bold text-slate-900">
                  5. Ma’lumotlarni tekshirish va tasdiqlash
                </h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  Barcha ma’lumotlarni sinchiklab ko‘rib chiqing. Yuborilgach, superadmin tomonidan
                  tekshiriladi.
                </p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
                <div className="p-4 bg-slate-50 rounded-xl space-y-2 border border-slate-100">
                  <div className="font-semibold text-slate-800 text-sm border-b border-slate-200 pb-1.5">
                    Asosiy ma’lumotlar
                  </div>
                  <div>
                    <span className="text-slate-500">Klinika nomi: </span>
                    <span className="font-semibold text-slate-900">{name}</span>
                  </div>
                  <div>
                    <span className="text-slate-500">Telefon: </span>
                    <span className="font-semibold text-slate-900">{phone}</span>
                  </div>
                  <div>
                    <span className="text-slate-500">Email: </span>
                    <span className="font-semibold text-slate-900">
                      {email || 'Belgilanmagan'}
                    </span>
                  </div>
                  {description && (
                    <div className="text-slate-600 italic line-clamp-2 pt-1 border-t border-slate-200">
                      "{description}"
                    </div>
                  )}
                </div>

                <div className="p-4 bg-slate-50 rounded-xl space-y-2 border border-slate-100">
                  <div className="font-semibold text-slate-800 text-sm border-b border-slate-200 pb-1.5">
                    Joylashuv va Rasmlar
                  </div>
                  <div>
                    <span className="text-slate-500">Hudud: </span>
                    <span className="font-semibold text-slate-900">
                      {district}, {region}
                    </span>
                  </div>
                  <div>
                    <span className="text-slate-500">Manzil: </span>
                    <span className="font-semibold text-slate-900">{address}</span>
                  </div>
                  <div className="flex items-center gap-3 pt-2">
                    {logoUrl ? (
                      <span className="inline-flex items-center gap-1 text-emerald-600 font-medium">
                        <Check className="w-3.5 h-3.5" /> Logotip yuklangan
                      </span>
                    ) : (
                      <span className="text-slate-400">Logotip yuklanmagan</span>
                    )}
                  </div>
                </div>
              </div>

              <div className="p-4 bg-blue-50/60 border border-blue-100 rounded-xl text-xs text-blue-900">
                <p className="font-semibold mb-1">Eslatma:</p>
                <p>
                  "Tasdiqlash uchun yuborish" tugmasini bosganingizdan so‘ng, arizangiz
                  ko‘rib chiqish bosqichiga o‘tadi (holat: <span className="font-bold">Pending</span>).
                  QabulGo moderatori arizani tekshirib tasdiqlagach, shifokorlar bazasini boshqarish
                  imkoniyati faollashadi.
                </p>
              </div>
            </div>
          )}

          {/* Navigation Controls */}
          <div className="flex items-center justify-between pt-6 border-t border-slate-100 mt-8">
            {currentStep > 1 ? (
              <Button
                variant="outline"
                size="md"
                onClick={handlePrev}
                className="gap-1.5"
                disabled={submitMutation.isPending}
              >
                <ArrowLeft className="w-4 h-4" />
                Orqaga
              </Button>
            ) : (
              <div />
            )}

            {currentStep < 5 ? (
              <Button
                variant="primary"
                size="md"
                onClick={handleNext}
                className="gap-1.5 bg-[#0866F5] hover:bg-blue-600"
                isLoading={saveDraftMutation.isPending}
              >
                Keyingisi
                <ArrowRight className="w-4 h-4" />
              </Button>
            ) : (
              <Button
                variant="primary"
                size="lg"
                onClick={() => submitMutation.mutate()}
                isLoading={submitMutation.isPending}
                className="gap-2 bg-emerald-600 hover:bg-emerald-700 text-white shadow-md"
              >
                <Send className="w-4 h-4" />
                Tasdiqlash uchun yuborish
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
