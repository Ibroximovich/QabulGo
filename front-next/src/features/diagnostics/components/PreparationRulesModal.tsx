'use client';

import React, { useState } from 'react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { AlertTriangle, CheckCircle, ShieldAlert, Calendar, Clock, DollarSign, Building2 } from 'lucide-react';
import { formatCurrency } from '@/lib/utils/formatters';

interface PreparationRulesModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  isLoading?: boolean;
  serviceName: string;
  clinicName?: string;
  date: string;
  startTime: string;
  endTime: string;
  price: number;
  preparationRules?: string;
}

export function PreparationRulesModal({
  isOpen,
  onClose,
  onConfirm,
  isLoading,
  serviceName,
  clinicName,
  date,
  startTime,
  endTime,
  price,
  preparationRules,
}: PreparationRulesModalProps) {
  const [acknowledged, setAcknowledged] = useState(false);

  const handleClose = () => {
    setAcknowledged(false);
    onClose();
  };

  const handleConfirm = () => {
    if (acknowledged) {
      onConfirm();
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Qabulni tasdiqlash va tayyorgarlik">
      <div className="space-y-5">
        {/* Summary Card */}
        <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 space-y-2.5 text-sm">
          <div className="font-bold text-slate-900 text-base">{serviceName}</div>
          {clinicName && (
            <div className="flex items-center gap-2 text-slate-600 text-xs">
              <Building2 className="w-4 h-4 text-slate-400" />
              <span>{clinicName}</span>
            </div>
          )}
          <div className="grid grid-cols-2 gap-2 pt-2 border-t border-slate-200/60 text-xs text-slate-700">
            <div className="flex items-center gap-1.5">
              <Calendar className="w-4 h-4 text-brand-blue" />
              <span>{date}</span>
            </div>
            <div className="flex items-center gap-1.5">
              <Clock className="w-4 h-4 text-brand-blue" />
              <span>{startTime} - {endTime}</span>
            </div>
          </div>
          <div className="flex items-center justify-between pt-1 border-t border-slate-200/60 text-xs">
            <span className="text-slate-500">To‘lov summasi:</span>
            <span className="font-bold text-slate-900 text-sm">{formatCurrency(price)}</span>
          </div>
        </div>

        {/* Preparation Rules Warning */}
        <div className="p-4 rounded-xl bg-amber-50/90 border border-amber-200 text-amber-900 space-y-2">
          <div className="flex items-center gap-2 font-semibold text-amber-800 text-sm">
            <ShieldAlert className="w-5 h-5 text-amber-600 flex-shrink-0" />
            <span>Muhim tayyorgarlik talablari!</span>
          </div>
          <p className="text-xs leading-relaxed text-amber-900/90 whitespace-pre-line pl-7">
            {preparationRules ||
              'Ushbu diagnostika tekshiruviga o‘z vaqtida, shaxsni tasdiqlovchi hujjat va mavjud bo‘lsa oldingi tibbiy xulosalar bilan birga kelishingiz so‘raladi.'}
          </p>
        </div>

        {/* Checkbox Acknowledgment */}
        <label className="flex items-start gap-3 p-3.5 rounded-xl border border-slate-200 hover:border-brand-blue/50 hover:bg-blue-50/20 cursor-pointer transition-colors">
          <input
            type="checkbox"
            checked={acknowledged}
            onChange={(e) => setAcknowledged(e.target.checked)}
            className="mt-0.5 h-4 w-4 rounded border-slate-300 text-brand-blue focus:ring-brand-blue"
          />
          <span className="text-xs text-slate-700 leading-normal select-none">
            Men yuqoridagi <strong>tayyorgarlik qoidalari bilan to‘liq tanishdim</strong> va qabul vaqtiga rioya qilishga roziman.
          </span>
        </label>

        {/* Actions */}
        <div className="flex items-center justify-end gap-3 pt-2">
          <Button variant="outline" onClick={handleClose} disabled={isLoading}>
            Bekor qilish
          </Button>
          <Button
            onClick={handleConfirm}
            disabled={!acknowledged || isLoading}
            isLoading={isLoading}
            className="gap-2"
          >
            <CheckCircle className="w-4 h-4" />
            Tasdiqlash va yozilish
          </Button>
        </div>
      </div>
    </Modal>
  );
}
