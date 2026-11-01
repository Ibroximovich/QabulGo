import { apiClient } from '@/lib/api/client';
import { ApiResponse, LoginData, RegisterResponseData } from '@/types/api';
import {
  ClinicDetail,
  CreateClinicPayload,
  UpdateClinicPayload,
  ClinicDoctorItem,
  CreateDoctorPayload,
  UpdateDoctorPayload,
  DoctorFilterParams,
  Specialty,
  UploadResponse,
} from '../types';

export interface ClinicRegisterPayload {
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  password: string;
  confirm_password: string;
  accept_terms: boolean;
}

export interface ClinicLoginPayload {
  email: string;
  password: string;
}

export interface DoctorListResponse {
  success: boolean;
  message: string;
  data: ClinicDoctorItem[];
  meta: {
    total: number;
    page: number;
    limit: number;
  };
}

export const clinicApi = {
  // Auth
  register(payload: ClinicRegisterPayload) {
    return apiClient.post<ApiResponse<RegisterResponseData>>('/auth/clinic/register', payload);
  },

  login(payload: ClinicLoginPayload) {
    return apiClient.post<ApiResponse<LoginData>>('/auth/clinic/login', payload);
  },

  // Clinic Management
  getMyClinic() {
    return apiClient.get<ApiResponse<ClinicDetail>>('/clinics/mine');
  },

  getClinicById(id: string) {
    return apiClient.get<ApiResponse<ClinicDetail>>(`/clinics/${id}`);
  },

  createClinic(payload: CreateClinicPayload) {
    return apiClient.post<ApiResponse<ClinicDetail>>('/clinics', payload);
  },

  updateClinic(id: string, payload: UpdateClinicPayload) {
    return apiClient.patch<ApiResponse<ClinicDetail>>(`/clinics/${id}`, payload);
  },

  submitClinic(id: string) {
    return apiClient.post<ApiResponse<{ status: string }>>(`/clinics/${id}/submit`);
  },

  // Specialties
  getSpecialties() {
    return apiClient.get<ApiResponse<Specialty[]>>('/specialties');
  },

  // Doctor Management
  listDoctors(clinicId: string, params?: DoctorFilterParams) {
    return apiClient.get<DoctorListResponse>(`/clinics/${clinicId}/doctors`, {
      params,
    });
  },

  getDoctor(clinicId: string, doctorId: string) {
    return apiClient.get<ApiResponse<ClinicDoctorItem>>(`/clinics/${clinicId}/doctors/${doctorId}`);
  },

  createDoctor(clinicId: string, payload: CreateDoctorPayload) {
    return apiClient.post<ApiResponse<ClinicDoctorItem>>(`/clinics/${clinicId}/doctors`, payload);
  },

  updateDoctor(clinicId: string, doctorId: string, payload: UpdateDoctorPayload) {
    return apiClient.patch<ApiResponse<ClinicDoctorItem>>(
      `/clinics/${clinicId}/doctors/${doctorId}`,
      payload
    );
  },

  deactivateDoctor(clinicId: string, doctorId: string) {
    return apiClient.delete<ApiResponse<null>>(`/clinics/${clinicId}/doctors/${doctorId}`);
  },

  // File Upload
  async uploadImage(file: File, folder: 'clinics' | 'doctors' = 'clinics') {
    const formData = new FormData();
    formData.append('file', file);
    const res = await apiClient.post<ApiResponse<UploadResponse>>(`/upload?folder=${folder}`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
    return res.data;
  },
};

