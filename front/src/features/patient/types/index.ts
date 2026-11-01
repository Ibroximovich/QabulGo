export interface PatientProfile {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  role: string;
  email_verified: boolean;
  created_at: string;
}

export interface UpdatePatientProfilePayload {
  first_name: string;
  last_name: string;
  phone: string;
}

