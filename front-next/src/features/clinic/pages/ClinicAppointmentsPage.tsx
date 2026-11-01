'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useClinicContext } from '@/components/layout/clinic/ClinicLayout';
import { clinicApi } from '../api/clinicApi';
import {
  useClinicAppointments,
  useCancelClinicAppointment,
  useCompleteClinicAppointment,
  useNoShowClinicAppointment,
} from '@/features/appointments/hooks/useAppointments';
import { AppointmentStatusBadge } from '@/features/appointments/components/AppointmentStatusBadge';
import { CancelConfirmModal } from '@/features/appointments/components/CancelConfirmModal';
import { Appointment, AppointmentStatus } from '@/features/appointments/types';
import { formatCurrency } from '@/lib/utils/formatters';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';
import {
  Calendar,
  Clock,
  User,
  Phone,
  Search,
  Filter,
  CheckCircle2,
  XCircle,
  AlertCircle,
  Eye,
  X,
  Stethoscope,
  CalendarCheck,
  UserX,
  Check,
} from 'lucide-react';

function getTashkentDateString(offsetDays = 0): string {
  const now = new Date();
  const tashkentOffsetMs = 5 * 60 * 60 * 1000;
  const utc = now.getTime() + now.getTimezoneOffset() * 60 * 1000;
  const target = new Date(utc + tashkentOffsetMs + offsetDays * 24 * 60 * 60 * 1000);
  return target.toISOString().split('T')[0];
}

export function ClinicAppointmentsPage() {
  const router = useRouter();
  const { clinic, isLoadingClinic } = useClinicContext();
  const clinicId = clinic?.id || '';

  // Filter States
  const todayStr = getTashkentDateString(0);
  const tomorrowStr = getTashkentDateString(1);

  const [dateFilter, setDateFilter] = useState<string>('');
  const [doctorIdFilter, setDoctorIdFilter] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [page, setPage] = useState<number>(1);
  const limit = 20;

  // Action Modals State
  const [selectedAppointment, setSelectedAppointment] = useState<Appointment | null>(null);
  const [actionType, setActionType] = useState<'cancel' | 'complete' | 'no_show' | null>(null);

  // Fetch Doctors for filter dropdown
  const { data: doctorsRes } = useQuery({
    queryKey: ['clinicDoctorsList', clinicId],
    queryFn: async () => {
      if (!clinicId) return [];
      const res = await clinicApi.listDoctors(clinicId);
      return res.data?.data || [];
    },
    enabled: Boolean(clinicId),
  });
  const clinicDoctors = doctorsRes || [];

  // Fetch Appointments
  const {
    data: appointmentsData,
    isLoading: isLoadingAppointments,
    isError,
    refetch,
  } = useClinicAppointments(
    clinicId,
    {
      date: dateFilter || undefined,
      doctor_id: doctorIdFilter || undefined,
      status: statusFilter ? (statusFilter as AppointmentStatus) : undefined,
      page,
      limit,
    },
    Boolean(clinicId)
  );

  const appointments = appointmentsData?.appointments || [];
  const total = appointmentsData?.total || 0;
  const totalPages = Math.ceil(total / limit) || 1;

  // Mutations
  const cancelMutation = useCancelClinicAppointment(clinicId);
  const completeMutation = useCompleteClinicAppointment(clinicId);
  const noShowMutation = useNoShowClinicAppointment(clinicId);

  const handleResetFilters = () => {
    setDateFilter('');
    setDoctorIdFilter('');
    setStatusFilter('');
    setPage(1);
  };

  const handleConfirmCancel = async (reason?: string) => {
    if (!selectedAppointment) return;
    try {
      await cancelMutation.mutateAsync({
        appointmentId: selectedAppointment.id,
        reason: reason || 'Klinika tomonidan bekor qilindi',
      });
      toast.success('Qabul bekor qilindi');
      setSelectedAppointment(null);
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
    if (!selectedAppointment) return;
    try {
      await completeMutation.mutateAsync(selectedAppointment.id);
      toast.success('Qabul muvaffaqiyatli yakunlandi deb belgilandi');
      setSelectedAppointment(null);
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
    if (!selectedAppointment) return;
    try {
      await noShowMutation.mutateAsync(selectedAppointment.id);
      toast.success('Bemor kelmadi deb belgilandi');
      setSelectedAppointment(null);
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
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-200">
        <div>
          <h1 className="text-2xl font-extrabold text-brand-navy flex items-center gap-3">
            <Calendar className="w-7 h-7 text-brand-blue" />
            <span>Qabul bronlari</span>
          </h1>
          <p className="text-xs text-slate-500 mt-1">
            Bemorlarning avtomatik tasdiqlangan qabullari, jadval nazorati va monitoringi
          </p>
        </div>

        <div className="flex items-center gap-2">
          <span className="text-xs font-semibold px-3 py-1.5 rounded-xl bg-blue-50 text-brand-blue border border-blue-200">
            Jami: {total} ta bron
          </span>
        </div>
      </div>

      {/* Filters Toolbar */}
      <div className="bg-white rounded-2xl border border-slate-200/80 p-4 space-y-3 shadow-2xs">
        <div className="flex flex-wrap items-center justify-between gap-3">
          {/* Quick Date Pills */}
          <div className="flex items-center gap-1.5">
            <span className="text-xs font-bold text-slate-400 mr-1">Sana:</span>
            <button
              type="button"
              onClick={() => {
                setDateFilter('');
                setPage(1);
              }}
              className={`px-3 py-1 rounded-lg text-xs font-bold transition-all border ${
                dateFilter === ''
                  ? 'bg-brand-blue text-white border-brand-blue'
                  : 'bg-slate-50 border-slate-200 text-slate-600 hover:bg-slate-100'
              }`}
            >
              Barchasi
            </button>
            <button
              type="button"
              onClick={() => {
                setDateFilter(todayStr);
                setPage(1);
              }}
              className={`px-3 py-1 rounded-lg text-xs font-bold transition-all border ${
                dateFilter === todayStr
                  ? 'bg-brand-blue text-white border-brand-blue'
                  : 'bg-slate-50 border-slate-200 text-slate-600 hover:bg-slate-100'
              }`}
            >
              Bugun
            </button>
            <button
              type="button"
              onClick={() => {
                setDateFilter(tomorrowStr);
                setPage(1);
              }}
              className={`px-3 py-1 rounded-lg text-xs font-bold transition-all border ${
                dateFilter === tomorrowStr
                  ? 'bg-brand-blue text-white border-brand-blue'
                  : 'bg-slate-50 border-slate-200 text-slate-600 hover:bg-slate-100'
              }`}
            >
              Ertaga
            </button>
          </div>

          {(dateFilter || doctorIdFilter || statusFilter) && (
            <button
              type="button"
              onClick={handleResetFilters}
              className="text-xs font-bold text-rose-600 hover:text-rose-700 flex items-center gap-1"
            >
              <X className="w-3.5 h-3.5" />
              <span>Filtrlarni tozalash</span>
            </button>
          )}
        </div>

        {/* Detailed Filter Inputs */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2 border-t border-slate-100">
          {/* Date Input */}
          <div>
            <label className="block text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-1">
              Aniq sana
            </label>
            <input
              type="date"
              value={dateFilter}
              onChange={(e) => {
                setDateFilter(e.target.value);
                setPage(1);
              }}
              className="w-full text-xs font-medium rounded-xl border border-slate-200 px-3 py-2 focus:outline-hidden focus:border-brand-blue"
            />
          </div>

          {/* Doctor Dropdown */}
          <div>
            <label className="block text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-1">
              Shifokor
            </label>
            <select
              value={doctorIdFilter}
              onChange={(e) => {
                setDoctorIdFilter(e.target.value);
                setPage(1);
              }}
              className="w-full text-xs font-medium rounded-xl border border-slate-200 px-3 py-2 focus:outline-hidden focus:border-brand-blue bg-white"
            >
              <option value="">Barcha shifokorlar</option>
              {clinicDoctors.map((doc: any) => (
                <option key={doc.id || doc.doctor_id} value={doc.id || doc.doctor_id}>
                  {doc.first_name} {doc.last_name}
                </option>
              ))}
            </select>
          </div>

          {/* Status Dropdown */}
          <div>
            <label className="block text-[11px] font-bold text-slate-500 uppercase tracking-wider mb-1">
              Holat
            </label>
            <select
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value);
                setPage(1);
              }}
              className="w-full text-xs font-medium rounded-xl border border-slate-200 px-3 py-2 focus:outline-hidden focus:border-brand-blue bg-white"
            >
              <option value="">Barcha holatlar</option>
              <option value="confirmed">Tasdiqlangan</option>
              <option value="completed">Yakunlangan</option>
              <option value="cancelled">Bekor qilingan</option>
              <option value="no_show">Kelmadi</option>
            </select>
          </div>
        </div>
      </div>

      {/* Appointments Table */}
      <div className="bg-white rounded-3xl border border-slate-200/80 shadow-subtle overflow-hidden">
        {isLoadingAppointments || isLoadingClinic ? (
          <div className="p-8 space-y-4">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-14 bg-slate-100 rounded-xl animate-pulse" />
            ))}
          </div>
        ) : isError ? (
          <div className="p-12 text-center space-y-3">
            <AlertCircle className="w-8 h-8 text-rose-500 mx-auto" />
            <h3 className="font-bold text-slate-800 text-sm">
              Bronlarni yuklashda xatolik yuz berdi
            </h3>
            <Button variant="secondary" size="sm" onClick={() => refetch()}>
              Qayta urinish
            </Button>
          </div>
        ) : appointments.length === 0 ? (
          <div className="p-12 text-center space-y-3">
            <CalendarCheck className="w-10 h-10 text-slate-300 mx-auto" />
            <h3 className="font-bold text-slate-800 text-base">Bronlar topilmadi</h3>
            <p className="text-xs text-slate-500 max-w-sm mx-auto">
              Tanlangan filtrlar bo‘yicha hozircha hech qanday qabul qayd etilmagan.
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs text-slate-600 border-collapse">
              <thead>
                <tr className="bg-slate-50/80 border-b border-slate-200 text-slate-400 uppercase font-bold text-[10px] tracking-wider">
                  <th className="py-3.5 px-4">Bemor</th>
                  <th className="py-3.5 px-4">Shifokor</th>
                  <th className="py-3.5 px-4">Sana va Vaqt</th>
                  <th className="py-3.5 px-4">Narx</th>
                  <th className="py-3.5 px-4">Holat</th>
                  <th className="py-3.5 px-4 text-right">Amallar</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {appointments.map((appt) => {
                  const patientName = appt.patient
                    ? `${appt.patient.first_name} ${appt.patient.last_name}`
                    : 'Bemor';
                  const doctorName = appt.doctor
                    ? `${appt.doctor.first_name} ${appt.doctor.last_name}`
                    : 'Shifokor';

                  return (
                    <tr key={appt.id} className="hover:bg-blue-50/30 transition-colors">
                      {/* Patient */}
                      <td className="py-3 px-4 font-semibold text-slate-900">
                        <div className="flex items-center gap-2">
                          <div className="w-7 h-7 rounded-full bg-blue-100 text-brand-blue flex items-center justify-center font-bold text-xs shrink-0">
                            {patientName[0]}
                          </div>
                          <div>
                            <div>{patientName}</div>
                            {appt.patient?.phone && (
                              <div className="text-[11px] text-slate-400 font-mono flex items-center gap-1">
                                <Phone className="w-2.5 h-2.5" />
                                {appt.patient.phone}
                              </div>
                            )}
                          </div>
                        </div>
                      </td>

                      {/* Doctor */}
                      <td className="py-3 px-4">
                        <div className="font-bold text-slate-900">{doctorName}</div>
                        {appt.doctor?.specialty && (
                          <div className="text-[11px] text-brand-blue">
                            {appt.doctor.specialty}
                          </div>
                        )}
                      </td>

                      {/* Date & Time */}
                      <td className="py-3 px-4">
                        <div className="font-bold text-slate-900 flex items-center gap-1.5">
                          <Calendar className="w-3.5 h-3.5 text-slate-400" />
                          <span>{appt.appointment_date}</span>
                        </div>
                        <div className="font-mono text-xs font-bold text-slate-700 flex items-center gap-1.5 mt-0.5">
                          <Clock className="w-3.5 h-3.5 text-brand-blue" />
                          <span>
                            {appt.start_time} – {appt.end_time}
                          </span>
                        </div>
                      </td>

                      {/* Price */}
                      <td className="py-3 px-4 font-bold text-slate-900">
                        {appt.consultation_price > 0
                          ? formatCurrency(appt.consultation_price)
                          : 'Bepul'}
                      </td>

                      {/* Status */}
                      <td className="py-3 px-4">
                        <AppointmentStatusBadge status={appt.status} size="sm" />
                      </td>

                      {/* Actions */}
                      <td className="py-3 px-4 text-right">
                        <div className="flex items-center justify-end gap-1.5">
                          <Link href={`/clinic/appointments/${appt.id}`}>
                            <button
                              type="button"
                              className="p-1.5 rounded-lg border border-slate-200 text-slate-600 hover:text-brand-blue hover:border-blue-200 transition-colors"
                              title="Batafsil ko‘rish"
                            >
                              <Eye className="w-3.5 h-3.5" />
                            </button>
                          </Link>

                          {appt.status === 'confirmed' && (
                            <>
                              <button
                                type="button"
                                onClick={() => {
                                  setSelectedAppointment(appt);
                                  setActionType('complete');
                                }}
                                className="p-1.5 rounded-lg border border-emerald-200 bg-emerald-50 text-emerald-700 hover:bg-emerald-100 transition-colors"
                                title="Yakunlangan deb belgilash"
                              >
                                <Check className="w-3.5 h-3.5" />
                              </button>

                              <button
                                type="button"
                                onClick={() => {
                                  setSelectedAppointment(appt);
                                  setActionType('no_show');
                                }}
                                className="p-1.5 rounded-lg border border-amber-200 bg-amber-50 text-amber-700 hover:bg-amber-100 transition-colors"
                                title="Kelmadi deb belgilash"
                              >
                                <UserX className="w-3.5 h-3.5" />
                              </button>

                              <button
                                type="button"
                                onClick={() => {
                                  setSelectedAppointment(appt);
                                  setActionType('cancel');
                                }}
                                className="p-1.5 rounded-lg border border-rose-200 bg-rose-50 text-rose-700 hover:bg-rose-100 transition-colors"
                                title="Bekor qilish"
                              >
                                <X className="w-3.5 h-3.5" />
                              </button>
                            </>
                          )}
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="p-4 border-t border-slate-100 flex items-center justify-between text-xs text-slate-500">
            <span>
              Jami <strong className="text-slate-800">{total}</strong> ta bron
            </span>
            <div className="flex items-center gap-2">
              <Button
                variant="secondary"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(p - 1, 1))}
              >
                Oldingi
              </Button>
              <span className="font-bold text-slate-700 px-1">
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

      {/* Cancel Modal */}
      <CancelConfirmModal
        isOpen={actionType === 'cancel'}
        onClose={() => {
          setSelectedAppointment(null);
          setActionType(null);
        }}
        onConfirm={handleConfirmCancel}
        isLoading={cancelMutation.isPending}
        title="Bronni bekor qilish"
        description="Ushbu qabulni bekor qilmoqchimisiz? Bemorga bekor qilinganligi haqida xabar ko‘rinadi."
        requireReason={true}
      />

      {/* Complete Modal */}
      <Modal
        isOpen={actionType === 'complete'}
        onClose={() => {
          setSelectedAppointment(null);
          setActionType(null);
        }}
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
              onClick={() => {
                setSelectedAppointment(null);
                setActionType(null);
              }}
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
        onClose={() => {
          setSelectedAppointment(null);
          setActionType(null);
        }}
        title="Bemor kelmadi deb belgilash"
        className="max-w-md"
      >
        <div className="space-y-4 pt-2">
          <p className="text-xs text-slate-600">
            Bemor belgilangan vaqtda qabulga kelmadi deb belgilansinmi?
          </p>
          <div className="flex items-center justify-end gap-2 pt-2">
            <Button
              variant="secondary"
              onClick={() => {
                setSelectedAppointment(null);
                setActionType(null);
              }}
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
