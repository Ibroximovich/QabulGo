'use client';

import { useMutation } from '@tanstack/react-query';
import { useRouter, useSearchParams } from 'next/navigation';
import { toast } from 'sonner';
import { authApi } from '../api/authApi';
import { LoginRequestPayload } from '../types';
import { useAuth } from '@/providers/AuthProvider';
import { normalizeApiError } from '@/lib/api/errorHandler';

export function useLogin() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login } = useAuth();

  return useMutation({
    mutationFn: (payload: LoginRequestPayload) => authApi.login(payload),
    onSuccess: (response) => {
      if (response.data) {
        login(response.data.access_token, response.data.user);
        toast.success(response.message || 'Muvaffaqiyatli tizimga kirdingiz');

        const from = searchParams.get('from') || '/dashboard';
        router.replace(from);
      }
    },
    onError: (err, variables) => {
      const normalized = normalizeApiError(err);
      if (
        normalized.statusCode === 403 &&
        (normalized.message.toLowerCase().includes('tasdiqlanmagan') ||
          normalized.message.toLowerCase().includes('verified'))
      ) {
        toast.error(normalized.message);
        router.push(`/verify-email?email=${encodeURIComponent(variables.email)}`);
        return;
      }
      toast.error(normalized.message);
    },
  });
}

