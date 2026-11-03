'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import {
  Building2,
  Users,
  Stethoscope,
  CalendarCheck,
  Clock4,
  CheckCircle2,
  XCircle,
  Ban,
  Calendar,
  DollarSign,
  TrendingUp,
  AlertCircle,
  ArrowRight,
  UserX,
  RotateCcw,
} from 'lucide-react';
import { adminApi } from '../api/adminApi';
import { Button } from '@/components/ui/Button';

const DATE_RANGES = [
  { id: 'today', label: 'Bugun' },
  { id: '7d', label: '7 kun' },
  { id: '30d', label: '30 kun' },
  { id: '90d', label: '90 kun' },
];

export function AdminDashboardPage() {
  const [selectedRange, setSelectedRange] = useState<string>('30d');

  // 1. Fetch KPI Summary
  const {
    data: summaryData,
    isLoading: isLoadingSummary,
    isError: isErrorSummary,
    refetch: refetchSummary,
  } = useQuery({
    queryKey: ['adminDashboardSummary'],
    queryFn: async () => {
      const res = await adminApi.getDashboardSummary();
      return res.data?.data;
    },
  });

  // 2. Fetch Chart Analytics
  const {
    data: analyticsData,
    isLoading: isLoadingAnalytics,
    isError: isErrorAnalytics,
    refetch: refetchAnalytics,
  } = useQuery({
    queryKey: ['adminDashboardAnalytics', selectedRange],
    queryFn: async () => {
      const res = await adminApi.getDashboardAnalytics({ range: selectedRange });
      return res.data?.data;
    },
  });

  const summary = summaryData;
  const analytics = analyticsData;

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('uz-UZ').format(amount) + " so'm";
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'confirmed':
        return 'bg-blue-500';
      case 'completed':
        return 'bg-emerald-500';
      case 'cancelled':
        return 'bg-rose-500';
      case 'no_show':
        return 'bg-amber-500';
      default:
        return 'bg-slate-400';
    }
  };

  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'confirmed':
        return 'Tasdiqlangan';
      case 'completed':
        return 'Yakunlangan';
      case 'cancelled':
        return 'Bekor qilingan';
      case 'no_show':
        return 'Kelmagan';
      default:
        return status;
    }
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-black text-slate-900 tracking-tight">Platforma Nazorat Markazi</h1>
          <p className="text-xs sm:text-sm text-slate-500 mt-1">
            QabulGo platformasining barcha asosiy ko‘rsatkichlari, o‘sish sur’ati va jonli statistikasi
          </p>
        </div>

        {/* Date Filter Tabs */}
        <div className="flex items-center gap-1.5 bg-white p-1 rounded-2xl border border-slate-200/80 shadow-2xs self-start sm:self-auto">
          {DATE_RANGES.map((r) => (
            <button
              key={r.id}
              type="button"
              onClick={() => setSelectedRange(r.id)}
              className={`px-3 py-1.5 rounded-xl text-xs font-bold transition-all ${
                selectedRange === r.id
                  ? 'bg-[#0866F5] text-white shadow-2xs'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
              }`}
            >
              {r.label}
            </button>
          ))}
        </div>
      </div>

      {/* Pending Clinics Attention Banner */}
      {summary && summary.pending_clinics > 0 && (
        <div className="bg-amber-50 border border-amber-200 rounded-2xl p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 shadow-2xs">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-amber-500/10 flex items-center justify-center shrink-0">
              <Clock4 className="w-5 h-5 text-amber-600 animate-pulse" />
            </div>
            <div>
              <h4 className="text-sm font-bold text-amber-900">
                {summary.pending_clinics} ta yangi klinika tasdiqlashni kutmoqda!
              </h4>
              <p className="text-xs text-amber-700 mt-0.5">
                Klinika arizalari ko‘rib chiqilishi va faollashtirilishi kerak
              </p>
            </div>
          </div>
          <Link
            href="/admin/clinics?status=pending"
            className="inline-flex items-center gap-1.5 text-xs font-bold text-amber-900 bg-amber-200/60 hover:bg-amber-200 px-3.5 py-2 rounded-xl transition-colors"
          >
            <span>Ko‘rib chiqish</span>
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
      )}

      {/* 13 KPI Cards Grid */}
      <div>
        <h2 className="text-xs font-bold text-slate-400 uppercase tracking-wider mb-3">
          Asosiy ko‘rsatkichlar (KPI)
        </h2>

        {isLoadingSummary ? (
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3 sm:gap-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="bg-white rounded-2xl p-4 border border-slate-200/80 animate-pulse h-28" />
            ))}
          </div>
        ) : isErrorSummary ? (
          <div className="bg-white rounded-2xl p-6 border border-rose-200 text-center">
            <AlertCircle className="w-8 h-8 text-rose-500 mx-auto mb-2" />
            <p className="text-sm font-bold text-slate-800">Statistikani yuklashda xatolik yuz berdi</p>
            <Button size="sm" variant="outline" onClick={() => refetchSummary()} className="mt-3 gap-2">
              <RotateCcw className="w-3.5 h-3.5" />
              <span>Qayta urinish</span>
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3 sm:gap-4">
            {/* 1. Total Clinics */}
            <Link
              href="/admin/clinics"
              className="bg-white rounded-2xl p-4 border border-slate-200/80 hover:border-blue-400 transition-all hover:shadow-sm group"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Jami klinikalar</span>
                <div className="w-8 h-8 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center group-hover:bg-blue-600 group-hover:text-white transition-colors">
                  <Building2 className="w-4 h-4" />
                </div>
              </div>
              <div className="mt-2 text-2xl font-black text-slate-900">{summary?.total_clinics || 0}</div>
              <div className="mt-1 flex items-center gap-1.5 text-[11px] text-slate-500">
                <span className="text-emerald-600 font-bold">{summary?.approved_clinics || 0} faol</span>
                <span>•</span>
                <span className="text-amber-600 font-bold">{summary?.pending_clinics || 0} kutilmoqda</span>
              </div>
            </Link>

            {/* 2. Total Patients */}
            <Link
              href="/admin/patients"
              className="bg-white rounded-2xl p-4 border border-slate-200/80 hover:border-emerald-400 transition-all hover:shadow-sm group"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Jami bemorlar</span>
                <div className="w-8 h-8 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center group-hover:bg-emerald-600 group-hover:text-white transition-colors">
                  <Users className="w-4 h-4" />
                </div>
              </div>
              <div className="mt-2 text-2xl font-black text-slate-900">{summary?.total_patients || 0}</div>
              <div className="mt-1 text-[11px] text-emerald-600 font-bold">Ro‘yxatdan o‘tgan profil</div>
            </Link>

            {/* 3. Total Doctors */}
            <Link
              href="/admin/doctors"
              className="bg-white rounded-2xl p-4 border border-slate-200/80 hover:border-indigo-400 transition-all hover:shadow-sm group"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Jami shifokorlar</span>
                <div className="w-8 h-8 rounded-lg bg-indigo-50 text-indigo-600 flex items-center justify-center group-hover:bg-indigo-600 group-hover:text-white transition-colors">
                  <Stethoscope className="w-4 h-4" />
                </div>
              </div>
              <div className="mt-2 text-2xl font-black text-slate-900">{summary?.total_doctors || 0}</div>
              <div className="mt-1 text-[11px] text-slate-500">Mutaxassis shifokorlar</div>
            </Link>

            {/* 4. Total Bookings */}
            <Link
              href="/admin/appointments"
              className="bg-white rounded-2xl p-4 border border-slate-200/80 hover:border-purple-400 transition-all hover:shadow-sm group"
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Jami bronlar</span>
                <div className="w-8 h-8 rounded-lg bg-purple-50 text-purple-600 flex items-center justify-center group-hover:bg-purple-600 group-hover:text-white transition-colors">
                  <CalendarCheck className="w-4 h-4" />
                </div>
              </div>
              <div className="mt-2 text-2xl font-black text-slate-900">{summary?.total_appointments || 0}</div>
              <div className="mt-1 text-[11px] text-purple-600 font-bold">Barcha qabul yozuvlari</div>
            </Link>

            {/* 5. Today's Bookings */}
            <div className="bg-white rounded-2xl p-4 border border-slate-200/80">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Bugungi bronlar</span>
                <div className="w-8 h-8 rounded-lg bg-sky-50 text-sky-600 flex items-center justify-center">
                  <Calendar className="w-4 h-4" />
                </div>
              </div>
              <div className="mt-2 text-2xl font-black text-sky-600">{summary?.today_appointments || 0}</div>
              <div className="mt-1 text-[11px] text-slate-500">Bugungi qabullar</div>
            </div>

            {/* 6. This Week's Bookings */}
            <div className="bg-white rounded-2xl p-4 border border-slate-200/80">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Haftalik bronlar</span>
                <div className="w-8 h-8 rounded-lg bg-teal-50 text-teal-600 flex items-center justify-center">
                  <TrendingUp className="w-4 h-4" />
                </div>
              </div>
              <div className="mt-2 text-2xl font-black text-slate-900">{summary?.this_week_appointments || 0}</div>
              <div className="mt-1 text-[11px] text-slate-500">So‘nggi 7 kunlik</div>
            </div>

            {/* 7. This Month's Bookings */}
            <div className="bg-white rounded-2xl p-4 border border-slate-200/80">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Oylik bronlar</span>
                <div className="w-8 h-8 rounded-lg bg-cyan-50 text-cyan-600 flex items-center justify-center">
                  <CalendarCheck className="w-4 h-4" />
                </div>
              </div>
              <div className="mt-2 text-2xl font-black text-slate-900">{summary?.this_month_appointments || 0}</div>
              <div className="mt-1 text-[11px] text-slate-500">Joriy oy ko‘rsatkichi</div>
            </div>

            {/* 8. Completed Bookings */}
            <div className="bg-white rounded-2xl p-4 border border-slate-200/80">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Yakunlangan bronlar</span>
                <div className="w-8 h-8 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center">
                  <CheckCircle2 className="w-4 h-4" />
                </div>
              </div>
              <div className="mt-2 text-2xl font-black text-emerald-600">{summary?.completed_appointments || 0}</div>
              <div className="mt-1 text-[11px] text-slate-500">Muvaffaqiyatli qabul</div>
            </div>

            {/* 9. Cancelled Bookings */}
            <div className="bg-white rounded-2xl p-4 border border-slate-200/80">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Bekor qilinganlar</span>
                <div className="w-8 h-8 rounded-lg bg-rose-50 text-rose-600 flex items-center justify-center">
                  <XCircle className="w-4 h-4" />
                </div>
              </div>
              <div className="mt-2 text-2xl font-black text-rose-600">{summary?.cancelled_appointments || 0}</div>
              <div className="mt-1 text-[11px] text-slate-500">Bekor bo‘lgan</div>
            </div>

            {/* 10. No-Show Bookings */}
            <div className="bg-white rounded-2xl p-4 border border-slate-200/80">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">Kelmagan bemorlar</span>
                <div className="w-8 h-8 rounded-lg bg-amber-50 text-amber-600 flex items-center justify-center">
                  <UserX className="w-4 h-4" />
                </div>
              </div>
              <div className="mt-2 text-2xl font-black text-amber-600">{summary?.no_show_appointments || 0}</div>
              <div className="mt-1 text-[11px] text-slate-500">Qabulga kelmadi</div>
            </div>

            {/* 11. Suspended / Blocked Clinics */}
            <div className="bg-white rounded-2xl p-4 border border-slate-200/80">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-slate-500">To‘xtatilgan klinikalar</span>
                <div className="w-8 h-8 rounded-lg bg-red-50 text-red-600 flex items-center justify-center">
                  <Ban className="w-4 h-4" />
                </div>
              </div>
              <div className="mt-2 text-2xl font-black text-red-600">
                {(summary?.rejected_clinics || 0) + (summary?.suspended_clinics || 0)}
              </div>
              <div className="mt-1 text-[11px] text-slate-500">Rad etilgan va bloklangan</div>
            </div>

            {/* 12 & 13. Estimated Booking Revenue (Platform volume) */}
            <div className="bg-linear-to-br from-[#08234A] to-[#0866F5] rounded-2xl p-4 text-white shadow-xs col-span-2 sm:col-span-1 lg:col-span-1">
              <div className="flex items-center justify-between text-blue-200">
                <span className="text-xs font-bold">Taxminiy aylanma</span>
                <DollarSign className="w-4 h-4" />
              </div>
              <div className="mt-2 text-xl font-black truncate">
                {formatCurrency(summary?.estimated_total_revenue || 0)}
              </div>
              <p className="mt-1 text-[10px] text-blue-200/80 leading-tight">
                *Bronlarning taxminiy umumiy qiymati (to‘lov klinikada qilinadi)
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Analytics Charts & Breakdowns */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left 2 Cols: Daily Booking Volume Chart */}
        <div className="lg:col-span-2 bg-white rounded-2xl border border-slate-200/80 p-5 shadow-2xs">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h3 className="text-base font-bold text-slate-900">Bronlar dinamikasi</h3>
              <p className="text-xs text-slate-500">Tanlangan vaqt oralig‘idagi kunlik bronlar soni</p>
            </div>
          </div>

          {isLoadingAnalytics ? (
            <div className="h-64 flex items-center justify-center text-xs text-slate-400">Grafik yuklanmoqda...</div>
          ) : !analytics?.daily_bookings || analytics.daily_bookings.length === 0 ? (
            <div className="h-64 flex flex-col items-center justify-center text-slate-400">
              <Calendar className="w-10 h-10 mb-2 opacity-30" />
              <p className="text-xs font-semibold">Ushbu oraliqda bronlar mavjud emas</p>
            </div>
          ) : (
            <div>
              {/* Responsive SVG Bar Chart */}
              <div className="h-56 flex items-end gap-1.5 sm:gap-3 pt-6 pb-2 border-b border-slate-100 overflow-x-auto">
                {analytics.daily_bookings.map((item, idx) => {
                  const maxCount = Math.max(...analytics.daily_bookings.map((d) => d.count), 1);
                  const heightPercent = Math.max(Math.round((item.count / maxCount) * 100), 10);
                  return (
                    <div key={idx} className="flex-1 min-w-[28px] flex flex-col items-center group relative">
                      {/* Tooltip on hover */}
                      <div className="absolute -top-8 bg-slate-900 text-white text-[10px] py-0.5 px-2 rounded-md font-bold opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-20">
                        {item.date}: {item.count} ta bron
                      </div>
                      <div
                        style={{ height: `${heightPercent}%` }}
                        className="w-full max-w-[24px] bg-blue-500 rounded-t-md group-hover:bg-[#0866F5] transition-all"
                      />
                      <span className="text-[9px] text-slate-400 mt-2 truncate w-full text-center">
                        {item.date.slice(5)}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>

        {/* Right 1 Col: Booking Status Distribution */}
        <div className="bg-white rounded-2xl border border-slate-200/80 p-5 shadow-2xs flex flex-col justify-between">
          <div>
            <h3 className="text-base font-bold text-slate-900 mb-1">Statuslar taqsimoti</h3>
            <p className="text-xs text-slate-500 mb-4">Bronlarning holati bo‘yicha nisbati</p>

            {isLoadingAnalytics ? (
              <div className="space-y-3 py-6">
                {Array.from({ length: 4 }).map((_, i) => (
                  <div key={i} className="h-6 bg-slate-100 rounded-lg animate-pulse" />
                ))}
              </div>
            ) : !analytics?.status_distribution || analytics.status_distribution.length === 0 ? (
              <div className="py-12 text-center text-xs text-slate-400">Ma’lumotlar yo‘q</div>
            ) : (
              <div className="space-y-3">
                {analytics.status_distribution.map((item, idx) => {
                  const total = analytics.status_distribution.reduce((acc, curr) => acc + curr.count, 0) || 1;
                  const percent = Math.round((item.count / total) * 100);
                  return (
                    <div key={idx} className="space-y-1">
                      <div className="flex items-center justify-between text-xs font-semibold">
                        <span className="flex items-center gap-2">
                          <span className={`w-2.5 h-2.5 rounded-full ${getStatusColor(item.status)}`} />
                          <span>{getStatusLabel(item.status)}</span>
                        </span>
                        <span className="text-slate-600">
                          {item.count} ta ({percent}%)
                        </span>
                      </div>
                      <div className="h-2 w-full bg-slate-100 rounded-full overflow-hidden">
                        <div
                          style={{ width: `${percent}%` }}
                          className={`h-full rounded-full ${getStatusColor(item.status)}`}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          <div className="mt-6 pt-4 border-t border-slate-100">
            <Link
              href="/admin/appointments"
              className="text-xs font-bold text-[#0866F5] hover:underline flex items-center justify-between"
            >
              <span>Barcha bronlarni ko‘rish</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </Link>
          </div>
        </div>
      </div>

      {/* Top Performing Clinics Table */}
      <div className="bg-white rounded-2xl border border-slate-200/80 p-5 shadow-2xs">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h3 className="text-base font-bold text-slate-900">Eng faol klinikalar</h3>
            <p className="text-xs text-slate-500">Eng ko‘p bron qabul qilgan yetakchi klinikalar</p>
          </div>
          <Link
            href="/admin/clinics"
            className="text-xs font-bold text-[#0866F5] hover:underline flex items-center gap-1"
          >
            <span>Barchasi</span>
            <ArrowRight className="w-3.5 h-3.5" />
          </Link>
        </div>

        {isLoadingAnalytics ? (
          <div className="p-8 text-center text-xs text-slate-400">Yuklanmoqda...</div>
        ) : !analytics?.top_clinics || analytics.top_clinics.length === 0 ? (
          <div className="p-8 text-center text-xs text-slate-400">Faol klinikalar hozircha mavjud emas</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs text-slate-600">
              <thead className="bg-slate-50/80 text-slate-700 uppercase font-semibold border-b border-slate-200">
                <tr>
                  <th className="px-4 py-3">№</th>
                  <th className="px-4 py-3">Klinika nomi</th>
                  <th className="px-4 py-3 text-center">Jami bronlar</th>
                  <th className="px-4 py-3 text-right">Taxminiy hajm</th>
                  <th className="px-4 py-3 text-right">Amal</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 font-medium">
                {analytics.top_clinics.map((clinic, idx) => (
                  <tr key={clinic.clinic_id} className="hover:bg-slate-50/70 transition-colors">
                    <td className="px-4 py-3.5 font-bold text-slate-400">{idx + 1}</td>
                    <td className="px-4 py-3.5 font-bold text-slate-900">{clinic.clinic_name}</td>
                    <td className="px-4 py-3.5 text-center">
                      <span className="inline-block px-2.5 py-0.5 rounded-full text-xs font-bold bg-blue-50 text-blue-700">
                        {clinic.bookings_count} ta
                      </span>
                    </td>
                    <td className="px-4 py-3.5 text-right font-semibold text-slate-800">
                      {formatCurrency(clinic.total_revenue)}
                    </td>
                    <td className="px-4 py-3.5 text-right">
                      <Link
                        href={`/admin/clinics?clinicId=${clinic.clinic_id}`}
                        className="text-[#0866F5] font-bold hover:underline"
                      >
                        Batafsil
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

