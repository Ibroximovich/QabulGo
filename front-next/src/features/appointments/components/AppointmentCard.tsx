'use client';

import React from 'react';
import Link from 'next/link';
import { Appointment } from '../types';
import { AppointmentStatusBadge } from './AppointmentStatusBadge';
import { formatCurrency, formatDateUz } from '@/lib/utils/formatters';
import {
  Calendar,
  Clock,
  Building2,
  User,
  MapPin,
  ChevronRight,
  XCircle,
  Phone,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';

interface AppointmentCardProps {
  appointment: Appointment;
  onCancel?: (appointment: Appointment) => void;
  isClinic?: boolean;
  role?: 'patient' | 'clinic' | string;
}

export function AppointmentCard({
  appointment,
  onCancel,
  isClinic = false,
  role = 'patient',
}: AppointmentCardProps) {
  const isClinicView = isClinic || role === 'clinic';
  const doctorName = appointment.doctor
    ? `${appointment.doctor.first_name} ${appointment.doctor.last_name}`
    : 'Shifokor';
  const clinicName = appointment.clinic?.name || 'Klinika';
  const clinicAddress = appointment.clinic
    ? `${appointment.clinic.region}, ${appointment.clinic.district}, ${appointment.clinic.address}`
    : '';

  // Check if future confirmed
  const isCancellable = appointment.status === 'confirmed';

  const detailUrl = isClinicView
    ? `/clinic/appointments/${appointment.id}`
    : `/appointments/${appointment.id}`;

  return (
    <div className="bg-white rounded-3xl border border-slate-200/90 p-5 sm:p-6 shadow-subtle hover:border-slate-300 transition-all space-y-4">
      {/* Top Header: Clinic / Doctor and Status */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-start gap-3">
          <div className="w-11 h-11 rounded-2xl bg-blue-50 text-brand-blue flex items-center justify-center shrink-0 border border-blue-100">
            {isClinic ? (
              <User className="w-5 h-5" />
            ) : (
              <Building2 className="w-5 h-5" />
            )}
          </div>
          <div>
            {isClinic ? (
              <>
                <div className="flex items-center gap-2">
                  <span className="text-xs font-semibold text-slate-400">Bemor:</span>
                  <span className="font-extrabold text-slate-900 text-sm">
                    {appointment.patient
                      ? `${appointment.patient.first_name} ${appointment.patient.last_name}`
                      : 'Bemor'}
                  </span>
                </div>
                {appointment.patient?.phone && (
                  <p className="text-xs text-slate-500 font-mono flex items-center gap-1 mt-0.5">
                    <Phone className="w-3 h-3 text-slate-400" />
                    {appointment.patient.phone}
                  </p>
                )}
              </>
            ) : (
              <>
                <h4 className="font-extrabold text-slate-900 text-base">{clinicName}</h4>
                <p className="text-xs text-slate-500 flex items-center gap-1 mt-0.5">
                  <User className="w-3.5 h-3.5 text-brand-blue shrink-0" />
                  <span className="font-bold text-slate-800">{doctorName}</span>
                  {appointment.doctor?.specialty && (
                    <span className="text-slate-400">({appointment.doctor.specialty})</span>
                  )}
                </p>
              </>
            )}
          </div>
        </div>

        <AppointmentStatusBadge status={appointment.status} />
      </div>

      {isClinic && (
        <div className="text-xs text-slate-600 bg-slate-50/80 p-2.5 rounded-xl border border-slate-100 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-slate-400 font-medium">Shifokor:</span>
            <span className="font-bold text-slate-800">{doctorName}</span>
            {appointment.doctor?.specialty && (
              <span className="text-slate-400">({appointment.doctor.specialty})</span>
            )}
          </div>
        </div>
      )}

      {/* Date, Time and Price Strip */}
      <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 bg-slate-50/70 p-3.5 rounded-2xl border border-slate-100 text-xs">
        <div className="flex items-center gap-2">
          <Calendar className="w-4 h-4 text-brand-blue shrink-0" />
          <div>
            <span className="text-[10px] text-slate-400 block font-medium">Qabul sanasi</span>
            <span className="font-bold text-slate-800">{formatDateUz(appointment.appointment_date)}</span>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Clock className="w-4 h-4 text-emerald-600 shrink-0" />
          <div>
            <span className="text-[10px] text-slate-400 block font-medium">Vaqt</span>
            <span className="font-bold text-slate-900 font-mono">
              {appointment.start_time} – {appointment.end_time}
            </span>
          </div>
        </div>

        <div className="col-span-2 sm:col-span-1 flex items-center justify-start sm:justify-end pt-2 sm:pt-0 border-t sm:border-t-0 border-slate-200/60">
          <div>
            <span className="text-[10px] text-slate-400 block sm:text-right font-medium">Ko‘rik narxi</span>
            <span className="font-extrabold text-brand-navy text-sm sm:text-right block">
              {formatCurrency(appointment.consultation_price)}
            </span>
          </div>
        </div>
      </div>

      {/* Clinic address for patient */}
      {!isClinic && clinicAddress && (
        <p className="text-xs text-slate-400 flex items-center gap-1.5 px-1 truncate">
          <MapPin className="w-3.5 h-3.5 text-slate-400 shrink-0" />
          <span className="truncate">{clinicAddress}</span>
        </p>
      )}

      {/* Cancellation note if cancelled */}
      {appointment.status === 'cancelled' && appointment.cancellation_reason && (
        <div className="text-xs bg-rose-50/80 border border-rose-200/60 p-2.5 rounded-xl text-rose-800">
          <span className="font-bold">Bekor qilish sababi:</span> {appointment.cancellation_reason}
        </div>
      )}

      {/* Card Footer Actions */}
      <div className="flex items-center justify-between pt-2 border-t border-slate-100">
        <div>
          {isCancellable && onCancel && (
            <button
              type="button"
              onClick={() => onCancel(appointment)}
              className="text-xs font-semibold text-rose-600 hover:text-rose-700 hover:underline inline-flex items-center gap-1"
            >
              <XCircle className="w-3.5 h-3.5" />
              <span>Bronni bekor qilish</span>
            </button>
          )}
        </div>

        <Link
          href={detailUrl}
          className="inline-flex items-center gap-1 text-xs font-bold text-brand-blue hover:text-brand-blue-600 hover:underline ml-auto"
        >
          <span>Tafsilotlar</span>
          <ChevronRight className="w-4 h-4" />
        </Link>
      </div>
    </div>
  );
}
