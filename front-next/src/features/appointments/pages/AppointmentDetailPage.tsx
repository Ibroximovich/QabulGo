'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { useParams, useRouter } from 'next/navigation';
import { AppLayout } from '@/components/layout/AppLayout';
import { useMyAppointmentDetail, useCancelMyAppointment } from '../hooks/useAppointments';
import { AppointmentStatusBadge } from '../components/AppointmentStatusBadge';
import { CancelConfirmModal } from '../components/CancelConfirmModal';
import { formatCurrency } from '@/lib/utils/formatters';
import { getMediaUrl } from '@/lib/utils/media';
import { Button } from '@/components/ui/Button';
import {
  Calendar,
  Clock,
  Building2,
  MapPin,
  Phone,
  User,
  CreditCard,
  ChevronLeft,
  AlertCircle,
  CheckCircle2,
  CalendarX,
  Stethoscope,
  ExternalLink,
  ShieldCheck,
} from 'lucide-react';
import { toast } from 'sonner';

const UZ_MONTHS = [
  'yanvar',
  'fevral',
  'mart',
  'aprel',
  'may',
  'iyun',
  'iyul',
  'avgust',
  'sentabr',
  'oktabr',
  'noyabr',
  'dekabr',
];

export function AppointmentDetailPage() {
  const params = useParams();
  const router = useRouter();
  const id = params?.id as string;

  const [isCancelModalOpen, setIsCancelModalOpen] = useState(false);

  const { data: appointment, isLoading, isError, refetch } = useMyAppointmentDetail(id);
  const cancelMutation = useCancelMyAppointment();

  if (isLoading) {
    return (
      <AppLayout>
        <div className="max-w-3xl mx-auto px-4 sm:px-6 py-10 space-y-6 animate-pulse">
          <div className="h-6 w-32 bg-slate-200 rounded" />
          <div className="h-48 bg-white rounded-3xl border border-slate-200" />
          <div className="h-64 bg-white rounded-3xl border border-slate-200" />
        </div>
      </AppLayout>
    );
  }

  if (isError || !appointment) {
    return (
      <AppLayout>
        <div className="max-w-md mx-auto px-4 py-16 text-center space-y-4">
          <div className="w-16 h-16 rounded-2xl bg-rose-50 text-rose-600 flex items-center justify-center mx-auto">
            <AlertCircle className="w-8 h-8" />
          </div>
          <h2 className="text-xl font-bold text-slate-800">Qabul topilmadi</h2>
          <p className="text-xs text-slate-500">
            Bunday ID ga ega qabul mavjud emas yoki sizga ko‘rish uchun ruxsat berilmagan.
          </p>
          <Link href="/appointments">
            <Button variant="primary">Qabullar ro‘yxatiga qaytish</Button>
          </Link>
        </div>
      </AppLayout>
    );
  }

  // Format Date
  let formattedDate = appointment.appointment_date;
  try {
    const [y, m, d] = appointment.appointment_date.split('-').map(Number);
    const dateObj = new Date(y, m - 1, d);
    const day = dateObj.getDate();
    const month = UZ_MONTHS[dateObj.getMonth()];
    const year = dateObj.getFullYear();
    const weekday = [
      'Yakshanba',
      'Dushanba',
      'Seshanba',
      'Chorshanba',
      'Payshanba',
      'Juma',
      'Shanba',
    ][dateObj.getDay()];
    formattedDate = `${day}-${month}, ${year}-yil (${weekday})`;
  } catch {
    // Keep raw
  }

  const doctorFullName = appointment.doctor
    ? `${appointment.doctor.first_name} ${appointment.doctor.last_name}`
    : 'Shifokor';

  const handleConfirmCancel = async (reason?: string) => {
    try {
      await cancelMutation.mutateAsync({
        id: appointment.id,
        reason,
      });
      toast.success('Qabul muvaffaqiyatli bekor qilindi');
      setIsCancelModalOpen(false);
      refetch();
    } catch (err: any) {
      const msg =
        err?.response?.data?.error ||
        err?.response?.data?.message ||
        'Qabulni bekor qilishda xatolik yuz berdi';
      toast.error(msg);
    }
  };

  const isCancellable = appointment.status === 'confirmed';

  return (
    <AppLayout>
      <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
        {/* Navigation & Status Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-200">
          <div className="space-y-1">
            <Link
              href="/appointments"
              className="inline-flex items-center gap-1.5 text-xs font-bold text-slate-500 hover:text-brand-blue transition-colors"
            >
              <ChevronLeft className="w-4 h-4" />
              <span>Barcha qabullarim</span>
            </Link>
            <h1 className="text-2xl font-extrabold text-brand-navy">Qabul tafsilotlari</h1>
          </div>

          <div>
            <AppointmentStatusBadge status={appointment.status} size="lg" />
          </div>
        </div>

        {/* Instant Confirmation Banner */}
        {appointment.status === 'confirmed' && (
          <div className="bg-emerald-50 border border-emerald-200 rounded-2xl p-4 flex items-start gap-3">
            <CheckCircle2 className="w-5 h-5 text-emerald-600 shrink-0 mt-0.5" />
            <div className="text-xs text-emerald-900 space-y-0.5">
              <span className="font-bold">Bron avtomatik tasdiqlangan</span>
              <p className="text-emerald-700">
                Klinika ma’muriyati ushbu qabulni avtomatik ravishda tasdiqladi. Shifokor sizni
                belgilangan vaqtda kutadi. Iltimos, 10 daqiqa oldin klinikaga yetib boring.
              </p>
            </div>
          </div>
        )}

        {/* Cancelled Banner */}
        {appointment.status === 'cancelled' && (
          <div className="bg-rose-50 border border-rose-200 rounded-2xl p-4 flex items-start gap-3">
            <CalendarX className="w-5 h-5 text-rose-600 shrink-0 mt-0.5" />
            <div className="text-xs text-rose-900 space-y-0.5">
              <span className="font-bold">Ushbu qabul bekor qilingan</span>
              {appointment.cancellation_reason && (
                <p className="text-rose-700">Sababi: {appointment.cancellation_reason}</p>
              )}
            </div>
          </div>
        )}

        {/* Main Details Card */}
        <div className="bg-white rounded-3xl border border-slate-200/80 p-6 sm:p-8 space-y-6 shadow-subtle">
          {/* Date & Time Highlight */}
          <div className="bg-gradient-to-r from-blue-50 to-indigo-50/50 rounded-2xl p-5 border border-blue-100 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-brand-blue">
                <Calendar className="w-5 h-5" />
                <span className="text-sm font-bold">{formattedDate}</span>
              </div>
              <div className="flex items-center gap-2 text-slate-800">
                <Clock className="w-5 h-5 text-slate-400" />
                <span className="font-mono text-lg font-extrabold">
                  {appointment.start_time} – {appointment.end_time}
                </span>
              </div>
            </div>

            <div className="sm:text-right border-t sm:border-t-0 pt-3 sm:pt-0 border-blue-100">
              <span className="text-[11px] font-bold text-slate-400 uppercase tracking-wider block">
                Qabul narxi
              </span>
              <span className="text-xl font-extrabold text-brand-navy">
                {appointment.consultation_price > 0
                  ? formatCurrency(appointment.consultation_price)
                  : 'Bepul'}
              </span>
            </div>
          </div>

          {/* Doctor Info */}
          <div className="pt-2">
            <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-3">
              Shifokor haqida
            </h3>
            <div className="flex items-start justify-between gap-4 p-4 rounded-2xl bg-slate-50 border border-slate-100">
              <div className="flex items-center gap-4">
                <div className="w-14 h-14 rounded-2xl bg-white border border-slate-200 flex items-center justify-center overflow-hidden shrink-0 shadow-2xs">
                  {appointment.doctor?.photo_url ? (
                    <img
                      src={getMediaUrl(appointment.doctor.photo_url)}
                      alt={doctorFullName}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <User className="w-7 h-7 text-slate-400" />
                  )}
                </div>

                <div>
                  <h4 className="font-bold text-slate-900 text-base">{doctorFullName}</h4>
                  <p className="text-xs text-brand-blue font-semibold flex items-center gap-1.5 mt-0.5">
                    <Stethoscope className="w-3.5 h-3.5" />
                    <span>{appointment.doctor?.specialty || 'Mutaxassis'}</span>
                  </p>
                </div>
              </div>

              {appointment.doctor_id && (
                <Link href={`/doctors/${appointment.doctor_id}`}>
                  <Button variant="secondary" size="sm" className="hidden sm:inline-flex items-center gap-1 text-xs">
                    <span>Profil</span>
                    <ExternalLink className="w-3 h-3" />
                  </Button>
                </Link>
              )}
            </div>
          </div>

          {/* Clinic Info */}
          <div>
            <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-3">
              Klinika manzili
            </h3>
            <div className="p-4 rounded-2xl bg-slate-50 border border-slate-100 space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Building2 className="w-4 h-4 text-brand-blue shrink-0" />
                  <span className="font-bold text-slate-900 text-sm">
                    {appointment.clinic?.name || 'Klinika'}
                  </span>
                </div>

                {appointment.clinic_id && (
                  <Link href={`/clinics/${appointment.clinic_id}`}>
                    <span className="text-xs font-bold text-brand-blue hover:underline inline-flex items-center gap-1">
                      Klinika sahifasi
                      <ExternalLink className="w-3 h-3" />
                    </span>
                  </Link>
                )}
              </div>

              <div className="text-xs text-slate-600 flex items-start gap-2">
                <MapPin className="w-4 h-4 text-slate-400 shrink-0 mt-0.5" />
                <span>
                  {appointment.clinic?.address || ''}
                  {appointment.clinic?.district ? `, ${appointment.clinic.district}` : ''}
                  {appointment.clinic?.region ? `, ${appointment.clinic.region}` : ''}
                </span>
              </div>

              {appointment.clinic?.phone && (
                <div className="text-xs text-slate-600 flex items-center gap-2">
                  <Phone className="w-4 h-4 text-slate-400 shrink-0" />
                  <a
                    href={`tel:${appointment.clinic.phone}`}
                    className="font-mono hover:text-brand-blue"
                  >
                    {appointment.clinic.phone}
                  </a>
                </div>
              )}
            </div>
          </div>

          {/* Cancellation section */}
          {isCancellable && (
            <div className="pt-4 border-t border-slate-100 flex items-center justify-between gap-4">
              <p className="text-xs text-slate-500">
                Agar rejalaringiz o‘zgarsa, shifokor va boshqa bemorlarga qulaylik yaratish uchun
                qabulni oldindan bekor qiling.
              </p>
              <Button
                variant="danger"
                size="sm"
                onClick={() => setIsCancelModalOpen(true)}
                className="shrink-0"
              >
                Qabulni bekor qilish
              </Button>
            </div>
          )}
        </div>

        {/* Cancel Confirmation Modal */}
        <CancelConfirmModal
          isOpen={isCancelModalOpen}
          onClose={() => setIsCancelModalOpen(false)}
          onConfirm={handleConfirmCancel}
          isLoading={cancelMutation.isPending}
          title="Qabulni bekor qilish"
          description={`Shifokor ${doctorFullName} qabuliga yozilgan navbatingizni bekor qilmoqchimisiz?`}
        />
      </div>
    </AppLayout>
  );
}
