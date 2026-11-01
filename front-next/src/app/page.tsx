import type { Metadata } from 'next';
import Link from 'next/link';
import { Building2, Stethoscope, MapPin, Search, ArrowRight, ShieldCheck, HeartPulse, Clock } from 'lucide-react';
import { AppLayout } from '@/components/layout/AppLayout';

export const metadata: Metadata = {
  title: 'QabulGo — Shifokor va klinikalarga onlayn qabul',
  description:
    'O‘zbekistondagi eng yaxshi xususiy klinikalar va malakali shifokorlarni toping, eng yaqin tibbiyot markazlarini xaritada ko‘ring va qabulga yoziling.',
};

export default function RootPage() {
  return (
    <AppLayout>
      <div className="space-y-16 py-6 sm:py-10">
        {/* Hero Section */}
        <div className="relative rounded-3xl gradient-primary p-8 sm:p-14 text-white overflow-hidden shadow-elevated">
          {/* Subtle background circles */}
          <div className="absolute -right-20 -bottom-20 w-96 h-96 rounded-full bg-white/10 blur-2xl pointer-events-none" />
          <div className="absolute right-1/3 -top-20 w-72 h-72 rounded-full bg-white/5 blur-xl pointer-events-none" />

          <div className="relative z-10 max-w-2xl space-y-6">
            <div className="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-white/20 backdrop-blur-md text-xs font-bold uppercase tracking-wider text-white border border-white/30">
              <HeartPulse className="w-4 h-4 animate-pulse" />
              <span>O‘zbekistondagi yagona tibbiyot platformasi</span>
            </div>

            <h1 className="text-3xl sm:text-5xl font-extrabold tracking-tight leading-tight sm:leading-tight">
              Sog‘lig‘ingiz uchun eng yaxshi shifokor va klinikalarni toping
            </h1>

            <p className="text-white/90 text-sm sm:text-base leading-relaxed">
              Klinikalarni lokatsiya va mutaxassislik bo‘yicha qidiring, interaktiv xarita orqali eng yaqin shifoxonalarni toping va shifokorlar qabuliga qulay vaqtni tanlang.
            </p>

            {/* Quick Action Buttons */}
            <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-3 pt-4">
              <Link
                href="/clinics"
                className="inline-flex items-center justify-center gap-2 px-6 py-3.5 rounded-2xl bg-white text-brand-blue font-bold text-sm shadow-card hover:bg-slate-50 transition-all transform hover:-translate-y-0.5"
              >
                <Building2 className="w-4 h-4" />
                <span>Klinikalarni qidirish</span>
                <ArrowRight className="w-4 h-4 ml-1" />
              </Link>

              <Link
                href="/doctors"
                className="inline-flex items-center justify-center gap-2 px-6 py-3.5 rounded-2xl bg-white/20 hover:bg-white/30 backdrop-blur-md text-white font-bold text-sm border border-white/40 transition-all"
              >
                <Stethoscope className="w-4 h-4" />
                <span>Shifokorlarni topish</span>
              </Link>
            </div>
          </div>
        </div>

        {/* Feature Highlights Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <Link
            href="/clinics"
            className="group p-6 rounded-3xl bg-white border border-slate-200/80 shadow-subtle hover:border-brand-blue hover:shadow-card transition-all"
          >
            <div className="w-12 h-12 rounded-2xl bg-brand-blue-50 text-brand-blue flex items-center justify-center mb-4 group-hover:bg-brand-blue group-hover:text-white transition-colors">
              <Building2 className="w-6 h-6" />
            </div>
            <h3 className="text-lg font-bold text-brand-navy group-hover:text-brand-blue transition-colors">
              Tasdiqlangan klinikalar
            </h3>
            <p className="text-xs text-slate-500 mt-2 leading-relaxed">
              O‘zbekiston bo‘ylab rasmiy litsenziyaga ega, tekshirilgan va tasdiqlangan barcha xususiy klinikalar.
            </p>
          </Link>

          <Link
            href="/clinics"
            className="group p-6 rounded-3xl bg-white border border-slate-200/80 shadow-subtle hover:border-brand-teal hover:shadow-card transition-all"
          >
            <div className="w-12 h-12 rounded-2xl bg-teal-50 text-brand-teal flex items-center justify-center mb-4 group-hover:bg-brand-teal group-hover:text-white transition-colors">
              <MapPin className="w-6 h-6" />
            </div>
            <h3 className="text-lg font-bold text-brand-navy group-hover:text-brand-teal transition-colors">
              Interaktiv xarita va geolokatsiya
            </h3>
            <p className="text-xs text-slate-500 mt-2 leading-relaxed">
              Bemor joylashuviga ko‘ra eng yaqin klinikalarni masofasi bilan xaritada aniqlash imkoniyati.
            </p>
          </Link>

          <Link
            href="/doctors"
            className="group p-6 rounded-3xl bg-white border border-slate-200/80 shadow-subtle hover:border-brand-navy hover:shadow-card transition-all"
          >
            <div className="w-12 h-12 rounded-2xl bg-slate-100 text-brand-navy flex items-center justify-center mb-4 group-hover:bg-brand-navy group-hover:text-white transition-colors">
              <Stethoscope className="w-6 h-6" />
            </div>
            <h3 className="text-lg font-bold text-brand-navy transition-colors">
              Shaffof narxlar va qabul
            </h3>
            <p className="text-xs text-slate-500 mt-2 leading-relaxed">
              Shifokorlarning ish staji, mutaxassisligi va har bir klinikadagi qabul narxlarini oldindan ko‘rish.
            </p>
          </Link>
        </div>
      </div>
    </AppLayout>
  );
}
