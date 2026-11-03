export interface AdminDashboardSummary {
  total_clinics: number;
  approved_clinics: number;
  pending_clinics: number;
  rejected_clinics: number;
  suspended_clinics: number;
  total_patients: number;
  total_doctors: number;
  total_appointments: number;
  today_appointments: number;
  this_week_appointments: number;
  this_month_appointments: number;
  confirmed_appointments: number;
  cancelled_appointments: number;
  completed_appointments: number;
  no_show_appointments: number;
  estimated_total_revenue: number;
}

export interface DailyMetric {
  date: string;
  count: number;
  amount?: number;
}

export interface StatusDistributionItem {
  status: string;
  count: number;
}

export interface TopClinicMetric {
  clinic_id: string;
  clinic_name: string;
  bookings_count: number;
  total_revenue: number;
}

export interface AdminDashboardAnalytics {
  daily_bookings: DailyMetric[];
  weekly_bookings: DailyMetric[];
  monthly_bookings: DailyMetric[];
  clinic_growth: DailyMetric[];
  patient_growth: DailyMetric[];
  status_distribution: StatusDistributionItem[];
  top_clinics: TopClinicMetric[];
}

export interface AdminPatientListItem {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  status: string;
  created_at: string;
  appointments_count: number;
  last_appointment_date?: string;
}

export interface AdminPatientDetail {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  status: string;
  created_at: string;
  total_appointments: number;
  upcoming_appointments: number;
  completed_appointments: number;
  cancelled_appointments: number;
  recent_appointments: AdminAppointmentListItem[];
}

export interface AdminDoctorListItem {
  id: string;
  first_name: string;
  last_name: string;
  bio: string;
  experience_years: number;
  photo_url: string;
  specialties: string[];
  clinics: string[];
  clinics_count: number;
  total_appointments: number;
  created_at: string;
}

export interface AdminDoctorClinicInfo {
  clinic_id: string;
  clinic_name: string;
  clinic_doctor_id: string;
  specialty_name: string;
  consultation_price: number;
  is_active: boolean;
}

export interface AdminDoctorDetail {
  id: string;
  first_name: string;
  last_name: string;
  bio: string;
  experience_years: number;
  photo_url: string;
  created_at: string;
  clinics: AdminDoctorClinicInfo[];
  total_appointments: number;
  upcoming_appointments: number;
  completed_appointments: number;
  cancelled_appointments: number;
  recent_appointments: AdminAppointmentListItem[];
}

export interface AdminAppointmentListItem {
  id: string;
  patient_id: string;
  patient_name: string;
  patient_phone: string;
  clinic_id: string;
  clinic_name: string;
  doctor_id: string;
  doctor_name: string;
  specialty_name: string;
  appointment_date: string;
  start_time: string;
  end_time: string;
  status: string;
  consultation_price: number;
  cancellation_reason?: string;
  created_at: string;
}

export interface AdminClinicDoctorItem {
  clinic_doctor_id: string;
  doctor_id: string;
  first_name: string;
  last_name: string;
  specialty_name: string;
  consultation_price: number;
  appointment_duration_minutes: number;
  experience_years: number;
  is_active: boolean;
  photo_url: string;
}

export interface AdminClinicTopDoctor {
  doctor_id: string;
  doctor_name: string;
  specialty_name: string;
  bookings_count: number;
}

export interface AdminClinicAnalytics {
  total_appointments: number;
  today_appointments: number;
  upcoming_appointments: number;
  completed_appointments: number;
  cancelled_appointments: number;
  no_show_appointments: number;
  estimated_total_revenue: number;
  status_distribution: StatusDistributionItem[];
  top_doctors: AdminClinicTopDoctor[];
}

