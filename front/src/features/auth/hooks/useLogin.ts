import { useMutation } from '@tanstack/react-query';
import { useNavigate, useLocation } from 'react-router-dom';
import { toast } from 'sonner';
import { authApi } from '../api/authApi';
import { LoginRequestPayload } from '../types';
import { useAuth } from '@/app/providers/AuthProvider';
import { normalizeApiError } from '@/lib/api/errorHandler';

export function useLogin() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login } = useAuth();

  return useMutation({
    mutationFn: (payload: LoginRequestPayload) => authApi.login(payload),
    onSuccess: (response) => {
      if (response.data) {
        login(response.data.access_token, response.data.user);
        toast.success(response.message || 'Muvaffaqiyatli tizimga kirdingiz');
        
        // Redirect to previous requested page or dashboard
        const from = (location.state as { from?: { pathname: string } })?.from?.pathname || '/dashboard';
        navigate(from, { replace: true });
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
        navigate(`/verify-email?email=${encodeURIComponent(variables.email)}`, {
          state: {
            email: variables.email,
            password: variables.password,
            justRegistered: false,
          },
        });
        return;
      }
      toast.error(normalized.message);
    },
  });
}

