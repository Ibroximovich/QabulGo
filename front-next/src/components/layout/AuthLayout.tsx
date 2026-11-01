import React from 'react';
import { BrandLogo } from '@/components/common/BrandLogo';
import { ShieldCheck, CalendarCheck, Clock } from 'lucide-react';

export interface AuthLayoutProps {
  children: React.ReactNode;
  title: string;
  subtitle: string;
  badgeText?: string;
}

export function AuthLayout({
  children,
  title,
  subtitle,
  badgeText = 'Bemor portali',
}: AuthLayoutProps) {
  return (
    <div className="min-h-screen w-full flex bg-brand-bg">
      {/* Form Section */}
      <div className="w-full lg:w-1/2 flex flex-col justify-between p-6 sm:p-10 lg:p-14 overflow-y-auto">
        {/* Top Header */}
        <div className="flex items-center justify-between">
          <BrandLogo size="md" subtitle={badgeText} />
          <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-brand-blue-50 text-brand-blue border border-brand-blue-100">
            {badgeText}
          </span>
        </div>

        {/* Form Container */}
        <div className="w-full max-w-md mx-auto my-8 sm:my-12">
          <div className="mb-8">
            <h1 className="text-2xl sm:text-3xl font-extrabold text-brand-navy tracking-tight">
              {title}
            </h1>
            <p className="mt-2 text-sm sm:text-base text-brand-muted leading-relaxed">
              {subtitle}
            </p>
          </div>

          <div className="bg-white rounded-2xl p-6 sm:p-8 shadow-card border border-slate-100">
            {children}
          </div>
        </div>

        {/* Footer */}
        <div className="pt-4 border-t border-slate-200/60 flex flex-col sm:flex-row items-center justify-between text-xs text-brand-muted gap-2 text-center sm:text-left">
          <p>© {new Date().getFullYear()} QabulGo. Barcha huquqlar himoyalangan.</p>
          <div className="flex items-center gap-4">
            <span className="text-slate-400">Ishonchli va xavfsiz platforma</span>
          </div>
        </div>
      </div>

      {/* Decorative Showcase Section (Desktop only) */}
      <div className="hidden lg:flex lg:w-1/2 relative bg-brand-navy overflow-hidden p-12 lg:p-16 flex-col justify-between">
        {/* Gradient backdrop */}
        <div className="absolute inset-0 bg-gradient-to-br from-brand-blue via-brand-navy to-brand-navy opacity-95" />
        <div className="absolute -top-32 -right-32 w-96 h-96 rounded-full bg-brand-teal/20 blur-3xl pointer-events-none" />
        <div className="absolute -bottom-32 -left-32 w-96 h-96 rounded-full bg-brand-blue/30 blur-3xl pointer-events-none" />

        {/* Pattern overlay */}
        <div className="absolute inset-0 opacity-[0.03] bg-[radial-gradient(#fff_1px,transparent_1px)] [background-size:16px_16px]" />

        {/* Top badge */}
        <div className="relative z-10">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-white/10 backdrop-blur-md border border-white/20 text-white text-xs font-medium">
            <span className="h-2 w-2 rounded-full bg-brand-teal animate-ping" />
            <span>O‘zbekistonda zamonaviy sog‘liqni saqlash</span>
          </div>
        </div>

        {/* Center Content */}
        <div className="relative z-10 max-w-lg space-y-6">
          <h2 className="text-3xl xl:text-4xl font-extrabold text-white leading-tight">
            Shifokor qabuliga navbatsiz va qulay yoziling
          </h2>
          <p className="text-slate-200 text-sm sm:text-base leading-relaxed">
            QabulGo — klinikalarni topish, malakali mutaxassislar haqida to‘liq ma’lumot olish va onlayn navbat olishning zamonaviy usuli.
          </p>

          {/* Value Props */}
          <div className="space-y-4 pt-4">
            <div className="flex items-start gap-3.5 text-white">
              <div className="p-2 rounded-xl bg-white/10 border border-white/15 text-brand-teal shrink-0">
                <CalendarCheck className="h-5 w-5" />
              </div>
              <div>
                <h4 className="text-sm font-semibold">Tezkor va aniq bron</h4>
                <p className="text-xs text-slate-300 mt-0.5">
                  Shifokorning bo‘sh vaqtlarini ko‘rib, o‘zingizga mos soatni tanlang.
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3.5 text-white">
              <div className="p-2 rounded-xl bg-white/10 border border-white/15 text-brand-teal shrink-0">
                <Clock className="h-5 w-5" />
              </div>
              <div>
                <h4 className="text-sm font-semibold">Vaqtingizni tejang</h4>
                <p className="text-xs text-slate-300 mt-0.5">
                  Klinika koridorlarida soatlab navbat kutishga hojat yo‘q.
                </p>
              </div>
            </div>

            <div className="flex items-start gap-3.5 text-white">
              <div className="p-2 rounded-xl bg-white/10 border border-white/15 text-brand-teal shrink-0">
                <ShieldCheck className="h-5 w-5" />
              </div>
              <div>
                <h4 className="text-sm font-semibold">Tibbiy ma’lumotlar xavfsizligi</h4>
                <p className="text-xs text-slate-300 mt-0.5">
                  Shaxsiy kabinetingiz va ma’lumotlaringiz xalqaro xavfsizlik standartlari bilan himoyalangan.
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Bottom card */}
        <div className="relative z-10 p-5 rounded-2xl bg-white/10 backdrop-blur-md border border-white/15 text-white text-xs leading-relaxed max-w-lg">
          <div className="flex items-center gap-3 mb-2">
            <div className="h-8 w-8 rounded-full bg-brand-teal flex items-center justify-center font-bold text-brand-navy">
              ✓
            </div>
            <div>
              <p className="font-semibold text-white">1-bosqich: Bemorlar uchun profil</p>
              <p className="text-slate-300 text-[11px]">Ro‘yxatdan o‘ting va tizim imkoniyatlaridan foydalaning</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

