export type ClinicStatus = 'draft' | 'pending' | 'approved' | 'rejected' | 'suspended';

export type UserRole = 'patient' | 'clinic_owner' | 'clinic_admin' | 'doctor' | 'super_admin';

export interface WorkingHourItem {
  weekday: number; // 1 = Dushanba, 7 = Yakshanba
  opening_time: string; // HH:mm
  closing_time: string; // HH:mm
  is_closed: boolean;
}

export interface Specialty {
  id: string;
  name: string;
  slug: string;
  icon_url?: string;
}

export interface ClinicDoctorItem {
  id: string; // clinic_doctor link id
  doctor_id: string;
  clinic_id: string;
  first_name: string;
  last_name: string;
  bio: string;
  experience_years: number;
  photo_url: string;
  specialty?: Specialty;
  consultation_price: number;
  appointment_duration_minutes: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ClinicDetail {
  id: string;
  name: string;
  description: string;
  phone: string;
  email: string;
  region: string;
  district: string;
  address: string;
  logo_url: string;
  cover_image_url: string;
  status: ClinicStatus;
  rejection_reason?: string;
  owner_id: string;
  working_hours: WorkingHourItem[];
  doctors_count: number;
  doctors?: ClinicDoctorItem[];
  created_at: string;
  updated_at: string;
}

export interface ClinicSummary {
  id: string;
  name: string;
  description: string;
  phone: string;
  email: string;
  region: string;
  district: string;
  address: string;
  logo_url: string;
  cover_image_url: string;
  status: ClinicStatus;
  created_at: string;
}

export interface CreateClinicPayload {
  name: string;
  description?: string;
  phone: string;
  email?: string;
  region: string;
  district: string;
  address: string;
  logo_url?: string;
  cover_image_url?: string;
  working_hours?: WorkingHourItem[];
}

export interface UpdateClinicPayload {
  name?: string;
  description?: string;
  phone?: string;
  email?: string;
  region?: string;
  district?: string;
  address?: string;
  logo_url?: string;
  cover_image_url?: string;
  working_hours?: WorkingHourItem[];
}

export interface CreateDoctorPayload {
  first_name: string;
  last_name: string;
  specialty_id: string;
  bio?: string;
  experience_years?: number;
  photo_url?: string;
  consultation_price: number;
  appointment_duration_minutes: number;
  is_active?: boolean;
}

export interface UpdateDoctorPayload {
  first_name?: string;
  last_name?: string;
  specialty_id?: string;
  bio?: string;
  experience_years?: number;
  photo_url?: string;
  consultation_price?: number;
  appointment_duration_minutes?: number;
  is_active?: boolean;
}

export interface DoctorFilterParams {
  specialty_id?: string;
  is_active?: boolean;
  search?: string;
  page?: number;
  limit?: number;
}

export interface UploadResponse {
  url: string;
  filename: string;
  size: number;
  mime_type: string;
}

export interface VerificationReview {
  id: string;
  clinic_id: string;
  reviewed_by: string;
  decision: 'approved' | 'rejected' | 'suspended';
  reason?: string;
  reviewed_at: string;
}

export interface AdminClinicDetailsResponse {
  clinic: ClinicDetail;
  reviews: VerificationReview[];
}

