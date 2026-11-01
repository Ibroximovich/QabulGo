'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { AppLayout } from '@/components/layout/AppLayout';
import { useMyAppointments, useCancelMyAppointment } from '../hooks/useAppointments';
import { AppointmentCard } from '../components/AppointmentCard';
import { CancelConfirmModal } from '../components/CancelConfirmModal';
import { Appointment } from '../types';
import { Button } from '@/components/ui/Button';
import {
  Calendar,
  Clock,
  CalendarCheck,
  CalendarX,
  History,
  Plus,
  ArrowRight,
  Search,
  AlertCircle,
  Loader2,
} from 'lucide-react';
import { toast } from 'sonner';

type TimeframeTab = 'upcoming' | 'past' | 'cancelled';

export function MyAppointmentsPage() {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState<TimeframeTab>('upcoming');
  const [page, setPage] = useState<number>(1);
  const limit = 10;

  // Selected appointment for cancellation
  const [cancellingAppointment, setCancellingAppointment] = useState<Appointment | null>(null);

  // Queries
  const { data, isLoading, isError, refetch } = useMyAppointments({
    timeframe: activeTab,
    page,
    limit,
  });

  const cancelMutation = useCancelMyAppointment();

  const appointments = data?.appointments || [];
  const total = data?.total || 0;
  const totalPages = Math.ceil(total / limit) || 1;

  const handleTabChange = (tab: TimeframeTab) => {
    setActiveTab(tab);
    setPage(1);
  };

  const handleOpenCancelModal = (appointment: Appointment) => {
    setCancellingAppointment(appointment);
  };

  const handleConfirmCancel = async (reason?: string) => {
    if (!cancellingAppointment) return;

    try {
      await cancelMutation.mutateAsync({
        id: cancellingAppointment.id,
        reason,
      });
      toast.success('Qabul muvaffaqiyatli bekor qilindi');
      setCancellingAppointment(null);
      refetch();
    } catch (err: any) {
      const msg =
        err?.response?.data?.error ||
        err?.response?.data?.message ||
        'Qabulni bekor qilishda xatolik yuz berdi';
      toast.error(msg);
    }
  };

  return (
    <AppLayout>
      <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-6 border-b border-slate-200">
          <div>
            <div className="flex items-center gap-2 text-xs text-slate-500 mb-1">
              <Link href="/dashboard" className="hover:text-brand-blue transition-colors">
                Kabinet
              </Link>
              <span>/</span>
              <span className="text-slate-700 font-semibold">Mening qabullarim</span>
            </div>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-brand-navy flex items-center gap-3">
              <Calendar className="w-8 h-8 text-brand-blue" />
              <span>Mening bronlarim</span>
            </h1>
            <p className="text-sm text-slate-500 mt-1">
              Shifokor qabuliga yozilgan barcha navbatlaringiz va ularning holati
            </p>
          </div>

          <Link href="/doctors">
            <Button variant="primary" className="inline-flex items-center gap-2 shadow-xs">
              <Plus className="w-4 h-4" />
              <span>Yangi qabulga yozilish</span>
            </Button>
          </Link>
        </div>

        {/* Tabs */}
        <div className="flex items-center gap-2 border-b border-slate-200 overflow-x-auto pb-px">
          <button
            type="button"
            onClick={() => handleTabChange('upcoming')}
            className={`flex items-center gap-2 py-3 px-4 font-bold text-sm border-b-2 transition-all whitespace-nowrap ${
              activeTab === 'upcoming'
                ? 'border-brand-blue text-brand-blue'
                : 'border-transparent text-slate-500 hover:text-slate-800 hover:border-slate-300'
            }`}
          >
            <CalendarCheck className="w-4 h-4" />
            <span>Kelgusi qabullar</span>
            {activeTab === 'upcoming' && total > 0 && (
              <span className="px-2 py-0.5 text-xs rounded-full bg-blue-100 text-brand-blue font-mono font-bold">
                {total}
              </span>
            )}
          </button>

          <button
            type="button"
            onClick={() => handleTabChange('past')}
            className={`flex items-center gap-2 py-3 px-4 font-bold text-sm border-b-2 transition-all whitespace-nowrap ${
              activeTab === 'past'
                ? 'border-brand-blue text-brand-blue'
                : 'border-transparent text-slate-500 hover:text-slate-800 hover:border-slate-300'
            }`}
          >
            <History className="w-4 h-4" />
            <span>O‘tgan qabullar</span>
            {activeTab === 'past' && total > 0 && (
              <span className="px-2 py-0.5 text-xs rounded-full bg-slate-100 text-slate-700 font-mono font-bold">
                {total}
              </span>
            )}
          </button>

          <button
            type="button"
            onClick={() => handleTabChange('cancelled')}
            className={`flex items-center gap-2 py-3 px-4 font-bold text-sm border-b-2 transition-all whitespace-nowrap ${
              activeTab === 'cancelled'
                ? 'border-brand-blue text-brand-blue'
                : 'border-transparent text-slate-500 hover:text-slate-800 hover:border-slate-300'
            }`}
          >
            <CalendarX className="w-4 h-4" />
            <span>Bekor qilinganlar</span>
            {activeTab === 'cancelled' && total > 0 && (
              <span className="px-2 py-0.5 text-xs rounded-full bg-rose-100 text-rose-700 font-mono font-bold">
                {total}
              </span>
            )}
          </button>
        </div>

        {/* Content */}
        {isLoading ? (
          <div className="space-y-4 py-8">
            {[1, 2, 3].map((i) => (
              <div
                key={i}
                className="h-44 bg-white rounded-2xl border border-slate-200/80 p-6 animate-pulse"
              >
                <div className="flex gap-4">
                  <div className="w-14 h-14 bg-slate-200 rounded-2xl" />
                  <div className="space-y-2 flex-1">
                    <div className="w-48 h-4 bg-slate-200 rounded" />
                    <div className="w-32 h-3 bg-slate-200 rounded" />
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : isError ? (
          <div className="bg-rose-50 border border-rose-200 rounded-2xl p-8 text-center space-y-3">
            <AlertCircle className="w-8 h-8 text-rose-600 mx-auto" />
            <h3 className="font-bold text-slate-800 text-base">
              Qabullarni yuklashda xatolik yuz berdi
            </h3>
            <p className="text-xs text-slate-500 max-w-sm mx-auto">
              Internet aloqasini tekshiring va sahifani qayta yangilang.
            </p>
            <Button variant="secondary" onClick={() => refetch()} className="text-xs">
              Qayta urinish
            </Button>
          </div>
        ) : appointments.length === 0 ? (
          <div className="bg-white border border-slate-200/80 rounded-3xl p-12 text-center space-y-4 shadow-subtle">
            <div className="w-16 h-16 rounded-2xl bg-blue-50 text-brand-blue flex items-center justify-center mx-auto">
              <Calendar className="w-8 h-8" />
            </div>
            <div className="space-y-1">
              <h3 className="text-lg font-bold text-slate-800">
                {activeTab === 'upcoming'
                  ? 'Rejalashtirilgan qabullar yo‘q'
                  : activeTab === 'past'
                  ? 'O‘tgan qabullar tarixi topilmadi'
                  : 'Bekor qilingan qabullar mavjud emas'}
              </h3>
              <p className="text-xs text-slate-500 max-w-md mx-auto">
                {activeTab === 'upcoming'
                  ? 'Siz hali birorta shifokor qabuliga yozilmagansiz. QabulGo orqali kerakli mutaxassisni toping va navbat oling.'
                  : 'Sizning ushbu toifadagi qabullaringiz ro‘yxati bo‘sh.'}
              </p>
            </div>
            {activeTab === 'upcoming' && (
              <div className="pt-2">
                <Link href="/doctors">
                  <Button variant="primary" className="inline-flex items-center gap-2">
                    <Search className="w-4 h-4" />
                    <span>Shifokorlarni ko‘rish</span>
                  </Button>
                </Link>
              </div>
            )}
          </div>
        ) : (
          <div className="space-y-4">
            {appointments.map((appointment) => (
              <AppointmentCard
                key={appointment.id}
                appointment={appointment}
                role="patient"
                onCancel={handleOpenCancelModal}
              />
            ))}

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-between pt-6 border-t border-slate-200">
                <p className="text-xs text-slate-500">
                  Jami <span className="font-bold text-slate-800">{total}</span> ta qabuldan{' '}
                  <span className="font-bold text-slate-800">
                    {(page - 1) * limit + 1}-{Math.min(page * limit, total)}
                  </span>{' '}
                  ko‘rsatilmoqda
                </p>

                <div className="flex items-center gap-2">
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={page <= 1}
                    onClick={() => setPage((p) => Math.max(p - 1, 1))}
                  >
                    Oldingi
                  </Button>
                  <span className="text-xs font-bold text-slate-600 px-2">
                    {page} / {totalPages}
                  </span>
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={page >= totalPages}
                    onClick={() => setPage((p) => Math.min(p + 1, totalPages))}
                  >
                    Keyingi
                  </Button>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Cancellation Modal */}
        <CancelConfirmModal
          isOpen={Boolean(cancellingAppointment)}
          onClose={() => setCancellingAppointment(null)}
          onConfirm={handleConfirmCancel}
          isLoading={cancelMutation.isPending}
          title="Qabulni bekor qilish"
          description={`Shifokor ${
            cancellingAppointment?.doctor
              ? `${cancellingAppointment.doctor.first_name} ${cancellingAppointment.doctor.last_name}`
              : ''
          } qabuliga yozilgan navbatingizni bekor qilmoqchimisiz?`}
          requireReason={false}
        />
      </div>
    </AppLayout>
  );
}
