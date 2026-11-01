import { useMutation } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { authApi } from '../api/authApi';
import { RegisterRequestPayload } from '../types';
import { normalizeApiError } from '@/lib/api/errorHandler';

export function useRegister() {
  const navigate = useNavigate();

  return useMutation({
    mutationFn: (payload: RegisterRequestPayload) => authApi.register(payload),
    onSuccess: (response, variables) => {
      toast.success(
        response.message || 'Muvaffaqiyatli ro‘yxatdan o‘tdingiz! Emailingizga tasdiqlash kodi yuborildi.'
      );
      // Pass email and password (in memory state) so OTP verification can perform automatic login!
      navigate(`/verify-email?email=${encodeURIComponent(variables.email)}`, {
        state: {
          email: variables.email,
          password: variables.password, // temporary in-memory router state for auto-login
          justRegistered: true,
        },
      });
    },
    onError: (err) => {
      const normalized = normalizeApiError(err);
      toast.error(normalized.message);
    },
  });
}

