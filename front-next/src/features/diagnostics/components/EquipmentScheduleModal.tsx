'use client';

import React, { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  Clock,
  Calendar,
  Save,
  RotateCcw,
  Sparkles,
  Check,
  X,
  AlertCircle,
  Copy,
  Info,
} from 'lucide-react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { diagnosticApi } from '../api/diagnosticApi';
import { DiagnosticEquipment, DiagnosticWeeklySlot } from '../types';
import { cn } from '@/lib/utils/cn';

interface EquipmentScheduleModalProps {
  isOpen: boolean;
  onClose: () => void;
  equipment: DiagnosticEquipment | null;
  clinicId: string;
}

const WEEKDAYS = [
  { id: 1, name: 'Dushanba', short: 'Dush' },
  { id: 2, name: 'Seshanba', short: 'Sesh' },
  { id: 3, name: 'Chorshanba', short: 'Chor' },
  { id: 4, name: 'Payshanba', short: 'Pay' },
  { id: 5, name: 'Juma', short: 'Jum' },
  { id: 6, name: 'Shanba', short: 'Shan' },
  { id: 7, name: 'Yakshanba', short: 'Yak' },
];

interface DaySchedule {
  weekday: number;
  isActive: boolean;
  startTime: string;
  endTime: string;
}

const DEFAULT_SCHEDULE: DaySchedule[] = [
  { weekday: 1, isActive: true, startTime: '09:00', endTime: '18:00' },
  { weekday: 2, isActive: true, startTime: '09:00', endTime: '18:00' },
  { weekday: 3, isActive: true, startTime: '09:00', endTime: '18:00' },
  { weekday: 4, isActive: true, startTime: '09:00', endTime: '18:00' },
  { weekday: 5, isActive: true, startTime: '09:00', endTime: '18:00' },
  { weekday: 6, isActive: true, startTime: '09:00', endTime: '15:00' },
  { weekday: 7, isActive: false, startTime: '09:00', endTime: '18:00' },
];

export function EquipmentScheduleModal({
  isOpen,
  onClose,
  equipment,
  clinicId,
}: EquipmentScheduleModalProps) {
  const queryClient = useQueryClient();
  const [schedule, setSchedule] = useState<DaySchedule[]>(DEFAULT_SCHEDULE);
  const [hasChanges, setHasChanges] = useState(false);

  // Fetch existing schedule
  const { data: scheduleData, isLoading } = useQuery({
    queryKey: ['equipmentWeeklySchedule', clinicId, equipment?.id],
    queryFn: async () => {
      if (!clinicId || !equipment?.id) return null;
      return await diagnosticApi.getWeeklySchedule(clinicId, equipment.id);
    },
    enabled: isOpen && !!clinicId && !!equipment?.id,
  });

  // Populate form with existing slots
  useEffect(() => {
    if (!scheduleData?.data) {
      setSchedule(DEFAULT_SCHEDULE);
      setHasChanges(false);
      return;
    }

    const fetchedSlots: DiagnosticWeeklySlot[] = scheduleData.data;
    if (fetchedSlots.length === 0) {
      setSchedule(DEFAULT_SCHEDULE);
      setHasChanges(false);
      return;
    }

    const newSchedule: DaySchedule[] = WEEKDAYS.map((w) => {
      const match = fetchedSlots.find((s) => s.weekday === w.id && s.status === 'active');
      if (match) {
        return {
          weekday: w.id,
          isActive: true,
          startTime: match.start_time,
          endTime: match.end_time,
        };
      }
      return {
        weekday: w.id,
        isActive: false,
        startTime: '09:00',
        endTime: '18:00',
      };
    });

    setSchedule(newSchedule);
    setHasChanges(false);
  }, [scheduleData, isOpen]);

  // Mutation to save schedule
  const saveMutation = useMutation({
    mutationFn: async () => {
      if (!equipment?.id) throw new Error('Uskuna topilmadi');

      // Validation
      for (const day of schedule) {
        if (day.isActive) {
          if (!day.startTime || !day.endTime) {
            throw new Error(`${WEEKDAYS.find((w) => w.id === day.weekday)?.name} uchun vaqt to‘liq kiritilmagan`);
          }
          if (day.startTime >= day.endTime) {
            throw new Error(
              `${WEEKDAYS.find((w) => w.id === day.weekday)?.name} uchun boshlanish vaqti tugash vaqtidan kichik bo‘lishi shart`
            );
          }
        }
      }

      const activeSlots = schedule
        .filter((d) => d.isActive)
        .map((d) => ({
          weekday: d.weekday,
          start_time: d.startTime,
          end_time: d.endTime,
          status: 'active',
        }));

      return await diagnosticApi.saveWeeklySchedule(clinicId, equipment.id, {
        slots: activeSlots,
      });
    },
    onSuccess: () => {
      toast.success('Uskuna ish jadvali muvaffaqiyatli saqlandi');
      queryClient.invalidateQueries({
        queryKey: ['equipmentWeeklySchedule', clinicId, equipment?.id],
      });
      queryClient.invalidateQueries({
        queryKey: ['clinicEquipments', clinicId],
      });
      queryClient.invalidateQueries({
        queryKey: ['diagnosticAvailability'],
      });
      queryClient.invalidateQueries({
        queryKey: ['equipmentAvailability'],
      });
      setHasChanges(false);
      onClose();
    },
    onError: (err: any) => {
      const msg = err?.response?.data?.message || err?.message || 'Jadvalni saqlashda xatolik yuz berdi';
      toast.error(msg);
    },
  });

  const handleToggleDay = (weekday: number) => {
    setSchedule((prev) =>
      prev.map((d) => (d.weekday === weekday ? { ...d, isActive: !d.isActive } : d))
    );
    setHasChanges(true);
  };

  const handleTimeChange = (weekday: number, field: 'startTime' | 'endTime', value: string) => {
    setSchedule((prev) =>
      prev.map((d) => (d.weekday === weekday ? { ...d, [field]: value } : d))
    );
    setHasChanges(true);
  };

  const applyStandardPreset = () => {
    setSchedule(DEFAULT_SCHEDULE);
    setHasChanges(true);
    toast.info('Standart ish vaqti o‘rnatildi (Dush-Jum 09:00-18:00, Shan 09:00-15:00)');
  };

  const apply247Preset = () => {
    setSchedule(
      WEEKDAYS.map((w) => ({
        weekday: w.id,
        isActive: true,
        startTime: '00:00',
        endTime: '23:59',
      }))
    );
    setHasChanges(true);
    toast.info('24/7 rejim o‘rnatildi');
  };

  const copyMondayToAllWorkdays = () => {
    const monday = schedule.find((d) => d.weekday === 1);
    if (!monday) return;
    setSchedule((prev) =>
      prev.map((d) => {
        if (d.weekday >= 1 && d.weekday <= 5) {
          return {
            ...d,
            isActive: monday.isActive,
            startTime: monday.startTime,
            endTime: monday.endTime,
          };
        }
        return d;
      })
    );
    setHasChanges(true);
    toast.info('Dushanba vaqti Dushanba-Juma kunlariga nusxalandi');
  };

  if (!isOpen || !equipment) return null;

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={`${equipment.name} — Ish jadvalini sozlash`}
      description="Apparatning haftalik qabul vaqtlarini belgilang. Tizim ushbu vaqtlar asosida 30 daqiqalik qabul slotlarini avtomatik yaratadi."
      className="max-w-2xl"
    >
      <div className="space-y-6 pt-2">
        {/* Equipment summary badge */}
        <div className="flex flex-wrap items-center justify-between gap-2 p-3 bg-slate-50 border border-slate-200 rounded-xl text-xs">
          <div className="flex items-center gap-2">
            <span className="font-bold text-slate-800">{equipment.name}</span>
            {equipment.room_number && (
              <span className="text-slate-500 font-medium">({equipment.room_number}-xona)</span>
            )}
            {equipment.category && (
              <span className="px-2 py-0.5 rounded bg-blue-100 text-brand-blue font-semibold">
                {equipment.category.name}
              </span>
            )}
          </div>
          <span className="text-slate-500 flex items-center gap-1">
            <Info className="w-3.5 h-3.5 text-brand-blue" />
            30 daqiqalik intervallar bilan qabul
          </span>
        </div>

        {/* Quick presets */}
        <div className="space-y-2">
          <div className="text-xs font-semibold text-slate-600 flex items-center gap-1.5">
            <Sparkles className="w-3.5 h-3.5 text-amber-500" />
            Tezkor andozalar
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={applyStandardPreset}
              className="text-xs text-slate-700 hover:text-brand-blue"
            >
              Standart (Dush-Jum 09:00-18:00, Shan 09:00-15:00)
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={apply247Preset}
              className="text-xs text-slate-700 hover:text-brand-blue"
            >
              24/7 (Har kuni tunu-kun)
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={copyMondayToAllWorkdays}
              className="text-xs text-slate-700 hover:text-brand-blue gap-1"
            >
              <Copy className="w-3 h-3" />
              Dush vaqtini Dush-Jum ga ko‘chirish
            </Button>
          </div>
        </div>

        {/* Days editor table/list */}
        <div className="space-y-2.5">
          {isLoading ? (
            <div className="space-y-3 py-6">
              {[...Array(7)].map((_, i) => (
                <div key={i} className="h-12 bg-slate-100 rounded-xl animate-pulse" />
              ))}
            </div>
          ) : (
            schedule.map((day) => {
              const weekdayMeta = WEEKDAYS.find((w) => w.id === day.weekday)!;
              return (
                <div
                  key={day.weekday}
                  className={cn(
                    'flex flex-col sm:flex-row sm:items-center justify-between p-3.5 rounded-xl border transition-all gap-3',
                    day.isActive
                      ? 'bg-white border-slate-200 shadow-2xs'
                      : 'bg-slate-50/70 border-slate-200/60 opacity-75'
                  )}
                >
                  {/* Left: Day name and active toggle */}
                  <div className="flex items-center gap-3 min-w-[150px]">
                    <button
                      type="button"
                      onClick={() => handleToggleDay(day.weekday)}
                      className={cn(
                        'w-5 h-5 rounded-md flex items-center justify-center border transition-all',
                        day.isActive
                          ? 'bg-brand-blue border-brand-blue text-white'
                          : 'border-slate-300 bg-white hover:border-slate-400'
                      )}
                    >
                      {day.isActive && <Check className="w-3.5 h-3.5 stroke-[3]" />}
                    </button>
                    <div>
                      <p className="text-sm font-bold text-slate-900">{weekdayMeta.name}</p>
                      <p className="text-[11px] font-medium text-slate-500">
                        {day.isActive ? 'Ish kuni' : 'Dam olish kuni'}
                      </p>
                    </div>
                  </div>

                  {/* Right: Time inputs or Dam olish kuni badge */}
                  {day.isActive ? (
                    <div className="flex items-center gap-2">
                      <div className="flex items-center gap-1.5">
                        <Clock className="w-3.5 h-3.5 text-slate-400 hidden sm:inline" />
                        <input
                          type="time"
                          value={day.startTime}
                          onChange={(e) => handleTimeChange(day.weekday, 'startTime', e.target.value)}
                          className="px-2.5 py-1.5 text-xs font-semibold text-slate-800 bg-slate-50 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-blue/30 focus:border-brand-blue"
                        />
                      </div>
                      <span className="text-slate-400 font-bold text-xs">—</span>
                      <div className="flex items-center gap-1.5">
                        <input
                          type="time"
                          value={day.endTime}
                          onChange={(e) => handleTimeChange(day.weekday, 'endTime', e.target.value)}
                          className="px-2.5 py-1.5 text-xs font-semibold text-slate-800 bg-slate-50 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-blue/30 focus:border-brand-blue"
                        />
                      </div>
                    </div>
                  ) : (
                    <div className="sm:self-center">
                      <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold bg-slate-200/80 text-slate-600">
                        Dam olish kuni
                      </span>
                    </div>
                  )}
                </div>
              );
            })
          )}
        </div>

        {/* Modal footer */}
        <div className="flex items-center justify-between pt-4 border-t border-slate-200">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onClose}
            disabled={saveMutation.isPending}
            className="text-slate-600"
          >
            Bekor qilish
          </Button>

          <Button
            type="button"
            size="sm"
            onClick={() => saveMutation.mutate()}
            disabled={saveMutation.isPending}
            className="gap-2 px-5 shadow-sm"
          >
            <Save className="w-4 h-4" />
            {saveMutation.isPending ? 'Saqlanmoqda...' : 'Jadvalni saqlash'}
          </Button>
        </div>
      </div>
    </Modal>
  );
}

