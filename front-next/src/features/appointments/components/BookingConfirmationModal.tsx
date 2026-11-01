'use client';

import React from 'react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { formatCurrency } from '@/lib/utils/formatters';
import {
  Calendar,
  Clock,
  Building2,
  User,
  CreditCard,
  AlertCircle,
  Loader2,
} from 'lucide-react';

interface BookingConfirmationModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  isLoading: boolean;
  doctorName: string;
  specialty?: string;
  clinicName: string;
  clinicAddress?: string;
  dateFormatted: string;
  startTime: string;
  endTime: string;
  price: number;
  error?: string | null;
}

export function BookingConfirmationModal({
  isOpen,
  onClose,
  onConfirm,
  isLoading,
  doctorName,
  specialty,
  clinicName,
  clinicAddress,
  dateFormatted,
  startTime,
  endTime,
  price,
  error,
}: BookingConfirmationModalProps) {
  return (
    <Modal
      isOpen={isOpen}
      onClose={isLoading ? () => {} : onClose}
      title="Qabulga yozilishni tasdiqlaysizmi?"
      description="Quyidagi ma’lumotlarni tekshiring va tasdiqlang."
      className="max-w-md"
    >
      <div className="space-y-4 pt-2">
        {error && (
          <div className="p-3.5 rounded-xl bg-rose-50 border border-rose-200 text-rose-700 text-xs flex items-start gap-2.5 animate-shake">
            <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
            <span className="leading-relaxed font-medium">{error}</span>
          </div>
        )}

        <div className="bg-slate-50/80 rounded-2xl p-4 border border-slate-200/80 space-y-3 text-xs">
          {/* Doctor */}
          <div className="flex items-start gap-3">
            <div className="w-8 h-8 rounded-xl bg-blue-100 text-brand-blue flex items-center justify-center shrink-0 mt-0.5">
              <User className="w-4 h-4" />
            </div>
            <div>
              <span className="text-[11px] font-semibold text-slate-400 block">Shifokor</span>
              <p className="font-bold text-slate-900 text-sm">{doctorName}</p>
              {specialty && <p className="text-slate-500 font-medium">{specialty}</p>}
            </div>
          </div>

          <div className="border-t border-slate-200/60" />

          {/* Clinic */}
          <div className="flex items-start gap-3">
            <div className="w-8 h-8 rounded-xl bg-emerald-100 text-emerald-700 flex items-center justify-center shrink-0 mt-0.5">
              <Building2 className="w-4 h-4" />
            </div>
            <div>
              <span className="text-[11px] font-semibold text-slate-400 block">Klinika</span>
              <p className="font-bold text-slate-900">{clinicName}</p>
              {clinicAddress && <p className="text-slate-500 text-[11px]">{clinicAddress}</p>}
            </div>
          </div>

          <div className="border-t border-slate-200/60" />

          {/* Date & Time */}
          <div className="grid grid-cols-2 gap-2">
            <div className="flex items-center gap-2">
              <Calendar className="w-4 h-4 text-slate-400 shrink-0" />
              <div>
                <span className="text-[10px] text-slate-400 block">Sana</span>
                <span className="font-bold text-slate-800">{dateFormatted}</span>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <Clock className="w-4 h-4 text-slate-400 shrink-0" />
              <div>
                <span className="text-[10px] text-slate-400 block">Vaqt</span>
                <span className="font-bold text-slate-800">
                  {startTime} – {endTime}
                </span>
              </div>
            </div>
          </div>

          <div className="border-t border-slate-200/60" />

          {/* Price */}
          <div className="flex items-center justify-between pt-1">
            <div className="flex items-center gap-2">
              <CreditCard className="w-4 h-4 text-slate-400" />
              <span className="text-slate-600 font-medium">Qabul narxi:</span>
            </div>
            <span className="text-base font-extrabold text-brand-navy">
              {formatCurrency(price)}
            </span>
          </div>
        </div>

        <p className="text-[11px] text-slate-400 text-center">
          Qabulga yozilish bilan bron darhol tasdiqlanadi. To‘lov klinikaning o‘zida amalga oshiriladi.
        </p>

        {/* Action Buttons */}
        <div className="flex items-center justify-end gap-3 pt-2">
          <Button
            type="button"
            variant="ghost"
            onClick={onClose}
            disabled={isLoading}
            className="text-xs font-semibold"
          >
            Bekor qilish
          </Button>

          <Button
            type="button"
            variant="primary"
            onClick={onConfirm}
            disabled={isLoading}
            className="text-xs font-bold px-5 bg-brand-blue hover:bg-brand-blue-600 shadow-sm"
          >
            {isLoading ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin mr-1.5" />
                Yozilmoqda...
              </>
            ) : (
              'Tasdiqlash va yozilish'
            )}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
