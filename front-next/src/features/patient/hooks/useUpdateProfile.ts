import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { patientApi } from '../api/patientApi';
import { UpdatePatientProfilePayload } from '../types';
import { PATIENT_PROFILE_QUERY_KEY } from './usePatientProfile';
import { useAuth } from '@/app/providers/AuthProvider';
import { normalizeApiError } from '@/lib/api/errorHandler';

export function useUpdateProfile() {
  const queryClient = useQueryClient();
  const { updateUser } = useAuth();

  return useMutation({
    mutationFn: (payload: UpdatePatientProfilePayload) => patientApi.updateProfile(payload),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: PATIENT_PROFILE_QUERY_KEY });
      if (response.data) {
        updateUser({
          first_name: response.data.first_name,
          last_name: response.data.last_name,
          phone: response.data.phone,
        });
      }
      toast.success(response.message || 'Profil ma’lumotlari muvaffaqiyatli yangilandi');
    },
    onError: (err) => {
      const normalized = normalizeApiError(err);
      toast.error(normalized.message);
    },
  });
}

