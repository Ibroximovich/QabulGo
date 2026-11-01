import { NavLink, Outlet } from 'react-router-dom';
import { ShieldCheck, Building2, LogOut } from 'lucide-react';
import { useAuth } from '@/app/providers/AuthProvider';
import { BrandLogo } from '@/components/common/BrandLogo';
import { Button } from '@/components/ui/Button';
import { cn } from '@/lib/utils/cn';

export function AdminLayout() {
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-[#F7FAFC] flex flex-col font-sans">
      {/* Top Header */}
      <header className="sticky top-0 z-40 bg-[#08234A] text-white shadow-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <BrandLogo size="md" subtitle="Superadmin" isDark />
            <span className="hidden sm:inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-500/20 text-blue-200 border border-blue-400/30">
              <ShieldCheck className="w-3.5 h-3.5 text-blue-400" />
              Nazorat markazi
            </span>
          </div>

          <div className="flex items-center gap-4">
            <div className="text-right hidden sm:block">
              <div className="text-sm font-semibold text-white">
                {user?.first_name ? `${user.first_name} ${user.last_name || ''}` : 'Superadmin'}
              </div>
              <div className="text-xs text-blue-200/70">{user?.email || 'admin@qabulgo.uz'}</div>
            </div>

            <div className="h-6 w-px bg-blue-800" />

            <Button
              variant="ghost"
              size="sm"
              onClick={logout}
              className="text-blue-100 hover:text-white hover:bg-white/10 gap-1.5"
            >
              <LogOut className="w-4 h-4" />
              <span className="hidden sm:inline">Chiqish</span>
            </Button>
          </div>
        </div>
      </header>

      {/* Main Container */}
      <div className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 flex gap-6">
        {/* Desktop Sidebar */}
        <aside className="hidden md:block w-60 shrink-0">
          <nav className="bg-white rounded-2xl border border-slate-200 p-3 shadow-sm sticky top-24 space-y-1">
            <NavLink
              to="/admin/clinics"
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-3 px-3.5 py-2.5 rounded-xl text-sm font-medium transition-colors',
                  isActive
                    ? 'bg-[#0866F5] text-white shadow-sm'
                    : 'text-slate-700 hover:bg-slate-100'
                )
              }
            >
              <Building2 className="w-5 h-5 shrink-0" />
              <span>Klinikalar</span>
            </NavLink>
          </nav>
        </aside>

        {/* Content Area */}
        <main className="flex-1 min-w-0">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
