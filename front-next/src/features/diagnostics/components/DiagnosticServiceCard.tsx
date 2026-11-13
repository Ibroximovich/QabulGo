'use client';

import React from 'react';
import Link from 'next/link';
import { DiagnosticService } from '../types';
import { formatCurrency } from '@/lib/utils/formatters';
import { Building2, Clock, AlertTriangle, Cpu, ChevronRight, MapPin } from 'lucide-react';
import { Button } from '@/components/ui/Button';

interface DiagnosticServiceCardProps {
  service: DiagnosticService;
}

export function DiagnosticServiceCard({ service }: DiagnosticServiceCardProps) {
  return (
    <div className="bg-white rounded-2xl border border-slate-200/80 p-5 hover:border-brand-blue/40 hover:shadow-lg hover:shadow-slate-100 transition-all duration-200 flex flex-col justify-between group">
      <div>
        {/* Header Badges */}
        <div className="flex items-center justify-between gap-2 mb-3">
          <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-semibold bg-blue-50 text-brand-blue border border-blue-100">
            {service.category?.name || 'Diagnostika'}
          </span>

          <span className="inline-flex items-center gap-1 text-xs text-slate-500 font-medium">
            <Clock className="w-3.5 h-3.5 text-slate-400" />
            {service.duration_minutes} daq
          </span>
        </div>

        {/* Service Name */}
        <Link href={`/diagnostics/services/${service.id}`}>
          <h3 className="text-base font-bold text-slate-900 group-hover:text-brand-blue transition-colors line-clamp-2 mb-2">
            {service.name}
          </h3>
        </Link>

        {/* Equipment Details */}
        {service.equipment && (
          <div className="flex items-center gap-2 p-2 rounded-xl bg-slate-50/80 border border-slate-100 mb-3 text-xs text-slate-700">
            <Cpu className="w-4 h-4 text-brand-blue flex-shrink-0" />
            <span className="font-medium truncate">{service.equipment.name}</span>
            {service.equipment.room_number && (
              <span className="text-slate-400 text-[11px] truncate">({service.equipment.room_number})</span>
            )}
          </div>
        )}

        {/* Clinic Name & Location */}
        {service.clinic && (
          <div className="space-y-1 mb-4 text-xs text-slate-600">
            <div className="flex items-center gap-1.5 font-medium text-slate-800">
              <Building2 className="w-3.5 h-3.5 text-slate-400 flex-shrink-0" />
              <span className="truncate">{service.clinic.name}</span>
            </div>
            <div className="flex items-center gap-1.5 text-slate-500 text-[11px]">
              <MapPin className="w-3.5 h-3.5 text-slate-400 flex-shrink-0" />
              <span className="truncate">{service.clinic.district}, {service.clinic.address}</span>
            </div>
          </div>
        )}

        {/* Preparation Rules Warning Badge */}
        {service.preparation_rules && (
          <div className="flex items-start gap-1.5 p-2 rounded-lg bg-amber-50/80 border border-amber-200/60 text-amber-800 text-xs mb-4">
            <AlertTriangle className="w-3.5 h-3.5 text-amber-600 flex-shrink-0 mt-0.5" />
            <span className="line-clamp-2 leading-relaxed text-[11px]">
              {service.preparation_rules}
            </span>
          </div>
        )}
      </div>

      {/* Footer / Price & CTA */}
      <div className="pt-3 border-t border-slate-100 flex items-center justify-between gap-3 mt-auto">
        <div>
          <span className="text-[11px] text-slate-400 block font-medium">Tekshiruv narxi</span>
          <span className="text-base font-extrabold text-slate-900">
            {formatCurrency(service.price)}
          </span>
        </div>

        <Link href={`/diagnostics/services/${service.id}`}>
          <Button size="sm" className="gap-1 shadow-sm">
            Vaqt tanlash
            <ChevronRight className="w-3.5 h-3.5" />
          </Button>
        </Link>
      </div>
    </div>
  );
}

