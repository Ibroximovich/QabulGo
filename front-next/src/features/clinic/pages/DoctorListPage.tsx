'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useClinicContext } from '@/components/layout/clinic/ClinicLayout';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  Users,
  Plus,
  Search,
  Edit,
  Trash2,
  Calendar,
  CheckCircle2,
  XCircle,
  AlertTriangle,
} from 'lucide-react';
import { ClinicDetail, ClinicDoctorItem } from '../types';
import { clinicApi } from '../api/clinicApi';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';
import { formatCurrency } from '@/lib/utils/formatters';
import { getMediaUrl } from '@/lib/utils/media';

interface ClinicContextType {
  clinic?: ClinicDetail;
  isLoadingClinic: boolean;
}

export function DoctorListPage() {
  const { clinic, isLoadingClinic } = useClinicContext();
  const router = useRouter();
  const queryClient = useQueryClient();

  const [searchTerm, setSearchTerm] = useState('');
  const [selectedSpecialty, setSelectedSpecialty] = useState<string>('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'inactive'>('all');
  const [selectedDoctorForDeactivate, setSelectedDoctorForDeactivate] =
    useState<ClinicDoctorItem | null>(null);

  // Fetch Specialties
  const { data: specialtiesRes } = useQuery({
    queryKey: ['specialties'],
    queryFn: async () => {
      const res = await clinicApi.getSpecialties();
      return res.data?.data || [];
    },
  });
  const specialties = specialtiesRes || [];

  // Fetch Doctors
  const { data: doctorsData, isLoading: isLoadingDoctors } = useQuery({
    queryKey: ['clinicDoctors', clinic?.id, searchTerm, selectedSpecialty, statusFilter],
    queryFn: async () => {
      if (!clinic?.id) return [];
      const res = await clinicApi.listDoctors(clinic.id, {
        search: searchTerm || undefined,
        specialty_id: selectedSpecialty || undefined,
        is_active:
          statusFilter === 'all' ? undefined : statusFilter === 'active',
      });
      return res.data?.data || [];
    },
    enabled: Boolean(clinic?.id),
  });

  // Deactivate doctor mutation
  const deactivateMutation = useMutation({
    mutationFn: async (doctorId: string) => {
      if (!clinic?.id) return;
      await clinicApi.deactivateDoctor(clinic.id, doctorId);
    },
    onSuccess: () => {
      toast.success('Shifokor faoliyati muvaffaqiyatli to‘xtatildi');
      setSelectedDoctorForDeactivate(null);
      queryClient.invalidateQueries({ queryKey: ['clinicDoctors'] });
      queryClient.invalidateQueries({ queryKey: ['myClinic'] });
    },
    onError: (err: unknown) => {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      toast.error(
        axiosErr.response?.data?.message || 'Shifokor statusini o‘zgartirishda xatolik'
      );
    },
  });

  if (isLoadingClinic) {
    return <div className="p-8 text-center text-slate-500">Yuklanmoqda...</div>;
  }

  if (clinic?.status !== 'approved') {
    return (
      <div className="bg-white rounded-2xl border border-slate-200 p-8 text-center max-w-lg mx-auto shadow-xs">
        <AlertTriangle className="w-12 h-12 text-amber-500 mx-auto mb-3" />
        <h3 className="text-base font-bold text-slate-900 mb-1">
          Klinika hali tasdiqlanmagan
        </h3>
        <p className="text-xs text-slate-500 mb-4">
          Shifokorlarni qo‘shish va boshqarish faqat klinika superadmin tomonidan
          tasdiqlangandan so‘ng ruxsat etiladi.
        </p>
        <Button variant="outline" onClick={() => router.push('/clinic/dashboard')}>
          Boshqaruv paneliga qaytish
        </Button>
      </div>
    );
  }

  const doctors = doctorsData || [];

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 className="text-xl font-extrabold text-slate-900">Shifokorlar bazasi</h2>
          <p className="text-xs text-slate-500 mt-0.5">
            Klinikada faoliyat yurituvchi mutaxassislar ro‘yxati va narxlar
          </p>
        </div>

        <Button
          variant="primary"
          onClick={() => router.push('/clinic/doctors/new')}
          className="bg-[#0866F5] hover:bg-blue-600 gap-1.5 shadow-xs"
        >
          <Plus className="w-4 h-4" />
          Yangi shifokor qo‘shish
        </Button>
      </div>

      {/* Filter and Search Bar */}
      <div className="bg-white rounded-2xl border border-slate-200 p-4 shadow-xs flex flex-col md:flex-row items-center gap-3">
        <div className="relative flex-1 w-full">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Shifokor ismi yoki familiyasi bo‘yicha qidirish..."
            className="w-full pl-10 pr-4 py-2 text-xs rounded-xl border border-slate-300 focus:outline-hidden focus:border-[#0866F5] focus:ring-2 focus:ring-blue-100"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>

        <div className="flex items-center gap-2 w-full md:w-auto">
          {/* Specialty filter */}
          <select
            className="h-9 px-3 text-xs rounded-xl border border-slate-300 bg-white focus:outline-hidden focus:border-[#0866F5] focus:ring-2 focus:ring-blue-100"
            value={selectedSpecialty}
            onChange={(e) => setSelectedSpecialty(e.target.value)}
          >
            <option value="">Barcha mutaxassisliklar</option>
            {specialties.map((spec) => (
              <option key={spec.id} value={spec.id}>
                {spec.name}
              </option>
            ))}
          </select>

          {/* Status filter */}
          <select
            className="h-9 px-3 text-xs rounded-xl border border-slate-300 bg-white focus:outline-hidden focus:border-[#0866F5] focus:ring-2 focus:ring-blue-100"
            value={statusFilter}
            onChange={(e) =>
              setStatusFilter(e.target.value as 'all' | 'active' | 'inactive')
            }
          >
            <option value="all">Barcha holatlar</option>
            <option value="active">Faqat faol</option>
            <option value="inactive">Nofaol</option>
          </select>
        </div>
      </div>

      {/* Doctor Cards / Grid */}
      {isLoadingDoctors ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-44 bg-slate-200 animate-pulse rounded-2xl" />
          ))}
        </div>
      ) : doctors.length === 0 ? (
        <div className="bg-white rounded-2xl border border-slate-200 p-8 text-center shadow-xs">
          <Users className="w-12 h-12 text-slate-300 mx-auto mb-3" />
          <h3 className="text-sm font-bold text-slate-800 mb-1">
            Hech qanday shifokor topilmadi
          </h3>
          <p className="text-xs text-slate-500 mb-4">
            Qidiruv mezonlarini o‘zgartiring yoki yangi shifokor qo‘shing.
          </p>
          <Button
            size="sm"
            variant="outline"
            onClick={() => router.push('/clinic/doctors/new')}
          >
            Shifokor qo‘shish
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {doctors.map((doc) => (
            <div
              key={doc.id}
              className={`bg-white rounded-2xl border transition-all p-5 shadow-xs flex flex-col justify-between ${
                doc.is_active ? 'border-slate-200' : 'border-slate-200/60 bg-slate-50/50 opacity-75'
              }`}
            >
              <div>
                <div className="flex items-start justify-between gap-3 mb-3">
                  <div className="flex items-center gap-3.5">
                    {doc.photo_url ? (
                      <img
                        src={getMediaUrl(doc.photo_url)}
                        alt={`${doc.first_name} ${doc.last_name}`}
                        className="w-14 h-14 rounded-2xl object-cover object-top border border-slate-200 shadow-2xs shrink-0"
                        style={{ imageRendering: '-webkit-optimize-contrast' }}
                      />
                    ) : (
                      <div className="w-14 h-14 rounded-2xl bg-blue-50 text-[#0866F5] font-bold flex items-center justify-center text-sm border border-blue-100 shrink-0">
                        {doc.first_name[0]}
                        {doc.last_name[0]}
                      </div>
                    )}

                    <div>
                      <h4 className="font-bold text-slate-900 text-sm leading-tight">
                        {doc.first_name} {doc.last_name}
                      </h4>
                      <span className="inline-block px-2 py-0.5 rounded-md text-[11px] font-semibold bg-blue-50 text-blue-700 mt-1">
                        {doc.specialty?.name || 'Mutaxassis'}
                      </span>
                    </div>
                  </div>

                  <div>
                    {doc.is_active ? (
                      <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded-full">
                        <CheckCircle2 className="w-3 h-3" /> Faol
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-slate-500 bg-slate-100 px-2 py-0.5 rounded-full">
                        <XCircle className="w-3 h-3" /> Nofaol
                      </span>
                    )}
                  </div>
                </div>

                <div className="space-y-1.5 text-xs text-slate-600 my-3 pt-2 border-t border-slate-100">
                  <div className="flex items-center justify-between">
                    <span className="text-slate-500">Tajriba:</span>
                    <span className="font-semibold text-slate-800">
                      {doc.experience_years} yil
                    </span>
                  </div>

                  <div className="flex items-center justify-between">
                    <span className="text-slate-500">Qabul narxi:</span>
                    <span className="font-semibold text-emerald-600">
                      {formatCurrency(doc.consultation_price)}
                    </span>
                  </div>

                  <div className="flex items-center justify-between">
                    <span className="text-slate-500">Qabul davomiyligi:</span>
                    <span className="font-semibold text-slate-800 font-mono">
                      {doc.appointment_duration_minutes} daqiqa
                    </span>
                  </div>
                </div>

                {doc.bio && (
                  <p className="text-[11px] text-slate-500 line-clamp-2 italic mb-3">
                    "{doc.bio}"
                  </p>
                )}
              </div>

              {/* Actions */}
              <div className="pt-3 border-t border-slate-100 flex items-center justify-between gap-2">
                <Button
                  size="sm"
                  variant="primary"
                  onClick={() => router.push(`/clinic/doctors/${doc.id || doc.doctor_id}/schedule`)}
                  className="flex-1 gap-1.5 text-xs font-semibold"
                >
                  <Calendar className="w-3.5 h-3.5" />
                  Qabul jadvali
                </Button>

                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => router.push(`/clinic/doctors/${doc.id || doc.doctor_id}/edit`)}
                  className="gap-1 text-xs px-2.5"
                  title="Tahrirlash"
                >
                  <Edit className="w-3.5 h-3.5 text-slate-500" />
                </Button>

                {doc.is_active && (
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => setSelectedDoctorForDeactivate(doc)}
                    className="text-rose-600 hover:bg-rose-50 text-xs px-2.5"
                    title="Faoliyatini to‘xtatish"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </Button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Confirmation Modal for Deactivate */}
      {selectedDoctorForDeactivate && (
        <Modal
          isOpen={Boolean(selectedDoctorForDeactivate)}
          onClose={() => setSelectedDoctorForDeactivate(null)}
          title="Shifokorni faolsizlantirish"
        >
          <div className="space-y-4">
            <p className="text-xs text-slate-600 leading-relaxed">
              Haqiqatan ham{' '}
              <span className="font-bold text-slate-900">
                {selectedDoctorForDeactivate.first_name}{' '}
                {selectedDoctorForDeactivate.last_name}
              </span>{' '}
              ning qabul faoliyatini vaqtincha to‘xtatmoqchimisiz? U klinika qidiruvida
              ko‘rinmay qoladi.
            </p>

            <div className="flex justify-end gap-2 pt-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setSelectedDoctorForDeactivate(null)}
              >
                Bekor qilish
              </Button>
              <Button
                variant="destructive"
                size="sm"
                isLoading={deactivateMutation.isPending}
                onClick={() =>
                  deactivateMutation.mutate(selectedDoctorForDeactivate.doctor_id)
                }
              >
                Faolsizlantirish
              </Button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}
