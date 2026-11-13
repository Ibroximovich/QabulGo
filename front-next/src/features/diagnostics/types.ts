export interface DiagnosticCategory {
  id: string;
  name: string;
  slug: string;
  icon_url?: string;
  description?: string;
}

export interface DiagnosticEquipment {
  id: string;
  clinic_id: string;
  category_id: string;
  category?: DiagnosticCategory;
  name: string;
  room_number?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface DiagnosticClinicInfo {
  id: string;
  name: string;
  address: string;
  region: string;
  district: string;
  phone?: string;
}

export interface DiagnosticService {
  id: string;
  clinic_id: string;
  category_id: string;
  equipment_id: string;
  category?: DiagnosticCategory;
  equipment?: DiagnosticEquipment;
  clinic?: DiagnosticClinicInfo;
  name: string;
  price: number;
  duration_minutes: number;
  preparation_rules?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface DiagnosticWeeklySlot {
  id: string;
  equipment_id: string;
  weekday: number; // 1=Dushanba .. 7=Yakshanba
  start_time: string;
  end_time: string;
  status: 'active' | 'inactive';
}

export interface DiagnosticDateOverride {
  id: string;
  equipment_id: string;
  date: string;
  start_time?: string;
  end_time?: string;
  status: 'closed' | 'available' | 'day_off';
  reason?: string;
}

export interface DiagnosticAvailabilitySlot {
  start_time: string;
  end_time: string;
}

export interface DiagnosticAvailabilityResponse {
  date: string;
  timezone: string;
  clinic_id: string;
  equipment_id: string;
  slots: DiagnosticAvailabilitySlot[];
}

export type DiagnosticAppointmentStatus = 'confirmed' | 'cancelled' | 'completed' | 'no_show';

export interface DiagnosticAppointment {
  id: string;
  patient_id: string;
  clinic_id: string;
  service_id: string;
  equipment_id: string;
  appointment_date: string;
  start_time: string;
  end_time: string;
  price: number;
  status: DiagnosticAppointmentStatus;
  preparation_acknowledged: boolean;
  preparation_rules?: string;
  cancellation_reason?: string | null;
  cancelled_at?: string | null;
  completed_at?: string | null;
  no_show_at?: string | null;
  created_at: string;
  updated_at: string;
  service?: DiagnosticService;
  equipment?: DiagnosticEquipment;
  clinic?: DiagnosticClinicInfo;
  patient?: {
    id: string;
    first_name: string;
    last_name: string;
    phone: string;
  };
}

export interface CreateDiagnosticAppointmentRequest {
  clinic_id: string;
  service_id: string;
  equipment_id: string;
  date: string;
  start_time: string;
  end_time: string;
  preparation_acknowledged: boolean;
}

export interface CreateEquipmentRequest {
  category_id: string;
  name: string;
  room_number?: string;
  is_active?: boolean;
}

export interface CreateDiagnosticServiceRequest {
  category_id: string;
  equipment_id: string;
  name: string;
  price: number;
  duration_minutes: number;
  preparation_rules?: string;
  is_active?: boolean;
}

export interface SaveWeeklySlotsRequest {
  slots: Array<{
    weekday: number;
    start_time: string;
    end_time: string;
    status: string;
  }>;
}

