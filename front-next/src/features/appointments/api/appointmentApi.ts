import { apiClient } from '@/lib/api/client';
import { ApiResponse } from '@/types/api';
import {
  Appointment,
  CreateAppointmentRequest,
  PatientAppointmentFilter,
  ClinicAppointmentFilter,
  AppointmentListResponse,
} from '../types';

export const appointmentApi = {
  // Patient: Book appointment with optional Idempotency-Key header
  async bookAppointment(
    payload: CreateAppointmentRequest,
    idempotencyKey?: string
  ): Promise<ApiResponse<Appointment>> {
    const headers: Record<string, string> = {};
    if (idempotencyKey) {
      headers['Idempotency-Key'] = idempotencyKey;
    }
    const response = await apiClient.post<ApiResponse<Appointment>>('/appointments', payload, {
      headers,
    });
    return response.data;
  },

  // Patient: List own appointments
  async getMyAppointments(
    filter?: PatientAppointmentFilter
  ): Promise<ApiResponse<AppointmentListResponse>> {
    const response = await apiClient.get<ApiResponse<AppointmentListResponse>>('/appointments/my', {
      params: filter,
    });
    return response.data;
  },

  // Patient: Get own appointment by ID
  async getMyAppointmentById(id: string): Promise<ApiResponse<Appointment>> {
    const response = await apiClient.get<ApiResponse<Appointment>>(`/appointments/${id}`);
    return response.data;
  },

  // Patient: Cancel own appointment
  async cancelMyAppointment(
    id: string,
    reason?: string
  ): Promise<ApiResponse<Appointment>> {
    const response = await apiClient.post<ApiResponse<Appointment>>(`/appointments/${id}/cancel`, {
      reason,
    });
    return response.data;
  },

  // Clinic: List appointments for clinic
  async getClinicAppointments(
    clinicId: string,
    filter?: ClinicAppointmentFilter
  ): Promise<ApiResponse<AppointmentListResponse>> {
    const response = await apiClient.get<ApiResponse<AppointmentListResponse>>(
      `/clinics/${clinicId}/appointments`,
      { params: filter }
    );
    return response.data;
  },

  // Clinic: Get appointment detail
  async getClinicAppointmentById(
    clinicId: string,
    appointmentId: string
  ): Promise<ApiResponse<Appointment>> {
    const response = await apiClient.get<ApiResponse<Appointment>>(
      `/clinics/${clinicId}/appointments/${appointmentId}`
    );
    return response.data;
  },

  // Clinic: Cancel appointment
  async cancelClinicAppointment(
    clinicId: string,
    appointmentId: string,
    reason?: string
  ): Promise<ApiResponse<Appointment>> {
    const response = await apiClient.post<ApiResponse<Appointment>>(
      `/clinics/${clinicId}/appointments/${appointmentId}/cancel`,
      { reason }
    );
    return response.data;
  },

  // Clinic: Complete appointment
  async completeClinicAppointment(
    clinicId: string,
    appointmentId: string
  ): Promise<ApiResponse<Appointment>> {
    const response = await apiClient.post<ApiResponse<Appointment>>(
      `/clinics/${clinicId}/appointments/${appointmentId}/complete`
    );
    return response.data;
  },

  // Clinic: Mark no-show
  async noShowClinicAppointment(
    clinicId: string,
    appointmentId: string
  ): Promise<ApiResponse<Appointment>> {
    const response = await apiClient.post<ApiResponse<Appointment>>(
      `/clinics/${clinicId}/appointments/${appointmentId}/no-show`
    );
    return response.data;
  },
};
