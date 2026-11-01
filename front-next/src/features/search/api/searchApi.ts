import { apiClient } from '@/lib/api/client';
import { ApiResponse } from '@/types/api';
import { ClinicDetail } from '@/features/clinic/types';
import {
  PublicClinicItem,
  NearbyClinicItem,
  PublicClinicFilter,
  NearbyClinicFilter,
  PublicDoctorItem,
  PublicDoctorFilter,
  PublicDoctorDetail,
  PublicClinicDoctorItem,
  RegionItem,
  SpecialtyItem,
  SearchListResponse,
  PublicAvailabilityResponse,
} from '../types';

export const searchApi = {
  // 1. Tasdiqlangan klinikalarni ommaviy qidirish va filtrlash
  async searchClinics(filter?: PublicClinicFilter): Promise<SearchListResponse<PublicClinicItem>> {
    const response = await apiClient.get<SearchListResponse<PublicClinicItem>>('/public/clinics', {
      params: filter,
    });
    return response.data;
  },

  // 2. Foydalanuvchi koordinatalari atrofidagi eng yaqin klinikalar
  async searchNearbyClinics(filter: NearbyClinicFilter): Promise<SearchListResponse<NearbyClinicItem>> {
    const response = await apiClient.get<SearchListResponse<NearbyClinicItem>>('/public/clinics/nearby', {
      params: filter,
    });
    return response.data;
  },

  // 3. Klinika profili
  async getClinicDetails(id: string): Promise<ApiResponse<ClinicDetail>> {
    const response = await apiClient.get<ApiResponse<ClinicDetail>>(`/public/clinics/${id}`);
    return response.data;
  },

  // 4. Klinikaning faol shifokorlari
  async getClinicDoctors(id: string): Promise<ApiResponse<PublicClinicDoctorItem[]>> {
    const response = await apiClient.get<ApiResponse<PublicClinicDoctorItem[]>>(`/public/clinics/${id}/doctors`);
    return response.data;
  },

  // 5. Shifokorlarni qidirish va filtrlash
  async searchDoctors(filter?: PublicDoctorFilter): Promise<SearchListResponse<PublicDoctorItem>> {
    const response = await apiClient.get<SearchListResponse<PublicDoctorItem>>('/public/doctors', {
      params: filter,
    });
    return response.data;
  },

  // 6. Shifokor profili va uning barcha klinikalardagi qabullari
  async getDoctorDetails(id: string): Promise<ApiResponse<PublicDoctorDetail>> {
    const response = await apiClient.get<ApiResponse<PublicDoctorDetail>>(`/public/doctors/${id}`);
    return response.data;
  },

  // 7. Klinikalari mavjud bo'lgan viloyat va tumanlar
  async getRegionsCatalog(): Promise<ApiResponse<RegionItem[]>> {
    const response = await apiClient.get<ApiResponse<RegionItem[]>>('/public/regions');
    return response.data;
  },

  // 8. Mutaxassisliklar katalogi
  async getSpecialties(): Promise<ApiResponse<SpecialtyItem[]>> {
    const response = await apiClient.get<ApiResponse<SpecialtyItem[]>>('/specialties');
    return response.data;
  },

  // 9. Shifokorning ma'lum sanadagi ochiq qabul vaqtlari (availability)
  async getDoctorAvailability(
    clinicId: string,
    doctorId: string,
    date: string
  ): Promise<ApiResponse<PublicAvailabilityResponse>> {
    const response = await apiClient.get<ApiResponse<PublicAvailabilityResponse>>(
      `/public/clinics/${clinicId}/doctors/${doctorId}/availability`,
      { params: { date } }
    );
    return response.data;
  },
};


