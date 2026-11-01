'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';
import { AppLayout } from '@/components/layout/AppLayout';
import { searchApi } from '../api/searchApi';
import { PublicDoctorDetail } from '../types';
import { getMediaUrl } from '@/lib/utils/media';
import { formatCurrency } from '@/lib/utils/formatters';
import {
  User,
  Stethoscope,
  Building2,
  MapPin,
  Phone,
  Clock,
  Calendar,
  ChevronLeft,
  Award,
  AlertCircle,
  ExternalLink,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';

interface DoctorDetailPageProps {
  initialDoctor?: PublicDoctorDetail | null;
}

export function DoctorDetailPage({ initialDoctor }: DoctorDetailPageProps) {
  const params = useParams();
  const router = useRouter();
  const doctorId = params?.id as string;

  const [doctor, setDoctor] = useState<PublicDoctorDetail | null>(initialDoctor || null);
  const [isLoading, setIsLoading] = useState<boolean>(!initialDoctor);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (initialDoctor) return;
    if (!doctorId) return;

    let active = true;

    async function loadDoctor() {
      setIsLoading(true);
      setError(null);

      try {
        const res = await searchApi.getDoctorDetails(doctorId);
        if (active) {
          if (res.data) {
            setDoctor(res.data);
          } else {
            setError('Shifokor topilmadi yoki tasdiqlangan klinikalarda faol emas.');
          }
        }
      } catch (err: any) {
        if (active) {
          console.error('Shifokor ma’lumotlarini yuklashda xatolik:', err);
          setError('Shifokor topilmadi yoki yuklashda xatolik yuz berdi.');
        }
      } finally {
        if (active) {
          setIsLoading(false);
        }
      }
    }

    loadDoctor();

    return () => {
      active = false;
    };
  }, [doctorId, initialDoctor]);

  if (isLoading) {
    return (
      <AppLayout>
        <div className="max-w-4xl mx-auto space-y-6 animate-pulse">
          <div className="h-64 bg-slate-200 rounded-3xl w-full" />
          <div className="h-10 bg-slate-200 rounded-xl w-1/3" />
          <div className="h-4 bg-slate-200 rounded-md w-1/2" />
        </div>
      </AppLayout>
    );
  }

  if (error || !doctor) {
    return (
      <AppLayout>
        <div className="max-w-md mx-auto my-16 text-center bg-white p-8 rounded-3xl border border-slate-200 shadow-subtle space-y-4">
          <div className="w-14 h-14 rounded-2xl bg-red-50 text-red-500 mx-auto flex items-center justify-center">
            <AlertCircle className="w-7 h-7" />
          </div>
          <h2 className="text-xl font-bold text-brand-navy">Shifokor topilmadi</h2>
          <p className="text-sm text-slate-500">
            {error || 'Siz qidirayotgan shifokor mavjud emas yoki faol ish joyiga ega emas.'}
          </p>
          <div className="pt-2">
            <Button onClick={() => router.push('/doctors')} className="gap-2">
              <ChevronLeft className="w-4 h-4" />
              <span>Shifokorlar ro‘yxatiga qaytish</span>
            </Button>
          </div>
        </div>
      </AppLayout>
    );
  }

  const photoUrl = doctor.photo_url ? getMediaUrl(doctor.photo_url) : null;
  const fullName = `${doctor.first_name} ${doctor.last_name}`;

  return (
    <AppLayout>
      <div className="max-w-4xl mx-auto space-y-8">
        {/* Breadcrumb */}
        <div className="flex items-center gap-2 text-xs font-semibold text-slate-500">
          <Link href="/doctors" className="hover:text-brand-blue flex items-center gap-1">
            <ChevronLeft className="w-3.5 h-3.5" />
            <span>Shifokorlar</span>
          </Link>
          <span>/</span>
          <span className="text-brand-navy truncate max-w-xs">{fullName}</span>
        </div>

        {/* Doctor Main Profile Card */}
        <div className="bg-white rounded-3xl border border-slate-200/80 p-6 sm:p-8 shadow-subtle">
          <div className="flex flex-col sm:flex-row items-center sm:items-start gap-6">
            {/* Large Photo */}
            <div className="w-36 h-36 sm:w-44 sm:h-44 rounded-3xl overflow-hidden bg-slate-100 border-2 border-white shadow-elevated shrink-0">
              {photoUrl ? (
                <img src={photoUrl} alt={fullName} className="w-full h-full object-cover" />
              ) : (
                <div className="w-full h-full flex items-center justify-center bg-brand-blue-50 text-brand-blue">
                  <User className="w-16 h-16 stroke-[1.5]" />
                </div>
              )}
            </div>

            {/* Profile Info */}
            <div className="flex-1 text-center sm:text-left space-y-3">
              <div className="flex flex-wrap items-center justify-center sm:justify-start gap-2">
                {doctor.primary_specialty && (
                  <span className="inline-flex items-center gap-1 px-3 py-1 rounded-xl bg-brand-blue-50 text-brand-blue text-xs font-bold">
                    <Stethoscope className="w-3.5 h-3.5" />
                    <span>{doctor.primary_specialty.name}</span>
                  </span>
                )}
                {doctor.experience_years > 0 && (
                  <span className="inline-flex items-center gap-1 px-3 py-1 rounded-xl bg-amber-50 text-amber-800 border border-amber-200/60 text-xs font-semibold">
                    <Award className="w-3.5 h-3.5 text-amber-600" />
                    <span>{doctor.experience_years} yil tajriba</span>
                  </span>
                )}
              </div>

              <h1 className="text-2xl sm:text-3xl font-extrabold text-brand-navy">{fullName}</h1>

              {doctor.bio ? (
                <p className="text-sm text-slate-600 leading-relaxed whitespace-pre-line max-w-2xl">
                  {doctor.bio}
                </p>
              ) : (
                <p className="text-xs text-slate-400 italic">
                  Shifokor haqida qo‘shimcha ma’lumot kiritilmagan.
                </p>
              )}
            </div>
          </div>
        </div>

        {/* Workplaces Section */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-xl font-bold text-brand-navy">Qabul joylari va klinikalar</h2>
              <p className="text-xs text-slate-500 mt-0.5">
                Shifokor quyidagi tasdiqlangan klinikalarda bemorlarni qabul qiladi:
              </p>
            </div>
            <span className="text-xs font-semibold px-3 py-1 rounded-full bg-white border border-slate-200 text-slate-600">
              {doctor.workplaces.length} ta klinika
            </span>
          </div>

          <div className="grid grid-cols-1 gap-4">
            {doctor.workplaces.map((wp) => (
              <div
                key={wp.clinic_doctor_id}
                className="bg-white rounded-3xl border border-slate-200/80 p-6 shadow-subtle flex flex-col md:flex-row items-start md:items-center justify-between gap-6 hover:border-slate-300 transition-all"
              >
                {/* Clinic Info */}
                <div className="space-y-2 flex-1">
                  <div className="flex items-center gap-2">
                    <Building2 className="w-5 h-5 text-brand-blue" />
                    <Link
                      href={`/clinics/${wp.clinic_id}`}
                      className="text-lg font-bold text-brand-navy hover:text-brand-blue hover:underline"
                    >
                      {wp.clinic_name}
                    </Link>
                  </div>

                  <div className="flex items-start gap-1.5 text-xs text-slate-500">
                    <MapPin className="w-3.5 h-3.5 text-slate-400 shrink-0 mt-0.5" />
                    <span>
                      {wp.address ? `${wp.address}, ` : ''}
                      {wp.district ? `${wp.district}, ` : ''}
                      {wp.region}
                    </span>
                  </div>

                  {wp.phone && (
                    <div className="flex items-center gap-1.5 text-xs text-slate-600 pt-1">
                      <Phone className="w-3.5 h-3.5 text-slate-400" />
                      <a href={`tel:${wp.phone}`} className="hover:text-brand-blue font-medium">
                        {wp.phone}
                      </a>
                    </div>
                  )}
                </div>

                {/* Price, Duration and Booking */}
                <div className="flex flex-col sm:flex-row md:flex-col items-start sm:items-center md:items-end justify-between w-full md:w-auto pt-4 md:pt-0 border-t md:border-t-0 md:border-l border-slate-100 md:pl-6 gap-3">
                  <div className="text-left md:text-right">
                    <span className="text-[11px] font-semibold text-slate-400 block">Qabul narxi</span>
                    <span className="text-xl font-extrabold text-brand-navy">
                      {formatCurrency(wp.consultation_price)}
                    </span>
                    {wp.appointment_duration_minutes > 0 && (
                      <div className="flex items-center gap-1 text-xs text-slate-400 md:justify-end mt-0.5">
                        <Clock className="w-3.5 h-3.5" />
                        <span>~{wp.appointment_duration_minutes} daqiqa</span>
                      </div>
                    )}
                  </div>

                  <button
                    type="button"
                    className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl bg-brand-blue text-white font-semibold text-xs shadow-xs hover:bg-brand-blue-600 cursor-not-allowed"
                    title="Qabulga yozilish tez kunda ishga tushadi"
                  >
                    <Calendar className="w-4 h-4" />
                    <span>Qabulga yozilish</span>
                    <span className="text-[9px] uppercase font-bold tracking-wider px-1.5 py-0.2 rounded bg-white/20 text-white ml-1">
                      Tez kunda
                    </span>
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </AppLayout>
  );
}

