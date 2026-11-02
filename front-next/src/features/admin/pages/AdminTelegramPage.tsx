'use client';

import React from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Send,
  CheckCircle2,
  AlertCircle,
  ExternalLink,
  RefreshCw,
  ShieldCheck,
  Building2,
  Check,
  X,
  FileText,
  UserCheck,
} from 'lucide-react';
import { adminApi } from '../api/adminApi';
import { Button } from '@/components/ui/Button';

export function AdminTelegramPage() {
  const {
    data: statusData,
    isLoading,
    isError,
    refetch,
    isRefetching,
  } = useQuery({
    queryKey: ['adminTelegramStatus'],
    queryFn: async () => {
      const res = await adminApi.getTelegramStatus();
      return res.data?.data;
    },
    retry: 1,
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-12">
        <RefreshCw className="w-8 h-8 animate-spin text-[#0866F5]" />
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-8 bg-amber-50 rounded-3xl border border-amber-200 text-amber-900 space-y-4 max-w-2xl">
        <div className="flex items-center gap-3">
          <AlertCircle className="w-6 h-6 text-amber-600 shrink-0" />
          <h3 className="text-lg font-bold">Backend serverni qayta ishga tushiring</h3>
        </div>
        <p className="text-sm text-amber-800 leading-relaxed">
          Yangi yaratilgan Telegram API endpointi ishlashi uchun terminalda ishlab turgan Go backendni to‘xtatib (<code className="font-mono font-bold bg-amber-100 px-1.5 py-0.5 rounded">Ctrl + C</code>), qaytadan ishga tushiring:
        </p>
        <pre className="p-3 bg-amber-100/80 rounded-xl font-mono text-xs font-semibold text-amber-950">
cd D:\QabulGo\back{'\n'}go run cmd/api/main.go
        </pre>
        <Button onClick={() => refetch()} className="gap-2 bg-[#0866F5] hover:bg-blue-600 text-white">
          <RefreshCw className="w-4 h-4" />
          <span>Qayta tekshirish</span>
        </Button>
      </div>
    );
  }

  const isConfigured = Boolean(statusData?.is_configured);
  const botUsername = statusData?.bot_username || 'qabulgo_admin_bot';
  const deepLink = statusData?.deep_link || `https://t.me/${botUsername}`;
  const allowedAdminIDs = statusData?.allowed_admin_ids || [];

  return (
    <div className="space-y-6 max-w-5xl">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2 border-b border-slate-200">
        <div>
          <h1 className="text-2xl font-extrabold text-slate-900 flex items-center gap-2.5">
            <Send className="w-6 h-6 text-[#229ED9]" />
            <span>Superadmin Telegram Boti</span>
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Yangi klinika arizalari bo‘yicha bildirishnomalar va Telegram orqali bir zumda tasdiqlash / rad etish markazi.
          </p>
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={() => refetch()}
          disabled={isRefetching}
          className="gap-2 shrink-0"
        >
          <RefreshCw className={`w-4 h-4 ${isRefetching ? 'animate-spin' : ''}`} />
          <span>Yangilash</span>
        </Button>
      </div>

      {/* Main Bot Status Card */}
      <div className="bg-white rounded-3xl border border-slate-200/80 p-6 sm:p-8 shadow-sm relative overflow-hidden">
        <div className="absolute top-0 right-0 -mt-10 -mr-10 w-44 h-44 bg-blue-50 rounded-full blur-2xl pointer-events-none" />

        <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div className="space-y-3 flex-1">
            <div className="flex items-center gap-3">
              <span
                className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold border ${
                  isConfigured
                    ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                    : 'bg-amber-50 text-amber-700 border-amber-200'
                }`}
              >
                {isConfigured ? (
                  <>
                    <CheckCircle2 className="w-3.5 h-3.5" />
                    Faol va ulangan
                  </>
                ) : (
                  <>
                    <AlertCircle className="w-3.5 h-3.5" />
                    Sozlanmagan
                  </>
                )}
              </span>

              <span className="text-xs text-slate-400 font-medium">
                Bot: <b className="text-slate-800 font-mono">@{botUsername}</b>
              </span>
            </div>

            <h2 className="text-xl font-bold text-slate-900">
              QabulGo Nazorat Markazi (Superadmin Bot)
            </h2>

            <p className="text-sm text-slate-600 leading-relaxed max-w-xl">
              Yangi klinika ro‘yxatdan o‘tib, tekshirishga yuborilganda barcha rekvizitlar bilan xabar darhol ushbu botga yuboriladi. 
              Siz Telegramning o‘zida tugmani bosib klinikani tasdiqlashingiz yoki sababini ko‘rsatib rad etishingiz mumkin.
            </p>
          </div>

          <div className="shrink-0 flex flex-col sm:flex-row gap-3">
            <a
              href={deepLink}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl text-sm font-semibold text-white bg-[#229ED9] hover:bg-[#1e8cc0] shadow-sm transition-all"
            >
              <Send className="w-4 h-4" />
              <span>Telegram botni ochish</span>
              <ExternalLink className="w-3.5 h-3.5 opacity-70" />
            </a>
          </div>
        </div>
      </div>

      {/* Allowed Admin IDs Section */}
      <div className="bg-white rounded-3xl border border-slate-200/80 p-6 sm:p-8 shadow-sm space-y-4">
        <div className="flex items-center justify-between border-b border-slate-100 pb-4">
          <div className="flex items-center gap-2.5">
            <ShieldCheck className="w-5 h-5 text-[#0866F5]" />
            <div>
              <h3 className="text-base font-bold text-slate-900">Ruxsat berilgan Superadmin hisoblari</h3>
              <p className="text-xs text-slate-500 mt-0.5">
                Faqat ushbu Telegram ID egalari bot orqali klinikani tasdiqlashi yoki rad etishi mumkin
              </p>
            </div>
          </div>
          <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-blue-50 text-blue-700 border border-blue-200">
            {allowedAdminIDs.length} ta admin
          </span>
        </div>

        {allowedAdminIDs.length === 0 ? (
          <div className="p-4 rounded-2xl bg-amber-50 border border-amber-200 text-amber-800 text-sm">
            ⚠️ Hozirda hech qanday Telegram ID ko‘rsatilmagan. Backend <code className="font-mono font-bold">.env</code> faylida <code className="font-mono font-bold">TELEGRAM_ADMIN_ALLOWED_IDS</code> sozlamasiga Telegram ID raqamingizni kiriting.
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {allowedAdminIDs.map((id) => (
              <div
                key={id}
                className="flex items-center gap-3 p-3.5 rounded-2xl border border-slate-200 bg-slate-50/50 hover:bg-slate-50 transition-colors"
              >
                <div className="w-9 h-9 rounded-xl bg-blue-100 text-[#0866F5] flex items-center justify-center font-bold text-sm shrink-0">
                  <UserCheck className="w-5 h-5" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-xs text-slate-400 font-medium">Telegram ID</div>
                  <div className="text-sm font-mono font-bold text-slate-800 truncate">{id}</div>
                </div>
                <span className="inline-flex items-center gap-1 text-[11px] font-semibold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-200">
                  <Check className="w-3 h-3" /> Faol
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Capabilities & Instructions */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Capabilities */}
        <div className="bg-white rounded-3xl border border-slate-200/80 p-6 shadow-sm space-y-4">
          <h3 className="text-base font-bold text-slate-900 flex items-center gap-2">
            <Building2 className="w-4 h-4 text-[#0866F5]" />
            <span>Bot imkoniyatlari</span>
          </h3>

          <ul className="space-y-3 text-sm text-slate-600">
            <li className="flex items-start gap-2.5">
              <span className="w-5 h-5 rounded-full bg-blue-50 text-[#0866F5] flex items-center justify-center shrink-0 mt-0.5 text-xs font-bold">1</span>
              <span><b>Tezkor bildirishnoma:</b> Klinika ariza yuborgan zahoti Telegram orqali barcha rekvizitlar va mas'ul shaxs ma'lumotlari keladi.</span>
            </li>
            <li className="flex items-start gap-2.5">
              <span className="w-5 h-5 rounded-full bg-emerald-50 text-emerald-600 flex items-center justify-center shrink-0 mt-0.5 text-xs font-bold">2</span>
              <span><b>✅ Tasdiqlash tugmasi:</b> Telegram botdagi tugma orqali klinika platformada darhol faollashadi.</span>
            </li>
            <li className="flex items-start gap-2.5">
              <span className="w-5 h-5 rounded-full bg-rose-50 text-rose-600 flex items-center justify-center shrink-0 mt-0.5 text-xs font-bold">3</span>
              <span><b>❌ Rad etish tugmasi:</b> Bot sizdan rad etish sababini so‘raydi va klinika arizasi sababi bilan rad etiladi.</span>
            </li>
            <li className="flex items-start gap-2.5">
              <span className="w-5 h-5 rounded-full bg-sky-50 text-[#229ED9] flex items-center justify-center shrink-0 mt-0.5 text-xs font-bold">4</span>
              <span><b>🌐 Saytda ko‘rish:</b> Tugma orqali to‘g‘ridan-to‘g‘ri admin panelda aynan o‘sha klinika ochiladi.</span>
            </li>
          </ul>
        </div>

        {/* Commands Guide */}
        <div className="bg-white rounded-3xl border border-slate-200/80 p-6 shadow-sm space-y-4">
          <h3 className="text-base font-bold text-slate-900 flex items-center gap-2">
            <FileText className="w-4 h-4 text-[#0866F5]" />
            <span>Bot buyruqlari</span>
          </h3>

          <div className="space-y-2 text-xs">
            <div className="p-2.5 rounded-xl bg-slate-50 border border-slate-200 flex items-center justify-between">
              <code className="font-mono font-bold text-[#0866F5] text-sm">/pending</code>
              <span className="text-slate-600">Kutilayotgan barcha arizalar ro‘yxati</span>
            </div>
            <div className="p-2.5 rounded-xl bg-slate-50 border border-slate-200 flex items-center justify-between">
              <code className="font-mono font-bold text-[#0866F5] text-sm">/approved</code>
              <span className="text-slate-600">Tasdiqlangan klinikalar</span>
            </div>
            <div className="p-2.5 rounded-xl bg-slate-50 border border-slate-200 flex items-center justify-between">
              <code className="font-mono font-bold text-[#0866F5] text-sm">/start</code>
              <span className="text-slate-600">Asosiy menyu va holat</span>
            </div>
            <div className="p-2.5 rounded-xl bg-slate-50 border border-slate-200 flex items-center justify-between">
              <code className="font-mono font-bold text-[#0866F5] text-sm">/cancel</code>
              <span className="text-slate-600">Joriy amalni bekor qilish</span>
            </div>
            <div className="p-2.5 rounded-xl bg-slate-50 border border-slate-200 flex items-center justify-between">
              <code className="font-mono font-bold text-[#0866F5] text-sm">/help</code>
              <span className="text-slate-600">Yordam ma’lumotlari</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

