import { WorkingHourItem } from '@/features/clinic/types';

export interface SpecialtyItem {
  id: string;
  name: string;
  slug: string;
  icon_url?: string;
  description?: string;
}

export interface PublicClinicItem {
  id: string;
  name: string;
  description: string;
  phone: string;
  email?: string;
  region: string;
  district: string;
  address: string;
  location_url?: string;
  logo_url: string;
  cover_image_url: string;
  latitude?: number;
  longitude?: number;
  working_hours?: WorkingHourItem[];
  doctors_count: number;
  specialties?: SpecialtyItem[];
  created_at?: string;
}

export interface NearbyClinicItem extends PublicClinicItem {
  distance_km: number;
}

export interface PublicClinicFilter {
  search?: string;
  region?: string;
  district?: string;
  specialty_id?: string;
  page?: number;
  limit?: number;
}

export interface NearbyClinicFilter {
  latitude: number;
  longitude: number;
  radius_km?: number;
  specialty_id?: string;
  page?: number;
  limit?: number;
}

export interface PublicDoctorClinic {
  id: string;
  name: string;
  region: string;
  district: string;
  address: string;
  phone: string;
  location_url?: string;
  logo_url: string;
  latitude?: number;
  longitude?: number;
}

export interface PublicDoctorItem {
  clinic_doctor_id: string;
  doctor_id: string;
  first_name: string;
  last_name: string;
  photo_url: string;
  bio: string;
  experience_years: number;
  specialty?: SpecialtyItem;
  consultation_price: number;
  appointment_duration_minutes: number;
  clinic: PublicDoctorClinic;
}

export interface PublicDoctorFilter {
  search?: string;
  region?: string;
  district?: string;
  specialty_id?: string;
  clinic_id?: string;
  min_price?: number;
  max_price?: number;
  sort?: 'name_asc' | 'experience_desc' | 'price_asc' | 'price_desc';
  page?: number;
  limit?: number;
}

export interface DoctorClinicWorkplace {
  clinic_doctor_id: string;
  clinic_id: string;
  clinic_name: string;
  region: string;
  district: string;
  address: string;
  phone: string;
  location_url?: string;
  logo_url: string;
  latitude?: number;
  longitude?: number;
  specialty?: SpecialtyItem;
  consultation_price: number;
  appointment_duration_minutes: number;
}

export interface PublicDoctorDetail {
  doctor_id: string;
  first_name: string;
  last_name: string;
  photo_url: string;
  bio: string;
  experience_years: number;
  primary_specialty?: SpecialtyItem;
  workplaces: DoctorClinicWorkplace[];
}

export interface RegionItem {
  region: string;
  districts: string[];
}

export interface PublicClinicDoctorItem {
  id: string; // clinic_doctor row id
  doctor_id: string;
  first_name: string;
  last_name: string;
  specialty?: SpecialtyItem;
  consultation_price: number;
  appointment_duration_minutes: number;
  is_active: boolean;
  photo_url: string;
  bio: string;
  experience_years: number;
}

export interface SearchMeta {
  total: number;
  page: number;
  limit: number;
}

export interface SearchListResponse<T> {
  success: boolean;
  message: string;
  data: T[];
  meta: SearchMeta;
}

export * from '@/features/clinic/types/schedule';


