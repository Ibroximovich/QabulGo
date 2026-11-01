'use client';

import { useMutation } from '@tanstack/react-query';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { authApi } from '../api/authApi';
import { ResetPasswordPayload } from '../types';
import { normalizeApiError } from '@/lib/api/errorHandler';

export function useResetPassword() {
  const router = useRouter();

  return useMutation({
    mutationFn: (payload: ResetPasswordPayload) => authApi.resetPassword(payload),
    onSuccess: (response) => {
      toast.success(
        response.message || 'Parolingiz muvaffaqiyatli yangilandi. Endi yangi parol bilan kiring.'
      );
      router.push('/login');
    },
    onError: (err) => {
      const normalized = normalizeApiError(err);
      toast.error(normalized.message);
    },
  });
}

