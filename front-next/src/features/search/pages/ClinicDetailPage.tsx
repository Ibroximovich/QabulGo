'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';
import { AppLayout } from '@/components/layout/AppLayout';
import { InteractiveMap } from '@/components/map/InteractiveMap';
import { searchApi } from '../api/searchApi';
import { ClinicDetail, WorkingHourItem } from '@/features/clinic/types';
import { PublicClinicDoctorItem } from '../types';
import { getMediaUrl } from '@/lib/utils/media';
import { formatCurrency } from '@/lib/utils/formatters';
import { extractCoordinatesFromUrl, REGION_DEFAULT_COORDINATES } from '@/lib/utils/location';
import {
  Building2,
  MapPin,
  Phone,
  Mail,
  Clock,
  ShieldCheck,
  Stethoscope,
  Calendar,
  ChevronLeft,
  Users,
  ExternalLink,
  ArrowRight,
  AlertCircle,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';

const DAYS_OF_WEEK = [
  'Dushanba',
  'Seshanba',
  'Chorshanba',
  'Payshanba',
  'Juma',
  'Shanba',
  'Yakshanba',
];

interface ClinicDetailPageProps {
  initialClinic?: ClinicDetail | null;
  initialDoctors?: PublicClinicDoctorItem[];
}

export function ClinicDetailPage({ initialClinic, initialDoctors }: ClinicDetailPageProps) {
  const params = useParams();
  const router = useRouter();
  const clinicId = params?.id as string;

  const [clinic, setClinic] = useState<ClinicDetail | null>(initialClinic || null);
  const [doctors, setDoctors] = useState<PublicClinicDoctorItem[]>(initialDoctors || []);
  const [isLoading, setIsLoading] = useState<boolean>(!initialClinic);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (initialClinic && initialDoctors) return;
    if (!clinicId) return;

    let active = true;

    async function loadData() {
      setIsLoading(true);
      setError(null);

      try {
        const [cRes, dRes] = await Promise.all([
          searchApi.getClinicDetails(clinicId),
          searchApi.getClinicDoctors(clinicId),
        ]);

        if (active) {
          if (cRes.data) {
            setClinic(cRes.data);
          } else {
            setError('Klinika topilmadi yoki tasdiqlanmagan.');
          }
          if (dRes.data) {
            setDoctors(dRes.data);
          }
        }
      } catch (err: any) {
        if (active) {
          console.error('Klinika ma’lumotlarini yuklashda xatolik:', err);
          setError('Klinika topilmadi yoki yuklashda xatolik yuz berdi.');
        }
      } finally {
        if (active) {
          setIsLoading(false);
        }
      }
    }

    loadData();

    return () => {
      active = false;
    };
  }, [clinicId, initialClinic, initialDoctors]);

  if (isLoading) {
    return (
      <AppLayout>
        <div className="max-w-5xl mx-auto space-y-6 animate-pulse">
          <div className="h-64 bg-slate-200 rounded-3xl w-full" />
          <div className="h-10 bg-slate-200 rounded-xl w-1/3" />
          <div className="h-4 bg-slate-200 rounded-md w-1/2" />
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 pt-6">
            <div className="md:col-span-2 h-96 bg-slate-200 rounded-2xl" />
            <div className="h-96 bg-slate-200 rounded-2xl" />
          </div>
        </div>
      </AppLayout>
    );
  }

  if (error || !clinic) {
    return (
      <AppLayout>
        <div className="max-w-md mx-auto my-16 text-center bg-white p-8 rounded-3xl border border-slate-200 shadow-subtle space-y-4">
          <div className="w-14 h-14 rounded-2xl bg-red-50 text-red-500 mx-auto flex items-center justify-center">
            <AlertCircle className="w-7 h-7" />
          </div>
          <h2 className="text-xl font-bold text-brand-navy">Klinika topilmadi</h2>
          <p className="text-sm text-slate-500">
            {error || 'Siz qidirayotgan klinika mavjud emas yoki admin tomonidan tasdiqlanish jarayonida.'}
          </p>
          <div className="pt-2">
            <Button onClick={() => router.push('/clinics')} className="gap-2">
              <ChevronLeft className="w-4 h-4" />
              <span>Klinikalar ro‘yxatiga qaytish</span>
            </Button>
          </div>
        </div>
      </AppLayout>
    );
  }

  const coverUrl = clinic.cover_image_url ? getMediaUrl(clinic.cover_image_url) : null;
  const logoUrl = clinic.logo_url ? getMediaUrl(clinic.logo_url) : null;
  // Koordinatalarni aniqlash: agar aniq latitude/longitude bo'lmasa, location_url yoki viloyatdan olish
  let effectiveLat =
    typeof clinic.latitude === 'number' && !isNaN(clinic.latitude) ? clinic.latitude : null;
  let effectiveLon =
    typeof clinic.longitude === 'number' && !isNaN(clinic.longitude) ? clinic.longitude : null;

  if (effectiveLat === null || effectiveLon === null) {
    const fromUrl = extractCoordinatesFromUrl(clinic.location_url);
    if (fromUrl) {
      effectiveLat = fromUrl.latitude;
      effectiveLon = fromUrl.longitude;
    } else if (clinic.region && REGION_DEFAULT_COORDINATES[clinic.region]) {
      effectiveLat = REGION_DEFAULT_COORDINATES[clinic.region].latitude;
      effectiveLon = REGION_DEFAULT_COORDINATES[clinic.region].longitude;
    }
  }

  const hasCoordinates =
    typeof effectiveLat === 'number' &&
    typeof effectiveLon === 'number' &&
    !isNaN(effectiveLat) &&
    !isNaN(effectiveLon);

  return (
    <AppLayout>
      <div className="max-w-5xl mx-auto space-y-8">
        {/* Breadcrumbs & Back Link */}
        <div className="flex items-center gap-2 text-xs font-semibold text-slate-500">
          <Link href="/clinics" className="hover:text-brand-blue flex items-center gap-1">
            <ChevronLeft className="w-3.5 h-3.5" />
            <span>Klinikalar</span>
          </Link>
          <span>/</span>
          <span className="text-brand-navy truncate max-w-xs">{clinic.name}</span>
        </div>

        {/* Hero Section: Cover, Logo, Name & Quick Info */}
        <div className="bg-white rounded-3xl border border-slate-200/80 shadow-subtle overflow-hidden">
          {/* Cover Banner */}
          <div className="relative h-60 sm:h-72 w-full bg-gradient-to-r from-brand-navy to-brand-blue/90 overflow-hidden">
            {coverUrl ? (
              <img
                src={coverUrl}
                alt={clinic.name}
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center text-white/20">
                <Building2 className="w-24 h-24 stroke-[1]" />
              </div>
            )}

            {/* Verified Badge */}
            <div className="absolute top-4 right-4 z-10">
              <span className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-white/95 backdrop-blur-md text-xs font-bold text-emerald-700 shadow-card border border-white">
                <ShieldCheck className="w-4 h-4 text-emerald-600" />
                <span>Tasdiqlangan klinika</span>
              </span>
            </div>
          </div>

          {/* Profile Header Bar */}
          <div className="p-6 sm:p-8 pt-0 relative">
            <div className="flex flex-col sm:flex-row items-start sm:items-end justify-between gap-4 -mt-12 sm:-mt-16 mb-6">
              {/* Logo */}
              <div className="relative z-10 w-24 h-24 sm:w-28 sm:h-28 rounded-3xl bg-white p-1.5 shadow-elevated border-2 border-white flex items-center justify-center overflow-hidden">
                {logoUrl ? (
                  <img src={logoUrl} alt="" className="w-full h-full object-contain rounded-2xl" />
                ) : (
                  <div className="w-full h-full rounded-2xl bg-brand-blue-50 text-brand-blue font-extrabold text-2xl flex items-center justify-center">
                    {clinic.name.charAt(0).toUpperCase()}
                  </div>
                )}
              </div>

              {/* Action buttons */}
              <div className="flex items-center gap-3 w-full sm:w-auto">
                {clinic.phone && (
                  <a
                    href={`tel:${clinic.phone}`}
                    className="flex-1 sm:flex-none inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl bg-brand-blue-50 text-brand-blue hover:bg-brand-blue-100 font-semibold text-xs transition-colors"
                  >
                    <Phone className="w-4 h-4" />
                    <span>Qo‘ng‘iroq qilish</span>
                  </a>
                )}
                {clinic.location_url && (
                  <a
                    href={clinic.location_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex-1 sm:flex-none inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl border border-slate-200 text-brand-navy hover:bg-slate-50 font-semibold text-xs transition-colors"
                  >
                    <ExternalLink className="w-4 h-4 text-slate-500" />
                    <span>Xaritada ochish</span>
                  </a>
                )}
              </div>
            </div>

            {/* Clinic Name & Main Description */}
            <div className="space-y-3">
              <h1 className="text-2xl sm:text-3xl font-extrabold text-brand-navy">{clinic.name}</h1>
              {clinic.description && (
                <p className="text-sm sm:text-base text-slate-600 leading-relaxed max-w-3xl">
                  {clinic.description}
                </p>
              )}

              {/* Meta pills */}
              <div className="flex flex-wrap items-center gap-y-2 gap-x-4 pt-2 text-xs text-slate-600">
                <span className="flex items-center gap-1.5">
                  <MapPin className="w-4 h-4 text-brand-blue shrink-0" />
                  <span>
                    {clinic.address ? `${clinic.address}, ` : ''}
                    {clinic.district ? `${clinic.district}, ` : ''}
                    {clinic.region}
                  </span>
                </span>
                {clinic.phone && (
                  <span className="flex items-center gap-1.5">
                    <Phone className="w-4 h-4 text-slate-400 shrink-0" />
                    <span>{clinic.phone}</span>
                  </span>
                )}
                {clinic.email && (
                  <span className="flex items-center gap-1.5">
                    <Mail className="w-4 h-4 text-slate-400 shrink-0" />
                    <span>{clinic.email}</span>
                  </span>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* 2-Column Content: Left Details / Right Schedule & Map */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
          {/* Left Column (Doctors & Specialties) */}
          <div className="lg:col-span-7 space-y-8">
            {/* Specialties Section */}
            {clinic.specialties && clinic.specialties.length > 0 && (
              <div className="bg-white rounded-3xl border border-slate-200/80 p-6 shadow-subtle space-y-4">
                <div className="flex items-center gap-2">
                  <Stethoscope className="w-5 h-5 text-brand-blue" />
                  <h2 className="text-lg font-bold text-brand-navy">Mavjud yo‘nalishlar va mutaxassisliklar</h2>
                </div>
                <div className="flex flex-wrap gap-2 pt-1">
                  {clinic.specialties.map((spec) => (
                    <span
                      key={spec.id}
                      className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-slate-100 text-slate-700 text-xs font-semibold border border-slate-200/70 hover:bg-brand-blue-50 hover:text-brand-blue transition-colors"
                    >
                      <span>{spec.name}</span>
                    </span>
                  ))}
                </div>
              </div>
            )}

            {/* Doctors in this Clinic */}
            <div id="doctors" className="bg-white rounded-3xl border border-slate-200/80 p-6 shadow-subtle space-y-6 scroll-mt-24">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Users className="w-5 h-5 text-brand-blue" />
                  <h2 className="text-lg font-bold text-brand-navy">Klinika shifokorlari</h2>
                </div>
                <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-slate-100 text-slate-600">
                  {doctors.length} nafar shifokor
                </span>
              </div>

              {doctors.length === 0 ? (
                <p className="text-xs text-slate-500 py-4 text-center">
                  Hozirda bu klinikaga biriktirilgan shifokorlar mavjud emas.
                </p>
              ) : (
                <div className="space-y-4">
                  {doctors.map((doc) => {
                    const docPhoto = doc.photo_url ? getMediaUrl(doc.photo_url) : null;
                    return (
                      <div
                        key={doc.id}
                        className="flex flex-col sm:flex-row items-start sm:items-center justify-between p-4 rounded-2xl bg-slate-50 border border-slate-100 hover:border-slate-200 transition-all gap-4"
                      >
                        <div className="flex items-center gap-3">
                          <div className="w-14 h-14 rounded-2xl bg-white overflow-hidden border border-slate-200 shrink-0 flex items-center justify-center">
                            {docPhoto ? (
                              <img src={docPhoto} alt="" className="w-full h-full object-cover" />
                            ) : (
                              <Stethoscope className="w-6 h-6 text-brand-blue" />
                            )}
                          </div>
                          <div>
                            <Link
                              href={`/doctors/${doc.doctor_id}`}
                              className="font-bold text-sm text-brand-navy hover:text-brand-blue hover:underline"
                            >
                              {doc.first_name} {doc.last_name}
                            </Link>
                            {doc.specialty && (
                              <p className="text-xs text-brand-blue font-medium mt-0.5">
                                {doc.specialty.name}
                              </p>
                            )}
                            {doc.experience_years > 0 && (
                              <p className="text-[11px] text-slate-400 mt-0.5">
                                {doc.experience_years} yil tajriba
                              </p>
                            )}
                          </div>
                        </div>

                        <div className="flex sm:flex-col items-center sm:items-end justify-between w-full sm:w-auto pt-2 sm:pt-0 border-t sm:border-t-0 border-slate-200/60 gap-2">
                          <span className="text-sm font-extrabold text-brand-navy">
                            {formatCurrency(doc.consultation_price)}
                          </span>
                          <Link
                            href={`/doctors/${doc.doctor_id}?clinicId=${clinic.id}`}
                            className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-brand-blue text-white hover:bg-brand-blue-600 text-xs font-bold shadow-xs transition-colors"
                          >
                            <Calendar className="w-3.5 h-3.5" />
                            <span>Qabulga yozilish</span>
                          </Link>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          </div>

          {/* Right Column (Working Hours & Interactive Map) */}
          <div className="lg:col-span-5 space-y-6">
            {/* Working Hours */}
            <div className="bg-white rounded-3xl border border-slate-200/80 p-6 shadow-subtle space-y-4">
              <div className="flex items-center gap-2">
                <Clock className="w-5 h-5 text-brand-blue" />
                <h3 className="font-bold text-base text-brand-navy">Ish vaqti jadvali</h3>
              </div>

              {clinic.working_hours && clinic.working_hours.length > 0 ? (
                <div className="space-y-2 text-xs">
                  {clinic.working_hours
                    .slice()
                    .sort((a, b) => {
                      const dayA = a.day_of_week ?? a.weekday ?? 0;
                      const dayB = b.day_of_week ?? b.weekday ?? 0;
                      return dayA - dayB;
                    })
                    .map((item) => {
                      const dayNum = item.day_of_week ?? item.weekday ?? 1;
                      const dayName = DAYS_OF_WEEK[dayNum - 1] || `${dayNum}-kun`;
                      const isOpen = item.is_open ?? !item.is_closed;
                      const openTime = item.open_time || item.opening_time || '08:00';
                      const closeTime = item.close_time || item.closing_time || '18:00';

                      return (
                        <div
                          key={dayNum}
                          className="flex items-center justify-between py-1.5 px-3 rounded-xl bg-slate-50 border border-slate-100"
                        >
                          <span className="font-semibold text-brand-navy">{dayName}</span>
                          {isOpen ? (
                            <span className="font-medium text-emerald-700">
                              {openTime} — {closeTime}
                            </span>
                          ) : (
                            <span className="font-medium text-slate-400">Dam olish kuni</span>
                          )}
                        </div>
                      );
                    })}
                </div>
              ) : (
                <p className="text-xs text-slate-500 py-2">
                  Ish vaqti ko‘rsatilmagan. Iltimos telefon orqali ma’lumot oling.
                </p>
              )}
            </div>

            {/* Clinic Map Location */}
            <div className="bg-white rounded-3xl border border-slate-200/80 p-6 shadow-subtle space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <MapPin className="w-5 h-5 text-brand-blue" />
                  <h3 className="font-bold text-base text-brand-navy">Xaritadagi joylashuvi</h3>
                </div>
                {hasCoordinates && (
                  <span className="text-[11px] font-mono text-slate-400">
                    {effectiveLat?.toFixed(4)}, {effectiveLon?.toFixed(4)}
                  </span>
                )}
              </div>

              {hasCoordinates && effectiveLat !== null && effectiveLon !== null ? (
                <InteractiveMap
                  clinics={[
                    {
                      id: clinic.id,
                      name: clinic.name,
                      address: clinic.address,
                      latitude: effectiveLat,
                      longitude: effectiveLon,
                      phone: clinic.phone,
                      logo_url: clinic.logo_url,
                    },
                  ]}
                  center={[effectiveLat, effectiveLon]}
                  zoom={15}
                  height="300px"
                />
              ) : (
                <div className="p-8 text-center bg-slate-50 rounded-2xl border border-slate-100 text-xs text-slate-500 space-y-2">
                  <MapPin className="w-8 h-8 text-slate-300 mx-auto" />
                  <p>Klinika koordinatalari kiritilmagan.</p>
                  <p className="text-slate-600 font-medium">{clinic.address || clinic.region}</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </AppLayout>
  );
}
