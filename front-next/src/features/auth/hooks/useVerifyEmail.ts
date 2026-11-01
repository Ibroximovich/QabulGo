'use client';

import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { authApi } from '../api/authApi';
import { VerifyEmailPayload, ResendVerificationPayload } from '../types';
import { useAuth } from '@/providers/AuthProvider';
import { normalizeApiError } from '@/lib/api/errorHandler';

export function useVerifyEmail(savedPassword?: string) {
  const router = useRouter();
  const { login } = useAuth();

  const verifyMutation = useMutation({
    mutationFn: (payload: VerifyEmailPayload) => authApi.verifyEmail(payload),
    onSuccess: async (response, variables) => {
      toast.success(response.message || 'Emailingiz muvaffaqiyatli tasdiqlandi!');

      // 1. Agar backend to‘g‘ridan-to‘g‘ri token qaytargan bo‘lsa
      if (response.data && response.data.access_token) {
        login(response.data.access_token, response.data.user);
        router.replace('/dashboard');
        return;
      }

      // 2. Agar registratsiyadan o‘tish vaqtida parol xotirada mavjud bo‘lsa -> Avtomatik Login
      const pwd = savedPassword || (typeof window !== 'undefined' ? sessionStorage.getItem('pending_reg_pass') : null);
      if (pwd) {
        try {
          const loginRes = await authApi.login({
            email: variables.email,
            password: pwd,
          });

          if (typeof window !== 'undefined') {
            sessionStorage.removeItem('pending_reg_pass');
          }

          if (loginRes.data) {
            login(loginRes.data.access_token, loginRes.data.user);
            toast.success('Xush kelibsiz! Shaxsiy kabinetingiz ochildi.');
            router.replace('/dashboard');
            return;
          }
        } catch (loginErr) {
          console.warn('Avtomatik kirish bajarilmadi, login sahifasiga yo‘naltiriladi', loginErr);
        }
      }

      // 3. Parol bo‘lmasa, login sahifasiga yo‘naltiramiz
      router.replace('/login?verified=true');
    },
    onError: (err) => {
      const normalized = normalizeApiError(err);
      toast.error(normalized.message);
    },
  });

  const resendMutation = useMutation({
    mutationFn: (payload: ResendVerificationPayload) => authApi.resendVerification(payload),
    onSuccess: (response) => {
      toast.success(response.message || 'Yangi tasdiqlash kodi emailingizga yuborildi.');
    },
    onError: (err) => {
      const normalized = normalizeApiError(err);
      toast.error(normalized.message);
    },
  });

  return {
    verifyEmail: verifyMutation.mutateAsync,
    isVerifying: verifyMutation.isPending,
    resendVerification: resendMutation.mutateAsync,
    isResending: resendMutation.isPending,
  };
}

