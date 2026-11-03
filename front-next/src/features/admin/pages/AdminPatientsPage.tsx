'use client';

import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Users,
  Search,
  Calendar,
  Phone,
  Mail,
  Eye,
  CalendarCheck,
  CheckCircle2,
  XCircle,
  Clock,
  UserCheck,
  ChevronLeft,
  ChevronRight,
  Shield,
} from 'lucide-react';
import { adminApi } from '../api/adminApi';
import { AdminPatientListItem, AdminPatientDetail } from '../types';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';
import { formatDate, formatDateTime, formatPrice } from '@/lib/utils/formatters';

export function AdminPatientsPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [appliedSearch, setAppliedSearch] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 15;

  const [selectedPatientId, setSelectedPatientId] = useState<string | null>(null);

  // List patients
  const { data: patientsData, isLoading } = useQuery({
    queryKey: ['adminPatients', appliedSearch, currentPage],
    queryFn: async () => {
      const res = await adminApi.listPatients({
        search: appliedSearch || undefined,
        page: currentPage,
        limit: pageSize,
      });
      return res.data;
    },
  });

  // Patient detail
  const { data: detailData, isLoading: isLoadingDetail } = useQuery({
    queryKey: ['adminPatientDetail', selectedPatientId],
    queryFn: async () => {
      if (!selectedPatientId) return null;
      const res = await adminApi.getPatientDetails(selectedPatientId);
      return res.data.data;
    },
    enabled: Boolean(selectedPatientId),
  });

  const patients = patientsData?.data || [];
  const meta = patientsData?.meta || { total: 0, page: 1, limit: pageSize };
  const totalPages = Math.ceil(meta.total / pageSize) || 1;

  const handleSearchSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setAppliedSearch(searchTerm.trim());
    setCurrentPage(1);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'confirmed':
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
            <CheckCircle2 className="w-3 h-3" /> Tasdiqlangan
          </span>
        );
      case 'completed':
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-blue-50 text-blue-700 border border-blue-200">
            <CheckCircle2 className="w-3 h-3" /> Yakunlangan
          </span>
        );
      case 'cancelled':
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-rose-50 text-rose-700 border border-rose-200">
            <XCircle className="w-3 h-3" /> Bekor qilingan
          </span>
        );
      case 'no_show':
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-slate-100 text-slate-700 border border-slate-300">
            Kelmagan
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-amber-50 text-amber-700 border border-amber-200">
            <Clock className="w-3 h-3" /> {status}
          </span>
        );
    }
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-xl sm:text-2xl font-black text-slate-900 tracking-tight flex items-center gap-2.5">
            <Users className="w-6 h-6 text-[#0866F5]" />
            Bemorlar boshqaruvi
          </h1>
          <p className="text-xs sm:text-sm text-slate-500 mt-0.5">
            Platformada ro‘yxatdan o‘tgan barcha bemorlar va ularning faolligi
          </p>
        </div>

        <div className="flex items-center gap-2">
          <div className="px-3.5 py-1.5 rounded-xl bg-blue-50 text-blue-700 font-semibold text-xs border border-blue-100 flex items-center gap-1.5">
            <UserCheck className="w-4 h-4 text-[#0866F5]" />
            Jami: <span className="font-bold">{meta.total}</span> nafar bemor
          </div>
        </div>
      </div>

      {/* Filter & Search Bar */}
      <div className="bg-white rounded-2xl border border-slate-200/80 p-3.5 shadow-2xs">
        <form onSubmit={handleSearchSubmit} className="flex gap-2">
          <div className="relative flex-1">
            <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Ism, familiya, telefon yoki email bo‘yicha qidirish..."
              className="w-full pl-9 pr-4 py-2 text-xs sm:text-sm rounded-xl border border-slate-200 bg-slate-50/50 focus:bg-white focus:outline-hidden focus:ring-2 focus:ring-[#0866F5]/20 focus:border-[#0866F5] transition-all"
            />
          </div>
          <Button type="submit" size="sm" className="bg-[#0866F5] text-white shrink-0 px-4">
            Qidirish
          </Button>
          {appliedSearch && (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => {
                setSearchTerm('');
                setAppliedSearch('');
                setCurrentPage(1);
              }}
              className="shrink-0"
            >
              Tozalash
            </Button>
          )}
        </form>
      </div>

      {/* Patients Table */}
      <div className="bg-white rounded-2xl border border-slate-200/80 shadow-2xs overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-slate-500 text-xs sm:text-sm">
            <div className="w-8 h-8 border-2 border-[#0866F5] border-t-transparent rounded-full animate-spin mx-auto mb-3" />
            Bemorlar ro‘yxati yuklanmoqda...
          </div>
        ) : patients.length === 0 ? (
          <div className="p-12 text-center text-slate-500">
            <Users className="w-12 h-12 text-slate-300 mx-auto mb-2" />
            <p className="text-sm font-semibold text-slate-700">Bemorlar topilmadi</p>
            <p className="text-xs text-slate-400 mt-0.5">
              {appliedSearch
                ? `"${appliedSearch}" so‘rovi bo‘yicha hech qanday natija yo‘q`
                : 'Platformada hozircha bemorlar mavjud emas'}
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs sm:text-sm text-slate-600">
              <thead className="bg-slate-50/80 text-slate-700 uppercase font-bold text-[11px] border-b border-slate-200">
                <tr>
                  <th className="px-5 py-3.5">Bemor</th>
                  <th className="px-5 py-3.5">Aloqa</th>
                  <th className="px-5 py-3.5">Ro‘yxatdan o‘tgan</th>
                  <th className="px-5 py-3.5 text-center">Jami bronlar</th>
                  <th className="px-5 py-3.5">Oxirgi bron</th>
                  <th className="px-5 py-3.5 text-right">Amal</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {patients.map((p) => {
                  const initial = (p.first_name?.[0] || 'B').toUpperCase();
                  return (
                    <tr key={p.id} className="hover:bg-slate-50/60 transition-colors">
                      <td className="px-5 py-3.5">
                        <div className="flex items-center gap-3">
                          <div className="w-9 h-9 rounded-xl bg-blue-100 text-[#0866F5] font-black text-xs flex items-center justify-center shrink-0">
                            {initial}
                          </div>
                          <div>
                            <div className="font-bold text-slate-900">
                              {p.first_name} {p.last_name}
                            </div>
                            <div className="text-[11px] text-slate-400">ID: {p.id.slice(0, 8)}...</div>
                          </div>
                        </div>
                      </td>

                      <td className="px-5 py-3.5">
                        <div className="flex flex-col gap-0.5">
                          <div className="flex items-center gap-1.5 text-slate-800 font-medium text-xs">
                            <Phone className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                            {p.phone || '—'}
                          </div>
                          <div className="flex items-center gap-1.5 text-slate-500 text-xs">
                            <Mail className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                            {p.email || '—'}
                          </div>
                        </div>
                      </td>

                      <td className="px-5 py-3.5 whitespace-nowrap text-slate-500 text-xs">
                        {formatDate(p.created_at)}
                      </td>

                      <td className="px-5 py-3.5 text-center">
                        <span
                          className={`inline-flex items-center justify-center px-2.5 py-0.5 rounded-full text-xs font-bold ${
                            p.appointments_count > 0
                              ? 'bg-blue-50 text-blue-700 border border-blue-200'
                              : 'bg-slate-100 text-slate-600'
                          }`}
                        >
                          {p.appointments_count} ta
                        </span>
                      </td>

                      <td className="px-5 py-3.5 whitespace-nowrap text-slate-500 text-xs">
                        {p.last_appointment_date ? formatDate(p.last_appointment_date) : '—'}
                      </td>

                      <td className="px-5 py-3.5 text-right whitespace-nowrap">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => setSelectedPatientId(p.id)}
                          className="text-[#0866F5] hover:bg-blue-50 text-xs gap-1.5 font-semibold"
                        >
                          <Eye className="w-3.5 h-3.5" />
                          <span>Tafsilot</span>
                        </Button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}

        {/* Pagination Bar */}
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

      {/* Patient Detail Modal */}
      {selectedPatientId && (
        <Modal
          isOpen={Boolean(selectedPatientId)}
          onClose={() => setSelectedPatientId(null)}
          title="Bemor profili va bronlar tarixi"
        >
          {isLoadingDetail || !detailData ? (
            <div className="p-8 text-center text-slate-500 text-xs">
              <div className="w-6 h-6 border-2 border-[#0866F5] border-t-transparent rounded-full animate-spin mx-auto mb-2" />
              Ma’lumotlar yuklanmoqda...
            </div>
          ) : (
            <div className="space-y-5 text-xs text-slate-600">
              {/* Header profile card */}
              <div className="flex items-center gap-3.5 p-3.5 rounded-2xl bg-blue-50/60 border border-blue-100">
                <div className="w-12 h-12 rounded-2xl bg-[#0866F5] text-white font-black text-lg flex items-center justify-center shrink-0">
                  {(detailData.first_name?.[0] || 'B').toUpperCase()}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="text-base font-bold text-slate-900 truncate">
                    {detailData.first_name} {detailData.last_name}
                  </div>
                  <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-slate-500 text-xs mt-0.5">
                    <span className="flex items-center gap-1">
                      <Phone className="w-3 h-3 text-slate-400" />
                      {detailData.phone || 'Telefon kiritilmagan'}
                    </span>
                    <span className="flex items-center gap-1">
                      <Mail className="w-3 h-3 text-slate-400" />
                      {detailData.email || 'Email yo‘q'}
                    </span>
                  </div>
                </div>
              </div>

              {/* Stats KPI */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
                <div className="p-3 rounded-xl bg-slate-50 border border-slate-200 text-center">
                  <div className="text-slate-400 text-[11px] font-medium">Jami bronlar</div>
                  <div className="text-lg font-black text-slate-900 mt-0.5">
                    {detailData.total_appointments}
                  </div>
                </div>
                <div className="p-3 rounded-xl bg-amber-50/60 border border-amber-200 text-center">
                  <div className="text-amber-700 text-[11px] font-medium">Kelgusi</div>
                  <div className="text-lg font-black text-amber-900 mt-0.5">
                    {detailData.upcoming_appointments}
                  </div>
                </div>
                <div className="p-3 rounded-xl bg-emerald-50/60 border border-emerald-200 text-center">
                  <div className="text-emerald-700 text-[11px] font-medium">Yakunlangan</div>
                  <div className="text-lg font-black text-emerald-900 mt-0.5">
                    {detailData.completed_appointments}
                  </div>
                </div>
                <div className="p-3 rounded-xl bg-rose-50/60 border border-rose-200 text-center">
                  <div className="text-rose-700 text-[11px] font-medium">Bekor qilingan</div>
                  <div className="text-lg font-black text-rose-900 mt-0.5">
                    {detailData.cancelled_appointments}
                  </div>
                </div>
              </div>

              {/* Recent appointments */}
              <div>
                <div className="flex items-center justify-between mb-2.5">
                  <h4 className="font-bold text-slate-800 text-xs flex items-center gap-1.5">
                    <CalendarCheck className="w-4 h-4 text-[#0866F5]" />
                    So‘nggi bronlar ({detailData.recent_appointments?.length || 0})
                  </h4>
                </div>

                {!detailData.recent_appointments || detailData.recent_appointments.length === 0 ? (
                  <div className="p-6 text-center text-slate-400 bg-slate-50 rounded-xl border border-slate-100">
                    Bemor hali bron qilmagan
                  </div>
                ) : (
                  <div className="space-y-2 max-h-60 overflow-y-auto pr-1">
                    {detailData.recent_appointments.map((app) => (
                      <div
                        key={app.id}
                        className="p-3 rounded-xl bg-slate-50 border border-slate-200/80 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2"
                      >
                        <div>
                          <div className="font-bold text-slate-900 text-xs">
                            {app.clinic_name} —{' '}
                            <span className="text-[#0866F5]">{app.doctor_name}</span>
                          </div>
                          <div className="text-[11px] text-slate-500 mt-0.5 flex items-center gap-2">
                            <span>{app.specialty_name || 'Shifokor'}</span>
                            <span>•</span>
                            <span>
                              {formatDate(app.appointment_date)} {app.start_time} - {app.end_time}
                            </span>
                          </div>
                        </div>

                        <div className="flex items-center justify-between sm:justify-end gap-3 shrink-0">
                          <span className="font-bold text-slate-900 text-xs">
                            {formatPrice(app.consultation_price)}
                          </span>
                          {getStatusBadge(app.status)}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Data protection note */}
              <div className="p-2.5 rounded-xl bg-slate-100 text-slate-500 text-[11px] flex items-center gap-2">
                <Shield className="w-4 h-4 text-slate-400 shrink-0" />
                <span>
                  Tibbiy ma’lumotlar maxfiyligi himoyalangan: bu yerda faqat bron qilish vaqti va holati ko‘rsatiladi.
                </span>
              </div>
            </div>
          )}
        </Modal>
      )}
    </div>
  );
}

