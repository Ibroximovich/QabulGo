'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { diagnosticApi } from '../api/diagnosticApi';
import { DiagnosticAppointment, DiagnosticAppointmentStatus } from '../types';
import { formatCurrency } from '@/lib/utils/formatters';
import {
  Calendar,
  Clock,
  Building2,
  MapPin,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Activity,
  ArrowRight,
  ShieldAlert,
  AlertCircle,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';
import { cn } from '@/lib/utils/cn';

export function DiagnosticAppointmentsPage() {
  const [activeTab, setActiveTab] = useState<'upcoming' | 'past' | 'cancelled'>('upcoming');
  const [cancellingAppt, setCancellingAppt] = useState<DiagnosticAppointment | null>(null);
  const queryClient = useQueryClient();

  const { data, isLoading, isError } = useQuery({
    queryKey: ['myDiagnosticAppointments', activeTab],
    queryFn: () => diagnosticApi.getMyAppointments({ timeframe: activeTab }),
  });

  const cancelMutation = useMutation({
    mutationFn: (appointmentId: string) =>
      diagnosticApi.cancelMyAppointment(appointmentId, 'Bemor tomonidan bekor qilindi'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['myDiagnosticAppointments'] });
      setCancellingAppt(null);
      toast.success('Diagnostika qabuli bekor qilindi');
    },
    onError: (err: any) => {
      const msg = err?.response?.data?.message || err?.message || 'Qabulni bekor qilishda xatolik yuz berdi';
      toast.error(msg);
    },
  });

  const appointments = data?.data?.appointments || [];

  const getStatusBadge = (status: DiagnosticAppointmentStatus) => {
    switch (status) {
      case 'confirmed':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
            <CheckCircle2 className="w-3.5 h-3.5" />
            Tasdiqlangan
          </span>
        );
      case 'completed':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-blue-50 text-brand-blue border border-blue-200">
            <CheckCircle2 className="w-3.5 h-3.5" />
            Yakunlangan
          </span>
        );
      case 'cancelled':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-rose-50 text-rose-700 border border-rose-200">
            <XCircle className="w-3.5 h-3.5" />
            Bekor qilingan
          </span>
        );
      case 'no_show':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200">
            <AlertTriangle className="w-3.5 h-3.5" />
            Kelmagan
          </span>
        );
    }
  };

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 py-8 space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-extrabold text-slate-900 tracking-tight">
            Mening diagnostika qabullarim
          </h1>
          <p className="text-xs sm:text-sm text-slate-500 mt-1">
            MRT, MSKT, UZI va boshqa diagnostika tekshiruvlariga yozilgan vaqtlaringiz
          </p>
        </div>

        <Link href="/diagnostics">
          <Button size="sm" className="gap-2 shadow-sm">
            <Activity className="w-4 h-4" />
            Yangi tekshiruvga yozilish
          </Button>
        </Link>
      </div>

      {/* Tabs */}
      <div className="flex gap-2 border-b border-slate-200 pb-2">
        {(
          [
            { id: 'upcoming', label: 'Kelgusi qabullar' },
            { id: 'past', label: 'O‘tgan qabullar' },
            { id: 'cancelled', label: 'Bekor qilinganlar' },
          ] as const
        ).map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={cn(
              'px-4 py-2 rounded-xl text-xs font-bold transition-all',
              activeTab === tab.id
                ? 'bg-brand-blue text-white shadow-xs'
                : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
            )}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Content */}
      {isLoading ? (
        <div className="space-y-4">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="h-44 bg-slate-100 rounded-3xl animate-pulse" />
          ))}
        </div>
      ) : isError ? (
        <div className="p-8 rounded-3xl bg-rose-50 border border-rose-200 text-center space-y-2">
          <AlertTriangle className="w-8 h-8 text-rose-500 mx-auto" />
          <p className="text-sm font-bold text-rose-900">Ma’lumotlarni yuklashda xatolik yuz berdi</p>
        </div>
      ) : appointments.length === 0 ? (
        <div className="p-12 rounded-3xl bg-slate-50 border border-slate-200/80 text-center space-y-3">
          <Activity className="w-12 h-12 text-slate-300 mx-auto" />
          <h3 className="text-base font-bold text-slate-800">Ushbu bo‘limda hozircha qabullar mavjud emas</h3>
          <p className="text-xs text-slate-500 max-w-sm mx-auto">
            Kerakli tekshiruvni topib, o‘zingizga ma’qul vaqtga onlayn yozilishingiz mumkin.
          </p>
          <div className="pt-2">
            <Link href="/diagnostics">
              <Button size="sm" variant="outline" className="gap-2">
                Tekshiruvlarni ko‘rish
                <ArrowRight className="w-4 h-4" />
              </Button>
            </Link>
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          {appointments.map((appt) => (
            <div
              key={appt.id}
              className="bg-white rounded-3xl border border-slate-200/80 p-5 sm:p-6 shadow-xs hover:border-slate-300 transition-all space-y-4"
            >
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-slate-100">
                <div className="space-y-1">
                  <span className="text-[11px] font-bold text-brand-blue uppercase tracking-wider block">
                    {appt.service?.category?.name || 'Diagnostika'}
                  </span>
                  <h3 className="text-base font-bold text-slate-900">{appt.service?.name}</h3>
                </div>
                <div>{getStatusBadge(appt.status)}</div>
              </div>

              {/* Grid info */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs text-slate-700">
                <div className="flex items-center gap-2 p-3 rounded-2xl bg-slate-50">
                  <Calendar className="w-4 h-4 text-brand-blue flex-shrink-0" />
                  <div>
                    <span className="text-[10px] text-slate-400 block font-medium">Sana va vaqt</span>
                    <span className="font-bold text-slate-900">
                      {appt.appointment_date} ({appt.start_time} - {appt.end_time})
                    </span>
                  </div>
                </div>

                <div className="flex items-center gap-2 p-3 rounded-2xl bg-slate-50">
                  <Building2 className="w-4 h-4 text-brand-blue flex-shrink-0" />
                  <div className="truncate">
                    <span className="text-[10px] text-slate-400 block font-medium">Klinika</span>
                    <span className="font-bold text-slate-900 truncate block">{appt.clinic?.name}</span>
                  </div>
                </div>

                <div className="flex items-center justify-between p-3 rounded-2xl bg-slate-50">
                  <div>
                    <span className="text-[10px] text-slate-400 block font-medium">To‘lov</span>
                    <span className="font-extrabold text-slate-900 text-sm">
                      {formatCurrency(appt.price)}
                    </span>
                  </div>
                </div>
              </div>

              {/* Equipment info */}
              {appt.equipment && (
                <div className="text-xs text-slate-600 flex items-center gap-2">
                  <span className="font-medium text-slate-500">Uskuna / xona:</span>
                  <span className="font-bold text-slate-800">{appt.equipment.name}</span>
                  {appt.equipment.room_number && (
                    <span className="text-slate-400">({appt.equipment.room_number})</span>
                  )}
                </div>
              )}

              {/* Preparation Rules Reminder */}
              {appt.preparation_rules && appt.status === 'confirmed' && (
                <div className="flex items-start gap-2 p-3 rounded-2xl bg-amber-50/80 border border-amber-200/60 text-amber-900 text-xs">
                  <ShieldAlert className="w-4 h-4 text-amber-600 flex-shrink-0 mt-0.5" />
                  <div className="leading-relaxed">
                    <strong>Tayyorgarlik eslatmasi:</strong> {appt.preparation_rules}
                  </div>
                </div>
              )}

              {/* Actions */}
              {appt.status === 'confirmed' && (
                <div className="pt-2 flex justify-end">
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-rose-600 border-rose-200 hover:bg-rose-50"
                    onClick={() => setCancellingAppt(appt)}
                    disabled={cancelMutation.isPending}
                  >
                    Qabulni bekor qilish
                  </Button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Cancel Appointment Confirmation Modal */}
      <Modal
        isOpen={!!cancellingAppt}
        onClose={() => setCancellingAppt(null)}
        title="Qabulni bekor qilish"
      >
        <div className="space-y-4">
          <div className="flex items-start gap-3 p-3 bg-amber-50 border border-amber-200 rounded-xl text-amber-900 text-sm">
            <AlertCircle className="w-5 h-5 flex-shrink-0 mt-0.5 text-amber-600" />
            <div>
              <p className="font-semibold">
                &ldquo;{cancellingAppt?.service?.name}&rdquo; qabulini bekor qilishga ishonchingiz komilmi?
              </p>
              <p className="text-xs text-amber-800 mt-1">
                Sana: <strong>{cancellingAppt?.appointment_date}</strong> ({cancellingAppt?.start_time} - {cancellingAppt?.end_time})
              </p>
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button
              variant="outline"
              onClick={() => setCancellingAppt(null)}
              disabled={cancelMutation.isPending}
            >
              Orqaga
            </Button>
            <Button
              className="bg-rose-600 hover:bg-rose-700 text-white"
              onClick={() => cancellingAppt && cancelMutation.mutate(cancellingAppt.id)}
              isLoading={cancelMutation.isPending}
            >
              Ha, bekor qilinsin
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}

