'use client';

import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useClinicContext } from '@/components/layout/clinic/ClinicLayout';
import {
  Users,
  Building2,
  Clock,
  Phone,
  MapPin,
  PlusCircle,
  Edit,
  ExternalLink,
  Calendar,
  CalendarCheck,
  CheckCircle2,
  ArrowRight,
} from 'lucide-react';
import { ClinicDetail } from '../types';
import { Button } from '@/components/ui/Button';
import { getMediaUrl } from '@/lib/utils/media';
import { useClinicAppointments } from '@/features/appointments/hooks/useAppointments';
import { AppointmentStatusBadge } from '@/features/appointments/components/AppointmentStatusBadge';

interface ClinicContextType {
  clinic?: ClinicDetail;
  isLoadingClinic: boolean;
}

const WEEKDAYS = [
  'Yakshanba',
  'Dushanba',
  'Seshanba',
  'Chorshanba',
  'Payshanba',
  'Juma',
  'Shanba',
];

export function ClinicDashboardPage() {
  const { clinic, isLoadingClinic } = useClinicContext();
  const router = useRouter();

  const todayStr = new Date(Date.now() + 5 * 60 * 60 * 1000).toISOString().split('T')[0];

  const { data: todayAppointmentsData, isLoading: isLoadingAppointments } = useClinicAppointments(
    clinic?.id || '',
    { date: todayStr, limit: 5 },
    Boolean(clinic?.id && clinic.status === 'approved')
  );

  const todayAppointments = todayAppointmentsData?.appointments || [];
  const todayTotal = todayAppointmentsData?.total || 0;

  if (isLoadingClinic) {
    return (
      <div className="space-y-4 animate-pulse">
        <div className="h-40 bg-slate-200 rounded-2xl" />
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="h-28 bg-slate-200 rounded-2xl" />
          <div className="h-28 bg-slate-200 rounded-2xl" />
          <div className="h-28 bg-slate-200 rounded-2xl" />
        </div>
      </div>
    );
  }

  if (!clinic) {
    return (
      <div className="bg-white rounded-2xl border border-slate-200 p-8 text-center max-w-lg mx-auto mt-12 shadow-sm">
        <Building2 className="w-12 h-12 text-[#0866F5] mx-auto mb-3" />
        <h3 className="text-lg font-bold text-slate-900 mb-1">Klinika profili topilmadi</h3>
        <p className="text-xs text-slate-500 mb-6">
          Klinika ma’lumotlarini to‘ldirish uchun qisqa onboarding bosqichidan o‘ting.
        </p>
        <Button
          variant="primary"
          onClick={() => router.push('/clinic/onboarding')}
          className="bg-[#0866F5]"
        >
          Klinika profilini yaratish
        </Button>
      </div>
    );
  }

  // Today's working hour calculation
  const todayJs = new Date().getDay();
  const todayWeekdayNumber = todayJs === 0 ? 7 : todayJs;
  const todayWorkingHour = clinic.working_hours?.find(
    (wh) => wh.weekday === todayWeekdayNumber
  );

  return (
    <div className="space-y-6">
      {/* Top Welcome / Hero Banner */}
      <div className="bg-white rounded-3xl border border-slate-200 p-6 sm:p-8 shadow-xs relative overflow-hidden">
        <div className="absolute top-0 right-0 -mr-16 -mt-16 w-64 h-64 bg-blue-50/60 rounded-full blur-2xl pointer-events-none" />

        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6 relative z-10">
          <div className="flex items-center gap-4">
            {clinic.logo_url ? (
              <img
                src={getMediaUrl(clinic.logo_url)}
                alt={clinic.name}
                className="w-16 h-16 sm:w-20 sm:h-20 object-contain rounded-2xl border border-slate-200 bg-white p-1.5 shadow-2xs"
              />
            ) : (
              <div className="w-16 h-16 sm:w-20 sm:h-20 rounded-2xl bg-blue-50 border border-blue-200 flex items-center justify-center text-[#0866F5]">
                <Building2 className="w-8 h-8" />
              </div>
            )}

            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-xl sm:text-2xl font-extrabold text-slate-900">
                  {clinic.name}
                </h1>
              </div>
              <p className="text-xs text-slate-500 mt-1 flex items-center gap-1.5">
                <MapPin className="w-3.5 h-3.5 text-slate-400" />
                {clinic.address}, {clinic.district}, {clinic.region}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2.5 w-full sm:w-auto">
            {clinic.status === 'approved' && (
              <Button
                variant="primary"
                size="md"
                onClick={() => router.push('/clinic/doctors/new')}
                className="bg-[#0866F5] hover:bg-blue-600 shadow-xs gap-1.5 w-full sm:w-auto"
              >
                <PlusCircle className="w-4 h-4" />
                Shifokor qo‘shish
              </Button>
            )}

            <Button
              variant="outline"
              size="md"
              onClick={() => router.push('/clinic/profile')}
              className="gap-1.5 w-full sm:w-auto"
            >
              <Edit className="w-4 h-4" />
              Profilni tahrirlash
            </Button>
          </div>
        </div>
      </div>

      {/* Metric Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Shifokorlar */}
        <div className="bg-white rounded-2xl border border-slate-200 p-5 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-500">Shifokorlar</span>
            <div className="w-9 h-9 rounded-xl bg-blue-50 text-[#0866F5] flex items-center justify-center">
              <Users className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-extrabold text-slate-900">
              {clinic.doctors_count || 0}
            </span>
            <span className="text-xs text-slate-500">nafar</span>
          </div>
          <div className="mt-2 text-xs">
            {clinic.status === 'approved' ? (
              <Link
                href="/clinic/doctors"
                className="text-[#0866F5] font-semibold hover:underline inline-flex items-center gap-1"
              >
                Barchasini ko‘rish <ExternalLink className="w-3 h-3" />
              </Link>
            ) : (
              <span className="text-slate-400">Tasdiqlangach boshqarish mumkin</span>
            )}
          </div>
        </div>

        {/* Bugungi Qabullar */}
        <div className="bg-white rounded-2xl border border-slate-200 p-5 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-500">Bugungi qabullar</span>
            <div className="w-9 h-9 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center">
              <Calendar className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3 flex items-baseline gap-2">
            <span className="text-2xl font-extrabold text-slate-900">
              {todayTotal}
            </span>
            <span className="text-xs text-slate-500">ta bron</span>
          </div>
          <div className="mt-2 text-xs">
            {clinic.status === 'approved' ? (
              <Link
                href="/clinic/appointments"
                className="text-[#0866F5] font-semibold hover:underline inline-flex items-center gap-1"
              >
                Bronlarni boshqarish <ExternalLink className="w-3 h-3" />
              </Link>
            ) : (
              <span className="text-slate-400">Tasdiqlangach ko‘rinadi</span>
            )}
          </div>
        </div>

        {/* Bugungi Ish vaqti */}
        <div className="bg-white rounded-2xl border border-slate-200 p-5 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-500">Bugungi ish rejimi</span>
            <div className="w-9 h-9 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center">
              <Clock className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3">
            {todayWorkingHour?.is_closed ? (
              <span className="text-base font-bold text-rose-600">Dam olish kuni</span>
            ) : (
              <div className="text-base font-extrabold text-slate-900 font-mono">
                {todayWorkingHour?.opening_time || '08:00'} —{' '}
                {todayWorkingHour?.closing_time || '18:00'}
              </div>
            )}
          </div>
          <div className="mt-2 text-xs text-slate-500">
            {WEEKDAYS[todayJs]} kuni bo‘yicha
          </div>
        </div>

        {/* Bog'lanish */}
        <div className="bg-white rounded-2xl border border-slate-200 p-5 shadow-xs">
          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-500">Aloqa</span>
            <div className="w-9 h-9 rounded-xl bg-teal-50 text-teal-600 flex items-center justify-center">
              <Phone className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-3 text-sm font-semibold text-slate-900 truncate">
            {clinic.phone}
          </div>
          <div className="mt-1 text-xs text-slate-500 truncate">
            {clinic.email || 'Email kiritilmagan'}
          </div>
        </div>
      </div>

      {/* Bugungi Qabullar Jadvali */}
      {clinic.status === 'approved' && (
        <div className="bg-white rounded-2xl border border-slate-200 p-6 shadow-xs space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-bold text-slate-900 flex items-center gap-2">
                <CalendarCheck className="w-4 h-4 text-brand-blue" />
                <span>Bugungi qabullar</span>
              </h3>
              <p className="text-xs text-slate-500 mt-0.5">
                Klinikada bugungi kun uchun tasdiqlangan qabullar ro‘yxati
              </p>
            </div>
            <Link
              href="/clinic/appointments"
              className="text-xs font-bold text-brand-blue hover:underline flex items-center gap-1"
            >
              <span>Barcha bronlar</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </Link>
          </div>

          {todayAppointments.length === 0 ? (
            <div className="py-8 text-center text-xs text-slate-400 bg-slate-50 rounded-xl border border-dashed border-slate-200">
              Bugungi kunga hozircha yangi bronlar mavjud emas.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs text-slate-600">
                <thead>
                  <tr className="border-b border-slate-100 text-slate-400 uppercase font-bold text-[10px]">
                    <th className="py-2.5 px-3">Bemor</th>
                    <th className="py-2.5 px-3">Shifokor</th>
                    <th className="py-2.5 px-3">Vaqt</th>
                    <th className="py-2.5 px-3">Holat</th>
                    <th className="py-2.5 px-3 text-right">Amal</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {todayAppointments.map((appt) => (
                    <tr key={appt.id} className="hover:bg-slate-50/60">
                      <td className="py-2.5 px-3 font-semibold text-slate-900">
                        {appt.patient
                          ? `${appt.patient.first_name} ${appt.patient.last_name}`
                          : 'Bemor'}
                      </td>
                      <td className="py-2.5 px-3">
                        {appt.doctor
                          ? `${appt.doctor.first_name} ${appt.doctor.last_name}`
                          : 'Shifokor'}
                      </td>
                      <td className="py-2.5 px-3 font-mono font-bold">
                        {appt.start_time} – {appt.end_time}
                      </td>
                      <td className="py-2.5 px-3">
                        <AppointmentStatusBadge status={appt.status} size="sm" />
                      </td>
                      <td className="py-2.5 px-3 text-right">
                        <Link
                          href={`/clinic/appointments/${appt.id}`}
                          className="text-brand-blue font-bold hover:underline"
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
      )}

      {/* Description & Details Section */}
      <div className="bg-white rounded-2xl border border-slate-200 p-6 shadow-xs">
        <h3 className="text-sm font-bold text-slate-900 mb-2">Klinika haqida</h3>
        <p className="text-xs text-slate-600 leading-relaxed">
          {clinic.description ||
            'Klinika haqida tavsif hali kiritilmagan. Profilingizni to‘ldirish orqali bemorlarga klinika imkoniyatlari haqida batafsil ma’lumot berishingiz mumkin.'}
        </p>
      </div>
    </div>
  );
}
