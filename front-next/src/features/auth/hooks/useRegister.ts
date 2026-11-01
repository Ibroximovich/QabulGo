'use client';

import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { authApi } from '../api/authApi';
import { RegisterRequestPayload } from '../types';
import { normalizeApiError } from '@/lib/api/errorHandler';

export function useRegister() {
  const router = useRouter();

  return useMutation({
    mutationFn: (payload: RegisterRequestPayload) => authApi.register(payload),
    onSuccess: (response, variables) => {
      toast.success(
        response.message || 'Muvaffaqiyatli ro‘yxatdan o‘tdingiz! Emailingizga tasdiqlash kodi yuborildi.'
      );
      // In browser, sessionStorage can also store temporary auth helper if needed
      if (typeof window !== 'undefined') {
        try {
          sessionStorage.setItem('pending_reg_pass', variables.password);
        } catch {
          // ignore
        }
      }
      router.push(`/verify-email?email=${encodeURIComponent(variables.email)}`);
    },
    onError: (err) => {
      const normalized = normalizeApiError(err);
      toast.error(normalized.message);
    },
  });
}

