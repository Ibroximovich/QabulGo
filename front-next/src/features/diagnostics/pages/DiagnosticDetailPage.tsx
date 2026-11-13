'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { diagnosticApi } from '../api/diagnosticApi';
import { DiagnosticAvailabilitySlot } from '../types';
import { EquipmentSlotPicker } from '../components/EquipmentSlotPicker';
import { PreparationRulesModal } from '../components/PreparationRulesModal';
import { formatCurrency } from '@/lib/utils/formatters';
import { useAuth } from '@/providers/AuthProvider';
import {
  Building2,
  MapPin,
  Clock,
  Cpu,
  AlertTriangle,
  ArrowLeft,
  ShieldCheck,
  Phone,
  CheckCircle2,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';

interface DiagnosticDetailPageProps {
  serviceId: string;
}

export function DiagnosticDetailPage({ serviceId }: DiagnosticDetailPageProps) {
  const router = useRouter();
  const { user } = useAuth();
  const queryClient = useQueryClient();

  const [selectedSlot, setSelectedSlot] = useState<DiagnosticAvailabilitySlot | null>(null);
  const [selectedDate, setSelectedDate] = useState<string>('');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [successBookingId, setSuccessBookingId] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // 1. Fetch service details
  const {
    data: serviceData,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ['diagnosticService', serviceId],
    queryFn: () => diagnosticApi.getService(serviceId),
    enabled: !!serviceId,
  });

  const service = serviceData?.data;

  // 2. Booking mutation
  const bookMutation = useMutation({
    mutationFn: () => {
      if (!service || !selectedSlot || !selectedDate) {
        throw new Error('Ma’lumotlar to‘liq emas');
      }
      return diagnosticApi.bookAppointment({
        clinic_id: service.clinic_id,
        service_id: service.id,
        equipment_id: service.equipment_id,
        date: selectedDate,
        start_time: selectedSlot.start_time,
        end_time: selectedSlot.end_time,
        preparation_acknowledged: true,
      });
    },
    onSuccess: (res) => {
      setIsModalOpen(false);
      queryClient.invalidateQueries({ queryKey: ['equipmentAvailability'] });
      queryClient.invalidateQueries({ queryKey: ['myDiagnosticAppointments'] });
      setSuccessBookingId(res.data?.id || 'ok');
    },
    onError: (err: unknown) => {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      setErrorMessage(
        axiosErr.response?.data?.message || 'Qabulga yozilishda xatolik yuz berdi. Iltimos, qaytadan urinib ko‘ring.'
      );
    },
  });

  const handleSlotSelect = (slot: DiagnosticAvailabilitySlot, date: string) => {
    setSelectedSlot(slot);
    setSelectedDate(date);
    setErrorMessage(null);

    // If user not logged in, redirect to login with callback
    if (!user) {
      router.push(`/login?redirect=/diagnostics/services/${serviceId}`);
      return;
    }

    setIsModalOpen(true);
  };

  if (isLoading) {
    return (
      <div className="max-w-5xl mx-auto px-4 sm:px-6 py-8 space-y-6">
        <div className="h-6 w-32 bg-slate-200 rounded animate-pulse" />
        <div className="h-48 bg-slate-100 rounded-3xl animate-pulse" />
        <div className="h-64 bg-slate-100 rounded-3xl animate-pulse" />
      </div>
    );
  }

  if (isError || !service) {
    return (
      <div className="max-w-md mx-auto my-16 p-8 text-center bg-white rounded-3xl border border-slate-200 space-y-4">
        <AlertTriangle className="w-10 h-10 text-amber-500 mx-auto" />
        <h2 className="text-lg font-bold text-slate-800">Diagnostika xizmati topilmadi</h2>
        <Link href="/diagnostics">
          <Button variant="outline" className="gap-2">
            <ArrowLeft className="w-4 h-4" />
            Xizmatlar ro‘yxatiga qaytish
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto px-4 sm:px-6 py-8 space-y-8">
      {/* Back button */}
      <Link
        href="/diagnostics"
        className="inline-flex items-center gap-1.5 text-xs font-semibold text-slate-500 hover:text-brand-blue transition-colors"
      >
        <ArrowLeft className="w-4 h-4" />
        Barcha diagnostika xizmatlariga qaytish
      </Link>

      {/* Success Banner */}
      {successBookingId && (
        <div className="p-6 rounded-3xl bg-emerald-50 border border-emerald-200 text-emerald-900 space-y-3">
          <div className="flex items-center gap-2.5 font-bold text-base text-emerald-800">
            <CheckCircle2 className="w-6 h-6 text-emerald-600 flex-shrink-0" />
            <span>Diagnostika qabuliga muvaffaqiyatli yozildingiz!</span>
          </div>
          <p className="text-xs text-emerald-800/90 leading-relaxed pl-8">
            Qabul vaqti: <strong>{selectedDate} ({selectedSlot?.start_time} - {selectedSlot?.end_time})</strong>.
            Iltimos, tekshiruvga tayyorgarlik qoidalariga amal qilib belgilangan vaqtdan 10 daqiqa oldin keling.
          </p>
          <div className="pl-8 pt-2">
            <Link href="/diagnostics/appointments">
              <Button size="sm" className="bg-emerald-600 hover:bg-emerald-700 text-white gap-2">
                Mening qabullarim sahifasiga o‘tish
              </Button>
            </Link>
          </div>
        </div>
      )}

      {/* Error Banner */}
      {errorMessage && (
        <div className="p-4 rounded-2xl bg-rose-50 border border-rose-200 text-rose-800 text-xs flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-rose-600 flex-shrink-0" />
            <span>{errorMessage}</span>
          </div>
          <button onClick={() => setErrorMessage(null)} className="text-rose-500 hover:text-rose-800 font-bold">
            ✕
          </button>
        </div>
      )}

      {/* Service Info Card */}
      <div className="bg-white rounded-3xl border border-slate-200/80 p-6 sm:p-8 shadow-sm space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4">
          <div className="space-y-2 max-w-2xl">
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold bg-blue-50 text-brand-blue border border-blue-100">
              {service.category?.name || 'Diagnostika'}
            </div>
            <h1 className="text-xl sm:text-2xl font-extrabold text-slate-900 tracking-tight">
              {service.name}
            </h1>
            {service.clinic && (
              <div className="flex items-center gap-2 text-xs text-slate-600 pt-1">
                <Building2 className="w-4 h-4 text-slate-400" />
                <span className="font-semibold text-slate-800">{service.clinic.name}</span>
                <span>•</span>
                <MapPin className="w-4 h-4 text-slate-400" />
                <span>{service.clinic.district}, {service.clinic.address}</span>
              </div>
            )}
          </div>

          <div className="p-4 rounded-2xl bg-slate-50 border border-slate-100 text-right sm:min-w-[180px]">
            <span className="text-xs text-slate-400 block font-medium">Tekshiruv narxi</span>
            <span className="text-xl font-black text-slate-900">{formatCurrency(service.price)}</span>
            <div className="flex items-center justify-end gap-1 text-[11px] text-slate-500 font-medium mt-1">
              <Clock className="w-3 h-3 text-slate-400" />
              <span>Davomiyligi: {service.duration_minutes} daqiqa</span>
            </div>
          </div>
        </div>

        {/* Equipment Specification Details */}
        {service.equipment && (
          <div className="p-4 rounded-2xl bg-blue-50/40 border border-blue-100/80 flex items-start sm:items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-brand-blue/10 flex items-center justify-center text-brand-blue flex-shrink-0">
                <Cpu className="w-5 h-5" />
              </div>
              <div>
                <span className="text-[11px] font-semibold text-brand-blue uppercase tracking-wider block">
                  Tibbiy Uskuna / Xona
                </span>
                <span className="text-sm font-bold text-slate-900">{service.equipment.name}</span>
                {service.equipment.room_number && (
                  <span className="text-xs text-slate-500 ml-2">({service.equipment.room_number})</span>
                )}
              </div>
            </div>
            <div className="hidden sm:flex items-center gap-1.5 text-xs text-emerald-700 font-semibold bg-emerald-50 px-2.5 py-1 rounded-full border border-emerald-200">
              <ShieldCheck className="w-3.5 h-3.5" />
              Kalibratsiyalangan
            </div>
          </div>
        )}

        {/* Preparation Rules */}
        {service.preparation_rules && (
          <div className="p-4 sm:p-5 rounded-2xl bg-amber-50/80 border border-amber-200 text-amber-900 space-y-2">
            <div className="flex items-center gap-2 font-bold text-xs uppercase tracking-wider text-amber-800">
              <AlertTriangle className="w-4 h-4 text-amber-600" />
              <span>Bemorlar uchun tayyorgarlik talablari</span>
            </div>
            <p className="text-xs sm:text-sm leading-relaxed text-amber-950/90 whitespace-pre-line pl-6">
              {service.preparation_rules}
            </p>
          </div>
        )}
      </div>

      {/* Slot Picker Component */}
      <div className="bg-white rounded-3xl border border-slate-200/80 p-6 sm:p-8 shadow-sm">
        <h2 className="text-base sm:text-lg font-bold text-slate-900 mb-6">
          Qabulga yozilish uchun vaqtni tanlang
        </h2>

        <EquipmentSlotPicker
          equipmentId={service.equipment_id}
          onSelectSlot={handleSlotSelect}
          selectedSlot={selectedSlot}
          selectedDate={selectedDate}
        />
      </div>

      {/* Confirmation & Rules Modal */}
      {selectedSlot && (
        <PreparationRulesModal
          isOpen={isModalOpen}
          onClose={() => setIsModalOpen(false)}
          onConfirm={() => bookMutation.mutate()}
          isLoading={bookMutation.isPending}
          serviceName={service.name}
          clinicName={service.clinic?.name}
          date={selectedDate}
          startTime={selectedSlot.start_time}
          endTime={selectedSlot.end_time}
          price={service.price}
          preparationRules={service.preparation_rules}
        />
      )}
    </div>
  );
}

