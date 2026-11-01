import { useQuery } from '@tanstack/react-query';
import { patientApi } from '../api/patientApi';
import { useAuth } from '@/app/providers/AuthProvider';

export const PATIENT_PROFILE_QUERY_KEY = ['patient', 'profile'];

export function usePatientProfile() {
  const { isAuthenticated } = useAuth();

  return useQuery({
    queryKey: PATIENT_PROFILE_QUERY_KEY,
    queryFn: async () => {
      const res = await patientApi.getProfile();
      return res.data;
    },
    enabled: isAuthenticated,
  });
}

