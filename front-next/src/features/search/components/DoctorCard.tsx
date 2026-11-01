'use client';

import React from 'react';
import Link from 'next/link';
import { User, Stethoscope, Building2, MapPin, Clock, Calendar, Award } from 'lucide-react';
import { PublicDoctorItem } from '../types';
import { formatCurrency } from '@/lib/utils/formatters';
import { getMediaUrl } from '@/lib/utils/media';

interface DoctorCardProps {
  doctor: PublicDoctorItem;
}

export function DoctorCard({ doctor }: DoctorCardProps) {
  const photoUrl = doctor.photo_url ? getMediaUrl(doctor.photo_url) : null;
  const fullName = `${doctor.first_name} ${doctor.last_name}`;

  return (
    <div className="group relative flex flex-col sm:flex-row bg-white rounded-2xl border border-slate-200/80 hover:border-slate-300 hover:shadow-card transition-all duration-200 overflow-hidden p-5 gap-5">
      {/* Doctor Photo */}
      <div className="shrink-0 flex items-center justify-center sm:self-start">
        <div className="relative w-28 h-28 sm:w-24 sm:h-24 rounded-2xl overflow-hidden bg-slate-100 border border-slate-200/80 shadow-xs">
          {photoUrl ? (
            <img
              src={photoUrl}
              alt={fullName}
              className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = 'none';
              }}
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center bg-brand-blue-50 text-brand-blue">
              <User className="w-10 h-10 stroke-[1.5]" />
            </div>
          )}
        </div>
      </div>

      {/* Middle: Details */}
      <div className="flex-1 min-w-0 flex flex-col justify-between">
        <div>
          {/* Header & Specialty */}
          <div className="flex flex-wrap items-center gap-2">
            {doctor.specialty && (
              <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-lg bg-brand-blue-50 text-brand-blue text-xs font-semibold">
                <Stethoscope className="w-3.5 h-3.5" />
                <span>{doctor.specialty.name}</span>
              </span>
            )}
            {doctor.experience_years > 0 && (
              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-lg bg-slate-100 text-slate-600 text-[11px] font-medium">
                <Award className="w-3 h-3 text-amber-500" />
                <span>{doctor.experience_years} yil tajriba</span>
              </span>
            )}
          </div>

          {/* Full Name */}
          <Link
            href={`/doctors/${doctor.doctor_id}`}
            className="block text-lg font-bold text-brand-navy hover:text-brand-blue transition-colors mt-1.5 group-hover:underline truncate"
          >
            {fullName}
          </Link>

          {/* Bio Preview */}
          {doctor.bio && (
            <p className="text-xs text-slate-500 line-clamp-2 mt-1">
              {doctor.bio}
            </p>
          )}

          {/* Clinic Information */}
          {doctor.clinic && (
            <div className="mt-3 p-2.5 rounded-xl bg-slate-50 border border-slate-100 flex flex-col gap-1 text-xs text-slate-600">
              <div className="flex items-center gap-1.5 font-semibold text-brand-navy">
                <Building2 className="w-3.5 h-3.5 text-brand-blue shrink-0" />
                <Link
                  href={`/clinics/${doctor.clinic.id}`}
                  className="hover:text-brand-blue hover:underline truncate"
                >
                  {doctor.clinic.name}
                </Link>
              </div>
              <div className="flex items-center gap-1.5 text-[11px] text-slate-500">
                <MapPin className="w-3 h-3 text-slate-400 shrink-0" />
                <span className="truncate">
                  {doctor.clinic.district ? `${doctor.clinic.district}, ` : ''}
                  {doctor.clinic.region}
                </span>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Right / Bottom: Price & Action */}
      <div className="shrink-0 flex sm:flex-col items-center sm:items-end justify-between sm:justify-center border-t sm:border-t-0 sm:border-l border-slate-100 pt-3 sm:pt-0 sm:pl-5 gap-3">
        <div className="text-left sm:text-right">
          <span className="text-[11px] font-medium text-slate-500 block">Qabul narxi</span>
          <span className="text-lg font-extrabold text-brand-navy">
            {formatCurrency(doctor.consultation_price)}
          </span>
          {doctor.appointment_duration_minutes > 0 && (
            <div className="flex items-center gap-1 text-[11px] text-slate-400 sm:justify-end mt-0.5">
              <Clock className="w-3 h-3" />
              <span>~{doctor.appointment_duration_minutes} daqiqa</span>
            </div>
          )}
        </div>

        <div className="flex flex-col gap-2 w-full sm:w-auto">
          {/* Booking Button (Tez kunda) */}
          <button
            type="button"
            className="inline-flex items-center justify-center gap-1.5 px-4 py-2 rounded-xl bg-brand-blue/10 text-brand-blue text-xs font-bold cursor-not-allowed border border-brand-blue/20"
            title="Qabulga yozilish tez kunda ishga tushadi"
          >
            <Calendar className="w-3.5 h-3.5" />
            <span>Qabulga yozilish</span>
            <span className="text-[9px] uppercase font-bold tracking-wider px-1.5 py-0.2 rounded bg-brand-blue text-white ml-1">
              Tez kunda
            </span>
          </button>

          {/* View Profile */}
          <Link
            href={`/doctors/${doctor.doctor_id}`}
            className="inline-flex items-center justify-center px-4 py-2 rounded-xl bg-slate-100 text-slate-700 hover:bg-slate-200 text-xs font-semibold transition-colors text-center"
          >
            Batafsil ko‘rish
          </Link>
        </div>
      </div>
    </div>
  );
}

