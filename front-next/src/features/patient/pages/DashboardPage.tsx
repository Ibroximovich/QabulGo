'use client';

import Link from 'next/link';
import {
  User,
  Mail,
  Phone,
  Calendar,
  CheckCircle2,
  Clock,
  ArrowRight,
  Building2,
  Stethoscope,
  CalendarDays,
  CalendarCheck,
  ShieldCheck,
  AlertCircle,
  RefreshCw,
  Search,
} from 'lucide-react';
import { AppLayout } from '@/components/layout/AppLayout';
import { Card, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Skeleton } from '@/components/ui/Skeleton';
import { usePatientProfile } from '../hooks/usePatientProfile';
import { formatDateUz, formatRoleUz } from '@/lib/utils/formatters';
import { formatUzbekPhoneDisplay } from '@/lib/utils/phone';
import { useMyAppointments } from '@/features/appointments/hooks/useAppointments';
import { AppointmentStatusBadge } from '@/features/appointments/components/AppointmentStatusBadge';

export function DashboardPage() {
  const { data: profile, isLoading, isError, refetch } = usePatientProfile();

  // Fetch upcoming appointment for quick preview
  const { data: appointmentsData, isLoading: isLoadingAppointments } = useMyAppointments({
    timeframe: 'upcoming',
    limit: 1,
  });

  const nextAppointment = appointmentsData?.appointments?.[0];

  const quickServices = [
    {
      title: 'Klinikalar katalogi',
      description: 'O‘zbekistondagi nufuzli xususiy klinikalarni joylashuvi va xizmatlari bo‘yicha qidiring.',
      icon: Building2,
      href: '/clinics',
      badge: 'Faol',
      badgeVariant: 'success' as const,
    },
    {
      title: 'Mutaxassis shifokorlar',
      description: 'Tajribali shifokorlar reytingi, ko‘rik narxlari va qabul jadvallari bilan tanishing.',
      icon: Stethoscope,
      href: '/doctors',
      badge: 'Faol',
      badgeVariant: 'success' as const,
    },
    {
      title: 'Onlayn navbat olish',
      description: 'Shifokorning bo‘sh vaqtlaridan o‘zingizga mos soatni tanlab, 1 daqiqada qabulga yoziling.',
      icon: CalendarDays,
      href: '/doctors',
      badge: 'Faol',
      badgeVariant: 'success' as const,
    },
    {
      title: 'Mening qabullarim',
      description: 'Rejalashtirilgan va avvalgi barcha qabullaringiz ro‘yxatini bitta joyda boshqaring.',
      icon: CalendarCheck,
      href: '/appointments',
      badge: 'Faol',
      badgeVariant: 'success' as const,
    },
  ];

  return (
    <AppLayout>
      <div className="space-y-8 animate-fade-in">
        {/* Welcome Banner */}
        <div className="relative overflow-hidden rounded-3xl bg-gradient-to-r from-brand-blue to-brand-teal p-6 sm:p-10 text-white shadow-elevated">
          <div className="relative z-10 max-w-2xl space-y-3">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-white/20 backdrop-blur-md text-xs font-semibold">
              <ShieldCheck className="h-3.5 w-3.5" />
              <span>Bemor shaxsiy kabineti</span>
            </div>
            {isLoading ? (
              <Skeleton className="h-9 w-64 bg-white/30 rounded-xl" />
            ) : (
              <h1 className="text-2xl sm:text-4xl font-extrabold tracking-tight">
                Salom, {profile?.first_name || 'Bemor'}!
              </h1>
            )}
            <p className="text-sm sm:text-base text-white/90 leading-relaxed">
              QabulGo platformasiga xush kelibsiz. Bu yerda siz shaxsiy ma’lumotlaringizni boshqarishingiz va kelgusida klinikalarga navbat olishingiz mumkin.
            </p>
          </div>

          {/* Decorative shapes */}
          <div className="absolute right-0 top-0 -mt-12 -mr-12 h-64 w-64 rounded-full bg-white/10 blur-2xl pointer-events-none" />
          <div className="absolute right-24 bottom-0 -mb-16 h-48 w-48 rounded-full bg-brand-navy/10 blur-xl pointer-events-none" />
        </div>

        {/* Profile Card Section */}
        <div>
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-lg sm:text-xl font-bold text-brand-navy">Shaxsiy ma’lumotlar</h2>
              <p className="text-xs sm:text-sm text-brand-muted">
                Profil va aloqa ma’lumotlaringiz
              </p>
            </div>
            <Link href="/profile">
              <Button variant="outline" size="sm" rightIcon={<ArrowRight className="h-3.5 w-3.5" />}>
                Tahrirlash
              </Button>
            </Link>
          </div>

          {isLoading ? (
            <Card>
              <CardContent className="p-6 grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                {[1, 2, 3, 4].map((i) => (
                  <div key={i} className="space-y-2">
                    <Skeleton className="h-4 w-20" />
                    <Skeleton className="h-6 w-36" />
                  </div>
                ))}
              </CardContent>
            </Card>
          ) : isError ? (
            <Card className="border-red-100 bg-red-50/20">
              <CardContent className="p-6 flex flex-col sm:flex-row items-center justify-between gap-4">
                <div className="flex items-center gap-3 text-brand-error">
                  <AlertCircle className="h-5 w-5 shrink-0" />
                  <p className="text-sm font-medium">Profil ma’lumotlarini yuklashda xatolik yuz berdi.</p>
                </div>
                <Button variant="outline" size="sm" onClick={() => refetch()} leftIcon={<RefreshCw className="h-3.5 w-3.5" />}>
                  Qayta urinish
                </Button>
              </CardContent>
            </Card>
          ) : (
            <Card elevated>
              <CardContent className="p-6 sm:p-8">
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                  {/* Ism va familiya */}
                  <div className="flex items-start gap-3.5">
                    <div className="p-2.5 rounded-xl bg-brand-blue-50 text-brand-blue shrink-0">
                      <User className="h-5 w-5" />
                    </div>
                    <div>
                      <p className="text-xs font-semibold text-brand-muted uppercase tracking-wider">
                        To‘liq ism
                      </p>
                      <p className="text-base font-bold text-brand-navy mt-0.5">
                        {profile?.first_name} {profile?.last_name}
                      </p>
                      <Badge variant="teal" size="sm" className="mt-1.5">
                        {formatRoleUz(profile?.role)}
                      </Badge>
                    </div>
                  </div>

                  {/* Email */}
                  <div className="flex items-start gap-3.5">
                    <div className="p-2.5 rounded-xl bg-brand-teal-50 text-brand-teal shrink-0">
                      <Mail className="h-5 w-5" />
                    </div>
                    <div className="overflow-hidden">
                      <p className="text-xs font-semibold text-brand-muted uppercase tracking-wider">
                        Email
                      </p>
                      <p className="text-sm sm:text-base font-bold text-brand-navy mt-0.5 truncate max-w-full">
                        {profile?.email}
                      </p>
                      <div className="mt-1.5">
                        {profile?.email_verified ? (
                          <Badge variant="success" size="sm" className="gap-1">
                            <CheckCircle2 className="h-3 w-3" />
                            <span>Tasdiqlangan</span>
                          </Badge>
                        ) : (
                          <Badge variant="warning" size="sm">
                            Tasdiqlanmagan
                          </Badge>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Telefon */}
                  <div className="flex items-start gap-3.5">
                    <div className="p-2.5 rounded-xl bg-emerald-50 text-emerald-600 shrink-0">
                      <Phone className="h-5 w-5" />
                    </div>
                    <div>
                      <p className="text-xs font-semibold text-brand-muted uppercase tracking-wider">
                        Telefon
                      </p>
                      <p className="text-base font-bold text-brand-navy mt-0.5">
                        {formatUzbekPhoneDisplay(profile?.phone || '')}
                      </p>
                      <p className="text-xs text-brand-muted mt-1">SMS xabarnomalar uchun</p>
                    </div>
                  </div>

                  {/* Ro‘yxatdan o‘tgan sana */}
                  <div className="flex items-start gap-3.5">
                    <div className="p-2.5 rounded-xl bg-purple-50 text-purple-600 shrink-0">
                      <Calendar className="h-5 w-5" />
                    </div>
                    <div>
                      <p className="text-xs font-semibold text-brand-muted uppercase tracking-wider">
                        A’zolik sanasi
                      </p>
                      <p className="text-base font-bold text-brand-navy mt-0.5">
                        {formatDateUz(profile?.created_at)}
                      </p>
                      <p className="text-xs text-emerald-600 font-medium mt-1">Faol akkaunt</p>
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </div>

        {/* Upcoming Appointment Section */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-lg sm:text-xl font-bold text-brand-navy">Kelgusi qabul</h2>
              <p className="text-xs sm:text-sm text-brand-muted">
                Yaqindagi rejalashtirilgan shifokor ko‘rigi
              </p>
            </div>
            <Link href="/appointments">
              <Button variant="outline" size="sm" rightIcon={<ArrowRight className="h-3.5 w-3.5" />}>
                Barcha qabullar
              </Button>
            </Link>
          </div>

          {isLoadingAppointments ? (
            <Card>
              <CardContent className="p-6">
                <Skeleton className="h-24 w-full rounded-2xl" />
              </CardContent>
            </Card>
          ) : nextAppointment ? (
            <Card className="border-blue-100 bg-gradient-to-r from-blue-50/50 via-white to-indigo-50/30 overflow-hidden">
              <CardContent className="p-6">
                <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                  <div className="flex items-start gap-4">
                    <div className="w-14 h-14 rounded-2xl bg-brand-blue text-white flex items-center justify-center font-bold text-lg shrink-0 shadow-sm">
                      <Stethoscope className="w-7 h-7" />
                    </div>
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <AppointmentStatusBadge status={nextAppointment.status} size="sm" />
                        <span className="text-xs font-bold text-brand-blue">
                          {nextAppointment.doctor?.specialty || 'Shifokor'}
                        </span>
                      </div>
                      <h3 className="text-lg font-bold text-slate-900">
                        {nextAppointment.doctor
                          ? `${nextAppointment.doctor.first_name} ${nextAppointment.doctor.last_name}`
                          : 'Shifokor'}
                      </h3>
                      <p className="text-xs text-slate-500 flex items-center gap-1.5">
                        <Building2 className="w-3.5 h-3.5 text-slate-400" />
                        <span>{nextAppointment.clinic?.name || 'Klinika'}</span>
                      </p>
                    </div>
                  </div>

                  <div className="flex sm:flex-col items-center sm:items-end justify-between w-full sm:w-auto pt-3 sm:pt-0 border-t sm:border-t-0 border-slate-100 gap-2">
                    <div className="text-left sm:text-right">
                      <div className="text-xs font-bold text-slate-800 flex items-center gap-1 sm:justify-end">
                        <Calendar className="w-3.5 h-3.5 text-brand-blue" />
                        <span>{nextAppointment.appointment_date}</span>
                      </div>
                      <div className="text-sm font-mono font-extrabold text-brand-navy flex items-center gap-1 sm:justify-end mt-0.5">
                        <Clock className="w-3.5 h-3.5 text-slate-400" />
                        <span>
                          {nextAppointment.start_time} – {nextAppointment.end_time}
                        </span>
                      </div>
                    </div>

                    <Link href={`/appointments/${nextAppointment.id}`} className="mt-2">
                      <Button variant="primary" size="sm" rightIcon={<ArrowRight className="h-3 w-3" />}>
                        Batafsil
                      </Button>
                    </Link>
                  </div>
                </div>
              </CardContent>
            </Card>
          ) : (
            <Card className="border-dashed">
              <CardContent className="p-8 text-center space-y-3">
                <CalendarDays className="w-10 h-10 text-slate-300 mx-auto" />
                <div className="space-y-1">
                  <h4 className="font-bold text-slate-800 text-sm">
                    Siz hali qabulga yozilmagansiz
                  </h4>
                  <p className="text-xs text-slate-500 max-w-sm mx-auto">
                    Klinikalar yoki mutaxassis shifokorlarni toping va o‘zingizga qulay vaqtga navbat oling.
                  </p>
                </div>
                <div className="pt-1">
                  <Link href="/doctors">
                    <Button variant="primary" size="sm" rightIcon={<Search className="h-3.5 w-3.5" />}>
                      Shifokor qidirish
                    </Button>
                  </Link>
                </div>
              </CardContent>
            </Card>
          )}
        </div>

        {/* Quick Services Grid */}
        <div className="space-y-4">
          <div>
            <h2 className="text-lg sm:text-xl font-bold text-brand-navy">
              QabulGo xizmatlari
            </h2>
            <p className="text-xs sm:text-sm text-brand-muted">
              Platforma orqali foydalanishingiz mumkin bo‘lgan tezkor imkoniyatlar
            </p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
            {quickServices.map((service) => {
              const Icon = service.icon;
              return (
                <Link key={service.title} href={service.href} className="group block h-full">
                  <Card className="h-full border hover:border-brand-blue/50 hover:shadow-subtle transition-all duration-200">
                    <CardContent className="p-6 space-y-4 flex flex-col justify-between h-full">
                      <div className="space-y-4">
                        <div className="flex items-center justify-between">
                          <div className="p-3 rounded-2xl bg-blue-50 text-brand-blue group-hover:bg-brand-blue group-hover:text-white transition-colors shrink-0">
                            <Icon className="h-6 w-6" />
                          </div>
                          <Badge variant={service.badgeVariant} size="sm">
                            {service.badge}
                          </Badge>
                        </div>

                        <div className="space-y-1.5">
                          <h3 className="font-bold text-brand-navy text-base group-hover:text-brand-blue transition-colors">
                            {service.title}
                          </h3>
                          <p className="text-xs text-brand-muted leading-relaxed">
                            {service.description}
                          </p>
                        </div>
                      </div>

                      <div className="pt-2 flex items-center gap-1.5 text-xs font-bold text-brand-blue group-hover:underline">
                        <span>O‘tish</span>
                        <ArrowRight className="h-3.5 w-3.5" />
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              );
            })}
          </div>
        </div>
      </div>
    </AppLayout>
  );
}
