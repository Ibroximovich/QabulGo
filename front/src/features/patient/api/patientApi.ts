import { apiClient } from '@/lib/api/client';
import { ApiResponse } from '@/types/api';
import { PatientProfile, UpdatePatientProfilePayload } from '../types';

export const patientApi = {
  async getProfile(): Promise<ApiResponse<PatientProfile>> {
    const response = await apiClient.get<ApiResponse<PatientProfile>>('/patient/me');
    return response.data;
  },

  async updateProfile(payload: UpdatePatientProfilePayload): Promise<ApiResponse<PatientProfile>> {
    const response = await apiClient.patch<ApiResponse<PatientProfile>>('/patient/me', payload);
    return response.data;
  },
};

