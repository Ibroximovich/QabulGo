'use client';

import React from 'react';
import { useRouter } from 'next/navigation';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Appointment } from '../types';
import { CheckCircle2, Calendar, Clock, Building2, User, ArrowRight } from 'lucide-react';
import { formatCurrency } from '@/lib/utils/formatters';

interface BookingSuccessModalProps {
  isOpen: boolean;
  onClose: () => void;
  appointment: Appointment | null;
}

export function BookingSuccessModal({ isOpen, onClose, appointment }: BookingSuccessModalProps) {
  const router = useRouter();

  if (!appointment) return null;

  const doctorName = appointment.doctor
    ? `${appointment.doctor.first_name} ${appointment.doctor.last_name}`
    : 'Shifokor';
  const clinicName = appointment.clinic?.name || 'Klinika';

  const handleGoToAppointments = () => {
    onClose();
    router.push('/appointments');
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} className="max-w-md text-center">
      <div className="space-y-5 pt-2">
        {/* Animated Green Checkmark Icon */}
        <div className="w-16 h-16 rounded-full bg-emerald-100 text-emerald-600 flex items-center justify-center mx-auto shadow-inner animate-scale-up">
          <CheckCircle2 className="w-10 h-10" />
        </div>

        <div>
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-[11px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200 mb-2">
            Status: Tasdiqlangan
          </span>
          <h3 className="text-xl font-extrabold text-slate-900">
            Qabulga muvaffaqiyatli yozildingiz!
          </h3>
          <p className="text-xs text-slate-500 mt-1">
            Qabulingiz tizimda avtomatik tasdiqlandi. Shifokor sizni belgilangan vaqtda kutadi.
          </p>
        </div>

        {/* Appointment Summary Box */}
        <div className="bg-slate-50 rounded-2xl p-4 border border-slate-200/80 text-left text-xs space-y-2.5">
          <div className="flex items-center gap-2.5">
            <User className="w-4 h-4 text-brand-blue shrink-0" />
            <span className="font-bold text-slate-900">{doctorName}</span>
            {appointment.doctor?.specialty && (
              <span className="text-slate-400">({appointment.doctor.specialty})</span>
            )}
          </div>

          <div className="flex items-center gap-2.5">
            <Building2 className="w-4 h-4 text-emerald-600 shrink-0" />
            <span className="font-medium text-slate-800">{clinicName}</span>
          </div>

          <div className="flex items-center justify-between pt-1 border-t border-slate-200/60">
            <div className="flex items-center gap-2 text-slate-700">
              <Calendar className="w-3.5 h-3.5 text-slate-400" />
              <span className="font-semibold">{appointment.appointment_date}</span>
            </div>
            <div className="flex items-center gap-2 text-slate-700">
              <Clock className="w-3.5 h-3.5 text-slate-400" />
              <span className="font-bold text-brand-navy">
                {appointment.start_time} – {appointment.end_time}
              </span>
            </div>
          </div>

          {appointment.consultation_price > 0 && (
            <div className="flex items-center justify-between pt-2 border-t border-slate-200/60 text-[11px]">
              <span className="text-slate-500">Qabul narxi:</span>
              <span className="font-extrabold text-slate-900">
                {formatCurrency(appointment.consultation_price)}
              </span>
            </div>
          )}
        </div>

        {/* Actions */}
        <div className="flex flex-col sm:flex-row items-center justify-center gap-3 pt-2">
          <Button
            type="button"
            variant="ghost"
            onClick={onClose}
            className="w-full sm:w-auto text-xs font-semibold"
          >
            Yopish
          </Button>

          <Button
            type="button"
            variant="primary"
            onClick={handleGoToAppointments}
            className="w-full sm:w-auto text-xs font-bold bg-brand-blue hover:bg-brand-blue-600 shadow-sm flex items-center justify-center gap-1.5"
          >
            <span>Mening bronlarim</span>
            <ArrowRight className="w-3.5 h-3.5" />
          </Button>
        </div>
      </div>
    </Modal>
  );
}
