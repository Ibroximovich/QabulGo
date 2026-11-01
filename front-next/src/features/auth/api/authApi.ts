import { apiClient } from '@/lib/api/client';
import { ApiResponse, LoginData, RefreshTokenResponse } from '@/types/api';
import {
  RegisterRequestPayload,
  RegisterResponseData,
  LoginRequestPayload,
  VerifyEmailPayload,
  ResendVerificationPayload,
  ForgotPasswordPayload,
  ResetPasswordPayload,
} from '../types';

export const authApi = {
  async register(payload: RegisterRequestPayload): Promise<ApiResponse<RegisterResponseData>> {
    const response = await apiClient.post<ApiResponse<RegisterResponseData>>(
      '/auth/patient/register',
      payload
    );
    return response.data;
  },

  async login(payload: LoginRequestPayload): Promise<ApiResponse<LoginData>> {
    const response = await apiClient.post<ApiResponse<LoginData>>(
      '/auth/patient/login',
      payload
    );
    return response.data;
  },

  async verifyEmail(payload: VerifyEmailPayload): Promise<ApiResponse<LoginData | null>> {
    const response = await apiClient.post<ApiResponse<LoginData | null>>(
      '/auth/verify-email',
      payload
    );
    return response.data;
  },

  async resendVerification(payload: ResendVerificationPayload): Promise<ApiResponse<null>> {
    const response = await apiClient.post<ApiResponse<null>>(
      '/auth/resend-verification',
      payload
    );
    return response.data;
  },

  async refreshToken(): Promise<ApiResponse<RefreshTokenResponse>> {
    const response = await apiClient.post<ApiResponse<RefreshTokenResponse>>(
      '/auth/refresh',
      {}
    );
    return response.data;
  },

  async logout(): Promise<ApiResponse<null>> {
    const response = await apiClient.post<ApiResponse<null>>(
      '/auth/logout',
      {}
    );
    return response.data;
  },

  async forgotPassword(payload: ForgotPasswordPayload): Promise<ApiResponse<null>> {
    const response = await apiClient.post<ApiResponse<null>>(
      '/auth/forgot-password',
      payload
    );
    return response.data;
  },

  async resetPassword(payload: ResetPasswordPayload): Promise<ApiResponse<null>> {
    const response = await apiClient.post<ApiResponse<null>>(
      '/auth/reset-password',
      payload
    );
    return response.data;
  },
};

