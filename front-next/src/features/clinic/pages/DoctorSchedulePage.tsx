'use client';

import React, { useState, useEffect, useMemo } from 'react';
import { useRouter, useParams } from 'next/navigation';
import Link from 'next/link';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  ChevronLeft,
  Calendar,
  Clock,
  Sparkles,
  Plus,
  Trash2,
  AlertTriangle,
  CheckCircle2,
  Save,
  Coffee,
  X,
  AlertCircle,
  HelpCircle,
  Check,
  CalendarDays,
  CalendarOff,
} from 'lucide-react';
import { useClinicContext } from '@/components/layout/clinic/ClinicLayout';
import { clinicApi } from '../api/clinicApi';
import { scheduleApi } from '../api/scheduleApi';
import {
  GenerateScheduleRequest,
  TimeBreakItem,
  SaveWeeklySlotItem,
  DoctorDateOverride,
  OverrideStatus,
} from '../types';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';
import { getMediaUrl } from '@/lib/utils/media';

const WEEKDAYS = [
  { id: 1, name: 'Dushanba', short: 'Dush' },
  { id: 2, name: 'Seshanba', short: 'Sesh' },
  { id: 3, name: 'Chorshanba', short: 'Chor' },
  { id: 4, name: 'Payshanba', short: 'Pay' },
  { id: 5, name: 'Juma', short: 'Jum' },
  { id: 6, name: 'Shanba', short: 'Shan' },
  { id: 7, name: 'Yakshanba', short: 'Yak' },
];

export function DoctorSchedulePage() {
  const { clinic, isLoadingClinic } = useClinicContext();
  const router = useRouter();
  const params = useParams();
  const doctorId = params?.doctorId as string;
  const queryClient = useQueryClient();

  // Active Tab: 'weekly' or 'exceptions'
  const [activeTab, setActiveTab] = useState<'weekly' | 'exceptions'>('weekly');

  // Selected Day in Weekly Editor (0 = All, 1..7 = specific day)
  const [selectedDayFilter, setSelectedDayFilter] = useState<number>(0);

  // Local state for weekly slots being edited
  const [localSlots, setLocalSlots] = useState<SaveWeeklySlotItem[]>([]);
  const [isDirty, setIsDirty] = useState<boolean>(false);

  // Auto Generate State
  const [showAutoModal, setShowAutoModal] = useState<boolean>(false);
  const [genWeekdays, setGenWeekdays] = useState<number[]>([1, 2, 3, 4, 5]);
  const [genStartTime, setGenStartTime] = useState<string>('09:00');
  const [genEndTime, setGenEndTime] = useState<string>('17:00');
  const [genDuration, setGenDuration] = useState<number>(60);
  const [genBreaks, setGenBreaks] = useState<TimeBreakItem[]>([
    { start_time: '13:00', end_time: '14:00' },
  ]);
  const [previewData, setPreviewData] = useState<{
    totalSlots: number;
    slots: SaveWeeklySlotItem[];
  } | null>(null);

  // Save Confirmation Modal
  const [showSaveConfirmModal, setShowSaveConfirmModal] = useState<boolean>(false);

  // Manual Add Slot Modal
  const [showAddSlotModal, setShowAddSlotModal] = useState<boolean>(false);
  const [newSlotWeekday, setNewSlotWeekday] = useState<number>(1);
  const [newSlotStartTime, setNewSlotStartTime] = useState<string>('09:00');
  const [newSlotEndTime, setNewSlotEndTime] = useState<string>('10:00');

  // Exception Modal
  const [showExceptionModal, setShowExceptionModal] = useState<boolean>(false);
  const [editingException, setEditingException] = useState<DoctorDateOverride | null>(null);
  const [exDate, setExDate] = useState<string>('');
  const [exStatus, setExStatus] = useState<OverrideStatus>('day_off');
  const [exStartTime, setExStartTime] = useState<string>('10:00');
  const [exEndTime, setExEndTime] = useState<string>('11:00');
  const [exReason, setExReason] = useState<string>('');
  const [deletingExceptionId, setDeletingExceptionId] = useState<string | null>(null);

  // 1. Fetch Doctor Info
  const { data: doctor, isLoading: isLoadingDoctor } = useQuery({
    queryKey: ['clinicDoctor', clinic?.id, doctorId],
    queryFn: async () => {
      if (!clinic?.id || !doctorId) return null;
      const res = await clinicApi.getDoctor(clinic.id, doctorId);
      return res.data?.data;
    },
    enabled: Boolean(clinic?.id && doctorId),
  });

  // 2. Fetch Weekly Schedule
  const { data: scheduleData, isLoading: isLoadingSchedule } = useQuery({
    queryKey: ['doctorWeeklySchedule', clinic?.id, doctorId],
    queryFn: async () => {
      if (!clinic?.id || !doctorId) return null;
      const res = await scheduleApi.getWeeklySchedule(clinic.id, doctorId);
      return res.data?.data;
    },
    enabled: Boolean(clinic?.id && doctorId),
  });

  // Sync server schedule to local state on load
  useEffect(() => {
    if (scheduleData?.slots) {
      const items: SaveWeeklySlotItem[] = scheduleData.slots.map((s) => ({
        weekday: s.weekday,
        start_time: s.start_time,
        end_time: s.end_time,
        status: s.status,
      }));
      setLocalSlots(items);
      setIsDirty(false);
    }
  }, [scheduleData]);

  // Warn user on leaving page with unsaved changes
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (isDirty) {
        e.preventDefault();
        e.returnValue = '';
      }
    };
    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [isDirty]);

  // 3. Fetch Exceptions
  const { data: exceptionsData, isLoading: isLoadingExceptions } = useQuery({
    queryKey: ['doctorDateExceptions', clinic?.id, doctorId],
    queryFn: async () => {
      if (!clinic?.id || !doctorId) return [];
      const res = await scheduleApi.getDateExceptions(clinic.id, doctorId);
      return res.data?.data || [];
    },
    enabled: Boolean(clinic?.id && doctorId),
  });

  // Overlap Detection on Frontend
  const overlapErrors = useMemo(() => {
    const errors: { [key: string]: string } = {};
    const slotsByDay: { [key: number]: Array<{ index: number; start: number; end: number; label: string }> } = {};

    localSlots.forEach((slot, index) => {
      const [sh, sm] = slot.start_time.split(':').map(Number);
      const [eh, em] = slot.end_time.split(':').map(Number);
      const sMins = sh * 60 + sm;
      const eMins = eh * 60 + em;

      if (sMins >= eMins) {
        errors[`${index}`] = "Boshlanish vaqti tugashidan kichik bo'lishi shart";
      }

      if (!slotsByDay[slot.weekday]) slotsByDay[slot.weekday] = [];
      slotsByDay[slot.weekday].push({
        index,
        start: sMins,
        end: eMins,
        label: `${slot.start_time}–${slot.end_time}`,
      });
    });

    Object.entries(slotsByDay).forEach(([_, daySlots]) => {
      daySlots.sort((a, b) => a.start - b.start);
      for (let i = 1; i < daySlots.length; i++) {
        if (daySlots[i].start < daySlots[i - 1].end) {
          errors[`${daySlots[i].index}`] = `To'qnashuv: ${daySlots[i - 1].label} bilan ustma-ust`;
          errors[`${daySlots[i - 1].index}`] = `To'qnashuv: ${daySlots[i].label} bilan ustma-ust`;
        }
      }
    });

    return errors;
  }, [localSlots]);

  const hasOverlapErrors = Object.keys(overlapErrors).length > 0;

  // Generate Preview Mutation
  const generateMutation = useMutation({
    mutationFn: async (payload: GenerateScheduleRequest) => {
      if (!clinic?.id || !doctorId) return null;
      const res = await scheduleApi.generateSchedulePreview(clinic.id, doctorId, payload);
      return res.data?.data;
    },
    onSuccess: (data) => {
      if (!data) return;
      const items: SaveWeeklySlotItem[] = data.slots.map((s) => ({
        weekday: s.weekday,
        start_time: s.start_time,
        end_time: s.end_time,
        status: s.status,
      }));
      setPreviewData({
        totalSlots: data.total_slots,
        slots: items,
      });
      toast.success(`${data.total_slots} ta slot muvaffaqiyatli hisoblandi`);
    },
    onError: (err: any) => {
      const msg = err.response?.data?.message || 'Avtomatik slotlarni hisoblashda xatolik';
      toast.error(msg);
    },
  });

  // Save Schedule Mutation
  const saveScheduleMutation = useMutation({
    mutationFn: async (slots: SaveWeeklySlotItem[]) => {
      if (!clinic?.id || !doctorId) return null;
      const res = await scheduleApi.saveWeeklySchedule(clinic.id, doctorId, { slots });
      return res.data?.data;
    },
    onSuccess: () => {
      toast.success('Haftalik qabul jadvali muvaffaqiyatli saqlandi');
      setIsDirty(false);
      setShowSaveConfirmModal(false);
      queryClient.invalidateQueries({ queryKey: ['doctorWeeklySchedule', clinic?.id, doctorId] });
    },
    onError: (err: any) => {
      const status = err.response?.status;
      const msg = err.response?.data?.message || 'Jadvalni saqlashda xatolik yuz berdi';
      if (status === 409) {
        toast.error(`Klinikalararo to'qnashuv: ${msg}`, { duration: 6000 });
      } else {
        toast.error(msg);
      }
    },
  });

  // Create/Update Exception Mutation
  const saveExceptionMutation = useMutation({
    mutationFn: async () => {
      if (!clinic?.id || !doctorId) return;
      const payload = {
        date: exDate,
        status: exStatus,
        start_time: exStatus === 'day_off' ? undefined : exStartTime,
        end_time: exStatus === 'day_off' ? undefined : exEndTime,
        reason: exReason || undefined,
      };

      if (editingException) {
        await scheduleApi.updateDateException(clinic.id, doctorId, editingException.id, payload);
      } else {
        await scheduleApi.createDateException(clinic.id, doctorId, payload);
      }
    },
    onSuccess: () => {
      toast.success(editingException ? 'Istisno yangilandi' : 'Yangi istisno qo‘shildi');
      setShowExceptionModal(false);
      setEditingException(null);
      queryClient.invalidateQueries({ queryKey: ['doctorDateExceptions', clinic?.id, doctorId] });
    },
    onError: (err: any) => {
      toast.error(err.response?.data?.message || 'Istisnoni saqlashda xatolik');
    },
  });

  // Delete Exception Mutation
  const deleteExceptionMutation = useMutation({
    mutationFn: async (id: string) => {
      if (!clinic?.id || !doctorId) return;
      await scheduleApi.deleteDateException(clinic.id, doctorId, id);
    },
    onSuccess: () => {
      toast.success('Istisno o‘chirildi');
      setDeletingExceptionId(null);
      queryClient.invalidateQueries({ queryKey: ['doctorDateExceptions', clinic?.id, doctorId] });
    },
    onError: (err: any) => {
      toast.error(err.response?.data?.message || 'Istisnoni o‘chirishda xatolik');
    },
  });

  // Handlers for Local Slots
  const handleRemoveSlot = (index: number) => {
    setLocalSlots((prev) => prev.filter((_, i) => i !== index));
    setIsDirty(true);
  };

  const handleApplyPreview = () => {
    if (!previewData) return;
    setLocalSlots(previewData.slots);
    setIsDirty(true);
    setPreviewData(null);
    setShowAutoModal(false);
    toast.success('Hisoblangan slotlar tahrirlovchiga yuklandi. Saqlashni unutmang!');
  };

  const handleAddManualSlot = () => {
    if (newSlotStartTime >= newSlotEndTime) {
      toast.error("Boshlanish vaqti tugash vaqtidan kichik bo'lishi shart");
      return;
    }
    const newSlot: SaveWeeklySlotItem = {
      weekday: newSlotWeekday,
      start_time: newSlotStartTime,
      end_time: newSlotEndTime,
      status: 'active',
    };
    setLocalSlots((prev) => [...prev, newSlot]);
    setIsDirty(true);
    setShowAddSlotModal(false);
    toast.success('Yangi vaqt qo‘shildi');
  };

  // Breaks Management in Auto Modal
  const handleAddBreak = () => {
    setGenBreaks((prev) => [...prev, { start_time: '13:00', end_time: '14:00' }]);
  };

  const handleRemoveBreak = (idx: number) => {
    setGenBreaks((prev) => prev.filter((_, i) => i !== idx));
  };

  const handleBreakChange = (idx: number, field: 'start_time' | 'end_time', val: string) => {
    setGenBreaks((prev) => {
      const copy = [...prev];
      copy[idx] = { ...copy[idx], [field]: val };
      return copy;
    });
  };

  const handleToggleGenWeekday = (wd: number) => {
    setGenWeekdays((prev) =>
      prev.includes(wd) ? prev.filter((w) => w !== wd) : [...prev, wd].sort((a, b) => a - b)
    );
  };

  // Open Add Exception
  const handleOpenAddException = () => {
    setEditingException(null);
    const today = new Date().toISOString().split('T')[0];
    setExDate(today);
    setExStatus('day_off');
    setExStartTime('10:00');
    setExEndTime('11:00');
    setExReason('');
    setShowExceptionModal(true);
  };

  // Open Edit Exception
  const handleOpenEditException = (ex: DoctorDateOverride) => {
    setEditingException(ex);
    setExDate(ex.date);
    setExStatus(ex.status);
    setExStartTime(ex.start_time || '10:00');
    setExEndTime(ex.end_time || '11:00');
    setExReason(ex.reason || '');
    setShowExceptionModal(true);
  };

  if (isLoadingClinic || isLoadingDoctor || isLoadingSchedule) {
    return (
      <div className="p-8 text-center text-slate-500 flex flex-col items-center justify-center min-h-[400px]">
        <Clock className="w-8 h-8 animate-spin text-brand-blue mb-3" />
        <p className="text-sm font-medium">Shifokor jadvali yuklanmoqda...</p>
      </div>
    );
  }

  const doctorFullName = doctor ? `${doctor.first_name} ${doctor.last_name}` : 'Shifokor';

  return (
    <div className="max-w-6xl mx-auto space-y-6 pb-16">
      {/* Top Header & Breadcrumbs */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-5 rounded-2xl border border-slate-200 shadow-xs">
        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            onClick={() => router.push('/clinic/doctors')}
            className="h-9 w-9 p-0 rounded-xl"
            title="Shifokorlar ro‘yxatiga qaytish"
          >
            <ChevronLeft className="w-4 h-4" />
          </Button>

          <div className="flex items-center gap-3">
            {doctor?.photo_url ? (
              <img
                src={getMediaUrl(doctor.photo_url)}
                alt={doctorFullName}
                className="w-12 h-12 rounded-xl object-cover border border-slate-200 shrink-0"
              />
            ) : (
              <div className="w-12 h-12 rounded-xl bg-blue-50 text-brand-blue font-bold flex items-center justify-center text-sm border border-blue-100 shrink-0">
                {doctor?.first_name?.[0] || 'D'}
                {doctor?.last_name?.[0] || ''}
              </div>
            )}

            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-lg font-extrabold text-slate-900 leading-tight">
                  {doctorFullName}
                </h1>
                <span className="text-xs px-2 py-0.5 rounded-md font-semibold bg-blue-50 text-brand-blue">
                  {doctor?.specialty?.name || 'Mutaxassis'}
                </span>
              </div>
              <p className="text-xs text-slate-500 mt-0.5">
                Qabul davomiyligi: ~{doctor?.appointment_duration_minutes || 30} daqiqa | Qabul jadvalini boshqarish
              </p>
            </div>
          </div>
        </div>

        {/* Unsaved indicator & Save button */}
        <div className="flex items-center gap-2">
          {isDirty && (
            <span className="inline-flex items-center gap-1.5 text-xs text-amber-600 font-semibold bg-amber-50 border border-amber-200 px-3 py-1.5 rounded-xl">
              <AlertCircle className="w-3.5 h-3.5" />
              Saqlanmagan o‘zgarishlar
            </span>
          )}

          <Button
            variant="primary"
            size="sm"
            onClick={() => setShowSaveConfirmModal(true)}
            disabled={!isDirty || hasOverlapErrors || saveScheduleMutation.isPending}
            isLoading={saveScheduleMutation.isPending}
            className="gap-1.5 px-4 h-9 font-semibold shadow-xs"
          >
            <Save className="w-4 h-4" />
            Jadvalni saqlash
          </Button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-2 border-b border-slate-200">
        <button
          onClick={() => setActiveTab('weekly')}
          className={`flex items-center gap-2 px-4 py-2.5 text-sm font-bold border-b-2 transition-all ${
            activeTab === 'weekly'
              ? 'border-brand-blue text-brand-blue'
              : 'border-transparent text-slate-500 hover:text-slate-900'
          }`}
        >
          <CalendarDays className="w-4 h-4" />
          <span>Haftalik asosiy jadval</span>
          <span className="ml-1 text-xs px-2 py-0.5 rounded-full bg-slate-100 text-slate-600 font-semibold">
            {localSlots.length} ta slot
          </span>
        </button>

        <button
          onClick={() => setActiveTab('exceptions')}
          className={`flex items-center gap-2 px-4 py-2.5 text-sm font-bold border-b-2 transition-all ${
            activeTab === 'exceptions'
              ? 'border-brand-blue text-brand-blue'
              : 'border-transparent text-slate-500 hover:text-slate-900'
          }`}
        >
          <CalendarOff className="w-4 h-4" />
          <span>Istisno sanalar va ta‘tillar</span>
          {exceptionsData && exceptionsData.length > 0 && (
            <span className="ml-1 text-xs px-2 py-0.5 rounded-full bg-amber-100 text-amber-800 font-semibold">
              {exceptionsData.length}
            </span>
          )}
        </button>
      </div>

      {/* Tab 1: Weekly Schedule */}
      {activeTab === 'weekly' && (
        <div className="space-y-6">
          {/* Action Toolbar */}
          <div className="bg-gradient-to-r from-blue-50/70 via-indigo-50/50 to-white p-4 sm:p-5 rounded-2xl border border-blue-100 flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
            <div>
              <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                <Sparkles className="w-4 h-4 text-brand-blue" />
                Haftalik qabul vaqtlarini boshqarish
              </h3>
              <p className="text-xs text-slate-500 mt-0.5">
                Avtomatik tarzda tanaffuslar bilan hisoblang yoki qo‘lda nostandart vaqtlarni kiriting.
              </p>
            </div>

            <div className="flex flex-wrap items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowAutoModal(true)}
                className="gap-1.5 text-xs bg-white text-slate-700 hover:text-brand-blue hover:border-brand-blue shadow-2xs font-semibold"
              >
                <Sparkles className="w-3.5 h-3.5 text-brand-blue" />
                Avtomatik yaratish
              </Button>

              <Button
                variant="primary"
                size="sm"
                onClick={() => setShowAddSlotModal(true)}
                className="gap-1.5 text-xs font-semibold shadow-2xs"
              >
                <Plus className="w-3.5 h-3.5" />
                Qo‘lda vaqt qo‘shish
              </Button>
            </div>
          </div>

          {/* Overlap Error Warning Banner */}
          {hasOverlapErrors && (
            <div className="bg-rose-50 border border-rose-200 text-rose-800 p-4 rounded-2xl flex items-start gap-3">
              <AlertTriangle className="w-5 h-5 text-rose-600 shrink-0 mt-0.5" />
              <div className="text-xs">
                <p className="font-bold">Qabul vaqtlari to‘qnashuvi aniqlandi!</p>
                <p className="mt-0.5">
                  Bitta kunda slotlar bir-birini qoplamasligi shart. Iltimos, belgilangan qizil vaqtlarni to‘g‘rilang yoki o‘chiring.
                </p>
              </div>
            </div>
          )}

          {/* Day Filter Pills */}
          <div className="flex items-center gap-1.5 overflow-x-auto pb-1 scrollbar-none">
            <button
              onClick={() => setSelectedDayFilter(0)}
              className={`px-3 py-1.5 rounded-xl text-xs font-bold transition-all shrink-0 ${
                selectedDayFilter === 0
                  ? 'bg-slate-900 text-white shadow-2xs'
                  : 'bg-white border border-slate-200 text-slate-600 hover:bg-slate-50'
              }`}
            >
              Barcha kunlar ({localSlots.length})
            </button>
            {WEEKDAYS.map((wd) => {
              const count = localSlots.filter((s) => s.weekday === wd.id).length;
              return (
                <button
                  key={wd.id}
                  onClick={() => setSelectedDayFilter(wd.id)}
                  className={`px-3 py-1.5 rounded-xl text-xs font-bold transition-all shrink-0 flex items-center gap-1.5 ${
                    selectedDayFilter === wd.id
                      ? 'bg-brand-blue text-white shadow-2xs'
                      : 'bg-white border border-slate-200 text-slate-600 hover:bg-slate-50'
                  }`}
                >
                  <span>{wd.short}</span>
                  <span
                    className={`text-[10px] px-1.5 py-0.2 rounded-full font-semibold ${
                      selectedDayFilter === wd.id ? 'bg-white/20 text-white' : 'bg-slate-100 text-slate-500'
                    }`}
                  >
                    {count}
                  </span>
                </button>
              );
            })}
          </div>

          {/* Days Slots Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {WEEKDAYS.filter((wd) => selectedDayFilter === 0 || selectedDayFilter === wd.id).map((wd) => {
              const daySlots = localSlots
                .map((slot, originalIndex) => ({ ...slot, originalIndex }))
                .filter((slot) => slot.weekday === wd.id)
                .sort((a, b) => a.start_time.localeCompare(b.start_time));

              return (
                <div
                  key={wd.id}
                  className="bg-white rounded-2xl border border-slate-200 p-4 shadow-xs flex flex-col justify-between"
                >
                  <div>
                    {/* Day Header */}
                    <div className="flex items-center justify-between pb-3 border-b border-slate-100 mb-3">
                      <div>
                        <h4 className="font-bold text-slate-900 text-sm">{wd.name}</h4>
                        <span className="text-[11px] text-slate-400">
                          {daySlots.length > 0 ? `${daySlots.length} ta slot` : 'Dam olish kuni'}
                        </span>
                      </div>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => {
                          setNewSlotWeekday(wd.id);
                          setShowAddSlotModal(true);
                        }}
                        className="h-7 px-2 text-xs text-brand-blue hover:bg-blue-50"
                        title={`${wd.name}ga yangi vaqt qo‘shish`}
                      >
                        <Plus className="w-3.5 h-3.5" />
                        Qo‘shish
                      </Button>
                    </div>

                    {/* Slots List */}
                    {daySlots.length === 0 ? (
                      <div className="py-8 text-center text-slate-400 text-xs">
                        <Calendar className="w-6 h-6 mx-auto mb-1.5 opacity-40" />
                        <p>Qabul vaqtlari belgilanmagan</p>
                      </div>
                    ) : (
                      <div className="space-y-1.5">
                        {daySlots.map((slot) => {
                          const err = overlapErrors[`${slot.originalIndex}`];
                          return (
                            <div
                              key={slot.originalIndex}
                              className={`flex items-center justify-between p-2 rounded-xl text-xs font-mono transition-all border ${
                                err
                                  ? 'bg-rose-50 border-rose-300 text-rose-900 shadow-2xs'
                                  : 'bg-slate-50 border-slate-200/80 text-slate-800 hover:border-slate-300'
                              }`}
                            >
                              <div className="flex items-center gap-2">
                                <Clock className={`w-3.5 h-3.5 ${err ? 'text-rose-500' : 'text-slate-400'}`} />
                                <span className="font-semibold">
                                  {slot.start_time} – {slot.end_time}
                                </span>
                              </div>

                              <button
                                onClick={() => handleRemoveSlot(slot.originalIndex)}
                                className="text-slate-400 hover:text-rose-600 transition-colors p-1"
                                title="O‘chirish"
                              >
                                <Trash2 className="w-3.5 h-3.5" />
                              </button>
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>

                  {/* Day Footer Summary */}
                  {daySlots.length > 0 && (
                    <div className="pt-3 mt-3 border-t border-slate-100 flex items-center justify-between text-[11px] text-slate-400">
                      <span>Ish vaqti:</span>
                      <span className="font-semibold text-slate-700 font-mono">
                        {daySlots[0].start_time} – {daySlots[daySlots.length - 1].end_time}
                      </span>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Tab 2: Exceptions & Overrides */}
      {activeTab === 'exceptions' && (
        <div className="space-y-6">
          {/* Header */}
          <div className="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div>
              <h3 className="text-sm font-bold text-slate-900">
                Alohida sana va istisnolar (Overrides)
              </h3>
              <p className="text-xs text-slate-500 mt-0.5">
                Ma’lum bir sanada kunni yopish, bitta vaqtni yopish yoki yangi qabul ochish. Bu haftalik asosiy jadvalni o‘zgartirmaydi.
              </p>
            </div>

            <Button
              variant="primary"
              size="sm"
              onClick={handleOpenAddException}
              className="gap-1.5 font-semibold text-xs shadow-xs"
            >
              <Plus className="w-3.5 h-3.5" />
              Yangi istisno qo‘shish
            </Button>
          </div>

          {/* Exceptions List */}
          {isLoadingExceptions ? (
            <div className="p-8 text-center text-slate-400 text-xs">Istisnolar yuklanmoqda...</div>
          ) : !exceptionsData || exceptionsData.length === 0 ? (
            <div className="bg-white rounded-2xl border border-slate-200 p-12 text-center max-w-md mx-auto shadow-xs">
              <CalendarOff className="w-10 h-10 text-slate-300 mx-auto mb-3" />
              <h4 className="font-bold text-slate-800 text-sm mb-1">
                Istisno sanalar belgilanmagan
              </h4>
              <p className="text-xs text-slate-500 mb-4">
                Hozircha shifokor uchun ta’til, bayram yoki maxsus yopilgan qabul kunlari kiritilmagan.
              </p>
              <Button size="sm" variant="outline" onClick={handleOpenAddException} className="gap-1.5 text-xs">
                <Plus className="w-3.5 h-3.5" />
                Birinchi istisnoni qo‘shish
              </Button>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {exceptionsData.map((ex) => (
                <div
                  key={ex.id}
                  className="bg-white rounded-2xl border border-slate-200 p-4 shadow-xs flex flex-col justify-between"
                >
                  <div className="space-y-2.5">
                    <div className="flex items-center justify-between">
                      <span className="font-mono text-xs font-bold text-slate-900 px-2.5 py-1 rounded-lg bg-slate-100">
                        {ex.date}
                      </span>

                      {ex.status === 'day_off' && (
                        <span className="text-[11px] font-bold text-rose-700 bg-rose-50 border border-rose-200 px-2 py-0.5 rounded-full">
                          Kun to‘liq yopilgan
                        </span>
                      )}
                      {ex.status === 'closed' && (
                        <span className="text-[11px] font-bold text-amber-700 bg-amber-50 border border-amber-200 px-2 py-0.5 rounded-full">
                          Slot yopilgan
                        </span>
                      )}
                      {ex.status === 'available' && (
                        <span className="text-[11px] font-bold text-emerald-700 bg-emerald-50 border border-emerald-200 px-2 py-0.5 rounded-full">
                          Qo‘shimcha slot
                        </span>
                      )}
                    </div>

                    {ex.status !== 'day_off' && ex.start_time && ex.end_time && (
                      <div className="flex items-center gap-1.5 text-xs font-mono font-semibold text-slate-700">
                        <Clock className="w-3.5 h-3.5 text-slate-400" />
                        <span>
                          {ex.start_time} – {ex.end_time}
                        </span>
                      </div>
                    )}

                    {ex.reason && (
                      <p className="text-xs text-slate-600 bg-slate-50 p-2.5 rounded-xl border border-slate-100 italic">
                        "{ex.reason}"
                      </p>
                    )}
                  </div>

                  <div className="pt-3 mt-3 border-t border-slate-100 flex items-center justify-end gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => handleOpenEditException(ex)}
                      className="h-8 px-2.5 text-xs"
                    >
                      Tahrirlash
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => setDeletingExceptionId(ex.id)}
                      className="h-8 px-2 text-rose-600 hover:bg-rose-50"
                      title="O‘chirish"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* AUTO GENERATE MODAL */}
      <Modal
        isOpen={showAutoModal}
        onClose={() => {
          setShowAutoModal(false);
          setPreviewData(null);
        }}
        title="Haftalik qabul jadvalini avtomatik yaratish"
      >
        <div className="space-y-5 max-h-[75vh] overflow-y-auto pr-1">
          {/* Step 1: Form */}
          <div className="space-y-4 text-xs">
            {/* Weekdays */}
            <div>
              <div className="flex items-center justify-between mb-2">
                <label className="font-bold text-slate-800">Qaysi kunlar uchun yaratilsin?</label>
                <div className="flex gap-2 text-[11px] text-brand-blue font-semibold">
                  <button
                    type="button"
                    onClick={() => setGenWeekdays([1, 2, 3, 4, 5])}
                    className="hover:underline"
                  >
                    Dush–Juma
                  </button>
                  <span>•</span>
                  <button
                    type="button"
                    onClick={() => setGenWeekdays([1, 2, 3, 4, 5, 6, 7])}
                    className="hover:underline"
                  >
                    Barcha kunlar
                  </button>
                </div>
              </div>
              <div className="grid grid-cols-4 sm:grid-cols-7 gap-1.5">
                {WEEKDAYS.map((wd) => {
                  const isSelected = genWeekdays.includes(wd.id);
                  return (
                    <button
                      key={wd.id}
                      type="button"
                      onClick={() => handleToggleGenWeekday(wd.id)}
                      className={`py-2 rounded-xl text-center font-bold text-xs transition-all border ${
                        isSelected
                          ? 'bg-brand-blue text-white border-brand-blue shadow-2xs'
                          : 'bg-white border-slate-200 text-slate-600 hover:bg-slate-50'
                      }`}
                    >
                      {wd.short}
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Hours Range */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="font-bold text-slate-800 block mb-1">Ish boshlanishi</label>
                <Input
                  type="time"
                  value={genStartTime}
                  onChange={(e) => setGenStartTime(e.target.value)}
                  className="font-mono text-xs"
                />
              </div>
              <div>
                <label className="font-bold text-slate-800 block mb-1">Ish tugashi</label>
                <Input
                  type="time"
                  value={genEndTime}
                  onChange={(e) => setGenEndTime(e.target.value)}
                  className="font-mono text-xs"
                />
              </div>
            </div>

            {/* Duration */}
            <div>
              <label className="font-bold text-slate-800 block mb-1">Bitta qabul davomiyligi</label>
              <div className="grid grid-cols-4 gap-2">
                {[20, 30, 45, 60].map((dur) => (
                  <button
                    key={dur}
                    type="button"
                    onClick={() => setGenDuration(dur)}
                    className={`py-2 rounded-xl text-xs font-bold transition-all border ${
                      genDuration === dur
                        ? 'bg-brand-blue text-white border-brand-blue shadow-2xs'
                        : 'bg-white border-slate-200 text-slate-600 hover:bg-slate-50'
                    }`}
                  >
                    {dur} daqiqa
                  </button>
                ))}
              </div>
            </div>

            {/* Breaks */}
            <div className="pt-2 border-t border-slate-100">
              <div className="flex items-center justify-between mb-2">
                <label className="font-bold text-slate-800 flex items-center gap-1.5">
                  <Coffee className="w-3.5 h-3.5 text-amber-600" />
                  Tanaffuslar (Breaks)
                </label>
                <button
                  type="button"
                  onClick={handleAddBreak}
                  className="text-brand-blue font-semibold hover:underline text-xs flex items-center gap-1"
                >
                  <Plus className="w-3.5 h-3.5" />
                  Tanaffus qo‘shish
                </button>
              </div>

              {genBreaks.length === 0 ? (
                <p className="text-slate-400 text-xs italic">Tanaffus belgilanmagan</p>
              ) : (
                <div className="space-y-2">
                  {genBreaks.map((brk, idx) => (
                    <div key={idx} className="flex items-center gap-2">
                      <Input
                        type="time"
                        value={brk.start_time}
                        onChange={(e) => handleBreakChange(idx, 'start_time', e.target.value)}
                        className="font-mono text-xs flex-1"
                      />
                      <span className="text-slate-400">–</span>
                      <Input
                        type="time"
                        value={brk.end_time}
                        onChange={(e) => handleBreakChange(idx, 'end_time', e.target.value)}
                        className="font-mono text-xs flex-1"
                      />
                      <button
                        type="button"
                        onClick={() => handleRemoveBreak(idx)}
                        className="text-slate-400 hover:text-rose-600 p-1.5"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Preview Results Banner */}
          {previewData && (
            <div className="bg-emerald-50 border border-emerald-200 rounded-2xl p-4 text-xs space-y-2">
              <div className="flex items-center gap-2 text-emerald-800 font-bold">
                <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                <span>Hisob-kitob muvaffaqiyatli: Jami {previewData.totalSlots} ta slot tayyor!</span>
              </div>
              <p className="text-slate-600">
                Ushbu preview ma’lumotlar bazasiga yozilmadi. Quyidagi tugma orqali tahrirlovchiga qabul qilishingiz mumkin.
              </p>
            </div>
          )}

          {/* Modal Actions */}
          <div className="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                setShowAutoModal(false);
                setPreviewData(null);
              }}
            >
              Bekor qilish
            </Button>

            {!previewData ? (
              <Button
                variant="primary"
                size="sm"
                isLoading={generateMutation.isPending}
                onClick={() =>
                  generateMutation.mutate({
                    weekdays: genWeekdays,
                    start_time: genStartTime,
                    end_time: genEndTime,
                    slot_duration_minutes: genDuration,
                    breaks: genBreaks,
                  })
                }
                className="gap-1.5"
              >
                <Sparkles className="w-3.5 h-3.5" />
                Preview hisoblash
              </Button>
            ) : (
              <Button variant="primary" size="sm" onClick={handleApplyPreview} className="gap-1.5 bg-emerald-600 hover:bg-emerald-700">
                <Check className="w-3.5 h-3.5" />
                Jadvalga yuklash ({previewData.totalSlots} ta slot)
              </Button>
            )}
          </div>
        </div>
      </Modal>

      {/* MANUAL ADD SLOT MODAL */}
      <Modal
        isOpen={showAddSlotModal}
        onClose={() => setShowAddSlotModal(false)}
        title="Qo‘lda qabul vaqti qo‘shish"
      >
        <div className="space-y-4 text-xs">
          <div>
            <label className="font-bold text-slate-800 block mb-1">Hafta kuni</label>
            <select
              value={newSlotWeekday}
              onChange={(e) => setNewSlotWeekday(Number(e.target.value))}
              className="w-full px-3 py-2 rounded-xl border border-slate-200 text-xs font-semibold text-slate-800 focus:outline-none focus:ring-2 focus:ring-brand-blue"
            >
              {WEEKDAYS.map((wd) => (
                <option key={wd.id} value={wd.id}>
                  {wd.name}
                </option>
              ))}
            </select>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="font-bold text-slate-800 block mb-1">Boshlanish vaqti</label>
              <Input
                type="time"
                value={newSlotStartTime}
                onChange={(e) => setNewSlotStartTime(e.target.value)}
                className="font-mono text-xs"
              />
            </div>
            <div>
              <label className="font-bold text-slate-800 block mb-1">Tugash vaqti</label>
              <Input
                type="time"
                value={newSlotEndTime}
                onChange={(e) => setNewSlotEndTime(e.target.value)}
                className="font-mono text-xs"
              />
            </div>
          </div>

          <div className="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
            <Button variant="outline" size="sm" onClick={() => setShowAddSlotModal(false)}>
              Bekor qilish
            </Button>
            <Button variant="primary" size="sm" onClick={handleAddManualSlot}>
              Qo‘shish
            </Button>
          </div>
        </div>
      </Modal>

      {/* SAVE CONFIRMATION MODAL */}
      <Modal
        isOpen={showSaveConfirmModal}
        onClose={() => setShowSaveConfirmModal(false)}
        title="Haftalik jadvalni saqlash"
      >
        <div className="space-y-4 text-xs">
          <p className="text-slate-600 leading-relaxed">
            Haqiqatan ham <span className="font-bold text-slate-900">{doctorFullName}</span> uchun yangi{' '}
            <span className="font-bold text-brand-blue">{localSlots.length} ta</span> qabul slotini saqlamoqchimisiz?
          </p>

          <div className="bg-slate-50 border border-slate-200 rounded-xl p-3 space-y-1">
            <span className="font-bold text-slate-800 block mb-1">Kunlar bo‘yicha taqsimot:</span>
            {WEEKDAYS.map((wd) => {
              const c = localSlots.filter((s) => s.weekday === wd.id).length;
              if (c === 0) return null;
              return (
                <div key={wd.id} className="flex justify-between text-slate-600">
                  <span>{wd.name}:</span>
                  <span className="font-semibold text-slate-900 font-mono">{c} ta slot</span>
                </div>
              );
            })}
          </div>

          <div className="flex items-center justify-end gap-2 pt-2">
            <Button variant="outline" size="sm" onClick={() => setShowSaveConfirmModal(false)}>
              Bekor qilish
            </Button>
            <Button
              variant="primary"
              size="sm"
              isLoading={saveScheduleMutation.isPending}
              onClick={() => saveScheduleMutation.mutate(localSlots)}
              className="gap-1.5"
            >
              <Save className="w-4 h-4" />
              Tasdiqlash va saqlash
            </Button>
          </div>
        </div>
      </Modal>

      {/* EXCEPTION CREATE/EDIT MODAL */}
      <Modal
        isOpen={showExceptionModal}
        onClose={() => setShowExceptionModal(false)}
        title={editingException ? "Istisno sanani tahrirlash" : "Yangi istisno sana qo'shish"}
      >
        <div className="space-y-4 text-xs">
          <div>
            <label className="font-bold text-slate-800 block mb-1">Sana</label>
            <Input
              type="date"
              value={exDate}
              onChange={(e) => setExDate(e.target.value)}
              className="font-mono text-xs"
            />
          </div>

          <div>
            <label className="font-bold text-slate-800 block mb-1">Istisno turi</label>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-2">
              <button
                type="button"
                onClick={() => setExStatus('day_off')}
                className={`p-2.5 rounded-xl border text-left transition-all ${
                  exStatus === 'day_off'
                    ? 'bg-rose-50 border-rose-400 text-rose-900 font-bold'
                    : 'bg-white border-slate-200 text-slate-600 hover:bg-slate-50'
                }`}
              >
                <span className="block font-bold">Kunni yopish</span>
                <span className="text-[10px] text-slate-400 font-normal">Butun kun qabul yo‘q</span>
              </button>

              <button
                type="button"
                onClick={() => setExStatus('closed')}
                className={`p-2.5 rounded-xl border text-left transition-all ${
                  exStatus === 'closed'
                    ? 'bg-amber-50 border-amber-400 text-amber-900 font-bold'
                    : 'bg-white border-slate-200 text-slate-600 hover:bg-slate-50'
                }`}
              >
                <span className="block font-bold">Slotni yopish</span>
                <span className="text-[10px] text-slate-400 font-normal">Aynan bitta vaqtni yopish</span>
              </button>

              <button
                type="button"
                onClick={() => setExStatus('available')}
                className={`p-2.5 rounded-xl border text-left transition-all ${
                  exStatus === 'available'
                    ? 'bg-emerald-50 border-emerald-400 text-emerald-900 font-bold'
                    : 'bg-white border-slate-200 text-slate-600 hover:bg-slate-50'
                }`}
              >
                <span className="block font-bold">Qo‘shimcha slot</span>
                <span className="text-[10px] text-slate-400 font-normal">Yangi qabul ochish</span>
              </button>
            </div>
          </div>

          {exStatus !== 'day_off' && (
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="font-bold text-slate-800 block mb-1">Boshlanish vaqti</label>
                <Input
                  type="time"
                  value={exStartTime}
                  onChange={(e) => setExStartTime(e.target.value)}
                  className="font-mono text-xs"
                />
              </div>
              <div>
                <label className="font-bold text-slate-800 block mb-1">Tugash vaqti</label>
                <Input
                  type="time"
                  value={exEndTime}
                  onChange={(e) => setExEndTime(e.target.value)}
                  className="font-mono text-xs"
                />
              </div>
            </div>
          )}

          <div>
            <label className="font-bold text-slate-800 block mb-1">Sabab (ixtiyoriy)</label>
            <Input
              type="text"
              placeholder="Masalan: Malaka oshirish, Konsilium, Bayram..."
              value={exReason}
              onChange={(e) => setExReason(e.target.value)}
              className="text-xs"
            />
          </div>

          <div className="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
            <Button variant="outline" size="sm" onClick={() => setShowExceptionModal(false)}>
              Bekor qilish
            </Button>
            <Button
              variant="primary"
              size="sm"
              isLoading={saveExceptionMutation.isPending}
              onClick={() => saveExceptionMutation.mutate()}
            >
              Saqlash
            </Button>
          </div>
        </div>
      </Modal>

      {/* DELETE EXCEPTION CONFIRM MODAL */}
      {deletingExceptionId && (
        <Modal
          isOpen={Boolean(deletingExceptionId)}
          onClose={() => setDeletingExceptionId(null)}
          title="Istisnoni o‘chirish"
        >
          <div className="space-y-4 text-xs">
            <p className="text-slate-600">
              Haqiqatan ham ushbu istisnoni o‘chirmoqchimisiz? Shifokor ushbu sanada yana asosiy haftalik jadvali bo‘yicha qabulga qaytadi.
            </p>
            <div className="flex items-center justify-end gap-2 pt-2">
              <Button variant="outline" size="sm" onClick={() => setDeletingExceptionId(null)}>
                Bekor qilish
              </Button>
              <Button
                variant="destructive"
                size="sm"
                isLoading={deleteExceptionMutation.isPending}
                onClick={() => deleteExceptionMutation.mutate(deletingExceptionId)}
              >
                O‘chirish
              </Button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}

