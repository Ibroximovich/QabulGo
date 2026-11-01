import { UserSummary } from '@/types/api';

export interface RegisterRequestPayload {
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  password: string;
  confirm_password: string;
  accept_terms: boolean;
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

export interface LoginRequestPayload {
  email: string;
  password: string;
}

export interface VerifyEmailPayload {
  email: string;
  code: string;
}

export interface ResendVerificationPayload {
  email: string;
}

export interface ForgotPasswordPayload {
  email: string;
}

export interface ResetPasswordPayload {
  email: string;
  code: string;
  new_password: string;
  confirm_password: string;
}

export interface AuthState {
  user: UserSummary | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

