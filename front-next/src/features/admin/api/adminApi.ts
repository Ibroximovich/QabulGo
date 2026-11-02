import { apiClient } from '@/lib/api/client';
import { ApiResponse, LoginData } from '@/types/api';
import {
  AdminClinicDetailsResponse,
  ClinicSummary,
} from '@/features/clinic/types';

export interface AdminLoginPayload {
  email: string;
  password: string;
}

export interface AdminClinicFilterParams {
  status?: string;
  search?: string;
  page?: number;
  limit?: number;
}

export interface AdminClinicListResponse {
  success: boolean;
  message: string;
  data: ClinicSummary[];
  meta: {
    total: number;
    page: number;
    limit: number;
  };
}

export interface AdminTelegramStatusResponse {
  is_configured: boolean;
  bot_username: string;
  bot_name: string;
  deep_link: string;
  allowed_admin_ids: number[];
}

export const adminApi = {
  login(payload: AdminLoginPayload) {
    return apiClient.post<ApiResponse<LoginData>>('/auth/admin/login', payload);
  },

  listClinics(params?: AdminClinicFilterParams) {
    return apiClient.get<AdminClinicListResponse>('/admin/clinics', {
      params,
    });
  },

  getClinicDetails(id: string) {
    return apiClient.get<ApiResponse<AdminClinicDetailsResponse>>(`/admin/clinics/${id}`);
  },

  approveClinic(id: string) {
    return apiClient.post<ApiResponse<null>>(`/admin/clinics/${id}/approve`);
  },

  rejectClinic(id: string, reason: string) {
    return apiClient.post<ApiResponse<null>>(`/admin/clinics/${id}/reject`, { reason });
  },

  suspendClinic(id: string, reason: string) {
    return apiClient.post<ApiResponse<null>>(`/admin/clinics/${id}/suspend`, { reason });
  },

  getTelegramStatus() {
    return apiClient.get<ApiResponse<AdminTelegramStatusResponse>>('/admin/telegram/status');
  },
};

