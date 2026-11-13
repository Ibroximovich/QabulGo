'use client';

import React, { useState, useMemo, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { diagnosticApi } from '../api/diagnosticApi';
import { DiagnosticAvailabilitySlot } from '../types';
import { Calendar as CalendarIcon, Clock, AlertCircle } from 'lucide-react';
import { cn } from '@/lib/utils/cn';

interface EquipmentSlotPickerProps {
  equipmentId: string;
  onSelectSlot: (slot: DiagnosticAvailabilitySlot, date: string) => void;
  selectedSlot?: DiagnosticAvailabilitySlot | null;
  selectedDate?: string;
}

const UZ_WEEKDAYS_UPPER = [
  'YAKSHANBA',
  'DUSHANBA',
  'SESHANBA',
  'CHORSHANBA',
  'PAYSHANBA',
  'JUMA',
  'SHANBA',
];

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

// Tashkent timezone (UTC+5) bugungi sanasini hisoblash
function getTashkentTodayString(): string {
  const now = new Date();
  const tashkentOffsetMs = 5 * 60 * 60 * 1000;
  const utc = now.getTime() + now.getTimezoneOffset() * 60 * 1000;
  const tashkentDate = new Date(utc + tashkentOffsetMs);
  const y = tashkentDate.getFullYear();
  const m = String(tashkentDate.getMonth() + 1).padStart(2, '0');
  const d = String(tashkentDate.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

export function EquipmentSlotPicker({
  equipmentId,
  onSelectSlot,
  selectedSlot,
  selectedDate: initialDate,
}: EquipmentSlotPickerProps) {
  const todayStr = useMemo(() => getTashkentTodayString(), []);

  // Tashkent vaqti asosida 14 kunlik ro'yxatni shakllantirish
  const days = useMemo(() => {
    const [y, m, d] = todayStr.split('-').map(Number);
    const startDate = new Date(y, m - 1, d);
    const result = [];

    for (let i = 0; i < 14; i++) {
      const cur = new Date(startDate);
      cur.setDate(startDate.getDate() + i);

      const curYear = cur.getFullYear();
      const curMonth = String(cur.getMonth() + 1).padStart(2, '0');
      const curDay = String(cur.getDate()).padStart(2, '0');
      const dateStr = `${curYear}-${curMonth}-${curDay}`;

      const dayOfWeek = cur.getDay(); // 0=Yakshanba, 1=Dushanba .. 6=Shanba
      const dayNum = cur.getDate();
      const monthName = UZ_MONTHS[cur.getMonth()];

      result.push({
        dateStr,
        weekdayUpper: i === 0 ? 'BUGUN' : UZ_WEEKDAYS_UPPER[dayOfWeek],
        dayMonth: `${dayNum}-${monthName}`,
        isToday: i === 0,
      });
    }
    return result;
  }, [todayStr]);

  const [activeDate, setActiveDate] = useState<string>(initialDate || days[0]?.dateStr || todayStr);

  useEffect(() => {
    if (initialDate) {
      setActiveDate(initialDate);
    }
  }, [initialDate]);

  const { data: availabilityData, isLoading, isError } = useQuery({
    queryKey: ['equipmentAvailability', equipmentId, activeDate],
    queryFn: () => diagnosticApi.getEquipmentAvailability(equipmentId, activeDate),
    enabled: !!equipmentId && !!activeDate,
  });

  const slots = availabilityData?.data?.slots || [];

  return (
    <div className="space-y-6">
      {/* 1. Date strip */}
      <div>
        <div className="flex items-center justify-between text-xs mb-3">
          <label className="flex items-center gap-1.5 font-bold text-slate-700 uppercase tracking-wider">
            <CalendarIcon className="w-4 h-4 text-brand-blue" />
            Qabul sanasini tanlang
          </label>
          <div className="flex items-center gap-2">
            <span className="text-slate-400 text-[11px]">Boshqa sana:</span>
            <input
              type="date"
              min={todayStr}
              value={activeDate}
              onChange={(e) => {
                if (e.target.value) {
                  setActiveDate(e.target.value);
                }
              }}
              className="px-2.5 py-1 rounded-lg border border-slate-200 text-xs font-mono text-slate-700 focus:outline-none focus:ring-1 focus:ring-brand-blue bg-white"
            />
          </div>
        </div>

        <div className="flex gap-2 overflow-x-auto pb-2 scrollbar-none">
          {days.map((item) => {
            const isSelected = activeDate === item.dateStr;
            return (
              <button
                key={item.dateStr}
                type="button"
                onClick={() => setActiveDate(item.dateStr)}
                className={cn(
                  'flex flex-col items-center justify-center min-w-[80px] py-2.5 px-3 rounded-2xl border text-center transition-all shrink-0',
                  isSelected
                    ? 'bg-brand-blue text-white border-brand-blue shadow-sm shadow-blue-500/20 scale-[1.02]'
                    : 'bg-white text-slate-700 border-slate-200 hover:border-brand-blue/40 hover:bg-slate-50'
                )}
              >
                <span
                  className={cn(
                    'text-[10px] font-bold uppercase tracking-wider',
                    isSelected ? 'text-blue-100' : 'text-slate-400'
                  )}
                >
                  {item.weekdayUpper}
                </span>
                <span className="text-xs font-extrabold mt-0.5 whitespace-nowrap">{item.dayMonth}</span>
              </button>
            );
          })}
        </div>
      </div>

      {/* 2. Time slots grid */}
      <div>
        <div className="flex items-center justify-between text-xs mb-3">
          <label className="flex items-center gap-1.5 font-bold text-slate-700 uppercase tracking-wider">
            <Clock className="w-4 h-4 text-brand-blue" />
            Bo‘sh vaqt oralig‘ini tanlang
          </label>
          {slots.length > 0 && (
            <span className="text-slate-500 text-[11px] font-medium">
              Mavjud slotlar: <strong className="text-brand-blue font-bold">{slots.length} ta</strong>
            </span>
          )}
        </div>

        {isLoading ? (
          <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-6 gap-2.5">
            {[...Array(12)].map((_, i) => (
              <div key={i} className="h-11 bg-slate-100 rounded-xl animate-pulse" />
            ))}
          </div>
        ) : isError ? (
          <div className="p-4 rounded-xl bg-rose-50 border border-rose-200 text-rose-700 text-xs flex items-center gap-2">
            <AlertCircle className="w-4 h-4 flex-shrink-0" />
            <span>Bo‘sh vaqtlarni yuklashda xatolik yuz berdi. Qaytadan urinib ko‘ring.</span>
          </div>
        ) : slots.length === 0 ? (
          <div className="p-6 rounded-2xl bg-slate-50 border border-slate-200/80 text-center space-y-1.5">
            <Clock className="w-8 h-8 text-slate-300 mx-auto" />
            <p className="text-sm font-semibold text-slate-700">
              {activeDate === todayStr
                ? 'Bugungi barcha qabul vaqtlari yakunlangan'
                : 'Ushbu kunda bo‘sh qabul vaqtlari mavjud emas'}
            </p>
            <p className="text-xs text-slate-400">
              {activeDate === todayStr
                ? 'Iltimos, ertangi yoki keyingi kunlar uchun bo‘sh vaqtni tanlang.'
                : 'Ushbu kun dam olish kuni bo‘lishi mumkin. Boshqa sanani tanlab ko‘ring.'}
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-6 gap-2.5">
            {slots.map((slot) => {
              const isSelected = selectedSlot?.start_time === slot.start_time;
              return (
                <button
                  key={slot.start_time}
                  type="button"
                  onClick={() => onSelectSlot(slot, activeDate)}
                  className={cn(
                    'py-2.5 px-3 rounded-xl text-xs font-bold transition-all border text-center shadow-xs',
                    isSelected
                      ? 'bg-brand-blue text-white border-brand-blue ring-2 ring-brand-blue/30 scale-105'
                      : 'bg-white text-slate-800 border-slate-200 hover:border-brand-blue hover:text-brand-blue hover:bg-blue-50/20'
                  )}
                >
                  {slot.start_time}
                </button>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
