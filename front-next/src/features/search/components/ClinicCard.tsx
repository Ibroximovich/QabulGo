'use client';

import React from 'react';
import Link from 'next/link';
import { Building2, MapPin, Phone, Users, ShieldCheck, Clock, ExternalLink } from 'lucide-react';
import { PublicClinicItem, NearbyClinicItem } from '../types';
import { getMediaUrl } from '@/lib/utils/media';

interface ClinicCardProps {
  clinic: PublicClinicItem | NearbyClinicItem;
  isSelected?: boolean;
  onSelect?: () => void;
}

export function ClinicCard({ clinic, isSelected = false, onSelect }: ClinicCardProps) {
  const isNearby = 'distance_km' in clinic && typeof (clinic as any).distance_km === 'number';
  const distanceKm = isNearby ? (clinic as any).distance_km : null;

  // Cover image URL or fallback
  const coverUrl = clinic.cover_image_url ? getMediaUrl(clinic.cover_image_url) : null;
  const logoUrl = clinic.logo_url ? getMediaUrl(clinic.logo_url) : null;

  return (
    <div
      onClick={onSelect}
      className={`group relative flex flex-col bg-white rounded-2xl border transition-all duration-200 overflow-hidden ${
        isSelected
          ? 'border-brand-blue ring-2 ring-brand-blue/20 shadow-elevated'
          : 'border-slate-200/80 hover:border-slate-300 hover:shadow-card'
      }`}
    >
      {/* Cover Image & Badges */}
      <div className="relative h-44 w-full bg-gradient-to-r from-slate-100 to-slate-200 overflow-hidden">
        {coverUrl ? (
          <img
            src={coverUrl}
            alt={clinic.name}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
            onError={(e) => {
              (e.target as HTMLImageElement).style.display = 'none';
            }}
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center bg-brand-navy/5 text-slate-400">
            <Building2 className="w-12 h-12 stroke-[1.5]" />
          </div>
        )}

        {/* Distance Badge */}
        {distanceKm !== null && (
          <div className="absolute top-3 left-3 z-10">
            <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-white/95 backdrop-blur-md text-xs font-bold text-brand-teal shadow-xs border border-white">
              <span className="w-2 h-2 rounded-full bg-brand-teal animate-pulse" />
              <span>{distanceKm < 1 ? `${Math.round(distanceKm * 1000)} m` : `${distanceKm.toFixed(1)} km`} yaqin</span>
            </span>
          </div>
        )}

        {/* Verified Badge */}
        <div className="absolute top-3 right-3 z-10">
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-white/90 backdrop-blur-md text-xs font-medium text-emerald-700 shadow-xs border border-emerald-100">
            <ShieldCheck className="w-3.5 h-3.5 text-emerald-600" />
            <span>Tasdiqlangan</span>
          </span>
        </div>

        {/* Clinic Logo */}
        <div className="absolute -bottom-5 left-5 z-10">
          <div className="w-14 h-14 rounded-2xl bg-white p-1 shadow-card border border-slate-100 flex items-center justify-center overflow-hidden">
            {logoUrl ? (
              <img src={logoUrl} alt="" className="w-full h-full object-contain rounded-xl" />
            ) : (
              <div className="w-full h-full rounded-xl bg-brand-blue-50 text-brand-blue flex items-center justify-center font-bold text-lg">
                {clinic.name.charAt(0).toUpperCase()}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Content Area */}
      <div className="flex-1 p-5 pt-8 flex flex-col justify-between gap-4">
        <div>
          {/* Clinic Name */}
          <Link
            href={`/clinics/${clinic.id}`}
            className="text-lg font-bold text-brand-navy hover:text-brand-blue transition-colors line-clamp-1 group-hover:underline"
          >
            {clinic.name}
          </Link>

          {/* Location / Address */}
          <div className="flex items-start gap-1.5 text-xs text-slate-500 mt-2">
            <MapPin className="w-3.5 h-3.5 text-slate-400 shrink-0 mt-0.5" />
            <span className="line-clamp-2">
              {clinic.address ? `${clinic.address}, ` : ''}
              {clinic.district ? `${clinic.district}, ` : ''}
              {clinic.region}
            </span>
          </div>

          {/* Specialties Pills */}
          {clinic.specialties && clinic.specialties.length > 0 && (
            <div className="flex flex-wrap gap-1.5 mt-3">
              {clinic.specialties.slice(0, 3).map((spec) => (
                <span
                  key={spec.id}
                  className="inline-flex items-center text-[11px] font-medium px-2.5 py-0.5 rounded-lg bg-slate-100 text-slate-700 border border-slate-200/60"
                >
                  {spec.name}
                </span>
              ))}
              {clinic.specialties.length > 3 && (
                <span className="inline-flex items-center text-[11px] font-medium px-2 py-0.5 rounded-lg bg-slate-50 text-slate-500 border border-slate-200/40">
                  +{clinic.specialties.length - 3}
                </span>
              )}
            </div>
          )}
        </div>

        {/* Footer Meta & Actions */}
        <div className="pt-3 border-t border-slate-100 flex items-center justify-between gap-2">
          {/* Doctors Count */}
          <div className="flex items-center gap-1.5 text-xs font-medium text-slate-600">
            <Users className="w-3.5 h-3.5 text-brand-blue" />
            <span>{clinic.doctors_count || 0} nafar shifokor</span>
          </div>

          {/* View Profile Link */}
          <Link
            href={`/clinics/${clinic.id}`}
            className="inline-flex items-center gap-1 text-xs font-semibold text-brand-blue hover:text-brand-blue-700 transition-colors"
          >
            <span>Batafsil</span>
            <ExternalLink className="w-3.5 h-3.5" />
          </Link>
        </div>
      </div>
    </div>
  );
}

