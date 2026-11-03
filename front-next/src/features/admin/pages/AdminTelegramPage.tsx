'use client';

import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  Send,
  CheckCircle2,
  AlertCircle,
  ExternalLink,
  RefreshCw,
  ShieldCheck,
  Building2,
  Check,
  FileText,
  UserCheck,
  Trash2,
  Link as LinkIcon,
  Copy,
} from 'lucide-react';
import { adminApi } from '../api/adminApi';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';

export function AdminTelegramPage() {
  const queryClient = useQueryClient();
  const [isLinking, setIsLinking] = useState(false);
  const [showDisconnectModal, setShowDisconnectModal] = useState(false);
  const [generatedLink, setGeneratedLink] = useState<{ deepLink: string; token: string } | null>(null);

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

  // Link Token Mutation
  const linkMutation = useMutation({
    mutationFn: async () => {
      const res = await adminApi.generateTelegramLinkToken();
      return res.data?.data;
    },
    onSuccess: (data) => {
      if (data?.deep_link) {
        setGeneratedLink({ deepLink: data.deep_link, token: data.token });
        if (typeof window !== 'undefined') {
          try {
            const opened = window.open(data.deep_link, '_blank');
            if (opened) {
              toast.success('Telegram ochilmoqda. Iltimos botda "Start" tugmasini bosing!');
            } else {
              toast.info('Telegram ulanish havolasi tayyor', {
                action: {
                  label: 'Botga o‘tish',
                  onClick: () => {
                    window.location.href = data.deep_link;
                  },
                },
              });
            }
          } catch {
            // ignore window open errors
          }
        }
      }
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.error || 'Ulanish havolasini yaratishda xatolik yuz berdi');
    },
    onSettled: () => {
      setIsLinking(false);
    },
  });

  // Disconnect Mutation
  const disconnectMutation = useMutation({
    mutationFn: async () => {
      await adminApi.disconnectTelegram();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['adminTelegramStatus'] });
      toast.success('Telegram hisobingiz muvaffaqiyatli uzildi');
      setShowDisconnectModal(false);
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.error || 'Ulanishni uzishda xatolik yuz berdi');
    },
  });

  const handleConnect = () => {
    setIsLinking(true);
    linkMutation.mutate();
  };

  const copyToClipboard = (text: string) => {
    if (typeof navigator !== 'undefined') {
      navigator.clipboard.writeText(text);
      toast.success('Nusxalandi');
    }
  };

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
          Telegram API endpointi bilan aloqa o‘rnatib bo‘lmadi. Qaytadan tekshirib ko‘ring:
        </p>
        <Button onClick={() => refetch()} className="gap-2 bg-[#0866F5] hover:bg-blue-600 text-white">
          <RefreshCw className="w-4 h-4" />
          <span>Qayta tekshirish</span>
        </Button>
      </div>
    );
  }

  const isConfigured = Boolean(statusData?.is_configured);
  const isConnected = Boolean(statusData?.is_connected);
  const botUsername = statusData?.bot_username || 'qabulgo_admin_bot';
  const deepLink = statusData?.deep_link || `https://t.me/${botUsername}`;
  const connectedAdmin = statusData?.connected_admin;
  const recipients = statusData?.recipients || [];

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

      {/* Main Bot Status & Action Card */}
      <div className="bg-white rounded-3xl border border-slate-200/80 p-6 sm:p-8 shadow-sm relative overflow-hidden">
        <div className="absolute top-0 right-0 -mt-10 -mr-10 w-44 h-44 bg-blue-50 rounded-full blur-2xl pointer-events-none" />

        <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div className="space-y-3 flex-1">
            <div className="flex items-center gap-3">
              <span
                className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold border ${
                  isConnected
                    ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
                    : 'bg-amber-50 text-amber-700 border-amber-200'
                }`}
              >
                {isConnected ? (
                  <>
                    <CheckCircle2 className="w-3.5 h-3.5" />
                    Ulangan
                  </>
                ) : (
                  <>
                    <AlertCircle className="w-3.5 h-3.5" />
                    Ulanmagan
                  </>
                )}
              </span>

              <span className="text-xs text-slate-400 font-medium">
                Bot: <b className="text-slate-800 font-mono">@{botUsername}</b>
              </span>
            </div>

            <h2 className="text-xl font-bold text-slate-900">
              {isConnected
                ? 'Superadmin hisobingiz botga ulangan'
                : 'Telegram hisobingizni botga ulang'}
            </h2>

            <p className="text-sm text-slate-600 leading-relaxed max-w-xl">
              {isConnected
                ? 'Yangi klinika ro‘yxatdan o‘tib ariza yuborganida, uning rekvizitlari bilan xabar darhol Telegram botingizga yuboriladi. Botdagi tugmalar orqali klinikani tasdiqlashingiz yoki rad etishingiz mumkin.'
                : 'Klinika arizalari bo‘yicha tezkor xabarlarni olish va ularni Telegram orqali tasdiqlash uchun shaxsiy Telegram hisobingizni bir martalik xavfsiz havola orqali botga ulang.'}
            </p>
          </div>

          <div className="shrink-0 flex flex-col sm:flex-row gap-3">
            {!isConnected ? (
              <Button
                onClick={handleConnect}
                disabled={isLinking || linkMutation.isPending}
                className="gap-2 bg-[#0866F5] hover:bg-blue-600 text-white px-5 py-2.5 rounded-xl shadow-sm text-sm font-semibold"
              >
                <LinkIcon className={`w-4 h-4 ${isLinking ? 'animate-spin' : ''}`} />
                <span>{isLinking ? 'Havola tayyorlanmoqda...' : 'Telegram botni ulash'}</span>
              </Button>
            ) : (
              <div className="flex items-center gap-2">
                <a
                  href={deepLink}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center justify-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold text-white bg-[#229ED9] hover:bg-[#1e8cc0] shadow-sm transition-all"
                >
                  <Send className="w-4 h-4" />
                  <span>Botni ochish</span>
                  <ExternalLink className="w-3.5 h-3.5 opacity-70" />
                </a>

                <Button
                  variant="outline"
                  onClick={() => setShowDisconnectModal(true)}
                  className="gap-1.5 text-rose-600 hover:text-rose-700 hover:bg-rose-50 border-rose-200"
                >
                  <Trash2 className="w-4 h-4" />
                  <span>Ulanishni uzish</span>
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Connected Superadmin Profile Card */}
      {isConnected && (connectedAdmin || recipients.length > 0) && (
        <div className="bg-white rounded-3xl border border-slate-200/80 p-6 sm:p-8 shadow-sm space-y-4">
          <div className="flex items-center justify-between border-b border-slate-100 pb-4">
            <div className="flex items-center gap-2.5">
              <ShieldCheck className="w-5 h-5 text-[#0866F5]" />
              <div>
                <h3 className="text-base font-bold text-slate-900">Ulangan Superadmin hisobi</h3>
                <p className="text-xs text-slate-500 mt-0.5">
                  Ushbu Telegram hisobi bot orqali klinikani tasdiqlash va rad etish huquqiga ega
                </p>
              </div>
            </div>
            <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200">
              <Check className="w-3 h-3 inline mr-1" />
              Faol
            </span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {(connectedAdmin ? [connectedAdmin] : recipients).map((r) => (
              <div
                key={r.id || r.telegram_user_id}
                className="flex items-center gap-3 p-3.5 rounded-2xl border border-slate-200 bg-slate-50/50 hover:bg-slate-50 transition-colors"
              >
                <div className="w-10 h-10 rounded-xl bg-blue-100 text-[#0866F5] flex items-center justify-center font-bold text-sm shrink-0">
                  <UserCheck className="w-5 h-5" />
                </div>
                <div className="min-w-0 flex-1">
                  <div className="text-sm font-bold text-slate-900 truncate">
                    {r.telegram_first_name || 'Superadmin'}
                    {r.telegram_username ? (
                      <span className="text-xs text-slate-400 font-normal ml-1">(@{r.telegram_username})</span>
                    ) : null}
                  </div>
                  <div className="text-xs font-mono font-semibold text-slate-600 mt-0.5">
                    Telegram ID: <span className="text-[#0866F5]">{r.telegram_user_id}</span>
                  </div>
                  {r.created_at && (
                    <div className="text-[11px] text-slate-400 mt-0.5">
                      Ulangan: {new Date(r.created_at).toLocaleDateString('uz-UZ')}
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Manual Link Modal if Popup was blocked */}
      {generatedLink && (
        <Modal
          isOpen={Boolean(generatedLink)}
          onClose={() => setGeneratedLink(null)}
          title="Telegram Botga Ulanish"
        >
          <div className="space-y-4 pt-2">
            <p className="text-sm text-slate-600 leading-relaxed">
              Botga ulanish havolasi tayyorlandi. Havolani bosib Telegram ilovasida botni oching va <b>Start</b> tugmasini bosing:
            </p>

            <div className="p-3 bg-slate-50 border border-slate-200 rounded-xl flex items-center justify-between gap-2">
              <span className="font-mono text-xs text-slate-700 truncate">{generatedLink.deepLink}</span>
              <Button
                variant="outline"
                size="sm"
                onClick={() => copyToClipboard(generatedLink.deepLink)}
                className="gap-1.5 shrink-0"
              >
                <Copy className="w-3.5 h-3.5" />
                <span>Nusxalash</span>
              </Button>
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <Button variant="outline" onClick={() => setGeneratedLink(null)}>
                Yopish
              </Button>
              <a
                href={generatedLink.deepLink}
                target="_blank"
                rel="noopener noreferrer"
                onClick={() => setGeneratedLink(null)}
                className="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl text-sm font-semibold text-white bg-[#229ED9] hover:bg-[#1e8cc0]"
              >
                <Send className="w-4 h-4" />
                <span>Telegramda ochish</span>
              </a>
            </div>
          </div>
        </Modal>
      )}

      {/* Disconnect Confirmation Modal */}
      <Modal
        isOpen={showDisconnectModal}
        onClose={() => setShowDisconnectModal(false)}
        title="Telegram Bot ulanishini uzish"
      >
        <div className="space-y-4 pt-2">
          <p className="text-sm text-slate-600 leading-relaxed">
            Haqiqatan ham Telegram bot ulanishini uzmoqchimisiz? Uzilgandan so‘ng yangi klinika arizalari bo‘yicha Telegram xabarnomalari kelmaydi va bot orqali tasdiqlash imkoni to‘xtatiladi.
          </p>

          <div className="flex justify-end gap-2 pt-2">
            <Button
              variant="outline"
              onClick={() => setShowDisconnectModal(false)}
              disabled={disconnectMutation.isPending}
            >
              Bekor qilish
            </Button>
            <Button
              variant="primary"
              onClick={() => disconnectMutation.mutate()}
              disabled={disconnectMutation.isPending}
              className="bg-rose-600 hover:bg-rose-700 text-white"
            >
              {disconnectMutation.isPending ? 'Uzilmoqda...' : 'Ha, ulanishni uzish'}
            </Button>
          </div>
        </div>
      </Modal>

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
