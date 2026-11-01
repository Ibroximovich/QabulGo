import { useMutation } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { authApi } from '../api/authApi';
import { ForgotPasswordPayload } from '../types';
import { normalizeApiError } from '@/lib/api/errorHandler';

export function useForgotPassword() {
  const navigate = useNavigate();

  return useMutation({
    mutationFn: (payload: ForgotPasswordPayload) => authApi.forgotPassword(payload),
    onSuccess: (response, variables) => {
      toast.success(
        response.message || 'Agar ushbu email tizimda mavjud bo‘lsa, tiklash kodi yuborildi.'
      );
      navigate(`/reset-password?email=${encodeURIComponent(variables.email)}`);
    },
    onError: (err) => {
      const normalized = normalizeApiError(err);
      toast.error(normalized.message);
    },
  });
}

