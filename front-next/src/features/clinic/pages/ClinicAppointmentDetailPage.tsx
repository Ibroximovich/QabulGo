'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';
import { useClinicContext } from '@/components/layout/clinic/ClinicLayout';
import {
  useClinicAppointmentDetail,
  useCancelClinicAppointment,
  useCompleteClinicAppointment,
  useNoShowClinicAppointment,
} from '@/features/appointments/hooks/useAppointments';
import { AppointmentStatusBadge } from '@/features/appointments/components/AppointmentStatusBadge';
import { CancelConfirmModal } from '@/features/appointments/components/CancelConfirmModal';
import { formatCurrency } from '@/lib/utils/formatters';
import { getMediaUrl } from '@/lib/utils/media';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';
import {
  Calendar,
  Clock,
  User,
  Phone,
  ChevronLeft,
  AlertCircle,
  CheckCircle2,
  CalendarX,
  Stethoscope,
  ExternalLink,
  Check,
  UserX,
} from 'lucide-react';
import { toast } from 'sonner';

export function ClinicAppointmentDetailPage() {
  const params = useParams();
  const router = useRouter();
  const appointmentId = params?.appointmentId as string;
  const { clinic } = useClinicContext();
  const clinicId = clinic?.id || '';

  const [actionType, setActionType] = useState<'cancel' | 'complete' | 'no_show' | null>(null);

  const {
    data: appointment,
    isLoading,
    isError,
    refetch,
  } = useClinicAppointmentDetail(clinicId, appointmentId, Boolean(clinicId && appointmentId));

  const cancelMutation = useCancelClinicAppointment(clinicId);
  const completeMutation = useCompleteClinicAppointment(clinicId);
  const noShowMutation = useNoShowClinicAppointment(clinicId);

  if (isLoading) {
    return (
      <div className="space-y-6 max-w-3xl animate-pulse">
        <div className="h-6 w-32 bg-slate-200 rounded" />
        <div className="h-64 bg-white rounded-3xl border border-slate-200" />
      </div>
    );
  }

  if (isError || !appointment) {
    return (
      <div className="max-w-md mx-auto py-16 text-center space-y-4">
        <AlertCircle className="w-10 h-10 text-rose-500 mx-auto" />
        <h2 className="text-lg font-bold text-slate-800">Bron topilmadi</h2>
        <p className="text-xs text-slate-500">
          Ushbu qabul mavjud emas yoki bu klinikaga tegishli emas.
        </p>
        <Link href="/clinic/appointments">
          <Button variant="primary" size="sm">
            Bronlar ro‘yxatiga qaytish
          </Button>
        </Link>
      </div>
    );
  }

  const patientFullName = appointment.patient
    ? `${appointment.patient.first_name} ${appointment.patient.last_name}`
    : 'Bemor';
  const doctorFullName = appointment.doctor
    ? `${appointment.doctor.first_name} ${appointment.doctor.last_name}`
    : 'Shifokor';

  const handleConfirmCancel = async (reason?: string) => {
    try {
      await cancelMutation.mutateAsync({
        appointmentId: appointment.id,
        reason: reason || 'Klinika tomonidan bekor qilindi',
      });
      toast.success('Qabul bekor qilindi');
      setActionType(null);
      refetch();
    } catch (err: any) {
      toast.error(
        err?.response?.data?.error ||
          err?.response?.data?.message ||
          'Qabulni bekor qilishda xatolik yuz berdi'
      );
    }
  };

  const handleConfirmComplete = async () => {
    try {
      await completeMutation.mutateAsync(appointment.id);
      toast.success('Qabul muvaffaqiyatli yakunlandi deb belgilandi');
      setActionType(null);
      refetch();
    } catch (err: any) {
      toast.error(
        err?.response?.data?.error ||
          err?.response?.data?.message ||
          'Qabulni yakunlashda xatolik yuz berdi'
      );
    }
  };

  const handleConfirmNoShow = async () => {
    try {
      await noShowMutation.mutateAsync(appointment.id);
      toast.success('Bemor kelmadi deb belgilandi');
      setActionType(null);
      refetch();
    } catch (err: any) {
      toast.error(
        err?.response?.data?.error ||
          err?.response?.data?.message ||
          'Kelmadi deb belgilashda xatolik yuz berdi'
      );
    }
  };

  return (
    <div className="space-y-6 max-w-3xl">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-200">
        <div className="space-y-1">
          <Link
            href="/clinic/appointments"
            className="inline-flex items-center gap-1.5 text-xs font-bold text-slate-500 hover:text-brand-blue transition-colors"
          >
            <ChevronLeft className="w-4 h-4" />
            <span>Barcha bronlar</span>
          </Link>
          <h1 className="text-2xl font-extrabold text-brand-navy">Bron tafsilotlari</h1>
        </div>

        <AppointmentStatusBadge status={appointment.status} size="lg" />
      </div>

      {/* Main Info Card */}
      <div className="bg-white rounded-3xl border border-slate-200/80 p-6 sm:p-8 space-y-6 shadow-subtle">
        {/* Date and Time Header */}
        <div className="bg-gradient-to-r from-blue-50 to-indigo-50/50 rounded-2xl p-5 border border-blue-100 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="space-y-1">
            <div className="flex items-center gap-2 text-brand-blue font-bold text-sm">
              <Calendar className="w-5 h-5" />
              <span>{appointment.appointment_date}</span>
            </div>
            <div className="flex items-center gap-2 text-slate-900 font-mono text-lg font-extrabold">
              <Clock className="w-5 h-5 text-slate-400" />
              <span>
                {appointment.start_time} – {appointment.end_time}
              </span>
            </div>
          </div>

          <div className="sm:text-right border-t sm:border-t-0 pt-3 sm:pt-0 border-blue-100">
            <span className="text-[11px] font-bold text-slate-400 uppercase tracking-wider block">
              Konsultatsiya narxi
            </span>
            <span className="text-xl font-extrabold text-brand-navy">
              {appointment.consultation_price > 0
                ? formatCurrency(appointment.consultation_price)
                : 'Bepul'}
            </span>
          </div>
        </div>

        {/* Patient Details */}
        <div>
          <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-3">
            Bemor ma’lumotlari
          </h3>
          <div className="p-4 rounded-2xl bg-slate-50 border border-slate-100 flex items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-2xl bg-blue-100 text-brand-blue flex items-center justify-center font-bold text-lg shrink-0">
                {patientFullName[0]}
              </div>
              <div>
                <h4 className="font-bold text-slate-900 text-sm">{patientFullName}</h4>
                {appointment.patient?.phone && (
                  <a
                    href={`tel:${appointment.patient.phone}`}
                    className="text-xs font-mono text-slate-500 hover:text-brand-blue flex items-center gap-1.5 mt-0.5"
                  >
                    <Phone className="w-3.5 h-3.5 text-slate-400" />
                    <span>{appointment.patient.phone}</span>
                  </a>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Doctor Details */}
        <div>
          <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-3">
            Biriktirilgan shifokor
          </h3>
          <div className="p-4 rounded-2xl bg-slate-50 border border-slate-100 flex items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-2xl bg-white border border-slate-200 flex items-center justify-center overflow-hidden shrink-0">
                {appointment.doctor?.photo_url ? (
                  <img
                    src={getMediaUrl(appointment.doctor.photo_url)}
                    alt={doctorFullName}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <User className="w-6 h-6 text-slate-400" />
                )}
              </div>
              <div>
                <h4 className="font-bold text-slate-900 text-sm">{doctorFullName}</h4>
                <p className="text-xs text-brand-blue font-semibold flex items-center gap-1.5 mt-0.5">
                  <Stethoscope className="w-3.5 h-3.5" />
                  <span>{appointment.doctor?.specialty || 'Mutaxassis'}</span>
                </p>
              </div>
            </div>

            {appointment.doctor_id && (
              <Link href={`/clinic/doctors/${appointment.doctor_id}/schedule`}>
                <Button variant="secondary" size="sm" className="text-xs inline-flex items-center gap-1">
                  <span>Jadval</span>
                  <ExternalLink className="w-3 h-3" />
                </Button>
              </Link>
            )}
          </div>
        </div>

        {/* Cancellation or Status Notice */}
        {appointment.status === 'cancelled' && (
          <div className="p-4 rounded-2xl bg-rose-50 border border-rose-200 text-xs text-rose-800 space-y-1">
            <div className="font-bold flex items-center gap-1.5 text-rose-900">
              <CalendarX className="w-4 h-4 text-rose-600" />
              <span>Bekor qilingan</span>
            </div>
            {appointment.cancellation_reason && (
              <p>Sabab: {appointment.cancellation_reason}</p>
            )}
          </div>
        )}

        {/* Action Buttons for Confirmed Status */}
        {appointment.status === 'confirmed' && (
          <div className="pt-4 border-t border-slate-100 flex flex-wrap items-center justify-end gap-2.5">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setActionType('cancel')}
              className="text-rose-600 hover:bg-rose-50 hover:border-rose-200"
            >
              Bekor qilish
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setActionType('no_show')}
              className="text-amber-700 hover:bg-amber-50 hover:border-amber-200"
            >
              Kelmadi deb belgilash
            </Button>
            <Button
              variant="primary"
              size="sm"
              onClick={() => setActionType('complete')}
              className="inline-flex items-center gap-1.5"
            >
              <Check className="w-3.5 h-3.5" />
              <span>Yakunlash</span>
            </Button>
          </div>
        )}
      </div>

      {/* Cancel Modal */}
      <CancelConfirmModal
        isOpen={actionType === 'cancel'}
        onClose={() => setActionType(null)}
        onConfirm={handleConfirmCancel}
        isLoading={cancelMutation.isPending}
        title="Bronni bekor qilish"
        description="Ushbu qabulni bekor qilish sababini kiriting:"
        requireReason={true}
      />

      {/* Complete Modal */}
      <Modal
        isOpen={actionType === 'complete'}
        onClose={() => setActionType(null)}
        title="Qabulni yakunlash"
        className="max-w-md"
      >
        <div className="space-y-4 pt-2">
          <p className="text-xs text-slate-600">
            Shifokor qabuli muvaffaqiyatli yakunlanganligini tasdiqlaysizmi?
          </p>
          <div className="flex items-center justify-end gap-2 pt-2">
            <Button
              variant="secondary"
              onClick={() => setActionType(null)}
              disabled={completeMutation.isPending}
            >
              Bekor qilish
            </Button>
            <Button
              variant="primary"
              onClick={handleConfirmComplete}
              disabled={completeMutation.isPending}
            >
              Ha, yakunlandi
            </Button>
          </div>
        </div>
      </Modal>

      {/* No Show Modal */}
      <Modal
        isOpen={actionType === 'no_show'}
        onClose={() => setActionType(null)}
        title="Bemor kelmadi deb belgilash"
        className="max-w-md"
      >
        <div className="space-y-4 pt-2">
          <p className="text-xs text-slate-600">
            Bemor qabulga kelmadi deb qayd etilsinmi?
          </p>
          <div className="flex items-center justify-end gap-2 pt-2">
            <Button
              variant="secondary"
              onClick={() => setActionType(null)}
              disabled={noShowMutation.isPending}
            >
              Bekor qilish
            </Button>
            <Button
              variant="danger"
              onClick={handleConfirmNoShow}
              disabled={noShowMutation.isPending}
            >
              Ha, kelmadi
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
