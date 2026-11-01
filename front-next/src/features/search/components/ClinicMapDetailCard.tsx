'use client';

import React from 'react';
import Link from 'next/link';
import {
  Building2,
  MapPin,
  Phone,
  Clock,
  ExternalLink,
  ChevronRight,
  Stethoscope,
  ShieldCheck,
  Calendar,
  X,
} from 'lucide-react';
import { PublicClinicItem, NearbyClinicItem } from '../types';
import { getMediaUrl } from '@/lib/utils/media';
import { Button } from '@/components/ui/Button';

interface ClinicMapDetailCardProps {
  clinic: PublicClinicItem | NearbyClinicItem;
  onClose?: () => void;
}

const DAYS_OF_WEEK = [
  'Dushanba',
  'Seshanba',
  'Chorshanba',
  'Payshanba',
  'Juma',
  'Shanba',
  'Yakshanba',
];

export function ClinicMapDetailCard({ clinic, onClose }: ClinicMapDetailCardProps) {
  const logoSrc = clinic.logo_url ? getMediaUrl(clinic.logo_url) : null;
  const coverSrc = clinic.cover_image_url ? getMediaUrl(clinic.cover_image_url) : null;
  const distance =
    'distance_km' in clinic && typeof (clinic as any).distance_km === 'number'
      ? (clinic as any).distance_km
      : null;

  // Bugungi hafta kuni (1: Dushanba, ..., 7: Yakshanba)
  const todayJs = new Date().getDay();
  const currentDayOfWeek = todayJs === 0 ? 7 : todayJs;

  // Tashqi xarita havolasi (Yandex yoki Google)
  const mapLink =
    clinic.location_url ||
    (clinic.latitude && clinic.longitude
      ? `https://yandex.uz/maps/?ll=${clinic.longitude},${clinic.latitude}&z=16&pt=${clinic.longitude},${clinic.latitude}`
      : null);

  return (
    <div className="bg-white rounded-3xl border border-slate-200/90 shadow-card overflow-hidden flex flex-col max-h-[720px] transition-all animate-fadeIn">
      {/* Cover and Close Header */}
      <div className="relative h-44 sm:h-48 w-full bg-linear-to-r from-slate-900 via-brand-navy to-slate-900 shrink-0">
        {coverSrc ? (
          <img
            src={coverSrc}
            alt={clinic.name}
            className="w-full h-full object-cover opacity-90"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center bg-brand-navy/90">
            <Building2 className="w-12 h-12 text-white/30" />
          </div>
        )}

        {/* Gradient Overlay */}
        <div className="absolute inset-0 bg-linear-to-t from-slate-950/80 via-black/20 to-transparent" />

        {/* Close Button (if collapsible) */}
        {onClose && (
          <button
            type="button"
            onClick={onClose}
            className="absolute top-3 right-3 w-8 h-8 rounded-full bg-black/40 hover:bg-black/60 text-white backdrop-blur-md flex items-center justify-center transition-colors z-10"
            title="Yopish"
          >
            <X className="w-4 h-4" />
          </button>
        )}

        {/* Distance Badge */}
        {distance !== null && (
          <div className="absolute top-3 left-3 px-3 py-1 rounded-full bg-brand-teal text-white text-xs font-bold shadow-md flex items-center gap-1 backdrop-blur-xs">
            <MapPin className="w-3.5 h-3.5" />
            <span>{distance.toFixed(1)} km masofada</span>
          </div>
        )}

        {/* Logo and Quick Info Overlay */}
        <div className="absolute -bottom-6 left-5 flex items-end gap-3 z-10">
          <div className="w-16 h-16 rounded-2xl bg-white border-2 border-white shadow-elevated overflow-hidden shrink-0 flex items-center justify-center p-1">
            {logoSrc ? (
              <img src={logoSrc} alt="" className="w-full h-full object-contain" />
            ) : (
              <Building2 className="w-8 h-8 text-brand-navy" />
            )}
          </div>
        </div>
      </div>

      {/* Main Content Body */}
      <div className="p-5 pt-8 overflow-y-auto space-y-5 flex-1">
        {/* Name and Verification */}
        <div>
          <div className="flex items-center gap-1.5 text-xs font-semibold text-emerald-600 mb-0.5">
            <ShieldCheck className="w-4 h-4 text-emerald-600 shrink-0" />
            <span>Rasmiy tasdiqlangan klinika</span>
          </div>
          <h3 className="text-xl font-extrabold text-brand-navy leading-tight">{clinic.name}</h3>
          {clinic.description && (
            <p className="text-xs text-slate-500 mt-1 line-clamp-2 leading-relaxed">
              {clinic.description}
            </p>
          )}
        </div>

        {/* Address and Region */}
        <div className="p-3.5 rounded-2xl bg-slate-50 border border-slate-100 space-y-1.5 text-xs">
          <div className="flex items-start gap-2">
            <MapPin className="w-4 h-4 text-brand-blue shrink-0 mt-0.5" />
            <div>
              <p className="font-bold text-brand-navy">
                {clinic.region}{clinic.district ? `, ${clinic.district}` : ''}
              </p>
              <p className="text-slate-600 mt-0.5">{clinic.address}</p>
            </div>
          </div>
        </div>

        {/* Quick Contacts */}
        {clinic.phone && (
          <div className="flex items-center gap-2">
            <a
              href={`tel:${clinic.phone}`}
              className="flex-1 inline-flex items-center justify-center gap-2 px-4 py-2.5 rounded-xl bg-brand-blue text-white text-xs font-bold hover:bg-brand-blue-600 transition-colors shadow-sm"
            >
              <Phone className="w-3.5 h-3.5" />
              <span>Qo‘ng‘iroq: {clinic.phone}</span>
            </a>
            {mapLink && (
              <a
                href={mapLink}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center justify-center p-2.5 rounded-xl border border-slate-200 text-slate-700 hover:bg-slate-50 text-xs font-semibold transition-colors"
                title="Xaritada ochish"
              >
                <ExternalLink className="w-4 h-4" />
              </a>
            )}
          </div>
        )}

        {/* Working Hours */}
        {clinic.working_hours && clinic.working_hours.length > 0 && (
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold text-brand-navy flex items-center gap-1.5">
                <Clock className="w-3.5 h-3.5 text-brand-blue" />
                <span>Ish vaqti jadvali</span>
              </span>
            </div>

            <div className="space-y-1 text-xs">
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
                  const isToday = dayNum === currentDayOfWeek;

                  return (
                    <div
                      key={dayNum}
                      className={`flex items-center justify-between py-1 px-2.5 rounded-lg text-[11px] ${
                        isToday
                          ? 'bg-blue-50 border border-blue-200 font-semibold text-brand-blue'
                          : 'bg-slate-50 text-slate-600'
                      }`}
                    >
                      <span className="flex items-center gap-1">
                        {dayName}
                        {isToday && (
                          <span className="text-[9px] px-1.5 py-0.2 rounded bg-brand-blue text-white font-bold uppercase">
                            Bugun
                          </span>
                        )}
                      </span>
                      {isOpen ? (
                        <span className={isToday ? 'text-brand-blue font-bold' : 'text-slate-700'}>
                          {openTime} — {closeTime}
                        </span>
                      ) : (
                        <span className="text-slate-400">Dam olish</span>
                      )}
                    </div>
                  );
                })}
            </div>
          </div>
        )}

        {/* Specialties */}
        {clinic.specialties && clinic.specialties.length > 0 && (
          <div className="space-y-2">
            <span className="text-xs font-bold text-brand-navy flex items-center gap-1.5">
              <Stethoscope className="w-3.5 h-3.5 text-brand-blue" />
              <span>Mutaxassisliklar</span>
            </span>
            <div className="flex flex-wrap gap-1.5">
              {clinic.specialties.map((s) => (
                <span
                  key={s.id}
                  className="px-2 py-0.5 rounded-md bg-slate-100 text-slate-700 text-[11px] font-medium border border-slate-200/60"
                >
                  {s.name}
                </span>
              ))}
            </div>
          </div>
        )}

        {/* Doctors Count */}
        <div className="flex items-center justify-between py-2 px-3 rounded-xl bg-slate-50 border border-slate-100 text-xs">
          <span className="text-slate-500 font-medium">Klinika shifokorlari:</span>
          <span className="font-extrabold text-brand-navy">
            {clinic.doctors_count || 0} nafar mutaxassis
          </span>
        </div>
      </div>

      {/* Footer Action Buttons */}
      <div className="p-4 bg-slate-50 border-t border-slate-200/80 flex flex-col gap-2 shrink-0">
        <Link
          href={`/clinics/${clinic.id}`}
          className="w-full inline-flex items-center justify-center gap-2 py-2.5 px-4 rounded-xl bg-brand-navy text-white text-xs font-bold hover:bg-slate-800 transition-colors shadow-xs"
        >
          <span>Klinika profilini to‘liq ko‘rish</span>
          <ChevronRight className="w-4 h-4" />
        </Link>

        <button
          type="button"
          disabled
          className="w-full inline-flex items-center justify-center gap-1.5 py-2 px-4 rounded-xl bg-slate-200/70 text-slate-500 text-xs font-semibold cursor-not-allowed"
          title="Qabulga yozilish tez kunda ishga tushadi"
        >
          <Calendar className="w-3.5 h-3.5" />
          <span>Qabulga yozilish</span>
          <span className="text-[9px] uppercase font-bold px-1.5 py-0.2 rounded bg-slate-400 text-white ml-1">
            Tez kunda
          </span>
        </button>
      </div>
    </div>
  );
}

