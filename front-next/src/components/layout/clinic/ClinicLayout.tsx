'use client';

import React, { createContext, useContext, useState } from 'react';
import Link from 'next/link';
import { useRouter, usePathname } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import {
  LayoutDashboard,
  Users,
  Building2,
  LogOut,
  Menu,
  X,
  AlertCircle,
  CheckCircle2,
  Clock4,
  AlertTriangle,
} from 'lucide-react';
import { useAuth } from '@/providers/AuthProvider';
import { clinicApi } from '@/features/clinic/api/clinicApi';
import { ClinicDetail } from '@/features/clinic/types';
import { BrandLogo } from '@/components/common/BrandLogo';
import { Button } from '@/components/ui/Button';
import { cn } from '@/lib/utils/cn';

export interface ClinicContextType {
  clinic?: ClinicDetail;
  isLoadingClinic: boolean;
}

const ClinicContext = createContext<ClinicContextType>({
  clinic: undefined,
  isLoadingClinic: false,
});

export const useClinicContext = () => useContext(ClinicContext);

export interface ClinicLayoutProps {
  children?: React.ReactNode;
}

export function ClinicLayout({ children }: ClinicLayoutProps) {
  const { user, logout } = useAuth();
  const router = useRouter();
  const pathname = usePathname();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const { data: clinicRes, isLoading } = useQuery({
    queryKey: ['myClinic'],
    queryFn: async () => {
      try {
        const res = await clinicApi.getMyClinic();
        return res.data;
      } catch (err: unknown) {
        const axiosErr = err as { response?: { status?: number } };
        if (axiosErr.response?.status === 404) {
          if (!pathname.startsWith('/clinic/onboarding')) {
            router.replace('/clinic/onboarding');
          }
          return null;
        }
        throw err;
      }
    },
    retry: 1,
  });

  const clinic = clinicRes?.data;
  const status = clinic?.status || 'draft';

  const navItems = [
    {
      to: '/clinic/dashboard',
      label: 'Boshqaruv paneli',
      icon: LayoutDashboard,
      disabled: false,
    },
    {
      to: '/clinic/doctors',
      label: 'Shifokorlar',
      icon: Users,
      disabled: status !== 'approved',
      badge: clinic?.doctors_count !== undefined ? String(clinic.doctors_count) : undefined,
    },
    {
      to: '/clinic/profile',
      label: 'Klinika profili',
      icon: Building2,
      disabled: false,
    },
  ];

  const renderStatusBadge = () => {
    switch (status) {
      case 'approved':
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
            <CheckCircle2 className="w-3.5 h-3.5" />
            Tasdiqlangan
          </span>
        );
      case 'pending':
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200">
            <Clock4 className="w-3.5 h-3.5 animate-pulse" />
            Tekshirilmoqda
          </span>
        );
      case 'rejected':
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-rose-50 text-rose-700 border border-rose-200">
            <AlertCircle className="w-3.5 h-3.5" />
            Rad etilgan
          </span>
        );
      case 'suspended':
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-red-100 text-red-800 border border-red-300">
            <AlertTriangle className="w-3.5 h-3.5" />
            To‘xtatilgan
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-700 border border-slate-200">
            Qoralama
          </span>
        );
    }
  };

  return (
    <ClinicContext.Provider value={{ clinic, isLoadingClinic: isLoading }}>
      <div className="min-h-screen bg-[#F7FAFC] flex flex-col font-sans">
        {/* Top Header */}
        <header className="sticky top-0 z-40 bg-white border-b border-slate-200 shadow-sm">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
            <div className="flex items-center gap-4">
              <button
                type="button"
                onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                className="lg:hidden p-2 rounded-lg text-slate-600 hover:text-slate-900 hover:bg-slate-100"
                aria-label="Menyu"
              >
                {mobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
              </button>
              <BrandLogo size="md" subtitle="Klinika portali" href="/clinic/dashboard" />
            </div>

            <div className="flex items-center gap-3 sm:gap-4">
              <div className="hidden sm:flex items-center gap-2">
                {renderStatusBadge()}
              </div>

              {clinic && (
                <div className="hidden md:flex flex-col text-right">
                  <span className="text-sm font-semibold text-slate-900 line-clamp-1">
                    {clinic.name}
                  </span>
                  <span className="text-xs text-slate-500">
                    {clinic.district}, {clinic.region}
                  </span>
                </div>
              )}

              <div className="h-6 w-px bg-slate-200 hidden sm:block" />

              <div className="flex items-center gap-2">
                <div className="w-8 h-8 rounded-full bg-blue-100 text-blue-700 flex items-center justify-center font-bold text-xs uppercase border border-blue-200">
                  {user?.first_name ? user.first_name[0] : 'K'}
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={logout}
                  className="text-slate-600 hover:text-rose-600 gap-1.5"
                  title="Kabinetdan chiqish"
                >
                  <LogOut className="w-4 h-4" />
                  <span className="hidden sm:inline">Chiqish</span>
                </Button>
              </div>
            </div>
          </div>
        </header>

        {/* Main Body */}
        <div className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 flex gap-6">
          {/* Desktop Sidebar */}
          <aside className="hidden lg:block w-64 shrink-0">
            <nav className="bg-white rounded-2xl border border-slate-200 p-3 shadow-sm sticky top-24 space-y-1">
              {navItems.map((item) => {
                const Icon = item.icon;
                const isActive = pathname === item.to || pathname.startsWith(item.to + '/');
                return (
                  <Link
                    key={item.to}
                    href={item.disabled ? '#' : item.to}
                    className={cn(
                      'flex items-center justify-between px-3.5 py-2.5 rounded-xl text-sm font-medium transition-all',
                      isActive
                        ? 'bg-[#0866F5] text-white shadow-sm'
                        : item.disabled
                        ? 'text-slate-400 opacity-60 cursor-not-allowed hover:bg-transparent'
                        : 'text-slate-700 hover:bg-slate-100 hover:text-slate-900'
                    )}
                    onClick={(e) => {
                      if (item.disabled) {
                        e.preventDefault();
                      }
                    }}
                  >
                    <div className="flex items-center gap-3">
                      <Icon className="w-5 h-5 shrink-0" />
                      <span>{item.label}</span>
                    </div>
                    {item.badge && (
                      <span className="text-xs bg-slate-100 text-slate-700 px-2 py-0.5 rounded-full font-bold">
                        {item.badge}
                      </span>
                    )}
                  </Link>
                );
              })}
            </nav>

            {clinic && (
              <div className="mt-4 p-4 bg-gradient-to-br from-blue-50/70 to-teal-50/50 rounded-2xl border border-blue-100 text-xs text-slate-700 space-y-2">
                <p className="font-semibold text-slate-900">Yordam kerakmi?</p>
                <p className="text-slate-600">
                  Savollaringiz bo‘yicha QabulGo qo‘llab-quvvatlash xizmati doimo aloqada:
                </p>
                <a
                  href="tel:+998712000000"
                  className="font-semibold text-[#0866F5] hover:underline block"
                >
                  +998 (71) 200-00-00
                </a>
              </div>
            )}
          </aside>

          {/* Mobile Navigation Drawer */}
          {mobileMenuOpen && (
            <div className="lg:hidden fixed inset-0 z-50 flex">
              <div
                className="fixed inset-0 bg-slate-900/40 backdrop-blur-xs"
                onClick={() => setMobileMenuOpen(false)}
              />
              <div className="relative w-72 max-w-[80%] bg-white h-full p-5 shadow-2xl flex flex-col justify-between">
                <div>
                  <div className="flex items-center justify-between pb-4 border-b border-slate-200 mb-4">
                    <BrandLogo size="sm" subtitle="Klinika" href="/clinic/dashboard" />
                    <button
                      onClick={() => setMobileMenuOpen(false)}
                      className="p-1 rounded-md text-slate-500 hover:bg-slate-100"
                    >
                      <X className="w-5 h-5" />
                    </button>
                  </div>

                  <div className="mb-4">
                    <div className="text-xs text-slate-500 mb-1">Status:</div>
                    {renderStatusBadge()}
                  </div>

                  <nav className="space-y-1">
                    {navItems.map((item) => {
                      const Icon = item.icon;
                      const isActive = pathname === item.to || pathname.startsWith(item.to + '/');
                      return (
                        <Link
                          key={item.to}
                          href={item.disabled ? '#' : item.to}
                          onClick={(e) => {
                            if (item.disabled) {
                              e.preventDefault();
                              return;
                            }
                            setMobileMenuOpen(false);
                          }}
                          className={cn(
                            'flex items-center justify-between px-3.5 py-2.5 rounded-xl text-sm font-medium transition-colors',
                            isActive
                              ? 'bg-[#0866F5] text-white shadow-sm'
                              : item.disabled
                              ? 'text-slate-400 opacity-60 cursor-not-allowed'
                              : 'text-slate-700 hover:bg-slate-100'
                          )}
                        >
                          <div className="flex items-center gap-3">
                            <Icon className="w-5 h-5" />
                            <span>{item.label}</span>
                          </div>
                          {item.badge && (
                            <span className="text-xs bg-slate-100 text-slate-700 px-2 py-0.5 rounded-full font-bold">
                              {item.badge}
                            </span>
                          )}
                        </Link>
                      );
                    })}
                  </nav>
                </div>

                <div className="pt-4 border-t border-slate-200">
                  <Button
                    variant="outline"
                    fullWidth
                    onClick={logout}
                    className="text-rose-600 border-rose-200 hover:bg-rose-50 gap-2"
                  >
                    <LogOut className="w-4 h-4" />
                    Kabinetdan chiqish
                  </Button>
                </div>
              </div>
            </div>
          )}

          {/* Main Content Area */}
          <main className="flex-1 min-w-0">
            {status === 'pending' && (
              <div className="mb-6 p-4 rounded-2xl bg-amber-50 border border-amber-200 flex items-start gap-3.5 text-amber-900 shadow-xs">
                <Clock4 className="w-5 h-5 text-amber-600 shrink-0 mt-0.5 animate-pulse" />
                <div>
                  <h4 className="font-semibold text-sm">Klinikangiz tekshirilmoqda</h4>
                  <p className="text-xs text-amber-800 mt-0.5 leading-relaxed">
                    Arizangiz QabulGo superadmini tomonidan ko‘rib chiqilmoqda. Tasdiqlangach,
                    shifokorlar va qabullarni boshqarishingiz mumkin bo‘ladi.
                  </p>
                </div>
              </div>
            )}

            {status === 'rejected' && (
              <div className="mb-6 p-4 rounded-2xl bg-rose-50 border border-rose-200 flex items-start justify-between gap-4 text-rose-900 shadow-xs">
                <div className="flex items-start gap-3.5">
                  <AlertCircle className="w-5 h-5 text-rose-600 shrink-0 mt-0.5" />
                  <div>
                    <h4 className="font-semibold text-sm">Klinika arizasi rad etildi</h4>
                    <p className="text-xs text-rose-800 mt-1">
                      <span className="font-semibold">Sabab: </span>
                      {clinic?.rejection_reason || 'Kiritilgan ma’lumotlarda xatolik mavjud.'}
                    </p>
                  </div>
                </div>
                <Button
                  size="sm"
                  variant="outline"
                  className="shrink-0 bg-white text-rose-700 border-rose-300 hover:bg-rose-100"
                  onClick={() => router.push('/clinic/profile')}
                >
                  Tuzatish
                </Button>
              </div>
            )}

            {status === 'suspended' && (
              <div className="mb-6 p-4 rounded-2xl bg-red-100 border border-red-300 flex items-start gap-3.5 text-red-900 shadow-xs">
                <AlertTriangle className="w-5 h-5 text-red-600 shrink-0 mt-0.5" />
                <div>
                  <h4 className="font-semibold text-sm">Klinika faoliyati to‘xtatilgan</h4>
                  <p className="text-xs text-red-800 mt-0.5 leading-relaxed">
                    Platforma qoidalari yoki tekshiruv sababli klinika faoliyati vaqtinchalik
                    cheklangan. Qo‘llab-quvvatlash xizmati bilan bog‘laning.
                  </p>
                </div>
              </div>
            )}

            {status === 'draft' && clinic && (
              <div className="mb-6 p-4 rounded-2xl bg-blue-50 border border-blue-200 flex items-center justify-between gap-4 text-blue-900 shadow-xs">
                <div className="flex items-center gap-3">
                  <Building2 className="w-5 h-5 text-blue-600 shrink-0" />
                  <div>
                    <h4 className="font-semibold text-sm">Klinika qoralama holatida</h4>
                    <p className="text-xs text-blue-700">
                      Barcha ma’lumotlarni to‘ldirib, tasdiqlash uchun superadminga yuboring.
                    </p>
                  </div>
                </div>
                <Button
                  size="sm"
                  className="bg-[#0866F5] text-white hover:bg-blue-600 shrink-0"
                  onClick={() => router.push('/clinic/onboarding')}
                >
                  Davom ettirish
                </Button>
              </div>
            )}

            {children}
          </main>
        </div>
      </div>
    </ClinicContext.Provider>
  );
}
