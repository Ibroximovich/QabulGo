import { apiClient } from '@/lib/api/client';
import { ApiResponse } from '@/types/api';
import {
  GenerateScheduleRequest,
  GenerateScheduleResponse,
  DoctorWeeklySchedule,
  SaveWeeklyScheduleRequest,
  DoctorDateOverride,
  CreateDateOverrideRequest,
  UpdateDateOverrideRequest,
} from '../types';

export const scheduleApi = {
  // 1. Avtomatik slot preview hisoblash (bazaga yozilmaydi)
  generateSchedulePreview(
    clinicId: string,
    doctorId: string,
    payload: GenerateScheduleRequest
  ) {
    return apiClient.post<ApiResponse<GenerateScheduleResponse>>(
      `/clinics/${clinicId}/doctors/${doctorId}/schedule/generate`,
      payload
    );
  },

  // 2. Shifokorning haftalik jadvalini olish
  getWeeklySchedule(clinicId: string, doctorId: string) {
    return apiClient.get<ApiResponse<DoctorWeeklySchedule>>(
      `/clinics/${clinicId}/doctors/${doctorId}/schedule`
    );
  },

  // 3. Haftalik jadvalni to‘liq va atomik saqlash
  saveWeeklySchedule(
    clinicId: string,
    doctorId: string,
    payload: SaveWeeklyScheduleRequest
  ) {
    return apiClient.put<ApiResponse<DoctorWeeklySchedule>>(
      `/clinics/${clinicId}/doctors/${doctorId}/schedule`,
      payload
    );
  },

  // 4. Istisno sanalar ro‘yxatini olish
  getDateExceptions(
    clinicId: string,
    doctorId: string,
    params?: { date?: string; start_date?: string; end_date?: string }
  ) {
    return apiClient.get<ApiResponse<DoctorDateOverride[]>>(
      `/clinics/${clinicId}/doctors/${doctorId}/schedule/exceptions`,
      { params }
    );
  },

  // 5. Yangi istisno sana/slot yaratish
  createDateException(
    clinicId: string,
    doctorId: string,
    payload: CreateDateOverrideRequest
  ) {
    return apiClient.post<ApiResponse<DoctorDateOverride>>(
      `/clinics/${clinicId}/doctors/${doctorId}/schedule/exceptions`,
      payload
    );
  },

  // 6. Mavjud istisnoni tahrirlash
  updateDateException(
    clinicId: string,
    doctorId: string,
    exceptionId: string,
    payload: UpdateDateOverrideRequest
  ) {
    return apiClient.patch<ApiResponse<DoctorDateOverride>>(
      `/clinics/${clinicId}/doctors/${doctorId}/schedule/exceptions/${exceptionId}`,
      payload
    );
  },

  // 7. Istisnoni o‘chirish
  deleteDateException(clinicId: string, doctorId: string, exceptionId: string) {
    return apiClient.delete<ApiResponse<null>>(
      `/clinics/${clinicId}/doctors/${doctorId}/schedule/exceptions/${exceptionId}`
    );
  },
};

