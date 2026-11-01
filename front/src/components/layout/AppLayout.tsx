import React, { useState } from 'react';
import { NavLink, Link, useNavigate } from 'react-router-dom';
import {
  LayoutDashboard,
  User,
  Building2,
  Stethoscope,
  CalendarDays,
  CalendarCheck,
  LogOut,
  Menu,
  X,
  ChevronDown,
  ShieldCheck,
} from 'lucide-react';
import { useAuth } from '@/app/providers/AuthProvider';
import { BrandLogo } from '@/components/common/BrandLogo';
import { getInitials } from '@/lib/utils/formatters';
import { cn } from '@/lib/utils/cn';

export interface AppLayoutProps {
  children: React.ReactNode;
}

interface NavItem {
  name: string;
  href: string;
  icon: React.ElementType;
  isAvailable: boolean;
}

export function AppLayout({ children }: AppLayoutProps) {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isProfileMenuOpen, setIsProfileMenuOpen] = useState(false);

  const navItems: NavItem[] = [
    {
      name: 'Kabinet',
      href: '/dashboard',
      icon: LayoutDashboard,
      isAvailable: true,
    },
    {
      name: 'Profilim',
      href: '/profile',
      icon: User,
      isAvailable: true,
    },
    {
      name: 'Klinikalar',
      href: '#',
      icon: Building2,
      isAvailable: false,
    },
    {
      name: 'Shifokorlar',
      href: '#',
      icon: Stethoscope,
      isAvailable: false,
    },
    {
      name: 'Qabulga yozilish',
      href: '#',
      icon: CalendarDays,
      isAvailable: false,
    },
    {
      name: 'Mening qabullarim',
      href: '#',
      icon: CalendarCheck,
      isAvailable: false,
    },
  ];

  const handleLogout = async () => {
    setIsProfileMenuOpen(false);
    setIsMobileMenuOpen(false);
    await logout();
    navigate('/login');
  };

  const userInitials = getInitials(user?.first_name, user?.last_name);
  const userFullName = `${user?.first_name || ''} ${user?.last_name || ''}`.trim() || 'Bemor';

  return (
    <div className="min-h-screen flex flex-col bg-brand-bg">
      {/* Top Navigation Bar */}
      <header className="sticky top-0 z-40 w-full bg-white/95 backdrop-blur-md border-b border-slate-200/80 shadow-subtle">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex h-18 items-center justify-between gap-4">
            {/* Logo */}
            <div className="flex items-center gap-8">
              <BrandLogo size="md" to="/dashboard" />

              {/* Desktop Navigation Links */}
              <nav className="hidden md:flex items-center gap-1.5" aria-label="Asosiy navigatsiya">
                {navItems.map((item) => {
                  const Icon = item.icon;
                  if (!item.isAvailable) {
                    return (
                      <span
                        key={item.name}
                        className="inline-flex items-center gap-2 px-3 py-2 text-xs font-medium text-slate-400 cursor-not-allowed select-none group"
                        title="Tez kunda ishga tushadi"
                      >
                        <Icon className="h-4 w-4 text-slate-300" />
                        <span>{item.name}</span>
                        <span className="text-[10px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded bg-slate-100 text-slate-500 border border-slate-200">
                          Tez kunda
                        </span>
                      </span>
                    );
                  }

                  return (
                    <NavLink
                      key={item.name}
                      to={item.href}
                      className={({ isActive }) =>
                        cn(
                          'inline-flex items-center gap-2 px-3.5 py-2 text-sm font-medium rounded-xl transition-all duration-150',
                          isActive
                            ? 'bg-brand-blue-50 text-brand-blue font-semibold shadow-xs'
                            : 'text-brand-navy/80 hover:text-brand-navy hover:bg-slate-100/70'
                        )
                      }
                    >
                      <Icon className="h-4 w-4" />
                      <span>{item.name}</span>
                    </NavLink>
                  );
                })}
              </nav>
            </div>

            {/* Right Side: Profile Menu & Mobile Toggle */}
            <div className="flex items-center gap-3">
              {/* Profile Dropdown (Desktop) */}
              <div className="relative hidden sm:block">
                <button
                  type="button"
                  onClick={() => setIsProfileMenuOpen((prev) => !prev)}
                  className="flex items-center gap-3 p-1.5 pr-3 rounded-full hover:bg-slate-100 transition-colors border border-transparent hover:border-slate-200"
                  aria-expanded={isProfileMenuOpen}
                  aria-haspopup="true"
                >
                  <div className="h-9 w-9 rounded-full gradient-primary text-white font-bold text-xs flex items-center justify-center shadow-sm">
                    {userInitials}
                  </div>
                  <div className="text-left hidden md:block">
                    <p className="text-xs font-bold text-brand-navy leading-tight truncate max-w-[140px]">
                      {userFullName}
                    </p>
                    <p className="text-[11px] text-brand-muted leading-tight truncate max-w-[140px]">
                      {user?.email}
                    </p>
                  </div>
                  <ChevronDown className="h-4 w-4 text-brand-muted" />
                </button>

                {/* Dropdown Menu */}
                {isProfileMenuOpen && (
                  <>
                    <div
                      className="fixed inset-0 z-20"
                      onClick={() => setIsProfileMenuOpen(false)}
                    />
                    <div className="absolute right-0 mt-2 w-56 rounded-2xl bg-white shadow-elevated border border-slate-100 py-2 z-30 animate-slide-up">
                      <div className="px-4 py-2.5 border-b border-slate-100">
                        <p className="text-xs font-semibold text-brand-navy">{userFullName}</p>
                        <p className="text-xs text-brand-muted truncate">{user?.email}</p>
                        <span className="inline-block mt-1 text-[10px] font-bold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-100">
                          Bemor akkaunti
                        </span>
                      </div>

                      <div className="py-1">
                        <Link
                          to="/profile"
                          onClick={() => setIsProfileMenuOpen(false)}
                          className="flex items-center gap-2.5 px-4 py-2 text-sm text-brand-navy hover:bg-slate-50 transition-colors"
                        >
                          <User className="h-4 w-4 text-brand-muted" />
                          <span>Profil ma’lumotlari</span>
                        </Link>
                      </div>

                      <div className="pt-1 border-t border-slate-100">
                        <button
                          type="button"
                          onClick={handleLogout}
                          className="w-full flex items-center gap-2.5 px-4 py-2 text-sm text-brand-error hover:bg-red-50 transition-colors text-left"
                        >
                          <LogOut className="h-4 w-4 text-brand-error" />
                          <span>Chiqish</span>
                        </button>
                      </div>
                    </div>
                  </>
                )}
              </div>

              {/* Mobile Menu Button */}
              <button
                type="button"
                onClick={() => setIsMobileMenuOpen((prev) => !prev)}
                className="md:hidden p-2 rounded-xl text-brand-navy hover:bg-slate-100 transition-colors"
                aria-label={isMobileMenuOpen ? 'Menyuni yopish' : 'Menyuni ochish'}
              >
                {isMobileMenuOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6 text-brand-navy" />}
              </button>
            </div>
          </div>
        </div>

        {/* Mobile Navigation Drawer */}
        {isMobileMenuOpen && (
          <div className="md:hidden border-t border-slate-200 bg-white px-4 pt-3 pb-6 space-y-3 animate-fade-in">
            {/* User summary on mobile */}
            <div className="flex items-center gap-3 p-3 rounded-xl bg-brand-bg border border-slate-100">
              <div className="h-10 w-10 rounded-full gradient-primary text-white font-bold text-sm flex items-center justify-center shrink-0">
                {userInitials}
              </div>
              <div className="overflow-hidden">
                <p className="text-sm font-bold text-brand-navy truncate">{userFullName}</p>
                <p className="text-xs text-brand-muted truncate">{user?.email}</p>
              </div>
            </div>

            {/* Mobile Nav Links */}
            <nav className="space-y-1">
              {navItems.map((item) => {
                const Icon = item.icon;
                if (!item.isAvailable) {
                  return (
                    <div
                      key={item.name}
                      className="flex items-center justify-between px-3.5 py-2.5 text-sm font-medium text-slate-400 rounded-xl"
                    >
                      <div className="flex items-center gap-3">
                        <Icon className="h-5 w-5 text-slate-300" />
                        <span>{item.name}</span>
                      </div>
                      <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded bg-slate-100 text-slate-500 border border-slate-200">
                        Tez kunda
                      </span>
                    </div>
                  );
                }

                return (
                  <NavLink
                    key={item.name}
                    to={item.href}
                    onClick={() => setIsMobileMenuOpen(false)}
                    className={({ isActive }) =>
                      cn(
                        'flex items-center gap-3 px-3.5 py-2.5 text-sm font-medium rounded-xl transition-colors',
                        isActive
                          ? 'bg-brand-blue-50 text-brand-blue font-semibold'
                          : 'text-brand-navy hover:bg-slate-100'
                      )
                    }
                  >
                    <Icon className="h-5 w-5" />
                    <span>{item.name}</span>
                  </NavLink>
                );
              })}
            </nav>

            {/* Mobile Logout Button */}
            <div className="pt-3 border-t border-slate-200">
              <button
                type="button"
                onClick={handleLogout}
                className="w-full flex items-center gap-3 px-3.5 py-2.5 text-sm font-medium text-brand-error rounded-xl hover:bg-red-50 transition-colors"
              >
                <LogOut className="h-5 w-5 text-brand-error" />
                <span>Chiqish</span>
              </button>
            </div>
          </div>
        )}
      </header>

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-8">
        {children}
      </main>

      {/* Footer */}
      <footer className="mt-auto border-t border-slate-200/80 bg-white py-6">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-brand-muted text-center sm:text-left">
          <div className="flex items-center gap-2">
            <span className="font-bold text-brand-navy">QabulGo</span>
            <span>—</span>
            <span>O‘zbekistondagi xususiy klinikalar va shifokorlarga onlayn qabul platformasi</span>
          </div>
          <div className="flex items-center gap-6">
            <span className="flex items-center gap-1.5 text-emerald-700 font-medium">
              <ShieldCheck className="h-4 w-4" />
              <span>Xavfsiz tizim</span>
            </span>
            <span className="font-semibold text-brand-navy">O‘zbekcha (Lotin)</span>
          </div>
        </div>
      </footer>
    </div>
  );
}

