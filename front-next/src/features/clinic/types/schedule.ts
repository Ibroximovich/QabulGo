export interface TimeBreakItem {
  start_time: string; // "HH:mm"
  end_time: string;   // "HH:mm"
}

export interface GenerateScheduleRequest {
  weekdays: number[]; // 1=Dushanba .. 7=Yakshanba
  start_time: string;
  end_time: string;
  slot_duration_minutes: number;
  breaks: TimeBreakItem[];
}

export interface GeneratedSlotItem {
  weekday: number;
  start_time: string;
  end_time: string;
  status: 'active' | 'inactive';
}

export interface GenerateScheduleResponse {
  total_slots: number;
  slots: GeneratedSlotItem[];
}

export interface DoctorWeeklySlot {
  id: string;
  clinic_doctor_id: string;
  weekday: number;
  start_time: string;
  end_time: string;
  status: 'active' | 'inactive';
  created_at: string;
  updated_at: string;
}

export interface DoctorWeeklySchedule {
  clinic_doctor_id: string;
  doctor_id: string;
  clinic_id: string;
  slots: DoctorWeeklySlot[];
}

export interface SaveWeeklySlotItem {
  weekday: number;
  start_time: string;
  end_time: string;
  status?: string;
}

export interface SaveWeeklyScheduleRequest {
  slots: SaveWeeklySlotItem[];
}

export type OverrideStatus = 'closed' | 'available' | 'day_off';

export interface DoctorDateOverride {
  id: string;
  clinic_doctor_id: string;
  date: string; // "YYYY-MM-DD"
  start_time?: string;
  end_time?: string;
  status: OverrideStatus;
  reason?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateDateOverrideRequest {
  date: string;
  start_time?: string;
  end_time?: string;
  status: OverrideStatus;
  reason?: string;
}

export interface UpdateDateOverrideRequest {
  date?: string;
  start_time?: string;
  end_time?: string;
  status?: OverrideStatus;
  reason?: string;
}

export interface AvailabilitySlot {
  start_time: string;
  end_time: string;
}

export interface PublicAvailabilityResponse {
  date: string;
  timezone: string;
  clinic_id: string;
  doctor_id: string;
  clinic_doctor_id: string;
  slots: AvailabilitySlot[];
}

