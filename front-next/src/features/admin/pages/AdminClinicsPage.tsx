'use client';

import { useState, useEffect } from 'react';
import { useSearchParams } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  Building2,
  Search,
  CheckCircle2,
  XCircle,
  Clock4,
  Eye,
  Check,
  Ban,
  Stethoscope,
  CalendarCheck,
  TrendingUp,
  Info,
  Phone,
  Mail,
  FileText,
  User,
} from 'lucide-react';
import { adminApi } from '../api/adminApi';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';
import { formatDate, formatDateTime, formatPrice } from '@/lib/utils/formatters';
import { getMediaUrl } from '@/lib/utils/media';

const STATUS_TABS = [
  { id: '', label: 'Barchasi' },
  { id: 'pending', label: 'Kutilayotganlar' },
  { id: 'approved', label: 'Tasdiqlanganlar' },
  { id: 'rejected', label: 'Rad etilganlar' },
  { id: 'suspended', label: 'To‘xtatilganlar' },
];

export function AdminClinicsPage() {
  const queryClient = useQueryClient();
  const searchParams = useSearchParams();
  const queryClinicId = searchParams.get('clinicId');

  const [selectedStatus, setSelectedStatus] = useState<string>('');
  const [searchTerm, setSearchTerm] = useState<string>('');

  const [detailClinicId, setDetailClinicId] = useState<string | null>(queryClinicId || null);
  const [detailTab, setDetailTab] = useState<'info' | 'doctors' | 'appointments' | 'analytics'>('info');

  useEffect(() => {
    if (queryClinicId) {
      setDetailClinicId(queryClinicId);
    }
  }, [queryClinicId]);

  const [approveClinicId, setApproveClinicId] = useState<string | null>(null);
  const [rejectClinicId, setRejectClinicId] = useState<string | null>(null);
  const [suspendClinicId, setSuspendClinicId] = useState<string | null>(null);
  const [actionReason, setActionReason] = useState<string>('');

  const { data: clinicsResponse, isLoading: isLoadingList } = useQuery({
    queryKey: ['adminClinics', selectedStatus, searchTerm],
    queryFn: async () => {
      const res = await adminApi.listClinics({
        status: selectedStatus || undefined,
        search: searchTerm || undefined,
      });
      return res.data;
    },
  });

  const { data: detailsData, isLoading: isLoadingDetails } = useQuery({
    queryKey: ['adminClinicDetails', detailClinicId],
    queryFn: async () => {
      if (!detailClinicId) return null;
      const res = await adminApi.getClinicDetails(detailClinicId);
      return res.data?.data || null;
    },
    enabled: Boolean(detailClinicId),
  });

  const { data: clinicDoctors, isLoading: isLoadingDoctors } = useQuery({
    queryKey: ['adminClinicDoctors', detailClinicId],
    queryFn: async () => {
      if (!detailClinicId) return [];
      const res = await adminApi.getClinicDoctors(detailClinicId);
      return res.data?.data || [];
    },
    enabled: Boolean(detailClinicId && detailTab === 'doctors'),
  });

  const { data: clinicAppointments, isLoading: isLoadingAppointments } = useQuery({
    queryKey: ['adminClinicAppointments', detailClinicId],
    queryFn: async () => {
      if (!detailClinicId) return [];
      const res = await adminApi.getClinicAppointments(detailClinicId);
      return res.data?.data || [];
    },
    enabled: Boolean(detailClinicId && detailTab === 'appointments'),
  });

  const { data: clinicAnalytics, isLoading: isLoadingAnalytics } = useQuery({
    queryKey: ['adminClinicAnalytics', detailClinicId],
    queryFn: async () => {
      if (!detailClinicId) return null;
      const res = await adminApi.getClinicAnalytics(detailClinicId);
      return res.data?.data || null;
    },
    enabled: Boolean(detailClinicId && detailTab === 'analytics'),
  });

  const approveMutation = useMutation({
    mutationFn: async (id: string) => {
      await adminApi.approveClinic(id);
    },
    onSuccess: () => {
      toast.success('Klinika muvaffaqiyatli tasdiqlandi');
      setApproveClinicId(null);
      queryClient.invalidateQueries({ queryKey: ['adminClinics'] });
      queryClient.invalidateQueries({ queryKey: ['adminClinicDetails'] });
    },
    onError: (err: unknown) => {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      toast.error(axiosErr.response?.data?.message || 'Xatolik yuz berdi');
    },
  });

  const rejectMutation = useMutation({
    mutationFn: async ({ id, reason }: { id: string; reason: string }) => {
      await adminApi.rejectClinic(id, reason);
    },
    onSuccess: () => {
      toast.success('Klinika arizasi rad etildi');
      setRejectClinicId(null);
      setActionReason('');
      queryClient.invalidateQueries({ queryKey: ['adminClinics'] });
      queryClient.invalidateQueries({ queryKey: ['adminClinicDetails'] });
    },
    onError: (err: unknown) => {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      toast.error(axiosErr.response?.data?.message || 'Xatolik yuz berdi');
    },
  });

  const suspendMutation = useMutation({
    mutationFn: async ({ id, reason }: { id: string; reason: string }) => {
      await adminApi.suspendClinic(id, reason);
    },
    onSuccess: () => {
      toast.success('Klinika faoliyati to‘xtatildi (bloklandi)');
      setSuspendClinicId(null);
      setActionReason('');
      queryClient.invalidateQueries({ queryKey: ['adminClinics'] });
      queryClient.invalidateQueries({ queryKey: ['adminClinicDetails'] });
    },
    onError: (err: unknown) => {
      const axiosErr = err as { response?: { data?: { message?: string } } };
      toast.error(axiosErr.response?.data?.message || 'Xatolik yuz berdi');
    },
  });

  const clinics = clinicsResponse?.data || [];

  const renderStatusBadge = (status: string) => {
    switch (status) {
      case 'approved':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
            <CheckCircle2 className="w-3 h-3" />
            Tasdiqlangan
          </span>
        );
      case 'pending':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200">
            <Clock4 className="w-3 h-3" />
            Kutilmoqda
          </span>
        );
      case 'rejected':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-rose-50 text-rose-700 border border-rose-200">
            <XCircle className="w-3 h-3" />
            Rad etilgan
          </span>
        );
      case 'suspended':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-red-100 text-red-800 border border-red-200">
            <Ban className="w-3 h-3" />
            To‘xtatilgan
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-700">
            Qoralama
          </span>
        );
    }
  };

  return (
    <div className="space-y-6 font-sans">
      {/* Header */}
      <div>
        <h2 className="text-xl font-extrabold text-slate-900">Klinikalar boshqaruvi</h2>
        <p className="text-xs text-slate-500 mt-0.5">
          Ro‘yxatdan o‘tgan barcha klinikalarni ko‘rib chiqish, tasdiqlash yoki faoliyatini boshqarish
        </p>
      </div>

      {/* Filter Tabs & Search Bar */}
      <div className="bg-white rounded-2xl border border-slate-200 p-4 shadow-xs space-y-4">
        <div className="flex flex-wrap items-center gap-2 border-b border-slate-100 pb-3">
          {STATUS_TABS.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setSelectedStatus(tab.id)}
              className={`px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
                selectedStatus === tab.id
                  ? 'bg-[#0866F5] text-white shadow-2xs'
                  : 'text-slate-600 hover:bg-slate-100'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        <div className="relative">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Klinika nomi yoki manzili bo‘yicha qidirish..."
            className="w-full pl-10 pr-4 py-2 text-xs rounded-xl border border-slate-300 focus:outline-hidden focus:border-[#0866F5] focus:ring-2 focus:ring-blue-100"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>

      {/* Clinics Table */}
      <div className="bg-white rounded-2xl border border-slate-200 shadow-xs overflow-hidden">
        {isLoadingList ? (
          <div className="p-8 text-center text-slate-500">Klinikalar yuklanmoqda...</div>
        ) : clinics.length === 0 ? (
          <div className="p-12 text-center text-slate-500">
            <Building2 className="w-10 h-10 text-slate-300 mx-auto mb-2" />
            <p className="text-sm font-semibold text-slate-700">Klinikalar topilmadi</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs text-slate-600">
              <thead className="bg-slate-50/80 text-slate-700 uppercase font-semibold border-b border-slate-200">
                <tr>
                  <th className="px-5 py-3.5">Klinika</th>
                  <th className="px-5 py-3.5">Hudud</th>
                  <th className="px-5 py-3.5">Aloqa</th>
                  <th className="px-5 py-3.5">Sana</th>
                  <th className="px-5 py-3.5">Status</th>
                  <th className="px-5 py-3.5 text-right">Amallar</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {clinics.map((c) => (
                  <tr key={c.id} className="hover:bg-slate-50/50 transition-colors">
                    <td className="px-5 py-3.5">
                      <div className="flex items-center gap-3">
                        {c.logo_url ? (
                          <img
                            src={getMediaUrl(c.logo_url)}
                            alt={c.name}
                            className="w-9 h-9 rounded-xl object-contain border border-slate-200 bg-white"
                          />
                        ) : (
                          <div className="w-9 h-9 rounded-xl bg-blue-50 text-[#0866F5] flex items-center justify-center font-bold text-xs border border-blue-100">
                            {c.name[0]}
                          </div>
                        )}
                        <div>
                          <div className="font-bold text-slate-900 text-sm">{c.name}</div>
                          <div className="text-[11px] text-slate-400 line-clamp-1">
                            {c.description || 'Tavsif yo‘q'}
                          </div>
                        </div>
                      </div>
                    </td>

                    <td className="px-5 py-3.5">
                      <div className="font-medium text-slate-800">{c.district}</div>
                      <div className="text-[11px] text-slate-400">{c.region}</div>
                    </td>

                    <td className="px-5 py-3.5">
                      <div className="font-medium text-slate-800">{c.phone}</div>
                      <div className="text-[11px] text-slate-400">{c.email || '—'}</div>
                    </td>

                    <td className="px-5 py-3.5 whitespace-nowrap text-slate-500">
                      {formatDate(c.created_at)}
                    </td>

                    <td className="px-5 py-3.5 whitespace-nowrap">
                      {renderStatusBadge(c.status)}
                    </td>

                    <td className="px-5 py-3.5 text-right whitespace-nowrap">
                      <div className="inline-flex items-center gap-1.5">
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => setDetailClinicId(c.id)}
                          className="text-slate-600 hover:text-slate-900 text-xs px-2"
                          title="Tafsilotlarni ko‘rish"
                        >
                          <Eye className="w-3.5 h-3.5" />
                        </Button>

                        {c.status === 'pending' && (
                          <>
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => setApproveClinicId(c.id)}
                              className="text-emerald-600 hover:bg-emerald-50 text-xs px-2"
                              title="Tasdiqlash"
                            >
                              <Check className="w-3.5 h-3.5" />
                            </Button>

                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() => setRejectClinicId(c.id)}
                              className="text-rose-600 hover:bg-rose-50 text-xs px-2"
                              title="Rad etish"
                            >
                              <XCircle className="w-3.5 h-3.5" />
                            </Button>
                          </>
                        )}

                        {c.status === 'approved' && (
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => setSuspendClinicId(c.id)}
                            className="text-amber-600 hover:bg-amber-50 text-xs px-2"
                            title="Faoliyatini to‘xtatish (bloklash)"
                          >
                            <Ban className="w-3.5 h-3.5" />
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* 1. Clinic Details Modal */}
      {detailClinicId && (
        <Modal
          isOpen={Boolean(detailClinicId)}
          onClose={() => {
            setDetailClinicId(null);
            setDetailTab('info');
          }}
          title="Klinika boshqaruvi va ma’lumotlari"
        >
          {isLoadingDetails || !detailsData ? (
            <div className="p-8 text-center text-slate-500">
              <div className="w-6 h-6 border-2 border-[#0866F5] border-t-transparent rounded-full animate-spin mx-auto mb-2" />
              Yuklanmoqda...
            </div>
          ) : (
            <div className="space-y-4 text-xs">
              {/* Clinic Header */}
              <div className="flex items-center justify-between gap-3 pb-3 border-b border-slate-100">
                <div className="flex items-center gap-3">
                  {detailsData.clinic.logo_url ? (
                    <img
                      src={getMediaUrl(detailsData.clinic.logo_url)}
                      alt={detailsData.clinic.name}
                      className="w-12 h-12 rounded-xl object-contain border border-slate-200 bg-white"
                    />
                  ) : (
                    <div className="w-12 h-12 rounded-xl bg-blue-50 text-[#0866F5] font-black text-base flex items-center justify-center border border-blue-100">
                      {detailsData.clinic.name[0]}
                    </div>
                  )}
                  <div>
                    <h3 className="text-base font-bold text-slate-900">
                      {detailsData.clinic.name}
                    </h3>
                    <p className="text-slate-500 text-[11px]">{detailsData.clinic.address}</p>
                  </div>
                </div>
                <div>{renderStatusBadge(detailsData.clinic.status)}</div>
              </div>

              {/* Detail Tabs */}
              <div className="flex items-center gap-1.5 border-b border-slate-200 pb-2">
                <button
                  type="button"
                  onClick={() => setDetailTab('info')}
                  className={`px-3 py-1.5 rounded-xl font-semibold text-xs transition-all flex items-center gap-1.5 ${
                    detailTab === 'info'
                      ? 'bg-[#0866F5] text-white shadow-2xs'
                      : 'bg-slate-100 text-slate-600 hover:bg-slate-200/70'
                  }`}
                >
                  <Info className="w-3.5 h-3.5" />
                  Ma’lumot
                </button>

                <button
                  type="button"
                  onClick={() => setDetailTab('doctors')}
                  className={`px-3 py-1.5 rounded-xl font-semibold text-xs transition-all flex items-center gap-1.5 ${
                    detailTab === 'doctors'
                      ? 'bg-[#0866F5] text-white shadow-2xs'
                      : 'bg-slate-100 text-slate-600 hover:bg-slate-200/70'
                  }`}
                >
                  <Stethoscope className="w-3.5 h-3.5" />
                  Shifokorlar
                  {clinicDoctors && clinicDoctors.length > 0 && (
                    <span className="ml-1 px-1.5 py-0.2 rounded-full text-[10px] bg-white/20 text-current">
                      {clinicDoctors.length}
                    </span>
                  )}
                </button>

                <button
                  type="button"
                  onClick={() => setDetailTab('appointments')}
                  className={`px-3 py-1.5 rounded-xl font-semibold text-xs transition-all flex items-center gap-1.5 ${
                    detailTab === 'appointments'
                      ? 'bg-[#0866F5] text-white shadow-2xs'
                      : 'bg-slate-100 text-slate-600 hover:bg-slate-200/70'
                  }`}
                >
                  <CalendarCheck className="w-3.5 h-3.5" />
                  Bronlar
                </button>

                <button
                  type="button"
                  onClick={() => setDetailTab('analytics')}
                  className={`px-3 py-1.5 rounded-xl font-semibold text-xs transition-all flex items-center gap-1.5 ${
                    detailTab === 'analytics'
                      ? 'bg-[#0866F5] text-white shadow-2xs'
                      : 'bg-slate-100 text-slate-600 hover:bg-slate-200/70'
                  }`}
                >
                  <TrendingUp className="w-3.5 h-3.5" />
                  Tahlil
                </button>
              </div>

              {/* Tab 1: Overview */}
              {detailTab === 'info' && (
                <div className="space-y-4">
                  <div className="grid grid-cols-2 gap-3 bg-slate-50 p-3 rounded-xl border border-slate-100">
                    <div>
                      <span className="text-slate-400 block text-[11px]">Klinika nomi:</span>
                      <span className="font-semibold text-slate-900 block mt-0.5">
                        {detailsData.clinic.name}
                      </span>
                    </div>
                    <div>
                      <span className="text-slate-400 block text-[11px]">Shifokorlar soni:</span>
                      <span className="font-bold text-slate-900 block mt-0.5">
                        {detailsData.clinic.doctors_count} nafar
                      </span>
                    </div>
                    <div>
                      <span className="text-slate-400 block text-[11px]">Telefon:</span>
                      <span className="font-semibold text-slate-900 block mt-0.5">
                        {detailsData.clinic.phone}
                      </span>
                    </div>
                    <div>
                      <span className="text-slate-400 block text-[11px]">Email:</span>
                      <span className="font-semibold text-slate-900 block mt-0.5">
                        {detailsData.clinic.email || '—'}
                      </span>
                    </div>
                    <div>
                      <span className="text-slate-400 block text-[11px]">Hudud / Tuman:</span>
                      <span className="font-semibold text-slate-900 block mt-0.5">
                        {detailsData.clinic.region}, {detailsData.clinic.district}
                      </span>
                    </div>
                    <div>
                      <span className="text-slate-400 block text-[11px]">Ariza sanasi:</span>
                      <span className="font-semibold text-slate-900 block mt-0.5">
                        {formatDate(detailsData.clinic.created_at)}
                      </span>
                    </div>
                  </div>

                  {detailsData.clinic.rejection_reason && (
                    <div className="p-3 bg-rose-50 border border-rose-200 rounded-xl text-rose-800">
                      <span className="font-semibold block mb-0.5">Rad etish / Bloklash sababi:</span>
                      {detailsData.clinic.rejection_reason}
                    </div>
                  )}

                  {/* Review History */}
                  {detailsData.reviews && detailsData.reviews.length > 0 && (
                    <div className="space-y-2 pt-1">
                      <span className="font-bold text-slate-800 block text-xs">
                        Tekshiruvlar tarixi (Audit)
                      </span>
                      <div className="space-y-1.5 max-h-36 overflow-y-auto">
                        {detailsData.reviews.map((rev) => (
                          <div
                            key={rev.id}
                            className="p-2.5 rounded-lg bg-slate-50 border border-slate-200 flex flex-col gap-0.5"
                          >
                            <div className="flex items-center justify-between font-semibold">
                              <span className="uppercase text-[10px] text-slate-700">
                                Qaror: {rev.decision}
                              </span>
                              <span className="text-slate-400 text-[10px]">
                                {formatDate(rev.reviewed_at)}
                              </span>
                            </div>
                            {rev.reason && (
                              <p className="text-slate-600 italic">"{rev.reason}"</p>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}

              {/* Tab 2: Clinic Doctors */}
              {detailTab === 'doctors' && (
                <div className="space-y-3">
                  {isLoadingDoctors ? (
                    <div className="p-8 text-center text-slate-400">
                      <div className="w-6 h-6 border-2 border-[#0866F5] border-t-transparent rounded-full animate-spin mx-auto mb-2" />
                      Shifokorlar yuklanmoqda...
                    </div>
                  ) : !clinicDoctors || clinicDoctors.length === 0 ? (
                    <div className="p-8 text-center text-slate-400 bg-slate-50 rounded-xl border border-slate-100">
                      <Stethoscope className="w-8 h-8 text-slate-300 mx-auto mb-2" />
                      Bu klinikada hozircha shifokorlar mavjud emas
                    </div>
                  ) : (
                    <div className="space-y-2 max-h-72 overflow-y-auto pr-1">
                      {clinicDoctors.map((doc) => (
                        <div
                          key={doc.clinic_doctor_id}
                          className="p-3 rounded-xl bg-slate-50 border border-slate-200/80 flex items-center justify-between gap-3"
                        >
                          <div className="flex items-center gap-3">
                            {doc.photo_url ? (
                              <img
                                src={getMediaUrl(doc.photo_url)}
                                alt={doc.first_name}
                                className="w-10 h-10 rounded-xl object-cover border border-slate-200 bg-white"
                              />
                            ) : (
                              <div className="w-10 h-10 rounded-xl bg-blue-100 text-[#0866F5] font-black text-xs flex items-center justify-center">
                                {(doc.first_name?.[0] || 'D').toUpperCase()}
                              </div>
                            )}
                            <div>
                              <div className="font-bold text-slate-900 text-xs">
                                {doc.first_name} {doc.last_name}
                              </div>
                              <div className="text-[11px] text-slate-500">
                                {doc.specialty_name || 'Umumiy'} • {doc.experience_years} yil tajriba
                              </div>
                            </div>
                          </div>

                          <div className="text-right shrink-0">
                            <span className="font-bold text-slate-900 text-xs block">
                              {formatPrice(doc.consultation_price)}
                            </span>
                            <span
                              className={`inline-block mt-0.5 px-1.5 py-0.2 rounded-md text-[10px] font-semibold ${
                                doc.is_active
                                  ? 'bg-emerald-100 text-emerald-800'
                                  : 'bg-slate-200 text-slate-700'
                              }`}
                            >
                              {doc.is_active ? 'Faol' : 'Nofaol'}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}

              {/* Tab 3: Clinic Appointments */}
              {detailTab === 'appointments' && (
                <div className="space-y-3">
                  {isLoadingAppointments ? (
                    <div className="p-8 text-center text-slate-400">
                      <div className="w-6 h-6 border-2 border-[#0866F5] border-t-transparent rounded-full animate-spin mx-auto mb-2" />
                      Bronlar yuklanmoqda...
                    </div>
                  ) : !clinicAppointments || clinicAppointments.length === 0 ? (
                    <div className="p-8 text-center text-slate-400 bg-slate-50 rounded-xl border border-slate-100">
                      <CalendarCheck className="w-8 h-8 text-slate-300 mx-auto mb-2" />
                      Klinika bo‘yicha hozircha bronlar mavjud emas
                    </div>
                  ) : (
                    <div className="space-y-2 max-h-72 overflow-y-auto pr-1">
                      {clinicAppointments.map((app) => (
                        <div
                          key={app.id}
                          className="p-3 rounded-xl bg-slate-50 border border-slate-200/80 flex items-center justify-between gap-3"
                        >
                          <div>
                            <div className="font-bold text-slate-900 text-xs">
                              {app.patient_name} —{' '}
                              <span className="text-[#0866F5]">{app.doctor_name}</span>
                            </div>
                            <div className="text-[11px] text-slate-500 mt-0.5 flex items-center gap-2">
                              <span>{app.specialty_name || 'Shifokor'}</span>
                              <span>•</span>
                              <span>
                                {formatDate(app.appointment_date)} {app.start_time}
                              </span>
                            </div>
                          </div>

                          <div className="text-right shrink-0">
                            <span className="font-bold text-slate-900 text-xs block">
                              {formatPrice(app.consultation_price)}
                            </span>
                            <span className="text-[10px] uppercase font-semibold text-slate-500">
                              {app.status}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}

              {/* Tab 4: Clinic Analytics */}
              {detailTab === 'analytics' && (
                <div className="space-y-4">
                  {isLoadingAnalytics || !clinicAnalytics ? (
                    <div className="p-8 text-center text-slate-400">
                      <div className="w-6 h-6 border-2 border-[#0866F5] border-t-transparent rounded-full animate-spin mx-auto mb-2" />
                      Statistika yuklanmoqda...
                    </div>
                  ) : (
                    <div className="space-y-3">
                      {/* KPI cards */}
                      <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
                        <div className="p-2.5 rounded-xl bg-slate-50 border border-slate-200 text-center">
                          <span className="text-slate-400 text-[10px] block">Jami bronlar</span>
                          <span className="text-base font-black text-slate-900 block mt-0.5">
                            {clinicAnalytics.total_appointments}
                          </span>
                        </div>
                        <div className="p-2.5 rounded-xl bg-amber-50/70 border border-amber-200 text-center">
                          <span className="text-amber-700 text-[10px] block">Kelgusi</span>
                          <span className="text-base font-black text-amber-900 block mt-0.5">
                            {clinicAnalytics.upcoming_appointments}
                          </span>
                        </div>
                        <div className="p-2.5 rounded-xl bg-emerald-50/70 border border-emerald-200 text-center">
                          <span className="text-emerald-700 text-[10px] block">Yakunlangan</span>
                          <span className="text-base font-black text-emerald-900 block mt-0.5">
                            {clinicAnalytics.completed_appointments}
                          </span>
                        </div>
                        <div className="p-2.5 rounded-xl bg-blue-50/70 border border-blue-200 text-center">
                          <span className="text-blue-700 text-[10px] block">Taxminiy hajm</span>
                          <span className="text-xs font-black text-blue-900 block mt-1">
                            {formatPrice(clinicAnalytics.estimated_total_revenue)}
                          </span>
                        </div>
                      </div>

                      {/* Top Doctors */}
                      <div>
                        <span className="font-bold text-slate-800 text-xs block mb-2">
                          Yetakchi shifokorlar (bronlar soni bo‘yicha)
                        </span>
                        {!clinicAnalytics.top_doctors || clinicAnalytics.top_doctors.length === 0 ? (
                          <div className="p-4 text-center text-slate-400 bg-slate-50 rounded-xl text-xs">
                            Yetakchi shifokorlar mavjud emas
                          </div>
                        ) : (
                          <div className="space-y-1.5 max-h-48 overflow-y-auto">
                            {clinicAnalytics.top_doctors.map((td, idx) => (
                              <div
                                key={td.doctor_id}
                                className="p-2 rounded-xl bg-slate-50 border border-slate-200/80 flex items-center justify-between text-xs"
                              >
                                <div className="flex items-center gap-2">
                                  <span className="w-5 h-5 rounded-full bg-blue-100 text-[#0866F5] font-black text-[10px] flex items-center justify-center">
                                    {idx + 1}
                                  </span>
                                  <div>
                                    <span className="font-bold text-slate-900 block">
                                      {td.doctor_name}
                                    </span>
                                    <span className="text-[10px] text-slate-400">
                                      {td.specialty_name || 'Umumiy'}
                                    </span>
                                  </div>
                                </div>
                                <span className="font-bold text-[#0866F5]">
                                  {td.bookings_count} ta bron
                                </span>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </Modal>
      )}

      {/* 2. Approve Confirmation Modal */}
      {approveClinicId && (
        <Modal
          isOpen={Boolean(approveClinicId)}
          onClose={() => setApproveClinicId(null)}
          title="Klinikani tasdiqlash"
        >
          <div className="space-y-4 text-xs text-slate-600">
            <p>
              Ushbu klinikani rasman tasdiqlab, platformada faollashtirishni xohlaysizmi?
              Tasdiqlangach, klinika xodimlari shifokorlarni qo‘shish imkoniyatiga ega bo‘ladi.
            </p>

            <div className="flex justify-end gap-2 pt-3">
              <Button variant="outline" size="sm" onClick={() => setApproveClinicId(null)}>
                Bekor qilish
              </Button>
              <Button
                variant="primary"
                size="sm"
                className="bg-emerald-600 hover:bg-emerald-700 text-white"
                isLoading={approveMutation.isPending}
                onClick={() => approveMutation.mutate(approveClinicId)}
              >
                Tasdiqlash
              </Button>
            </div>
          </div>
        </Modal>
      )}

      {/* 3. Reject Modal with Reason */}
      {rejectClinicId && (
        <Modal
          isOpen={Boolean(rejectClinicId)}
          onClose={() => {
            setRejectClinicId(null);
            setActionReason('');
          }}
          title="Arizani rad etish"
        >
          <div className="space-y-4 text-xs">
            <p className="text-slate-600">
              Klinika arizasini rad etish sababini batafsil yozing. Bu sabab klinika egasining
              kabinetida ko‘rinadi.
            </p>

            <div>
              <label className="font-semibold text-slate-700 block mb-1">
                Rad etish sababi *
              </label>
              <textarea
                rows={3}
                className="w-full px-3 py-2 rounded-xl border border-slate-300 text-xs focus:outline-hidden focus:border-rose-500 focus:ring-2 focus:ring-rose-100"
                placeholder="Masalan: Tibbiy litsenziya ma’lumotlari yoki aniq manzil to‘liq ko‘rsatilmagan..."
                value={actionReason}
                onChange={(e) => setActionReason(e.target.value)}
              />
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setRejectClinicId(null);
                  setActionReason('');
                }}
              >
                Bekor qilish
              </Button>
              <Button
                variant="destructive"
                size="sm"
                disabled={actionReason.trim().length < 5}
                isLoading={rejectMutation.isPending}
                onClick={() =>
                  rejectMutation.mutate({ id: rejectClinicId, reason: actionReason })
                }
              >
                Rad etish
              </Button>
            </div>
          </div>
        </Modal>
      )}

      {/* 4. Suspend Modal with Reason */}
      {suspendClinicId && (
        <Modal
          isOpen={Boolean(suspendClinicId)}
          onClose={() => {
            setSuspendClinicId(null);
            setActionReason('');
          }}
          title="Klinika faoliyatini to‘xtatish"
        >
          <div className="space-y-4 text-xs">
            <p className="text-slate-600">
              Klinikani bloklash yoki faoliyatini vaqtincha to‘xtatish sababini yozing.
            </p>

            <div>
              <label className="font-semibold text-slate-700 block mb-1">
                To‘xtatish sababi *
              </label>
              <textarea
                rows={3}
                className="w-full px-3 py-2 rounded-xl border border-slate-300 text-xs focus:outline-hidden focus:border-red-500 focus:ring-2 focus:ring-red-100"
                placeholder="Qoidabuzarlik yoki tekshiruv sababini kiriting..."
                value={actionReason}
                onChange={(e) => setActionReason(e.target.value)}
              />
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  setSuspendClinicId(null);
                  setActionReason('');
                }}
              >
                Bekor qilish
              </Button>
              <Button
                variant="destructive"
                size="sm"
                disabled={actionReason.trim().length < 5}
                isLoading={suspendMutation.isPending}
                onClick={() =>
                  suspendMutation.mutate({ id: suspendClinicId, reason: actionReason })
                }
              >
                To‘xtatish (Bloklash)
              </Button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  );
}
