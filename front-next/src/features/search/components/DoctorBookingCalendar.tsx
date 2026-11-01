'use client';

import React, { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Calendar as CalendarIcon,
  Clock,
  Building2,
  MapPin,
  Phone,
  CheckCircle2,
  CalendarOff,
  ChevronRight,
  Info,
} from 'lucide-react';
import { searchApi } from '../api/searchApi';
import { DoctorClinicWorkplace, AvailabilitySlot } from '../types';
import { formatCurrency } from '@/lib/utils/formatters';
import { Button } from '@/components/ui/Button';

interface DoctorBookingCalendarProps {
  doctorId: string;
  workplaces: DoctorClinicWorkplace[];
}

const UZ_WEEKDAYS_SHORT = ['Yak', 'Dush', 'Sesh', 'Chor', 'Pay', 'Jum', 'Shan'];
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
const UZ_MONTHS_SHORT = [
  'Yan',
  'Fev',
  'Mar',
  'Apr',
  'May',
  'Iyun',
  'Iyul',
  'Avg',
  'Sen',
  'Okt',
  'Noy',
  'Dek',
];

// Get current date string in Tashkent timezone (UTC+5)
function getTashkentTodayString(): string {
  const now = new Date();
  const tashkentOffsetMs = 5 * 60 * 60 * 1000;
  const utc = now.getTime() + now.getTimezoneOffset() * 60 * 1000;
  const tashkentDate = new Date(utc + tashkentOffsetMs);
  return tashkentDate.toISOString().split('T')[0];
}

export function DoctorBookingCalendar({ doctorId, workplaces }: DoctorBookingCalendarProps) {
  // 1. Selected Workplace (default to first)
  const [selectedClinicId, setSelectedClinicId] = useState<string>(
    workplaces?.[0]?.clinic_id || ''
  );

  // 2. Selected Date
  const todayStr = useMemo(() => getTashkentTodayString(), []);
  const [selectedDate, setSelectedDate] = useState<string>(todayStr);

  // 3. Selected Slot
  const [selectedSlot, setSelectedSlot] = useState<AvailabilitySlot | null>(null);

  // Active Workplace object
  const activeWorkplace = useMemo(() => {
    return workplaces.find((w) => w.clinic_id === selectedClinicId) || workplaces[0];
  }, [workplaces, selectedClinicId]);

  // Generate next 14 days
  const upcomingDays = useMemo(() => {
    const days: Array<{
      dateStr: string;
      dayNum: number;
      weekdayShort: string;
      monthShort: string;
      isToday: boolean;
    }> = [];

    const [y, m, d] = todayStr.split('-').map(Number);
    const startDate = new Date(y, m - 1, d);

    for (let i = 0; i < 14; i++) {
      const cur = new Date(startDate);
      cur.setDate(startDate.getDate() + i);

      const curYear = cur.getFullYear();
      const curMonth = String(cur.getMonth() + 1).padStart(2, '0');
      const curDay = String(cur.getDate()).padStart(2, '0');
      const dateStr = `${curYear}-${curMonth}-${curDay}`;

      days.push({
        dateStr,
        dayNum: cur.getDate(),
        weekdayShort: UZ_WEEKDAYS_SHORT[cur.getDay()],
        monthShort: UZ_MONTHS_SHORT[cur.getMonth()],
        isToday: i === 0,
      });
    }

    return days;
  }, [todayStr]);

  // 4. Fetch Availability for selected clinic, doctor, and date
  const {
    data: availabilityData,
    isLoading: isLoadingSlots,
    isError,
  } = useQuery({
    queryKey: ['publicAvailability', selectedClinicId, doctorId, selectedDate],
    queryFn: async () => {
      if (!selectedClinicId || !doctorId || !selectedDate) return null;
      const res = await searchApi.getDoctorAvailability(
        selectedClinicId,
        doctorId,
        selectedDate
      );
      return res.data;
    },
    enabled: Boolean(selectedClinicId && doctorId && selectedDate),
  });

  const slots = availabilityData?.slots || [];

  // Formatted date string for Uzbek language (e.g. "9-noyabr, Dushanba")
  const formattedSelectedDate = useMemo(() => {
    if (!selectedDate) return '';
    const [y, m, d] = selectedDate.split('-').map(Number);
    const dateObj = new Date(y, m - 1, d);
    const day = dateObj.getDate();
    const month = UZ_MONTHS[dateObj.getMonth()];
    const weekday = [
      'Yakshanba',
      'Dushanba',
      'Seshanba',
      'Chorshanba',
      'Payshanba',
      'Juma',
      'Shanba',
    ][dateObj.getDay()];
    return `${day}-${month}, ${weekday}`;
  }, [selectedDate]);

  if (!workplaces || workplaces.length === 0) {
    return null;
  }

  return (
    <div className="bg-white rounded-3xl border border-slate-200/80 p-6 sm:p-8 shadow-subtle space-y-6">
      {/* Title */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 pb-4 border-b border-slate-100">
        <div>
          <h2 className="text-xl font-extrabold text-brand-navy flex items-center gap-2">
            <CalendarIcon className="w-5 h-5 text-brand-blue" />
            <span>Shifokor qabul vaqtlari</span>
          </h2>
          <p className="text-xs text-slate-500 mt-0.5">
            Qulay sana va vaqtni tanlab shifokor qabuliga navbat oling
          </p>
        </div>

        {/* Multi-Clinic Selection Pills */}
        {workplaces.length > 1 && (
          <div className="flex items-center gap-1.5 overflow-x-auto py-1">
            <span className="text-xs font-bold text-slate-400 mr-1 shrink-0">Klinika:</span>
            {workplaces.map((wp) => (
              <button
                key={wp.clinic_id}
                type="button"
                onClick={() => {
                  setSelectedClinicId(wp.clinic_id);
                  setSelectedSlot(null);
                }}
                className={`px-3 py-1.5 rounded-xl text-xs font-bold transition-all shrink-0 flex items-center gap-1.5 border ${
                  selectedClinicId === wp.clinic_id
                    ? 'bg-brand-blue text-white border-brand-blue shadow-2xs'
                    : 'bg-white border-slate-200 text-slate-600 hover:bg-slate-50'
                }`}
              >
                <Building2 className="w-3.5 h-3.5" />
                <span>{wp.clinic_name}</span>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* Selected Clinic Details Strip */}
      {activeWorkplace && (
        <div className="bg-slate-50 rounded-2xl p-4 border border-slate-100 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 text-xs">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <span className="font-bold text-slate-900 text-sm">
                {activeWorkplace.clinic_name}
              </span>
              <span className="text-slate-400">•</span>
              <span className="font-semibold text-emerald-600 font-mono">
                {formatCurrency(activeWorkplace.consultation_price)}
              </span>
            </div>

            <div className="flex items-center gap-1.5 text-slate-500">
              <MapPin className="w-3.5 h-3.5 text-slate-400 shrink-0" />
              <span className="truncate max-w-md">
                {activeWorkplace.address ? `${activeWorkplace.address}, ` : ''}
                {activeWorkplace.district ? `${activeWorkplace.district}, ` : ''}
                {activeWorkplace.region}
              </span>
            </div>
          </div>

          {activeWorkplace.phone && (
            <a
              href={`tel:${activeWorkplace.phone}`}
              className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-white border border-slate-200 text-slate-700 font-semibold text-xs hover:border-brand-blue hover:text-brand-blue transition-all shrink-0"
            >
              <Phone className="w-3.5 h-3.5 text-brand-blue" />
              <span>{activeWorkplace.phone}</span>
            </a>
          )}
        </div>
      )}

      {/* Date Strip (Next 14 Days) */}
      <div className="space-y-2">
        <div className="flex items-center justify-between text-xs">
          <span className="font-bold text-slate-700">Qabul sanasini tanlang:</span>
          <div className="flex items-center gap-2">
            <span className="text-slate-400 text-[11px]">Boshqa sana:</span>
            <input
              type="date"
              min={todayStr}
              value={selectedDate}
              onChange={(e) => {
                if (e.target.value) {
                  setSelectedDate(e.target.value);
                  setSelectedSlot(null);
                }
              }}
              className="px-2.5 py-1 rounded-lg border border-slate-200 text-xs font-mono text-slate-700 focus:outline-none focus:ring-1 focus:ring-brand-blue bg-white"
            />
          </div>
        </div>

        <div className="flex items-center gap-2 overflow-x-auto pb-2 pt-1 scrollbar-none">
          {upcomingDays.map((day) => {
            const isSelected = selectedDate === day.dateStr;
            return (
              <button
                key={day.dateStr}
                type="button"
                onClick={() => {
                  setSelectedDate(day.dateStr);
                  setSelectedSlot(null);
                }}
                className={`flex flex-col items-center justify-center min-w-[62px] py-2.5 px-2 rounded-2xl border transition-all shrink-0 ${
                  isSelected
                    ? 'bg-brand-blue text-white border-brand-blue shadow-sm shadow-brand-blue/20 scale-[1.02]'
                    : 'bg-white border-slate-200 text-slate-700 hover:bg-slate-50 hover:border-slate-300'
                }`}
              >
                <span
                  className={`text-[10px] font-bold uppercase tracking-wider ${
                    isSelected ? 'text-blue-100' : 'text-slate-400'
                  }`}
                >
                  {day.isToday ? 'Bugun' : day.weekdayShort}
                </span>
                <span className="text-base font-extrabold my-0.5 font-mono">{day.dayNum}</span>
                <span
                  className={`text-[10px] font-medium ${
                    isSelected ? 'text-blue-100' : 'text-slate-500'
                  }`}
                >
                  {day.monthShort}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {/* Available Slots Section */}
      <div className="space-y-3 pt-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Clock className="w-4 h-4 text-brand-blue" />
            <span className="text-sm font-bold text-slate-900">
              {formattedSelectedDate}
            </span>
          </div>

          {!isLoadingSlots && (
            <span className="text-xs font-semibold text-slate-400">
              {slots.length > 0 ? `${slots.length} ta bo‘sh vaqt` : 'Bo‘sh vaqt yo‘q'}
            </span>
          )}
        </div>

        {/* Loading State */}
        {isLoadingSlots && (
          <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-6 gap-2.5 py-4">
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <div key={i} className="h-11 rounded-xl bg-slate-100 animate-pulse" />
            ))}
          </div>
        )}

        {/* Error State */}
        {!isLoadingSlots && isError && (
          <div className="bg-rose-50 border border-rose-200 rounded-2xl p-4 text-center text-xs text-rose-800">
            Qabul vaqtlarini yuklashda xatolik yuz berdi. Iltimos, qayta urinib ko‘ring.
          </div>
        )}

        {/* Empty State */}
        {!isLoadingSlots && !isError && slots.length === 0 && (
          <div className="bg-slate-50 border border-slate-200/80 rounded-2xl p-8 text-center space-y-2">
            <CalendarOff className="w-8 h-8 text-slate-300 mx-auto" />
            <h4 className="font-bold text-slate-800 text-sm">
              Bu kunga qabul vaqtlari belgilanmagan
            </h4>
            <p className="text-xs text-slate-500 max-w-sm mx-auto">
              Shifokorning bu sanada ish vaqti yo‘q yoki dam olish kuni. Iltimos, boshqa sanani tanlab ko‘ring.
            </p>
          </div>
        )}

        {/* Slots Grid */}
        {!isLoadingSlots && !isError && slots.length > 0 && (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-2.5">
            {slots.map((slot: AvailabilitySlot) => {
              const isSlotSelected =
                selectedSlot?.start_time === slot.start_time &&
                selectedSlot?.end_time === slot.end_time;

              return (
                <button
                  key={`${slot.start_time}-${slot.end_time}`}
                  type="button"
                  onClick={() => setSelectedSlot(slot)}
                  className={`py-2.5 px-3 rounded-xl border text-center transition-all font-mono text-xs font-bold flex items-center justify-center gap-1.5 ${
                    isSlotSelected
                      ? 'bg-brand-blue text-white border-brand-blue shadow-sm shadow-brand-blue/20 scale-[1.02]'
                      : 'bg-white border-slate-200 text-slate-800 hover:bg-blue-50 hover:border-blue-200 hover:text-brand-blue'
                  }`}
                >
                  <Clock className={`w-3.5 h-3.5 ${isSlotSelected ? 'text-white' : 'text-slate-400'}`} />
                  <span>
                    {slot.start_time} – {slot.end_time}
                  </span>
                </button>
              );
            })}
          </div>
        )}
      </div>

      {/* Booking Action Card */}
      <div className="bg-gradient-to-r from-blue-50/80 via-slate-50 to-indigo-50/50 rounded-2xl p-5 border border-blue-100 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <span className="text-xs font-bold text-slate-900">
              {selectedSlot ? (
                <span className="text-brand-blue flex items-center gap-1.5">
                  <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                  Tanlangan vaqt: {formattedSelectedDate}, {selectedSlot.start_time} – {selectedSlot.end_time}
                </span>
              ) : (
                'Qabulga yozilish uchun vaqt tanlang'
              )}
            </span>
          </div>
          <p className="text-[11px] text-slate-500 flex items-center gap-1">
            <Info className="w-3.5 h-3.5 text-slate-400 shrink-0" />
            Onlayn to‘lov va bron qilish funksiyasi tez kunda ishga tushiriladi.
          </p>
        </div>

        <div className="w-full sm:w-auto">
          <button
            type="button"
            disabled
            className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-6 py-2.5 rounded-xl bg-brand-blue/70 text-white font-bold text-xs shadow-xs cursor-not-allowed select-none"
            title="Qabulga yozilish tez kunda ishga tushadi"
          >
            <CalendarIcon className="w-4 h-4" />
            <span>Qabulga yozilish</span>
            <span className="text-[9px] uppercase font-bold tracking-wider px-1.5 py-0.2 rounded bg-white/25 text-white ml-1">
              Tez kunda
            </span>
          </button>
        </div>
      </div>
    </div>
  );
}
