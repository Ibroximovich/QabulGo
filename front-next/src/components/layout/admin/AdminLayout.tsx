'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  ShieldCheck,
  Building2,
  LogOut,
  Send,
  LayoutDashboard,
  Users,
  Stethoscope,
  CalendarCheck,
  Menu,
  X,
  ChevronRight,
} from 'lucide-react';
import { useAuth } from '@/providers/AuthProvider';
import { BrandLogo } from '@/components/common/BrandLogo';
import { Button } from '@/components/ui/Button';
import { cn } from '@/lib/utils/cn';

export interface AdminLayoutProps {
  children?: React.ReactNode;
}

const NAV_ITEMS = [
  {
    href: '/admin/dashboard',
    label: 'Dashboard',
    icon: LayoutDashboard,
  },
  {
    href: '/admin/clinics',
    label: 'Klinikalar',
    icon: Building2,
  },
  {
    href: '/admin/patients',
    label: 'Bemorlar',
    icon: Users,
  },
  {
    href: '/admin/doctors',
    label: 'Shifokorlar',
    icon: Stethoscope,
  },
  {
    href: '/admin/appointments',
    label: 'Barcha bronlar',
    icon: CalendarCheck,
  },
  {
    href: '/admin/telegram',
    label: 'Telegram bot',
    icon: Send,
    isTelegram: true,
  },
];

export function AdminLayout({ children }: AdminLayoutProps) {
  const { user, logout } = useAuth();
  const pathname = usePathname();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  // Close mobile drawer on route change
  useEffect(() => {
    setMobileMenuOpen(false);
  }, [pathname]);

  // Lock body scroll when mobile drawer is open
  useEffect(() => {
    if (mobileMenuOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [mobileMenuOpen]);

  const isActive = (href: string) => {
    if (href === '/admin/dashboard') {
      return pathname === '/admin/dashboard' || pathname === '/admin';
    }
    return pathname.startsWith(href);
  };

  return (
    <div className="min-h-screen bg-[#F8FAFC] flex flex-col font-sans overflow-x-hidden text-slate-800">
      {/* Top Header */}
      <header className="sticky top-0 z-40 bg-[#08234A] text-white shadow-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            {/* Mobile Hamburger Button */}
            <button
              type="button"
              onClick={() => setMobileMenuOpen(true)}
              className="md:hidden p-2 -ml-2 rounded-xl text-blue-100 hover:text-white hover:bg-white/10 transition-colors focus:outline-hidden"
              aria-label="Menyuni ochish"
            >
              <Menu className="w-6 h-6" />
            </button>

            <BrandLogo size="md" subtitle="Superadmin" isDark href="/admin/dashboard" />
            <span className="hidden sm:inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-500/20 text-blue-200 border border-blue-400/30">
              <ShieldCheck className="w-3.5 h-3.5 text-blue-400" />
              Nazorat markazi
            </span>
          </div>

          <div className="flex items-center gap-3 sm:gap-4">
            <div className="text-right hidden sm:block">
              <div className="text-sm font-semibold text-white">
                {user?.first_name ? `${user.first_name} ${user.last_name || ''}` : 'Super Admin'}
              </div>
              <div className="text-xs text-blue-200/70">{user?.email || 'admin@qabulgo.uz'}</div>
            </div>

            <div className="hidden sm:block h-6 w-px bg-blue-800" />

            <Button
              variant="ghost"
              size="sm"
              onClick={logout}
              className="text-blue-100 hover:text-white hover:bg-white/10 gap-1.5 text-xs sm:text-sm px-2.5 sm:px-3"
            >
              <LogOut className="w-4 h-4" />
              <span className="hidden sm:inline">Chiqish</span>
            </Button>
          </div>
        </div>
      </header>

      {/* Main Container */}
      <div className="flex-1 max-w-7xl w-full mx-auto px-3 sm:px-6 lg:px-8 py-5 sm:py-6 flex gap-6">
        {/* Desktop Sidebar (visible on md+) */}
        <aside className="hidden md:block w-60 shrink-0">
          <nav className="bg-white rounded-2xl border border-slate-200/80 p-2.5 shadow-xs sticky top-24 space-y-1">
            {NAV_ITEMS.map((item) => {
              const active = isActive(item.href);
              const Icon = item.icon;
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    'flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-sm font-medium transition-all',
                    active
                      ? 'bg-[#0866F5] text-white shadow-2xs'
                      : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
                  )}
                >
                  <Icon
                    className={cn(
                      'w-4.5 h-4.5 shrink-0',
                      item.isTelegram && !active ? 'text-[#229ED9]' : ''
                    )}
                  />
                  <span>{item.label}</span>
                </Link>
              );
            })}
          </nav>
        </aside>

        {/* Content Area */}
        <main className="flex-1 min-w-0 pb-16 md:pb-6">{children}</main>
      </div>

      {/* Mobile Drawer (Responsive Slide-in Sidebar) */}
      {mobileMenuOpen && (
        <div className="md:hidden fixed inset-0 z-50 flex">
          {/* Backdrop */}
          <div
            className="fixed inset-0 bg-slate-900/60 backdrop-blur-xs transition-opacity animate-in fade-in"
            onClick={() => setMobileMenuOpen(false)}
          />

          {/* Drawer content */}
          <div className="relative w-4/5 max-w-xs bg-white h-full shadow-2xl flex flex-col z-10 animate-in slide-in-from-left duration-200">
            {/* Drawer Header */}
            <div className="p-4 bg-[#08234A] text-white flex items-center justify-between border-b border-blue-900/40">
              <div className="flex items-center gap-2">
                <BrandLogo size="sm" isDark href="/admin/dashboard" />
                <span className="text-xs font-bold text-blue-200">Admin</span>
              </div>
              <button
                type="button"
                onClick={() => setMobileMenuOpen(false)}
                className="p-1.5 rounded-lg text-blue-200 hover:text-white hover:bg-white/10"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Admin User Info */}
            <div className="p-4 bg-slate-50 border-b border-slate-100">
              <p className="text-xs font-bold text-slate-900">
                {user?.first_name ? `${user.first_name} ${user.last_name || ''}` : 'Super Admin'}
              </p>
              <p className="text-[11px] text-slate-500 truncate">{user?.email || 'admin@qabulgo.uz'}</p>
            </div>

            {/* Navigation Links */}
            <nav className="flex-1 overflow-y-auto p-3 space-y-1">
              {NAV_ITEMS.map((item) => {
                const active = isActive(item.href);
                const Icon = item.icon;
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    onClick={() => setMobileMenuOpen(false)}
                    className={cn(
                      'flex items-center justify-between px-3.5 py-3 rounded-xl text-sm font-semibold transition-all',
                      active
                        ? 'bg-[#0866F5] text-white shadow-2xs'
                        : 'text-slate-700 hover:bg-slate-100'
                    )}
                  >
                    <div className="flex items-center gap-3">
                      <Icon
                        className={cn(
                          'w-5 h-5 shrink-0',
                          item.isTelegram && !active ? 'text-[#229ED9]' : ''
                        )}
                      />
                      <span>{item.label}</span>
                    </div>
                    <ChevronRight className="w-4 h-4 opacity-50" />
                  </Link>
                );
              })}
            </nav>

            {/* Drawer Footer with Logout */}
            <div className="p-3 border-t border-slate-100">
              <Button
                variant="outline"
                size="sm"
                onClick={logout}
                className="w-full justify-center gap-2 text-rose-600 border-rose-200 hover:bg-rose-50"
              >
                <LogOut className="w-4 h-4" />
                <span>Chiqish</span>
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Mobile Bottom Quick-Nav Bar */}
      <div className="md:hidden fixed bottom-0 inset-x-0 bg-white/95 backdrop-blur-md border-t border-slate-200 z-30 px-2 py-1.5 flex items-center justify-around shadow-lg">
        {NAV_ITEMS.map((item) => {
          const active = isActive(item.href);
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={cn(
                'flex flex-col items-center justify-center py-1 px-2 rounded-lg text-[10px] font-semibold transition-colors',
                active ? 'text-[#0866F5]' : 'text-slate-500 hover:text-slate-800'
              )}
            >
              <Icon
                className={cn(
                  'w-4.5 h-4.5 mb-0.5',
                  item.isTelegram && !active ? 'text-[#229ED9]' : ''
                )}
              />
              <span className="truncate max-w-[50px]">{item.label}</span>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
