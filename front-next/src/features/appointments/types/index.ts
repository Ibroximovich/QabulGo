export type AppointmentStatus = 'confirmed' | 'cancelled' | 'completed' | 'no_show';

export interface AppointmentDoctor {
  id: string;
  first_name: string;
  last_name: string;
  specialty?: string;
  photo_url?: string;
}

export interface AppointmentClinic {
  id: string;
  name: string;
  address: string;
  region: string;
  district: string;
  phone?: string;
}

export interface AppointmentPatient {
  id: string;
  first_name: string;
  last_name: string;
  phone: string;
}

export interface Appointment {
  id: string;
  patient_id: string;
  clinic_id: string;
  doctor_id: string;
  clinic_doctor_id: string;
  appointment_date: string; // YYYY-MM-DD
  start_time: string; // HH:mm
  end_time: string; // HH:mm
  status: AppointmentStatus;
  consultation_price: number;
  cancellation_reason?: string | null;
  cancelled_at?: string | null;
  completed_at?: string | null;
  no_show_at?: string | null;
  created_at: string;
  updated_at: string;
  doctor?: AppointmentDoctor;
  clinic?: AppointmentClinic;
  patient?: AppointmentPatient;
}

export interface CreateAppointmentRequest {
  clinic_id: string;
  doctor_id: string;
  date: string;
  start_time: string;
  end_time: string;
}

export interface PatientAppointmentFilter {
  timeframe?: 'upcoming' | 'past' | 'cancelled';
  status?: string;
  page?: number;
  limit?: number;
}

export interface ClinicAppointmentFilter {
  date?: string;
  date_from?: string;
  date_to?: string;
  doctor_id?: string;
  status?: string;
  page?: number;
  limit?: number;
}

export interface AppointmentListResponse {
  appointments: Appointment[];
  total: number;
  page: number;
  limit: number;
}

export const APPOINTMENT_STATUS_LABELS: Record<AppointmentStatus, string> = {
  confirmed: 'Tasdiqlangan',
  cancelled: 'Bekor qilingan',
  completed: 'Yakunlangan',
  no_show: 'Bemor kelmagan',
};

export const APPOINTMENT_STATUS_COLORS: Record<AppointmentStatus, { bg: string; text: string; border: string }> = {
  confirmed: {
    bg: 'bg-emerald-50',
    text: 'text-emerald-700',
    border: 'border-emerald-200',
  },
  cancelled: {
    bg: 'bg-rose-50',
    text: 'text-rose-700',
    border: 'border-rose-200',
  },
  completed: {
    bg: 'bg-blue-50',
    text: 'text-brand-blue',
    border: 'border-blue-200',
  },
  no_show: {
    bg: 'bg-amber-50',
    text: 'text-amber-700',
    border: 'border-amber-200',
  },
};
