'use client';

import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  Send,
  CheckCircle2,
  AlertCircle,
  ExternalLink,
  Bell,
  BellOff,
  Trash2,
  RefreshCw,
  ShieldCheck,
  Calendar,
  Smartphone,
  Info,
} from 'lucide-react';
import { clinicApi, ClinicTelegramRecipient } from '../api/clinicApi';
import { useClinicContext } from '@/components/layout/clinic/ClinicLayout';
import { Button } from '@/components/ui/Button';
import { Modal } from '@/components/ui/Modal';

export function ClinicTelegramPage() {
  const queryClient = useQueryClient();
  const { clinic, isLoadingClinic } = useClinicContext();
  const clinicId = clinic?.id;

  const [isLinking, setIsLinking] = useState(false);
  const [recipientToDisconnect, setRecipientToDisconnect] = useState<ClinicTelegramRecipient | null>(null);

  // Fetch status & recipients
  const {
    data: statusData,
    isLoading: isLoadingStatus,
    refetch,
    isRefetching,
  } = useQuery({
    queryKey: ['clinicTelegramStatus', clinicId],
    queryFn: async () => {
      if (!clinicId) return null;
      const res = await clinicApi.getTelegramStatus(clinicId);
      return res.data?.data || null;
    },
    enabled: Boolean(clinicId),
  });

  // Generate Linking Token Mutation
  const linkMutation = useMutation({
    mutationFn: async () => {
      if (!clinicId) throw new Error('Klinika topilmadi');
      const res = await clinicApi.generateTelegramLinkToken(clinicId);
      return res.data?.data;
    },
    onSuccess: (data) => {
      if (data?.deep_link) {
        window.open(data.deep_link, '_blank');
        toast.success('Telegram ochilmoqda. Iltimos botda "Start" tugmasini bosing!');
      }
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.error || 'Ulanish havolasini yaratishda xatolik yuz berdi');
    },
    onSettled: () => {
      setIsLinking(false);
    },
  });

  // Toggle Notifications Mutation
  const toggleMutation = useMutation({
    mutationFn: async (recipientId: string) => {
      if (!clinicId) return;
      await clinicApi.toggleTelegramNotifications(clinicId, recipientId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clinicTelegramStatus', clinicId] });
      toast.success('Xabarnoma holati muvaffaqiyatli o‘zgartirildi');
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.error || 'Xatolik yuz berdi');
    },
  });

  // Disconnect Recipient Mutation
  const disconnectMutation = useMutation({
    mutationFn: async (recipientId: string) => {
      if (!clinicId) return;
      await clinicApi.disconnectTelegramRecipient(clinicId, recipientId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clinicTelegramStatus', clinicId] });
      toast.success('Telegram qabul qiluvchi muvaffaqiyatli uzildi');
      setRecipientToDisconnect(null);
    },
    onError: (err: any) => {
      toast.error(err?.response?.data?.error || 'Ulanishni uzishda xatolik yuz berdi');
    },
  });

  const handleConnect = () => {
    setIsLinking(true);
    linkMutation.mutate();
  };

  if (isLoadingClinic || isLoadingStatus) {
    return (
      <div className="flex items-center justify-center p-12">
        <RefreshCw className="w-8 h-8 animate-spin text-brand-blue" />
      </div>
    );
  }

  const isConnected = Boolean(statusData?.is_connected);
  const recipients = statusData?.recipients || [];
  const activeRecipients = recipients.filter((r) => r.is_active);
  const botUsername = statusData?.bot_username || 'QabulGoClinicBot';

  return (
    <div className="space-y-6 max-w-5xl">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2 border-b border-slate-200">
        <div>
          <h1 className="text-2xl font-extrabold text-brand-navy flex items-center gap-2.5">
            <Send className="w-6 h-6 text-[#229ED9]" />
            <span>Telegram Bot Integratsiyasi</span>
          </h1>
          <p className="text-sm text-slate-500 mt-1">
            Yangi bronlar va bekor qilingan qabullar haqida Telegram orqali bir zumda xabar oling.
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

      {/* Main Connection Status Card */}
      <div className="bg-white rounded-3xl border border-slate-200/80 p-6 sm:p-8 shadow-subtle relative overflow-hidden">
        <div className="absolute top-0 right-0 -mt-10 -mr-10 w-40 h-40 bg-sky-50 rounded-full blur-2xl pointer-events-none" />

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
                Bot: <b className="text-slate-700 font-mono">@{botUsername}</b>
              </span>
            </div>

            <h2 className="text-xl font-bold text-brand-navy">
              {isConnected
                ? 'Klinikangiz Telegram botga ulangan'
                : 'Klinikangizni Telegram botga ulang'}
            </h2>

            <p className="text-sm text-slate-600 leading-relaxed max-w-xl">
              {isConnected
                ? 'Bemorlar qabulga yozilganida, vaqtlar o‘zgarganda yoki bekor qilinganda xabarlar bir zumda ulangan Telegram akkauntlaringizga yuboriladi.'
                : 'Telegram orqali bildirishnomalarni qabul qilish uchun quyidagi tugmani bosing va Telegram botda Start tugmasini bosing.'}
            </p>
          </div>

          <div className="shrink-0 flex flex-col sm:flex-row gap-3">
            <Button
              onClick={handleConnect}
              disabled={isLinking || linkMutation.isPending}
              className="gap-2 bg-[#229ED9] hover:bg-[#1e8cc0] text-white shadow-xs"
            >
              <Send className="w-4 h-4" />
              <span>{isConnected ? 'Yana bitta akkaunt ulash' : 'Telegram botni ulash'}</span>
              <ExternalLink className="w-3.5 h-3.5 opacity-70" />
            </Button>
          </div>
        </div>
      </div>

      {/* Connected Recipients List */}
      <div className="bg-white rounded-3xl border border-slate-200/80 p-6 sm:p-8 shadow-subtle space-y-6">
        <div className="flex items-center justify-between border-b border-slate-100 pb-4">
          <div>
            <h3 className="text-lg font-bold text-brand-navy">Ulangan Telegram akkauntlari</h3>
            <p className="text-xs text-slate-500 mt-0.5">
              Bildirishnomalarni qabul qiluvchi klinika xodimlari ro‘yxati
            </p>
          </div>
          <span className="text-xs font-semibold px-2.5 py-1 rounded-full bg-slate-100 text-slate-600">
            {activeRecipients.length} ta faol
          </span>
        </div>

        {activeRecipients.length === 0 ? (
          <div className="text-center py-8 px-4 bg-slate-50 rounded-2xl border border-dashed border-slate-200">
            <Smartphone className="w-10 h-10 text-slate-300 mx-auto mb-2" />
            <p className="text-sm font-semibold text-slate-700">Hozircha hech qanday akkaunt ulanmagan</p>
            <p className="text-xs text-slate-400 mt-1 max-w-sm mx-auto">
              «Telegram botni ulash» tugmasi orqali klinika ma’murlari va resepshn xodimlarini botga bog‘lang.
            </p>
          </div>
        ) : (
          <div className="divide-y divide-slate-100">
            {activeRecipients.map((recipient) => (
              <div
                key={recipient.id}
                className="py-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4"
              >
                <div className="flex items-center gap-3.5">
                  <div className="w-11 h-11 rounded-2xl bg-sky-50 text-[#229ED9] border border-sky-100 flex items-center justify-center font-bold text-base shrink-0">
                    {recipient.telegram_first_name ? recipient.telegram_first_name.charAt(0).toUpperCase() : 'T'}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-bold text-sm text-brand-navy">
                        {recipient.telegram_first_name || 'Telegram foydalanuvchisi'}
                      </span>
                      {recipient.telegram_username && (
                        <span className="text-xs text-slate-400">
                          @{recipient.telegram_username}
                        </span>
                      )}
                    </div>
                    <p className="text-[11px] text-slate-400 mt-0.5">
                      Chat ID: {recipient.telegram_user_id}
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-2 self-end sm:self-center">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => toggleMutation.mutate(recipient.id)}
                    disabled={toggleMutation.isPending}
                    className={`gap-1.5 text-xs ${
                      recipient.notifications_enabled
                        ? 'text-emerald-700 border-emerald-200 bg-emerald-50/50 hover:bg-emerald-100'
                        : 'text-slate-500 border-slate-200 bg-slate-50 hover:bg-slate-100'
                    }`}
                  >
                    {recipient.notifications_enabled ? (
                      <>
                        <Bell className="w-3.5 h-3.5 text-emerald-600" />
                        <span>Xabarnoma: Yoqilgan</span>
                      </>
                    ) : (
                      <>
                        <BellOff className="w-3.5 h-3.5 text-slate-400" />
                        <span>Xabarnoma: O‘chirilgan</span>
                      </>
                    )}
                  </Button>

                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setRecipientToDisconnect(recipient)}
                    className="text-red-500 hover:text-red-700 hover:bg-red-50 p-2"
                    title="Ulanishni uzish"
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Instructions & Features Card */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-white rounded-3xl border border-slate-200/80 p-6 shadow-subtle space-y-3">
          <div className="flex items-center gap-2 text-brand-navy font-bold">
            <Calendar className="w-5 h-5 text-brand-blue" />
            <span>Bot buyruqlari</span>
          </div>
          <p className="text-xs text-slate-500">
            Bot ichida quyidagi buyruqlar orqali qabullarni ko‘rishingiz mumkin:
          </p>
          <ul className="text-xs text-slate-600 space-y-1.5 pt-1">
            <li><code className="font-mono bg-slate-100 px-1.5 py-0.5 rounded text-brand-blue">/today</code> — Bugungi kun qabullari</li>
            <li><code className="font-mono bg-slate-100 px-1.5 py-0.5 rounded text-brand-blue">/tomorrow</code> — Ertangi kun qabullari</li>
            <li><code className="font-mono bg-slate-100 px-1.5 py-0.5 rounded text-brand-blue">/appointments</code> — Kelgusi barcha tasdiqlangan qabullar</li>
            <li><code className="font-mono bg-slate-100 px-1.5 py-0.5 rounded text-brand-blue">/start</code> — Bot holatini yangilash</li>
          </ul>
        </div>

        <div className="bg-white rounded-3xl border border-slate-200/80 p-6 shadow-subtle space-y-3">
          <div className="flex items-center gap-2 text-brand-navy font-bold">
            <ShieldCheck className="w-5 h-5 text-emerald-600" />
            <span>Xavfsizlik va Maxfiylik</span>
          </div>
          <p className="text-xs text-slate-600 leading-relaxed">
            Telegram xabarnomalarida faqat zarur qabul ma’lumotlari (bemor ismi, telefon raqami, shifokor va qabul vaqti) yuboriladi. Hech qanday shaxsiy tibbiy ma’lumotlar, parollar yoki tokenlar oshkor etilmaydi.
          </p>
          <p className="text-xs text-slate-400">
            Ulanish bir martalik shifrlangan xavfsiz token orqali amalga oshiriladi.
          </p>
        </div>
      </div>

      {/* Disconnect Confirmation Modal */}
      <Modal
        isOpen={Boolean(recipientToDisconnect)}
        onClose={() => setRecipientToDisconnect(null)}
        title="Ulanishni uzish"
      >
        <div className="space-y-4">
          <p className="text-sm text-slate-600">
            Haqiqatan ham <b>{recipientToDisconnect?.telegram_first_name || 'ushbu foydalanuvchi'}</b> ({recipientToDisconnect?.telegram_username ? `@${recipientToDisconnect.telegram_username}` : recipientToDisconnect?.telegram_user_id}) Telegram akkauntini klinikadan uzmoqchimisiz?
          </p>
          <p className="text-xs text-slate-500">
            Ushbu akkaunt endi yangi bronlar haqida xabar ololmaydi.
          </p>
          <div className="flex justify-end gap-3 pt-3">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setRecipientToDisconnect(null)}
            >
              Bekor qilish
            </Button>
            <Button
              variant="danger"
              size="sm"
              onClick={() => {
                if (recipientToDisconnect) {
                  disconnectMutation.mutate(recipientToDisconnect.id);
                }
              }}
              disabled={disconnectMutation.isPending}
            >
              {disconnectMutation.isPending ? 'Uzilmoqda...' : 'Ha, uzish'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}

