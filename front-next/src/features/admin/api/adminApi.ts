import { apiClient } from '@/lib/api/client';
import { ApiResponse, LoginData } from '@/types/api';
import {
  AdminClinicDetailsResponse,
  ClinicSummary,
} from '@/features/clinic/types';
import {
  AdminDashboardSummary,
  AdminDashboardAnalytics,
  AdminPatientListItem,
  AdminPatientDetail,
  AdminDoctorListItem,
  AdminDoctorDetail,
  AdminAppointmentListItem,
  AdminClinicDoctorItem,
  AdminClinicAnalytics,
} from '../types';

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

export interface AdminPatientListResponse {
  success: boolean;
  message: string;
  data: AdminPatientListItem[];
  meta: {
    total: number;
    page: number;
    limit: number;
  };
}

export interface AdminDoctorListResponse {
  success: boolean;
  message: string;
  data: AdminDoctorListItem[];
  meta: {
    total: number;
    page: number;
    limit: number;
  };
}

export interface AdminAppointmentListResponse {
  success: boolean;
  message: string;
  data: AdminAppointmentListItem[];
  meta: {
    total: number;
    page: number;
    limit: number;
  };
}

export const adminApi = {
  login(payload: AdminLoginPayload) {
    return apiClient.post<ApiResponse<LoginData>>('/auth/admin/login', payload);
  },

  // Dashboard Summary & Analytics
  getDashboardSummary() {
    return apiClient.get<ApiResponse<AdminDashboardSummary>>('/admin/dashboard/summary');
  },

  getDashboardAnalytics(params?: { range?: string; start_date?: string; end_date?: string }) {
    return apiClient.get<ApiResponse<AdminDashboardAnalytics>>('/admin/dashboard/analytics', { params });
  },

  // Clinics
  listClinics(params?: AdminClinicFilterParams) {
    return apiClient.get<AdminClinicListResponse>('/admin/clinics', {
      params,
    });
  },

  getClinicDetails(id: string) {
    return apiClient.get<ApiResponse<AdminClinicDetailsResponse>>(`/admin/clinics/${id}`);
  },

  getClinicDoctors(id: string) {
    return apiClient.get<ApiResponse<AdminClinicDoctorItem[]>>(`/admin/clinics/${id}/doctors`);
  },

  getClinicAppointments(id: string, params?: { status?: string; page?: number; limit?: number }) {
    return apiClient.get<AdminAppointmentListResponse>(`/admin/clinics/${id}/appointments`, { params });
  },

  getClinicAnalytics(id: string) {
    return apiClient.get<ApiResponse<AdminClinicAnalytics>>(`/admin/clinics/${id}/analytics`);
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

  // Patients
  listPatients(params?: { search?: string; page?: number; limit?: number }) {
    return apiClient.get<AdminPatientListResponse>('/admin/patients', { params });
  },

  getPatientDetails(id: string) {
    return apiClient.get<ApiResponse<AdminPatientDetail>>(`/admin/patients/${id}`);
  },

  // Doctors
  listDoctors(params?: { search?: string; specialty_id?: string; clinic_id?: string; page?: number; limit?: number }) {
    return apiClient.get<AdminDoctorListResponse>('/admin/doctors', { params });
  },

  getDoctorDetails(id: string) {
    return apiClient.get<ApiResponse<AdminDoctorDetail>>(`/admin/doctors/${id}`);
  },

  // Appointments
  listAppointments(params?: {
    clinic_id?: string;
    doctor_id?: string;
    patient_search?: string;
    status?: string;
    start_date?: string;
    end_date?: string;
    page?: number;
    limit?: number;
  }) {
    return apiClient.get<AdminAppointmentListResponse>('/admin/appointments', { params });
  },

  getAppointmentDetails(id: string) {
    return apiClient.get<ApiResponse<AdminAppointmentListItem>>(`/admin/appointments/${id}`);
  },

  // Telegram Status
  getTelegramStatus() {
    return apiClient.get<ApiResponse<AdminTelegramStatusResponse>>('/admin/telegram/status');
  },
};
