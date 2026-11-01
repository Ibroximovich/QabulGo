export interface ApiResponse<T = unknown> {
  success: boolean;
  message: string;
  data?: T;
}

export interface ApiErrorResponse {
  success: boolean;
  message: string;
  errors?: Record<string, string>;
}

export interface UserSummary {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  role: string;
}

export interface RegisterResponseData {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  role: string;
  created_at: string;
}

export interface LoginData {
  access_token: string;
  token_type: string;
  expires_in: number;
  user: UserSummary;
}

export interface RefreshTokenResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
}

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

