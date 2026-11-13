'use client';

import React, { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { diagnosticApi } from '../api/diagnosticApi';
import { useClinicContext } from '@/components/layout/clinic/ClinicLayout';
import { DiagnosticEquipment, DiagnosticService, DiagnosticAppointment } from '../types';
import { formatCurrency } from '@/lib/utils/formatters';
import {
  Cpu,
  Layers,
  Calendar,
  Plus,
  Trash2,
  CheckCircle,
  AlertCircle,
  XCircle,
  Clock,
  ShieldCheck,
  Building2,
  Check,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';
import { Input } from '@/components/ui/Input';
import { cn } from '@/lib/utils/cn';
import { EquipmentScheduleModal } from '../components/EquipmentScheduleModal';

export function ClinicDiagnosticsPage() {
  const { clinic } = useClinicContext();
  const clinicId = clinic?.id || '';
  const queryClient = useQueryClient();

  const [activeTab, setActiveTab] = useState<'equipments' | 'services' | 'appointments'>('equipments');

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const params = new URLSearchParams(window.location.search);
      const tabParam = params.get('tab');
      if (tabParam === 'appointments' || tabParam === 'services' || tabParam === 'equipments') {
        setActiveTab(tabParam);
      }
    }
  }, []);

  // Modal states
  const [isEquipmentModalOpen, setIsEquipmentModalOpen] = useState(false);
  const [isServiceModalOpen, setIsServiceModalOpen] = useState(false);
  const [scheduleEquipment, setScheduleEquipment] = useState<DiagnosticEquipment | null>(null);
  const [deletingEquipment, setDeletingEquipment] = useState<DiagnosticEquipment | null>(null);
  const [deletingService, setDeletingService] = useState<DiagnosticService | null>(null);

  // Form states for new equipment
  const [eqName, setEqName] = useState('');
  const [eqRoom, setEqRoom] = useState('');
  const [eqCategoryId, setEqCategoryId] = useState('');

  // Form states for new service
  const [svcName, setSvcName] = useState('');
  const [svcPrice, setSvcPrice] = useState<number>(100000);
  const [svcDuration, setSvcDuration] = useState<number>(30);
  const [svcRules, setSvcRules] = useState('');
  const [svcEquipmentId, setSvcEquipmentId] = useState('');
  const [svcCategoryId, setSvcCategoryId] = useState('');

  // 1. Fetch categories
  const { data: categoriesData } = useQuery({
    queryKey: ['diagnosticCategories'],
    queryFn: diagnosticApi.getCategories,
  });
  const categories = categoriesData?.data?.categories || [];

  // 2. Fetch clinic equipments
  const { data: equipmentsData, isLoading: isLoadingEquipments } = useQuery({
    queryKey: ['clinicEquipments', clinicId],
    queryFn: () => diagnosticApi.getEquipments(clinicId),
    enabled: !!clinicId,
  });
  const equipments = equipmentsData?.data?.equipments || [];

  // 3. Fetch clinic services
  const { data: servicesData, isLoading: isLoadingServices } = useQuery({
    queryKey: ['clinicServices', clinicId],
    queryFn: () => diagnosticApi.getClinicServices(clinicId),
    enabled: !!clinicId,
  });
  const services = servicesData?.data?.services || [];

  // 4. Fetch clinic diagnostic appointments
  const { data: appointmentsData, isLoading: isLoadingAppointments } = useQuery({
    queryKey: ['clinicDiagnosticAppointments', clinicId],
    queryFn: () => diagnosticApi.getClinicAppointments(clinicId),
    enabled: !!clinicId,
  });
  const appointments = appointmentsData?.data?.appointments || [];

  // Mutations
  const createEquipmentMutation = useMutation({
    mutationFn: () =>
      diagnosticApi.createEquipment(clinicId, {
        category_id: eqCategoryId,
        name: eqName,
        room_number: eqRoom,
        is_active: true,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clinicEquipments', clinicId] });
      setIsEquipmentModalOpen(false);
      setEqName('');
      setEqRoom('');
      setEqCategoryId('');
    },
  });

  const deleteEquipmentMutation = useMutation({
    mutationFn: (id: string) => diagnosticApi.deleteEquipment(clinicId, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clinicEquipments', clinicId] });
      queryClient.invalidateQueries({ queryKey: ['clinicServices', clinicId] });
      setDeletingEquipment(null);
      toast.success('Apparat muvaffaqiyatli o‘chirildi');
    },
    onError: (err: any) => {
      const msg = err?.response?.data?.message || err?.message || 'Uskunani o‘chirishda xatolik yuz berdi';
      toast.error(msg);
    },
  });

  const createServiceMutation = useMutation({
    mutationFn: () =>
      diagnosticApi.createClinicService(clinicId, {
        category_id: svcCategoryId,
        equipment_id: svcEquipmentId,
        name: svcName,
        price: Number(svcPrice),
        duration_minutes: Number(svcDuration),
        preparation_rules: svcRules,
        is_active: true,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clinicServices', clinicId] });
      setIsServiceModalOpen(false);
      setSvcName('');
      setSvcPrice(100000);
      setSvcDuration(30);
      setSvcRules('');
      setSvcEquipmentId('');
      setSvcCategoryId('');
    },
  });

  const deleteServiceMutation = useMutation({
    mutationFn: (id: string) => diagnosticApi.deleteClinicService(clinicId, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clinicServices', clinicId] });
      setDeletingService(null);
      toast.success('Xizmat muvaffaqiyatli o‘chirildi');
    },
    onError: (err: any) => {
      const msg = err?.response?.data?.message || err?.message || 'Xizmatni o‘chirishda xatolik yuz berdi';
      toast.error(msg);
    },
  });

  // Appointment actions
  const completeAppointmentMutation = useMutation({
    mutationFn: (id: string) => diagnosticApi.completeClinicAppointment(clinicId, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clinicDiagnosticAppointments', clinicId] });
    },
  });

  const noShowAppointmentMutation = useMutation({
    mutationFn: (id: string) => diagnosticApi.noShowClinicAppointment(clinicId, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clinicDiagnosticAppointments', clinicId] });
    },
  });

  const cancelAppointmentMutation = useMutation({
    mutationFn: (id: string) => diagnosticApi.cancelClinicAppointment(clinicId, id, 'Klinika tomonidan bekor qilindi'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clinicDiagnosticAppointments', clinicId] });
    },
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-xl sm:text-2xl font-bold text-slate-900 tracking-tight">
            Diagnostika Moduli Boshqaruvi
          </h1>
          <p className="text-xs sm:text-sm text-slate-500 mt-1">
            MRT, MSKT, UZI apparatlari, tekshiruv xizmatlari va kelib tushgan bronlar
          </p>
        </div>

        {activeTab === 'equipments' && (
          <Button onClick={() => setIsEquipmentModalOpen(true)} className="gap-2 shadow-sm" size="sm">
            <Plus className="w-4 h-4" />
            Yangi apparat qo‘shish
          </Button>
        )}
        {activeTab === 'services' && (
          <Button onClick={() => setIsServiceModalOpen(true)} className="gap-2 shadow-sm" size="sm">
            <Plus className="w-4 h-4" />
            Diagnostika xizmati qo‘shish
          </Button>
        )}
      </div>

      {/* Tabs */}
      <div className="flex gap-2 border-b border-slate-200 pb-2">
        <button
          onClick={() => setActiveTab('equipments')}
          className={cn(
            'flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold transition-all',
            activeTab === 'equipments'
              ? 'bg-brand-blue text-white shadow-xs'
              : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
          )}
        >
          <Cpu className="w-4 h-4" />
          Uskunalar ({equipments.length})
        </button>

        <button
          onClick={() => setActiveTab('services')}
          className={cn(
            'flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold transition-all',
            activeTab === 'services'
              ? 'bg-brand-blue text-white shadow-xs'
              : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
          )}
        >
          <Layers className="w-4 h-4" />
          Xizmatlar ({services.length})
        </button>

        <button
          onClick={() => setActiveTab('appointments')}
          className={cn(
            'flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold transition-all',
            activeTab === 'appointments'
              ? 'bg-brand-blue text-white shadow-xs'
              : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
          )}
        >
          <Calendar className="w-4 h-4" />
          Bronlar ({appointments.length})
        </button>
      </div>

      {/* Tab 1: Equipments */}
      {activeTab === 'equipments' && (
        <div className="space-y-4">
          {isLoadingEquipments ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {[...Array(4)].map((_, i) => (
                <div key={i} className="h-32 bg-slate-100 rounded-2xl animate-pulse" />
              ))}
            </div>
          ) : equipments.length === 0 ? (
            <div className="p-10 rounded-2xl bg-white border border-slate-200 text-center space-y-3">
              <Cpu className="w-10 h-10 text-slate-300 mx-auto" />
              <p className="text-sm font-semibold text-slate-700">Hali hech qanday diagnostika apparati qo‘shilmagan</p>
              <Button size="sm" onClick={() => setIsEquipmentModalOpen(true)} className="gap-2">
                <Plus className="w-4 h-4" />
                Birinchi apparatni qo‘shish
              </Button>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {equipments.map((eq) => (
                <div
                  key={eq.id}
                  className="bg-white rounded-2xl border border-slate-200 p-5 flex items-start justify-between gap-4 shadow-xs"
                >
                  <div className="space-y-2">
                    <span className="inline-flex px-2 py-0.5 rounded text-[11px] font-bold bg-blue-50 text-brand-blue border border-blue-100">
                      {eq.category?.name || 'Kategoriya'}
                    </span>
                    <h3 className="text-base font-bold text-slate-900">{eq.name}</h3>
                    {eq.room_number && (
                      <p className="text-xs text-slate-500">Xona: <strong>{eq.room_number}</strong></p>
                    )}
                    <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-200">
                      <Check className="w-3 h-3" /> Faol
                    </span>
                  </div>

                  <div className="flex items-center gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      className="text-brand-blue border-blue-200 hover:bg-blue-50 gap-1.5"
                      onClick={() => setScheduleEquipment(eq)}
                      title="Ish jadvalini sozlash"
                    >
                      <Clock className="w-4 h-4" />
                      <span className="hidden sm:inline">Ish jadvali</span>
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      className="text-rose-600 border-rose-200 hover:bg-rose-50"
                      onClick={() => setDeletingEquipment(eq)}
                      disabled={deleteEquipmentMutation.isPending}
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Tab 2: Services */}
      {activeTab === 'services' && (
        <div className="space-y-4">
          {isLoadingServices ? (
            <div className="space-y-3">
              {[...Array(4)].map((_, i) => (
                <div key={i} className="h-20 bg-slate-100 rounded-2xl animate-pulse" />
              ))}
            </div>
          ) : services.length === 0 ? (
            <div className="p-10 rounded-2xl bg-white border border-slate-200 text-center space-y-3">
              <Layers className="w-10 h-10 text-slate-300 mx-auto" />
              <p className="text-sm font-semibold text-slate-700">Diagnostika xizmatlari mavjud emas</p>
              <Button size="sm" onClick={() => setIsServiceModalOpen(true)} className="gap-2">
                <Plus className="w-4 h-4" />
                Xizmat qo‘shish
              </Button>
            </div>
          ) : (
            <div className="bg-white rounded-2xl border border-slate-200 overflow-hidden shadow-xs">
              <div className="divide-y divide-slate-100">
                {services.map((svc) => (
                  <div key={svc.id} className="p-4 sm:p-5 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <span className="text-xs font-bold text-brand-blue bg-blue-50 px-2 py-0.5 rounded">
                          {svc.category?.name}
                        </span>
                        <h4 className="text-sm font-bold text-slate-900">{svc.name}</h4>
                      </div>
                      <div className="text-xs text-slate-500 flex items-center gap-3">
                        <span>Apparat: <strong>{svc.equipment?.name}</strong></span>
                        <span>•</span>
                        <span>Vaqt: <strong>{svc.duration_minutes} daqiqa</strong></span>
                      </div>
                      {svc.preparation_rules && (
                        <p className="text-xs text-amber-800 bg-amber-50 p-2 rounded-lg mt-1 border border-amber-200/60 max-w-xl">
                          <strong>Tayyorgarlik:</strong> {svc.preparation_rules}
                        </p>
                      )}
                    </div>

                    <div className="flex items-center justify-between sm:justify-end gap-4">
                      <span className="text-base font-extrabold text-slate-900">
                        {formatCurrency(svc.price)}
                      </span>
                      <Button
                        variant="outline"
                        size="sm"
                        className="text-rose-600 border-rose-200 hover:bg-rose-50"
                        onClick={() => setDeletingService(svc)}
                        disabled={deleteServiceMutation.isPending}
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Tab 3: Appointments */}
      {activeTab === 'appointments' && (
        <div className="space-y-4">
          {isLoadingAppointments ? (
            <div className="space-y-3">
              {[...Array(4)].map((_, i) => (
                <div key={i} className="h-24 bg-slate-100 rounded-2xl animate-pulse" />
              ))}
            </div>
          ) : appointments.length === 0 ? (
            <div className="p-10 rounded-2xl bg-white border border-slate-200 text-center space-y-2">
              <Calendar className="w-10 h-10 text-slate-300 mx-auto" />
              <p className="text-sm font-semibold text-slate-700">Hozircha diagnostika bronlari kelib tushmagan</p>
            </div>
          ) : (
            <div className="bg-white rounded-2xl border border-slate-200 overflow-hidden shadow-xs divide-y divide-slate-100">
              {appointments.map((appt) => (
                <div key={appt.id} className="p-4 sm:p-5 flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                  <div className="space-y-1.5">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-bold text-brand-blue bg-blue-50 px-2 py-0.5 rounded">
                        {appt.service?.name}
                      </span>
                      <span className="text-xs text-slate-400">({appt.equipment?.name})</span>
                    </div>

                    <div className="text-xs text-slate-700 font-medium">
                      Bemor: <strong>{appt.patient?.first_name} {appt.patient?.last_name}</strong> ({appt.patient?.phone})
                    </div>

                    <div className="text-xs text-slate-500 flex items-center gap-2">
                      <Clock className="w-3.5 h-3.5 text-slate-400" />
                      <span>{appt.appointment_date} ({appt.start_time} - {appt.end_time})</span>
                      <span>•</span>
                      <span className="font-bold text-slate-900">{formatCurrency(appt.price)}</span>
                    </div>
                  </div>

                  {/* Actions for clinic admin */}
                  <div className="flex items-center gap-2">
                    {appt.status === 'confirmed' ? (
                      <>
                        <Button
                          size="sm"
                          className="bg-emerald-600 hover:bg-emerald-700 text-white gap-1"
                          onClick={() => completeAppointmentMutation.mutate(appt.id)}
                        >
                          <CheckCircle className="w-3.5 h-3.5" />
                          Yakunlash
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          className="text-amber-700 border-amber-200 hover:bg-amber-50"
                          onClick={() => noShowAppointmentMutation.mutate(appt.id)}
                        >
                          Kelmadi
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          className="text-rose-600 border-rose-200 hover:bg-rose-50"
                          onClick={() => cancelAppointmentMutation.mutate(appt.id)}
                        >
                          Bekor qilish
                        </Button>
                      </>
                    ) : (
                      <span className="text-xs font-bold text-slate-500 uppercase px-3 py-1 bg-slate-100 rounded-full">
                        {appt.status}
                      </span>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Modal: Add Equipment */}
      <Modal
        isOpen={isEquipmentModalOpen}
        onClose={() => setIsEquipmentModalOpen(false)}
        title="Yangi diagnostika apparati qo‘shish"
      >
        <div className="space-y-4">
          <div>
            <label className="text-xs font-bold text-slate-700 block mb-1">Diagnostika kategoriyasi</label>
            <select
              value={eqCategoryId}
              onChange={(e) => setEqCategoryId(e.target.value)}
              className="w-full px-3 py-2 rounded-xl border border-slate-200 text-sm bg-white"
            >
              <option value="">Tanlang...</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="text-xs font-bold text-slate-700 block mb-1">Apparat modeli / Nomi</label>
            <input
              type="text"
              value={eqName}
              onChange={(e) => setEqName(e.target.value)}
              placeholder="Masalan: Siemens Magnetom Avanto 1.5 Tesla"
              className="w-full px-3 py-2 rounded-xl border border-slate-200 text-sm"
            />
          </div>

          <div>
            <label className="text-xs font-bold text-slate-700 block mb-1">Xona yoki Bino raqami</label>
            <input
              type="text"
              value={eqRoom}
              onChange={(e) => setEqRoom(e.target.value)}
              placeholder="Masalan: 104-xona, 1-qavat"
              className="w-full px-3 py-2 rounded-xl border border-slate-200 text-sm"
            />
          </div>

          <div className="flex justify-end gap-2 pt-3">
            <Button variant="outline" onClick={() => setIsEquipmentModalOpen(false)}>
              Bekor qilish
            </Button>
            <Button
              onClick={() => createEquipmentMutation.mutate()}
              disabled={!eqName || !eqCategoryId || createEquipmentMutation.isPending}
              isLoading={createEquipmentMutation.isPending}
            >
              Saqlash
            </Button>
          </div>
        </div>
      </Modal>

      {/* Modal: Add Service */}
      <Modal
        isOpen={isServiceModalOpen}
        onClose={() => setIsServiceModalOpen(false)}
        title="Yangi diagnostika xizmati qo‘shish"
      >
        <div className="space-y-4">
          <div>
            <label className="text-xs font-bold text-slate-700 block mb-1">Diagnostika kategoriyasi</label>
            <select
              value={svcCategoryId}
              onChange={(e) => setSvcCategoryId(e.target.value)}
              className="w-full px-3 py-2 rounded-xl border border-slate-200 text-sm bg-white"
            >
              <option value="">Tanlang...</option>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="text-xs font-bold text-slate-700 block mb-1">Biriktiriladigan apparat</label>
            <select
              value={svcEquipmentId}
              onChange={(e) => setSvcEquipmentId(e.target.value)}
              className="w-full px-3 py-2 rounded-xl border border-slate-200 text-sm bg-white"
            >
              <option value="">Tanlang...</option>
              {equipments.map((eq) => (
                <option key={eq.id} value={eq.id}>
                  {eq.name} ({eq.room_number})
                </option>
              ))}
            </select>
          </div>

          <div>
            <label className="text-xs font-bold text-slate-700 block mb-1">Xizmat nomi</label>
            <input
              type="text"
              value={svcName}
              onChange={(e) => setSvcName(e.target.value)}
              placeholder="Masalan: Bosh miya va qon tomirlari MRT tekshiruvi"
              className="w-full px-3 py-2 rounded-xl border border-slate-200 text-sm"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-xs font-bold text-slate-700 block mb-1">Narxi (so‘m)</label>
              <input
                type="number"
                value={svcPrice}
                onChange={(e) => setSvcPrice(Number(e.target.value))}
                className="w-full px-3 py-2 rounded-xl border border-slate-200 text-sm"
              />
            </div>
            <div>
              <label className="text-xs font-bold text-slate-700 block mb-1">Davomiyligi (daqiqa)</label>
              <input
                type="number"
                value={svcDuration}
                onChange={(e) => setSvcDuration(Number(e.target.value))}
                className="w-full px-3 py-2 rounded-xl border border-slate-200 text-sm"
              />
            </div>
          </div>

          <div>
            <label className="text-xs font-bold text-slate-700 block mb-1">
              Bemorlar uchun tayyorgarlik qoidalari (Majburiy ogohlantirish)
            </label>
            <textarea
              value={svcRules}
              onChange={(e) => setSvcRules(e.target.value)}
              rows={3}
              placeholder="Masalan: Tekshiruvdan 6 soat oldin ovqatlanmaslik, metall taqinchoqlarni yechish..."
              className="w-full px-3 py-2 rounded-xl border border-slate-200 text-sm"
            />
          </div>

          <div className="flex justify-end gap-2 pt-3">
            <Button variant="outline" onClick={() => setIsServiceModalOpen(false)}>
              Bekor qilish
            </Button>
            <Button
              onClick={() => createServiceMutation.mutate()}
              disabled={!svcName || !svcEquipmentId || !svcCategoryId || createServiceMutation.isPending}
              isLoading={createServiceMutation.isPending}
            >
              Saqlash
            </Button>
          </div>
        </div>
      </Modal>

      {/* Equipment Weekly Schedule Modal */}
      <EquipmentScheduleModal
        isOpen={!!scheduleEquipment}
        onClose={() => setScheduleEquipment(null)}
        equipment={scheduleEquipment}
        clinicId={clinicId}
      />

      {/* Delete Equipment Confirmation Modal */}
      <Modal
        isOpen={!!deletingEquipment}
        onClose={() => setDeletingEquipment(null)}
        title="Apparatni o‘chirish"
      >
        <div className="space-y-4">
          <div className="flex items-start gap-3 p-3 bg-rose-50 border border-rose-200 rounded-xl text-rose-800 text-sm">
            <AlertCircle className="w-5 h-5 flex-shrink-0 mt-0.5 text-rose-600" />
            <div>
              <p className="font-semibold">
                &ldquo;{deletingEquipment?.name}&rdquo; apparatini o‘chirmoqchimisiz?
              </p>
              <p className="text-xs text-rose-700 mt-1">
                Ushbu apparatga biriktirilgan xizmatlar ham o‘chiriladi. Agar apparatda o‘tmishdagi bronlar mavjud bo‘lsa, u arxivlanadi (faolsizlantiriladi).
              </p>
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button
              variant="outline"
              onClick={() => setDeletingEquipment(null)}
              disabled={deleteEquipmentMutation.isPending}
            >
              Bekor qilish
            </Button>
            <Button
              className="bg-rose-600 hover:bg-rose-700 text-white"
              onClick={() => deletingEquipment && deleteEquipmentMutation.mutate(deletingEquipment.id)}
              isLoading={deleteEquipmentMutation.isPending}
            >
              O‘chirish
            </Button>
          </div>
        </div>
      </Modal>

      {/* Delete Service Confirmation Modal */}
      <Modal
        isOpen={!!deletingService}
        onClose={() => setDeletingService(null)}
        title="Xizmatni o‘chirish"
      >
        <div className="space-y-4">
          <div className="flex items-start gap-3 p-3 bg-rose-50 border border-rose-200 rounded-xl text-rose-800 text-sm">
            <AlertCircle className="w-5 h-5 flex-shrink-0 mt-0.5 text-rose-600" />
            <div>
              <p className="font-semibold">
                &ldquo;{deletingService?.name}&rdquo; diagnostika xizmatini o‘chirmoqchimisiz?
              </p>
              <p className="text-xs text-rose-700 mt-1">
                Ushbu xizmat bemorlar qidiruvidan olib tashlanadi.
              </p>
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button
              variant="outline"
              onClick={() => setDeletingService(null)}
              disabled={deleteServiceMutation.isPending}
            >
              Bekor qilish
            </Button>
            <Button
              className="bg-rose-600 hover:bg-rose-700 text-white"
              onClick={() => deletingService && deleteServiceMutation.mutate(deletingService.id)}
              isLoading={deleteServiceMutation.isPending}
            >
              O‘chirish
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}

