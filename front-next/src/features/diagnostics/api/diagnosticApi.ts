import { apiClient } from '@/lib/api/client';
import { ApiResponse } from '@/types/api';
import {
  DiagnosticCategory,
  DiagnosticEquipment,
  DiagnosticService,
  DiagnosticWeeklySlot,
  DiagnosticAvailabilityResponse,
  DiagnosticAppointment,
  CreateDiagnosticAppointmentRequest,
  CreateEquipmentRequest,
  CreateDiagnosticServiceRequest,
  SaveWeeklySlotsRequest,
} from '../types';

export const diagnosticApi = {
  // Public
  getCategories: async () => {
    const res = await apiClient.get<ApiResponse<{ categories: DiagnosticCategory[]; total: number }>>(
      '/diagnostics/categories'
    );
    return res.data;
  },

  getServices: async (params?: { clinic_id?: string; category_id?: string }) => {
    const res = await apiClient.get<ApiResponse<{ services: DiagnosticService[]; total: number }>>(
      '/diagnostics/services',
      { params }
    );
    return res.data;
  },

  getService: async (serviceId: string) => {
    const res = await apiClient.get<ApiResponse<DiagnosticService>>(`/diagnostics/services/${serviceId}`);
    return res.data;
  },

  getEquipmentAvailability: async (equipmentId: string, date: string) => {
    const res = await apiClient.get<ApiResponse<DiagnosticAvailabilityResponse>>(
      `/diagnostics/equipments/${equipmentId}/availability`,
      { params: { date } }
    );
    return res.data;
  },

  // Patient Appointments
  bookAppointment: async (data: CreateDiagnosticAppointmentRequest) => {
    const res = await apiClient.post<ApiResponse<DiagnosticAppointment>>('/diagnostics/appointments', data);
    return res.data;
  },

  getMyAppointments: async (params?: { timeframe?: string; status?: string; page?: number; limit?: number }) => {
    const res = await apiClient.get<ApiResponse<{ appointments: DiagnosticAppointment[]; total: number }>>(
      '/diagnostics/appointments/my',
      { params }
    );
    return res.data;
  },

  cancelMyAppointment: async (appointmentId: string, reason?: string) => {
    const res = await apiClient.post<ApiResponse<DiagnosticAppointment>>(
      `/diagnostics/appointments/${appointmentId}/cancel`,
      { reason }
    );
    return res.data;
  },

  // Clinic Management
  getEquipments: async (clinicId: string) => {
    const res = await apiClient.get<ApiResponse<{ equipments: DiagnosticEquipment[]; total: number }>>(
      `/clinics/${clinicId}/diagnostic-equipments`
    );
    return res.data;
  },

  createEquipment: async (clinicId: string, data: CreateEquipmentRequest) => {
    const res = await apiClient.post<ApiResponse<DiagnosticEquipment>>(
      `/clinics/${clinicId}/diagnostic-equipments`,
      data
    );
    return res.data;
  },

  updateEquipment: async (clinicId: string, equipmentId: string, data: Partial<CreateEquipmentRequest>) => {
    const res = await apiClient.patch<ApiResponse<DiagnosticEquipment>>(
      `/clinics/${clinicId}/diagnostic-equipments/${equipmentId}`,
      data
    );
    return res.data;
  },

  deleteEquipment: async (clinicId: string, equipmentId: string) => {
    const res = await apiClient.delete<ApiResponse<null>>(
      `/clinics/${clinicId}/diagnostic-equipments/${equipmentId}`
    );
    return res.data;
  },

  getClinicServices: async (clinicId: string) => {
    const res = await apiClient.get<ApiResponse<{ services: DiagnosticService[]; total: number }>>(
      `/clinics/${clinicId}/diagnostic-services`
    );
    return res.data;
  },

  createClinicService: async (clinicId: string, data: CreateDiagnosticServiceRequest) => {
    const res = await apiClient.post<ApiResponse<DiagnosticService>>(
      `/clinics/${clinicId}/diagnostic-services`,
      data
    );
    return res.data;
  },

  updateClinicService: async (clinicId: string, serviceId: string, data: Partial<CreateDiagnosticServiceRequest>) => {
    const res = await apiClient.patch<ApiResponse<DiagnosticService>>(
      `/clinics/${clinicId}/diagnostic-services/${serviceId}`,
      data
    );
    return res.data;
  },

  deleteClinicService: async (clinicId: string, serviceId: string) => {
    const res = await apiClient.delete<ApiResponse<null>>(
      `/clinics/${clinicId}/diagnostic-services/${serviceId}`
    );
    return res.data;
  },

  getWeeklySchedule: async (clinicId: string, equipmentId: string) => {
    const res = await apiClient.get<ApiResponse<DiagnosticWeeklySlot[]>>(
      `/clinics/${clinicId}/diagnostic-equipments/${equipmentId}/schedule`
    );
    return res.data;
  },

  saveWeeklySchedule: async (clinicId: string, equipmentId: string, data: SaveWeeklySlotsRequest) => {
    const res = await apiClient.put<ApiResponse<DiagnosticWeeklySlot[]>>(
      `/clinics/${clinicId}/diagnostic-equipments/${equipmentId}/schedule`,
      data
    );
    return res.data;
  },

  getClinicAppointments: async (clinicId: string, params?: { date?: string; status?: string; page?: number }) => {
    const res = await apiClient.get<ApiResponse<{ appointments: DiagnosticAppointment[]; total: number }>>(
      `/clinics/${clinicId}/diagnostic-appointments`,
      { params }
    );
    return res.data;
  },

  cancelClinicAppointment: async (clinicId: string, appointmentId: string, reason?: string) => {
    const res = await apiClient.post<ApiResponse<DiagnosticAppointment>>(
      `/clinics/${clinicId}/diagnostic-appointments/${appointmentId}/cancel`,
      { reason }
    );
    return res.data;
  },

  completeClinicAppointment: async (clinicId: string, appointmentId: string) => {
    const res = await apiClient.post<ApiResponse<DiagnosticAppointment>>(
      `/clinics/${clinicId}/diagnostic-appointments/${appointmentId}/complete`
    );
    return res.data;
  },

  noShowClinicAppointment: async (clinicId: string, appointmentId: string) => {
    const res = await apiClient.post<ApiResponse<DiagnosticAppointment>>(
      `/clinics/${clinicId}/diagnostic-appointments/${appointmentId}/no-show`
    );
    return res.data;
  },
};

