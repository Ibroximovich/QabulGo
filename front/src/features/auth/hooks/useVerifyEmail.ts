import { useMutation } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';
import { authApi } from '../api/authApi';
import { VerifyEmailPayload, ResendVerificationPayload } from '../types';
import { useAuth } from '@/app/providers/AuthProvider';
import { normalizeApiError } from '@/lib/api/errorHandler';

export function useVerifyEmail(savedPassword?: string) {
  const navigate = useNavigate();
  const { login } = useAuth();

  const verifyMutation = useMutation({
    mutationFn: (payload: VerifyEmailPayload) => authApi.verifyEmail(payload),
    onSuccess: async (response, variables) => {
      toast.success(response.message || 'Emailingiz muvaffaqiyatli tasdiqlandi!');

      // 1. Agar backend to‘g‘ridan-to‘g‘ri token qaytargan bo‘lsa
      if (response.data && response.data.access_token) {
        login(response.data.access_token, response.data.user);
        navigate('/dashboard', { replace: true });
        return;
      }

      // 2. Agar registratsiyadan o‘tish vaqtida parol xotirada mavjud bo‘lsa -> Avtomatik Login
      if (savedPassword) {
        try {
          const loginRes = await authApi.login({
            email: variables.email,
            password: savedPassword,
          });

          if (loginRes.data) {
            login(loginRes.data.access_token, loginRes.data.user);
            toast.success('Xush kelibsiz! Shaxsiy kabinetingiz ochildi.');
            navigate('/dashboard', { replace: true });
            return;
          }
        } catch (loginErr) {
          console.warn('Avtomatik kirish bajarilmadi, login sahifasiga yo‘naltiriladi', loginErr);
        }
      }

      // 3. Parol bo‘lmasa (sahifa yangilangan bo‘lsa), login sahifasiga yo‘naltiramiz
      navigate('/login?verified=true', { replace: true });
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

