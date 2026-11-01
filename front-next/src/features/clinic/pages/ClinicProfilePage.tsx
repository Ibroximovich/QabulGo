'use client';

import { useState, useEffect } from 'react';
import { useClinicContext } from '@/components/layout/clinic/ClinicLayout';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  Building2,
  Image as ImageIcon,
  Save,
  UploadCloud,
  Loader2,
  MapPin,
  ExternalLink,
  Navigation,
} from 'lucide-react';
import { InteractiveMap } from '@/components/map/InteractiveMap';
import { UZBEKISTAN_REGIONS } from '@/lib/constants/regions';
import { isValidMapLocationUrl } from '@/lib/utils/location';
import { ClinicDetail, WorkingHourItem, UpdateClinicPayload } from '../types';
import { clinicApi } from '../api/clinicApi';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { formatPhone } from '@/lib/utils/phone';
import { getMediaUrl } from '@/lib/utils/media';

interface ClinicContextType {
  clinic?: ClinicDetail;
  isLoadingClinic: boolean;
}

const WEEKDAYS = [
  { id: 1, name: 'Dushanba' },
  { id: 2, name: 'Seshanba' },
  { id: 3, name: 'Chorshanba' },
  { id: 4, name: 'Payshanba' },
  { id: 5, name: 'Juma' },
  { id: 6, name: 'Shanba' },
  { id: 7, name: 'Yakshanba' },
];

export function ClinicProfilePage() {
  const { clinic, isLoadingClinic } = useClinicContext();
  const queryClient = useQueryClient();

  const [activeTab, setActiveTab] = useState<'info' | 'location' | 'hours' | 'images'>('info');

  const [name, setName] = useState('');
  const [phone, setPhone] = useState('');
  const [email, setEmail] = useState('');
  const [description, setDescription] = useState('');

  const [region, setRegion] = useState('');
  const [district, setDistrict] = useState('');
  const [address, setAddress] = useState('');
  const [locationUrl, setLocationUrl] = useState('');
  const [locationUrlError, setLocationUrlError] = useState<string | null>(null);
  const [latitude, setLatitude] = useState<number | null>(null);
  const [longitude, setLongitude] = useState<number | null>(null);

  const [workingHours, setWorkingHours] = useState<WorkingHourItem[]>([]);

  const [logoUrl, setLogoUrl] = useState('');
  const [coverImageUrl, setCoverImageUrl] = useState('');
  const [isUploadingLogo, setIsUploadingLogo] = useState(false);
  const [isUploadingCover, setIsUploadingCover] = useState(false);

  useEffect(() => {
    if (clinic) {
      setName(clinic.name || '');
      setPhone(clinic.phone || '');
      setEmail(clinic.email || '');
      setDescription(clinic.description || '');
      setRegion(clinic.region || '');
      setDistrict(clinic.district || '');
      setAddress(clinic.address || '');
      setLocationUrl(clinic.location_url || '');
      setLatitude(typeof clinic.latitude === 'number' ? clinic.latitude : null);
      setLongitude(typeof clinic.longitude === 'number' ? clinic.longitude : null);
      setLogoUrl(clinic.logo_url || '');
      setCoverImageUrl(clinic.cover_image_url || '');
      if (clinic.working_hours && clinic.working_hours.length > 0) {
        setWorkingHours(clinic.working_hours);
      }
    }
  }, [clinic]);

  const updateMutation = useMutation({
    mutationFn: async () => {
      if (!clinic?.id) throw new Error('Klinika ID topilmadi');
      if (locationUrl && !isValidMapLocationUrl(locationUrl)) {
        setLocationUrlError('Faqat haqiqiy Google Maps yoki Yandex Maps lokatsiya linki qabul qilinadi');
        throw new Error('Xarita lokatsiyasi linki noto‘g‘ri kiritilgan');
      }
      const payload: UpdateClinicPayload = {
        name,
        phone,
        email: email || undefined,
        description: description || undefined,
        region,
        district,
        address,
        location_url: locationUrl ? locationUrl.trim() : undefined,
        latitude: latitude !== null ? latitude : undefined,
        longitude: longitude !== null ? longitude : undefined,
        logo_url: logoUrl || undefined,
        cover_image_url: coverImageUrl || undefined,
        working_hours: workingHours,
      };
      const res = await clinicApi.updateClinic(clinic.id, payload);
      return res.data;
    },
    onSuccess: () => {
      toast.success('Klinika ma’lumotlari muvaffaqiyatli yangilandi');
      queryClient.invalidateQueries({ queryKey: ['myClinic'] });
    },
    onError: (err: unknown) => {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      toast.error(
        axiosErr.response?.data?.message || 'Ma’lumotlarni saqlashda xatolik yuz berdi'
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
        toast.success('Rasm yuklandi');
      }
    } catch {
      toast.error('Rasmni yuklashda xatolik yuz berdi');
    } finally {
      if (type === 'logo') setIsUploadingLogo(false);
      else setIsUploadingCover(false);
    }
  };

  if (isLoadingClinic) {
    return <div className="p-8 text-center text-slate-500">Yuklanmoqda...</div>;
  }

  return (
    <div className="space-y-6">
      {/* Page Title & Save Action */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-xl font-extrabold text-slate-900">Klinika profili</h2>
          <p className="text-xs text-slate-500 mt-0.5">
            Klinika rekvizitlari, aloqa vositalari va ish soatlarini boshqarish
          </p>
        </div>

        <Button
          variant="primary"
          onClick={() => updateMutation.mutate()}
          isLoading={updateMutation.isPending}
          className="bg-[#0866F5] hover:bg-blue-600 gap-1.5 shadow-xs"
        >
          <Save className="w-4 h-4" />
          O‘zgarishlarni saqlash
        </Button>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b border-slate-200 pb-2">
        <button
          onClick={() => setActiveTab('info')}
          className={`px-4 py-2 text-xs font-semibold rounded-xl transition-all ${
            activeTab === 'info'
              ? 'bg-[#0866F5] text-white shadow-2xs'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          Asosiy
        </button>
        <button
          onClick={() => setActiveTab('location')}
          className={`px-4 py-2 text-xs font-semibold rounded-xl transition-all ${
            activeTab === 'location'
              ? 'bg-[#0866F5] text-white shadow-2xs'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          Joylashuv
        </button>
        <button
          onClick={() => setActiveTab('hours')}
          className={`px-4 py-2 text-xs font-semibold rounded-xl transition-all ${
            activeTab === 'hours'
              ? 'bg-[#0866F5] text-white shadow-2xs'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          Ish vaqti
        </button>
        <button
          onClick={() => setActiveTab('images')}
          className={`px-4 py-2 text-xs font-semibold rounded-xl transition-all ${
            activeTab === 'images'
              ? 'bg-[#0866F5] text-white shadow-2xs'
              : 'text-slate-600 hover:bg-slate-100'
          }`}
        >
          Rasmlar
        </button>
      </div>

      {/* Tab Panels */}
      <div className="bg-white rounded-2xl border border-slate-200 p-6 sm:p-8 shadow-xs">
        {activeTab === 'info' && (
          <div className="space-y-4 max-w-2xl">
            <Input
              label="Klinika nomi"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <Input
                label="Telefon"
                value={phone}
                onChange={(e) => setPhone(formatPhone(e.target.value))}
              />
              <Input
                type="email"
                label="Email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div>
              <label className="text-xs font-semibold text-slate-700 block mb-1.5">
                Klinika haqida tavsif
              </label>
              <textarea
                rows={4}
                className="w-full px-3.5 py-2.5 rounded-xl border border-slate-300 text-sm focus:outline-hidden focus:border-[#0866F5] focus:ring-2 focus:ring-blue-100 resize-none"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
              />
            </div>
          </div>
        )}

        {activeTab === 'location' && (
          <div className="space-y-5 max-w-2xl">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="text-xs font-semibold text-slate-700 block mb-1.5">
                  Viloyat *
                </label>
                <select
                  className="w-full h-11 px-3.5 rounded-xl border border-slate-300 text-sm focus:outline-hidden focus:border-[#0866F5] focus:ring-2 focus:ring-blue-100 bg-white"
                  value={region}
                  onChange={(e) => {
                    const newReg = e.target.value;
                    setRegion(newReg);
                    setDistrict(UZBEKISTAN_REGIONS[newReg]?.[0] || '');
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
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              placeholder="Masalan: Amir Temur ko‘chasi, 45-uy"
            />

            <div>
              <Input
                label="Xarita lokatsiya havolasi (Google Maps yoki Yandex Maps)"
                placeholder="https://maps.app.goo.gl/... yoki https://yandex.uz/maps/..."
                leftIcon={<MapPin className="w-4 h-4 text-slate-400" />}
                value={locationUrl}
                error={locationUrlError || undefined}
                onChange={(e) => {
                  const val = e.target.value;
                  setLocationUrl(val);
                  if (!val || isValidMapLocationUrl(val)) {
                    setLocationUrlError(null);
                  } else {
                    setLocationUrlError('Faqat haqiqiy Google Maps yoki Yandex Maps lokatsiya linki bo‘lishi kerak');
                  }
                }}
                helperText="Mijozlar klinikangizni xaritada oson topishi uchun Google Maps yoki Yandex Maps’dan ulashish (Share link) havolasini kiriting."
              />

              {locationUrl && isValidMapLocationUrl(locationUrl) && (
                <div className="mt-2 flex items-center gap-2">
                  <a
                    href={locationUrl.startsWith('http') ? locationUrl : `https://${locationUrl}`}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-flex items-center gap-1.5 text-xs font-semibold text-[#0866F5] hover:underline"
                  >
                    <ExternalLink className="w-3.5 h-3.5" />
                    <span>Havolani xaritada ochib ko‘rish</span>
                  </a>
                </div>
              )}

              {/* Interactive Map Picker for Clinic Coordinates */}
              <div className="mt-6 pt-6 border-t border-slate-100 space-y-3">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                  <div>
                    <label className="text-xs font-bold text-brand-navy flex items-center gap-1.5">
                      <MapPin className="w-4 h-4 text-brand-blue" />
                      <span>Klinika koordinatalari (Xaritada belgilash)</span>
                    </label>
                    <p className="text-[11px] text-slate-500 mt-0.5">
                      Xaritadagi nuqtani bosing yoki markerni surib aniq lokatsiyani belgilang.
                    </p>
                  </div>

                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => {
                        if (!navigator.geolocation) {
                          toast.error('Brauzer geolokatsiyani qo‘llab-quvvatlamaydi');
                          return;
                        }
                        navigator.geolocation.getCurrentPosition(
                          (pos) => {
                            setLatitude(parseFloat(pos.coords.latitude.toFixed(6)));
                            setLongitude(parseFloat(pos.coords.longitude.toFixed(6)));
                            toast.success('Joriy joylashuv koordinatalari o‘rnatildi');
                          },
                          () => {
                            toast.error('Geolokatsiyani aniqlashga ruxsat berilmadi');
                          }
                        );
                      }}
                      className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-teal-50 text-brand-teal text-xs font-semibold hover:bg-teal-100 transition-colors border border-teal-100"
                    >
                      <Navigation className="w-3.5 h-3.5" />
                      <span>Joriy lokatsiyamni belgilash</span>
                    </button>

                    {latitude !== null && longitude !== null && (
                      <button
                        type="button"
                        onClick={() => {
                          setLatitude(null);
                          setLongitude(null);
                        }}
                        className="text-xs text-red-500 hover:text-red-700 font-medium px-2 py-1"
                      >
                        Tozalash
                      </button>
                    )}
                  </div>
                </div>

                <div className="rounded-2xl overflow-hidden border border-slate-200">
                  <InteractiveMap
                    pickerMode={true}
                    pickedLocation={
                      latitude !== null && longitude !== null ? { latitude, longitude } : null
                    }
                    onLocationPick={(coords) => {
                      setLatitude(coords.latitude);
                      setLongitude(coords.longitude);
                    }}
                    height="320px"
                    zoom={latitude !== null && longitude !== null ? 15 : 12}
                    center={
                      latitude !== null && longitude !== null
                        ? [latitude, longitude]
                        : [41.311081, 69.240562]
                    }
                  />
                </div>

                {latitude !== null && longitude !== null && (
                  <div className="flex items-center gap-4 text-xs font-mono bg-slate-50 p-2.5 rounded-xl border border-slate-100 text-slate-700">
                    <span>
                      <strong>Kenglik (Lat):</strong> {latitude}
                    </span>
                    <span>
                      <strong>Uzunlik (Lon):</strong> {longitude}
                    </span>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'hours' && (
          <div className="space-y-4 max-w-2xl">
            <div className="space-y-3 divide-y divide-slate-100">
              {workingHours.map((wh) => {
                const dayName = WEEKDAYS.find((d) => d.id === wh.weekday)?.name || '';
                return (
                  <div
                    key={wh.weekday}
                    className="pt-3 first:pt-0 flex items-center justify-between gap-4"
                  >
                    <span className="w-28 font-semibold text-xs text-slate-800">
                      {dayName}
                    </span>
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
                        className="rounded border-slate-300 text-rose-500"
                      />
                      Dam olish
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
                          className="px-2 py-1 border border-slate-300 rounded-lg text-xs"
                        />
                        <span className="text-slate-400">—</span>
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
                        className="px-2 py-1 border border-slate-300 rounded-lg text-xs"
                      />
                      </div>
                    ) : (
                      <span className="text-xs text-rose-500 font-medium px-2 py-1 bg-rose-50 rounded">
                        Yopiq
                      </span>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {activeTab === 'images' && (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6 max-w-2xl">
            <div className="border border-slate-200 rounded-2xl p-4 flex flex-col items-center justify-center text-center bg-slate-50/50">
              <span className="text-xs font-semibold text-slate-700 mb-3">Logotip</span>
              {logoUrl ? (
                <img
                  src={getMediaUrl(logoUrl)}
                  alt="Logo"
                  className="w-24 h-24 object-contain rounded-xl border border-slate-200 bg-white p-1 mb-3"
                />
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
                <span>Yuklash</span>
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
              <span className="text-xs font-semibold text-slate-700 mb-3">Muqova rasmi</span>
              {coverImageUrl ? (
                <img
                  src={getMediaUrl(coverImageUrl)}
                  alt="Cover"
                  className="w-40 h-24 object-cover rounded-xl border border-slate-200 bg-white mb-3"
                />
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
                <span>Yuklash</span>
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
        )}
      </div>
    </div>
  );
}
