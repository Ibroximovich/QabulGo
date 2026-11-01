'use client';

import React, { useState } from 'react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { AlertCircle, Loader2 } from 'lucide-react';

interface CancelConfirmModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (reason?: string) => Promise<void> | void;
  isLoading: boolean;
  title?: string;
  description?: string;
  appointmentInfo?: string;
  isClinic?: boolean;
  requireReason?: boolean;
}

export function CancelConfirmModal({
  isOpen,
  onClose,
  onConfirm,
  isLoading,
  title = 'Qabulni bekor qilish',
  description = 'Haqiqatan ham ushbu qabulni bekor qilmoqchimisiz?',
  appointmentInfo,
  isClinic = false,
  requireReason = false,
}: CancelConfirmModalProps) {
  const [reason, setReason] = useState('');

  const handleConfirm = async () => {
    await onConfirm(reason.trim() || undefined);
    setReason('');
  };

  const handleClose = () => {
    if (!isLoading) {
      setReason('');
      onClose();
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title={title}
      description={description}
      className="max-w-md"
    >
      <div className="space-y-4 pt-2">
        <div className="p-3.5 rounded-xl bg-amber-50 border border-amber-200/80 text-amber-800 text-xs flex items-start gap-2.5">
          <AlertCircle className="w-4 h-4 shrink-0 mt-0.5 text-amber-600" />
          <div className="leading-relaxed">
            <span className="font-semibold block mb-0.5">Diqqat:</span>
            Bekor qilingan qabul qayta tasdiqlanmaydi va ajratilgan vaqt boshqa bemorlar uchun bo‘sh deb ochiladi.
          </div>
        </div>

        {appointmentInfo && (
          <div className="p-3 rounded-xl bg-slate-50 border border-slate-200 text-xs font-medium text-slate-700">
            {appointmentInfo}
          </div>
        )}

        {/* Reason input */}
        <div className="space-y-1.5">
          <label htmlFor="cancel-reason" className="block text-xs font-semibold text-slate-700">
            Bekor qilish sababi <span className="text-slate-400 font-normal">(ixtiyoriy)</span>:
          </label>
          <textarea
            id="cancel-reason"
            rows={3}
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            disabled={isLoading}
            placeholder={
              isClinic
                ? 'Masalan: Shifokor favqulodda operatsiyada'
                : 'Masalan: Rejalarim o‘zgarib qoldi'
            }
            className="w-full text-xs rounded-xl border border-slate-200 p-3 focus:outline-none focus:ring-2 focus:ring-brand-blue/30 focus:border-brand-blue resize-none disabled:bg-slate-100 disabled:text-slate-400"
          />
        </div>

        {/* Actions */}
        <div className="flex items-center justify-end gap-3 pt-2">
          <Button
            type="button"
            variant="ghost"
            onClick={handleClose}
            disabled={isLoading}
            className="text-xs font-semibold"
          >
            Orqaga
          </Button>

          <Button
            type="button"
            variant="danger"
            onClick={handleConfirm}
            disabled={isLoading}
            className="text-xs font-bold px-4"
          >
            {isLoading ? (
              <>
                <Loader2 className="w-3.5 h-3.5 animate-spin mr-1.5" />
                Bekor qilinmoqda...
              </>
            ) : (
              'Ha, bekor qilinsin'
            )}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
