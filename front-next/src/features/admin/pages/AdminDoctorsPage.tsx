'use client';

import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Stethoscope,
  Search,
  Building2,
  CalendarCheck,
  CheckCircle2,
  XCircle,
  Clock,
  Eye,
  ChevronLeft,
  ChevronRight,
  Award,
  DollarSign,
} from 'lucide-react';
import { adminApi } from '../api/adminApi';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';
import { formatDate, formatPrice } from '@/lib/utils/formatters';
import { getMediaUrl } from '@/lib/utils/media';

export function AdminDoctorsPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [appliedSearch, setAppliedSearch] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 15;

  const [selectedDoctorId, setSelectedDoctorId] = useState<string | null>(null);

  // Doctors list
  const { data: doctorsData, isLoading } = useQuery({
    queryKey: ['adminDoctors', appliedSearch, currentPage],
    queryFn: async () => {
      const res = await adminApi.listDoctors({
        search: appliedSearch || undefined,
        page: currentPage,
        limit: pageSize,
      });
      return res.data;
    },
  });

  // Doctor detail
  const { data: detailData, isLoading: isLoadingDetail } = useQuery({
    queryKey: ['adminDoctorDetail', selectedDoctorId],
    queryFn: async () => {
      if (!selectedDoctorId) return null;
      const res = await adminApi.getDoctorDetails(selectedDoctorId);
      return res.data.data;
    },
    enabled: Boolean(selectedDoctorId),
  });

  const doctors = doctorsData?.data || [];
  const meta = doctorsData?.meta || { total: 0, page: 1, limit: pageSize };
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
      default:
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-slate-100 text-slate-700">
            <Clock className="w-3 h-3" /> {status}
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
            <Stethoscope className="w-6 h-6 text-[#0866F5]" />
            Shifokorlar boshqaruvi
          </h1>
          <p className="text-xs sm:text-sm text-slate-500 mt-0.5">
            Platformadagi barcha unikal shifokorlar, ularning ixtisosliklari va klinikalari
          </p>
        </div>

        <div className="flex items-center gap-2">
          <div className="px-3.5 py-1.5 rounded-xl bg-blue-50 text-blue-700 font-semibold text-xs border border-blue-100 flex items-center gap-1.5">
            <Award className="w-4 h-4 text-[#0866F5]" />
            Jami unikal shifokorlar: <span className="font-bold">{meta.total}</span> nafar
          </div>
        </div>
      </div>

      {/* Search */}
      <div className="bg-white rounded-2xl border border-slate-200/80 p-3.5 shadow-2xs">
        <form onSubmit={handleSearchSubmit} className="flex gap-2">
          <div className="relative flex-1">
            <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Shifokor ismi, familiyasi yoki ixtisosligi bo‘yicha qidirish..."
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

      {/* Doctors Table */}
      <div className="bg-white rounded-2xl border border-slate-200/80 shadow-2xs overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-slate-500 text-xs sm:text-sm">
            <div className="w-8 h-8 border-2 border-[#0866F5] border-t-transparent rounded-full animate-spin mx-auto mb-3" />
            Shifokorlar ro‘yxati yuklanmoqda...
          </div>
        ) : doctors.length === 0 ? (
          <div className="p-12 text-center text-slate-500">
            <Stethoscope className="w-12 h-12 text-slate-300 mx-auto mb-2" />
            <p className="text-sm font-semibold text-slate-700">Shifokorlar topilmadi</p>
            <p className="text-xs text-slate-400 mt-0.5">
              {appliedSearch
                ? `"${appliedSearch}" so‘rovi bo‘yicha hech narsa topilmadi`
                : 'Platformada shifokorlar mavjud emas'}
            </p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs sm:text-sm text-slate-600">
              <thead className="bg-slate-50/80 text-slate-700 uppercase font-bold text-[11px] border-b border-slate-200">
                <tr>
                  <th className="px-5 py-3.5">Shifokor</th>
                  <th className="px-5 py-3.5">Ixtisoslik(lar)</th>
                  <th className="px-5 py-3.5">Klinikalar</th>
                  <th className="px-5 py-3.5 text-center">Jami bronlar</th>
                  <th className="px-5 py-3.5">Qo‘shilgan</th>
                  <th className="px-5 py-3.5 text-right">Amal</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {doctors.map((d) => {
                  return (
                    <tr key={d.id} className="hover:bg-slate-50/60 transition-colors">
                      <td className="px-5 py-3.5">
                        <div className="flex items-center gap-3">
                          {d.photo_url ? (
                            <img
                              src={getMediaUrl(d.photo_url)}
                              alt={d.first_name}
                              className="w-10 h-10 rounded-xl object-cover border border-slate-200 bg-slate-100 shrink-0"
                            />
                          ) : (
                            <div className="w-10 h-10 rounded-xl bg-blue-100 text-[#0866F5] font-black text-sm flex items-center justify-center shrink-0">
                              {(d.first_name?.[0] || 'D').toUpperCase()}
                            </div>
                          )}
                          <div>
                            <div className="font-bold text-slate-900">
                              {d.first_name} {d.last_name}
                            </div>
                            <div className="text-[11px] text-slate-400 flex items-center gap-1.5 mt-0.5">
                              <span>Tajriba: {d.experience_years} yil</span>
                            </div>
                          </div>
                        </div>
                      </td>

                      <td className="px-5 py-3.5">
                        <div className="flex flex-wrap gap-1 max-w-xs">
                          {d.specialties && d.specialties.length > 0 ? (
                            d.specialties.map((sp, idx) => (
                              <span
                                key={idx}
                                className="px-2 py-0.5 rounded-md text-[11px] font-semibold bg-blue-50 text-blue-700 border border-blue-100"
                              >
                                {sp}
                              </span>
                            ))
                          ) : (
                            <span className="text-slate-400 text-xs">—</span>
                          )}
                        </div>
                      </td>

                      <td className="px-5 py-3.5">
                        <div className="flex flex-wrap gap-1 max-w-xs">
                          {d.clinics && d.clinics.length > 0 ? (
                            d.clinics.map((cl, idx) => (
                              <span
                                key={idx}
                                className="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-[11px] font-medium bg-slate-100 text-slate-700"
                              >
                                <Building2 className="w-3 h-3 text-slate-400" />
                                {cl}
                              </span>
                            ))
                          ) : (
                            <span className="text-slate-400 text-xs">Biriktirilmagan</span>
                          )}
                        </div>
                      </td>

                      <td className="px-5 py-3.5 text-center">
                        <span
                          className={`inline-flex items-center justify-center px-2.5 py-0.5 rounded-full text-xs font-bold ${
                            d.total_appointments > 0
                              ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                              : 'bg-slate-100 text-slate-600'
                          }`}
                        >
                          {d.total_appointments} ta
                        </span>
                      </td>

                      <td className="px-5 py-3.5 whitespace-nowrap text-slate-500 text-xs">
                        {formatDate(d.created_at)}
                      </td>

                      <td className="px-5 py-3.5 text-right whitespace-nowrap">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => setSelectedDoctorId(d.id)}
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

      {/* Doctor Detail Modal */}
      {selectedDoctorId && (
        <Modal
          isOpen={Boolean(selectedDoctorId)}
          onClose={() => setSelectedDoctorId(null)}
          title="Shifokor to‘liq profili va faoliyati"
        >
          {isLoadingDetail || !detailData ? (
            <div className="p-8 text-center text-slate-500 text-xs">
              <div className="w-6 h-6 border-2 border-[#0866F5] border-t-transparent rounded-full animate-spin mx-auto mb-2" />
              Ma’lumotlar yuklanmoqda...
            </div>
          ) : (
            <div className="space-y-5 text-xs text-slate-600">
              {/* Doctor Profile card */}
              <div className="flex items-start gap-4 p-4 rounded-2xl bg-blue-50/60 border border-blue-100">
                {detailData.photo_url ? (
                  <img
                    src={getMediaUrl(detailData.photo_url)}
                    alt={detailData.first_name}
                    className="w-14 h-14 rounded-2xl object-cover border border-slate-200 bg-white shrink-0"
                  />
                ) : (
                  <div className="w-14 h-14 rounded-2xl bg-[#0866F5] text-white font-black text-xl flex items-center justify-center shrink-0">
                    {(detailData.first_name?.[0] || 'D').toUpperCase()}
                  </div>
                )}
                <div className="flex-1 min-w-0">
                  <div className="text-base font-bold text-slate-900">
                    {detailData.first_name} {detailData.last_name}
                  </div>
                  <div className="text-xs text-slate-500 mt-0.5">
                    Ish tajribasi: <span className="font-semibold text-slate-800">{detailData.experience_years} yil</span>
                  </div>
                  {detailData.bio && (
                    <p className="text-xs text-slate-600 mt-2 line-clamp-3 bg-white/70 p-2 rounded-xl border border-blue-100/60">
                      {detailData.bio}
                    </p>
                  )}
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

              {/* Clinic Affiliations & Prices */}
              <div>
                <h4 className="font-bold text-slate-800 text-xs flex items-center gap-1.5 mb-2.5">
                  <Building2 className="w-4 h-4 text-[#0866F5]" />
                  Ishlayotgan klinikalari ({detailData.clinics?.length || 0})
                </h4>
                {!detailData.clinics || detailData.clinics.length === 0 ? (
                  <div className="p-4 text-center text-slate-400 bg-slate-50 rounded-xl border border-slate-100">
                    Klinikalar biriktirilmagan
                  </div>
                ) : (
                  <div className="space-y-2 max-h-48 overflow-y-auto pr-1">
                    {detailData.clinics.map((cl) => (
                      <div
                        key={cl.clinic_doctor_id}
                        className="p-3 rounded-xl bg-slate-50 border border-slate-200 flex items-center justify-between gap-3"
                      >
                        <div>
                          <div className="font-bold text-slate-900 text-xs flex items-center gap-2">
                            {cl.clinic_name}
                            {cl.is_active ? (
                              <span className="px-1.5 py-0.5 rounded-md text-[10px] font-semibold bg-emerald-100 text-emerald-800">
                                Faol
                              </span>
                            ) : (
                              <span className="px-1.5 py-0.5 rounded-md text-[10px] font-semibold bg-slate-200 text-slate-700">
                                Faol emas
                              </span>
                            )}
                          </div>
                          <div className="text-[11px] text-slate-500 mt-0.5">
                            Yo‘nalish: {cl.specialty_name || 'Umumiy'}
                          </div>
                        </div>

                        <div className="text-right">
                          <span className="font-bold text-slate-900 text-xs">
                            {formatPrice(cl.consultation_price)}
                          </span>
                          <div className="text-[10px] text-slate-400">qabul narxi</div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Recent appointments */}
              <div>
                <h4 className="font-bold text-slate-800 text-xs flex items-center gap-1.5 mb-2.5">
                  <CalendarCheck className="w-4 h-4 text-[#0866F5]" />
                  So‘nggi bronlar ({detailData.recent_appointments?.length || 0})
                </h4>

                {!detailData.recent_appointments || detailData.recent_appointments.length === 0 ? (
                  <div className="p-4 text-center text-slate-400 bg-slate-50 rounded-xl border border-slate-100">
                    Bronlar tarixi mavjud emas
                  </div>
                ) : (
                  <div className="space-y-2 max-h-48 overflow-y-auto pr-1">
                    {detailData.recent_appointments.map((app) => (
                      <div
                        key={app.id}
                        className="p-2.5 rounded-xl bg-slate-50 border border-slate-200/80 flex items-center justify-between gap-2"
                      >
                        <div>
                          <div className="font-bold text-slate-900 text-xs">
                            {app.patient_name} ({app.patient_phone})
                          </div>
                          <div className="text-[11px] text-slate-500 mt-0.5">
                            {app.clinic_name} • {formatDate(app.appointment_date)} {app.start_time}
                          </div>
                        </div>

                        <div className="flex items-center gap-2 shrink-0">
                          <span className="font-bold text-slate-800 text-xs">
                            {formatPrice(app.consultation_price)}
                          </span>
                          {getStatusBadge(app.status)}
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          )}
        </Modal>
      )}
    </div>
  );
}

