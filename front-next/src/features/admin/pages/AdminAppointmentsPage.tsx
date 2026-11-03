'use client';

import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  CalendarCheck,
  Search,
  Building2,
  Stethoscope,
  User,
  Phone,
  CheckCircle2,
  XCircle,
  Clock,
  Eye,
  ChevronLeft,
  ChevronRight,
  Filter,
  AlertCircle,
  Calendar,
} from 'lucide-react';
import { adminApi } from '../api/adminApi';
import { AdminAppointmentListItem } from '../types';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';
import { formatDate, formatDateTime, formatPrice } from '@/lib/utils/formatters';

const STATUS_OPTIONS = [
  { value: '', label: 'Barchasi' },
  { value: 'confirmed', label: 'Tasdiqlangan' },
  { value: 'completed', label: 'Yakunlangan' },
  { value: 'cancelled', label: 'Bekor qilingan' },
  { value: 'no_show', label: 'Kelmagan' },
];

export function AdminAppointmentsPage() {
  const [statusFilter, setStatusFilter] = useState('');
  const [patientSearch, setPatientSearch] = useState('');
  const [appliedSearch, setAppliedSearch] = useState('');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 15;

  const [selectedAppointment, setSelectedAppointment] = useState<AdminAppointmentListItem | null>(null);

  // Appointments query
  const { data: appointmentsData, isLoading } = useQuery({
    queryKey: ['adminAppointments', statusFilter, appliedSearch, startDate, endDate, currentPage],
    queryFn: async () => {
      const res = await adminApi.listAppointments({
        status: statusFilter || undefined,
        patient_search: appliedSearch || undefined,
        start_date: startDate || undefined,
        end_date: endDate || undefined,
        page: currentPage,
        limit: pageSize,
      });
      return res.data;
    },
  });

  const appointments = appointmentsData?.data || [];
  const meta = appointmentsData?.meta || { total: 0, page: 1, limit: pageSize };
  const totalPages = Math.ceil(meta.total / pageSize) || 1;

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setAppliedSearch(patientSearch.trim());
    setCurrentPage(1);
  };

  const resetFilters = () => {
    setStatusFilter('');
    setPatientSearch('');
    setAppliedSearch('');
    setStartDate('');
    setEndDate('');
    setCurrentPage(1);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'confirmed':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
            <CheckCircle2 className="w-3.5 h-3.5" /> Tasdiqlangan
          </span>
        );
      case 'completed':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-50 text-blue-700 border border-blue-200">
            <CheckCircle2 className="w-3.5 h-3.5" /> Yakunlangan
          </span>
        );
      case 'cancelled':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-rose-50 text-rose-700 border border-rose-200">
            <XCircle className="w-3.5 h-3.5" /> Bekor qilingan
          </span>
        );
      case 'no_show':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-700 border border-slate-300">
            Kelmagan
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200">
            <Clock className="w-3.5 h-3.5" /> {status}
          </span>
        );
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl sm:text-2xl font-black text-slate-900 tracking-tight flex items-center gap-2.5">
            <CalendarCheck className="w-6 h-6 text-[#0866F5]" />
            Barcha bronlar nazorati
          </h1>
          <p className="text-xs sm:text-sm text-slate-500 mt-0.5">
            Platforma miqyosidagi barcha qabullar monitoringi va filtrlari (Read-only)
          </p>
        </div>

        <div className="flex items-center gap-2">
          <div className="px-3.5 py-1.5 rounded-xl bg-blue-50 text-blue-700 font-semibold text-xs border border-blue-100 flex items-center gap-1.5">
            <Filter className="w-4 h-4 text-[#0866F5]" />
            Jami: <span className="font-bold">{meta.total}</span> ta bron
          </div>
        </div>
      </div>

      {/* Filter Toolbar */}
      <div className="bg-white rounded-2xl border border-slate-200/80 p-4 shadow-2xs space-y-3">
        {/* Status quick tabs */}
        <div className="flex flex-wrap items-center gap-1.5 pb-2 border-b border-slate-100">
          <span className="text-xs font-bold text-slate-500 mr-2">Status:</span>
          {STATUS_OPTIONS.map((st) => (
            <button
              key={st.value}
              onClick={() => {
                setStatusFilter(st.value);
                setCurrentPage(1);
              }}
              className={`px-3 py-1 rounded-xl text-xs font-semibold transition-all ${
                statusFilter === st.value
                  ? 'bg-[#0866F5] text-white shadow-2xs'
                  : 'bg-slate-100 text-slate-600 hover:bg-slate-200/70'
              }`}
            >
              {st.label}
            </button>
          ))}
        </div>

        {/* Inputs row */}
        <form onSubmit={handleSearchSubmit} className="grid grid-cols-1 sm:grid-cols-12 gap-2.5">
          <div className="sm:col-span-5 relative">
            <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={patientSearch}
              onChange={(e) => setPatientSearch(e.target.value)}
              placeholder="Bemor ismi yoki telefon raqami..."
              className="w-full pl-9 pr-4 py-2 text-xs sm:text-sm rounded-xl border border-slate-200 bg-slate-50/50 focus:bg-white focus:outline-hidden focus:ring-2 focus:ring-[#0866F5]/20 focus:border-[#0866F5] transition-all"
            />
          </div>

          <div className="sm:col-span-3 flex items-center gap-1.5">
            <input
              type="date"
              value={startDate}
              onChange={(e) => {
                setStartDate(e.target.value);
                setCurrentPage(1);
              }}
              className="w-full px-3 py-2 text-xs rounded-xl border border-slate-200 bg-slate-50/50 focus:bg-white focus:outline-hidden focus:ring-2 focus:ring-[#0866F5]/20 text-slate-700"
              title="Boshlanish sanasi"
            />
            <span className="text-slate-400 text-xs">-</span>
            <input
              type="date"
              value={endDate}
              onChange={(e) => {
                setEndDate(e.target.value);
                setCurrentPage(1);
              }}
              className="w-full px-3 py-2 text-xs rounded-xl border border-slate-200 bg-slate-50/50 focus:bg-white focus:outline-hidden focus:ring-2 focus:ring-[#0866F5]/20 text-slate-700"
              title="Tugash sanasi"
            />
          </div>

          <div className="sm:col-span-4 flex items-center gap-2 justify-end">
            <Button type="submit" size="sm" className="bg-[#0866F5] text-white px-4 text-xs">
              Qidirish
            </Button>
            {(statusFilter || appliedSearch || startDate || endDate) && (
              <Button type="button" variant="outline" size="sm" onClick={resetFilters} className="text-xs">
                Filtrni tozalash
              </Button>
            )}
          </div>
        </form>
      </div>

      {/* Appointments Table */}
      <div className="bg-white rounded-2xl border border-slate-200/80 shadow-2xs overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-slate-500 text-xs sm:text-sm">
            <div className="w-8 h-8 border-2 border-[#0866F5] border-t-transparent rounded-full animate-spin mx-auto mb-3" />
            Bronlar yuklanmoqda...
          </div>
        ) : appointments.length === 0 ? (
          <div className="p-12 text-center text-slate-500">
            <CalendarCheck className="w-12 h-12 text-slate-300 mx-auto mb-2" />
            <p className="text-sm font-semibold text-slate-700">Bronlar topilmadi</p>
            <p className="text-xs text-slate-400 mt-0.5">
              Tanlangan filtrlar bo‘yicha hech qanday bron qayd etilmagan
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs sm:text-sm text-slate-600">
              <thead className="bg-slate-50/80 text-slate-700 uppercase font-bold text-[11px] border-b border-slate-200">
                <tr>
                  <th className="px-5 py-3.5">Qabul vaqti</th>
                  <th className="px-5 py-3.5">Bemor</th>
                  <th className="px-5 py-3.5">Klinika</th>
                  <th className="px-5 py-3.5">Shifokor</th>
                  <th className="px-5 py-3.5">Narxi</th>
                  <th className="px-5 py-3.5">Status</th>
                  <th className="px-5 py-3.5 text-right">Amal</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {appointments.map((app) => (
                  <tr key={app.id} className="hover:bg-slate-50/60 transition-colors">
                    <td className="px-5 py-3.5 whitespace-nowrap">
                      <div className="font-bold text-slate-900 text-xs">
                        {formatDate(app.appointment_date)}
                      </div>
                      <div className="text-[11px] text-[#0866F5] font-semibold">
                        {app.start_time} - {app.end_time}
                      </div>
                    </td>

                    <td className="px-5 py-3.5">
                      <div className="font-bold text-slate-900 text-xs">
                        {app.patient_name || 'Bemor ismi yo‘q'}
                      </div>
                      <div className="text-[11px] text-slate-400 flex items-center gap-1 mt-0.5">
                        <Phone className="w-3 h-3 text-slate-400" />
                        {app.patient_phone || '—'}
                      </div>
                    </td>

                    <td className="px-5 py-3.5">
                      <div className="font-semibold text-slate-800 text-xs flex items-center gap-1">
                        <Building2 className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                        {app.clinic_name}
                      </div>
                    </td>

                    <td className="px-5 py-3.5">
                      <div className="font-bold text-slate-900 text-xs">
                        {app.doctor_name}
                      </div>
                      <div className="text-[11px] text-slate-500">
                        {app.specialty_name || 'Mutaxassis'}
                      </div>
                    </td>

                    <td className="px-5 py-3.5 whitespace-nowrap font-bold text-slate-900 text-xs">
                      {formatPrice(app.consultation_price)}
                    </td>

                    <td className="px-5 py-3.5 whitespace-nowrap">
                      {getStatusBadge(app.status)}
                    </td>

                    <td className="px-5 py-3.5 text-right whitespace-nowrap">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => setSelectedAppointment(app)}
                        className="text-[#0866F5] hover:bg-blue-50 text-xs gap-1.5 font-semibold"
                      >
                        <Eye className="w-3.5 h-3.5" />
                        <span>Tafsilot</span>
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="px-5 py-3.5 bg-slate-50/60 border-t border-slate-200/80 flex items-center justify-between text-xs text-slate-600">
            <div>
              Ko‘rsatilmoqda: <span className="font-bold">{(currentPage - 1) * pageSize + 1}</span> -{' '}
              <span className="font-bold">{Math.min(currentPage * pageSize, meta.total)}</span> / {meta.total} ta
            </div>
            <div className="flex items-center gap-1.5">
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage <= 1}
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                className="h-8 px-2 text-xs"
              >
                <ChevronLeft className="w-4 h-4" />
                <span className="hidden sm:inline">Oldingi</span>
              </Button>
              <span className="px-2 font-semibold text-slate-700">
                {currentPage} / {totalPages}
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage >= totalPages}
                onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                className="h-8 px-2 text-xs"
              >
                <span className="hidden sm:inline">Keyingi</span>
                <ChevronRight className="w-4 h-4" />
              </Button>
            </div>
          </div>
        )}
      </div>

      {/* Appointment Detail Modal */}
      {selectedAppointment && (
        <Modal
          isOpen={Boolean(selectedAppointment)}
          onClose={() => setSelectedAppointment(null)}
          title="Bron tafsilotlari (Read-only)"
        >
          <div className="space-y-4 text-xs text-slate-600">
            {/* Status card */}
            <div className="flex items-center justify-between p-3.5 rounded-2xl bg-slate-50 border border-slate-200">
              <div>
                <span className="text-[11px] text-slate-400 block font-medium">Bron ID:</span>
                <span className="font-mono font-bold text-slate-800 text-xs">
                  {selectedAppointment.id}
                </span>
              </div>
              <div>{getStatusBadge(selectedAppointment.status)}</div>
            </div>

            {/* Grid specs */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div className="p-3 rounded-xl bg-slate-50 border border-slate-100">
                <span className="text-[11px] text-slate-400 block font-medium">Bemor:</span>
                <span className="font-bold text-slate-900 text-sm block mt-0.5">
                  {selectedAppointment.patient_name}
                </span>
                <span className="text-slate-500 text-xs flex items-center gap-1 mt-1">
                  <Phone className="w-3.5 h-3.5 text-slate-400" />
                  {selectedAppointment.patient_phone || 'Telefon yo‘q'}
                </span>
              </div>

              <div className="p-3 rounded-xl bg-slate-50 border border-slate-100">
                <span className="text-[11px] text-slate-400 block font-medium">Klinika:</span>
                <span className="font-bold text-slate-900 text-sm block mt-0.5">
                  {selectedAppointment.clinic_name}
                </span>
              </div>

              <div className="p-3 rounded-xl bg-slate-50 border border-slate-100">
                <span className="text-[11px] text-slate-400 block font-medium">Shifokor:</span>
                <span className="font-bold text-slate-900 text-sm block mt-0.5">
                  {selectedAppointment.doctor_name}
                </span>
                <span className="text-slate-500 text-xs block mt-0.5">
                  Ixtisoslik: {selectedAppointment.specialty_name || 'Umumiy'}
                </span>
              </div>

              <div className="p-3 rounded-xl bg-slate-50 border border-slate-100">
                <span className="text-[11px] text-slate-400 block font-medium">Qabul vaqti va narxi:</span>
                <span className="font-bold text-slate-900 text-sm block mt-0.5">
                  {formatDate(selectedAppointment.appointment_date)} ({selectedAppointment.start_time} - {selectedAppointment.end_time})
                </span>
                <span className="font-bold text-[#0866F5] text-xs block mt-1">
                  {formatPrice(selectedAppointment.consultation_price)}
                </span>
              </div>
            </div>

            {/* Cancellation reason */}
            {selectedAppointment.cancellation_reason && (
              <div className="p-3 rounded-xl bg-rose-50 border border-rose-200 text-rose-800">
                <div className="flex items-center gap-1.5 font-bold text-xs mb-1">
                  <AlertCircle className="w-4 h-4 text-rose-600" />
                  Bekor qilinish sababi:
                </div>
                <p className="text-xs">{selectedAppointment.cancellation_reason}</p>
              </div>
            )}

            <div className="pt-2 text-[11px] text-slate-400 flex items-center justify-between">
              <span>Yaratilgan sana: {formatDateTime(selectedAppointment.created_at)}</span>
              <span className="font-medium text-slate-500">Superadmin nazorati</span>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}

