import axios, { AxiosError } from 'axios';
import { ApiErrorResponse } from '@/types/api';

export interface NormalizedError {
  message: string;
  fieldErrors?: Record<string, string>;
  statusCode?: number;
  isNetworkError: boolean;
}

export function normalizeApiError(error: unknown): NormalizedError {
  if (!axios.isAxiosError(error)) {
    if (error instanceof Error) {
      return {
        message: error.message,
        isNetworkError: false,
      };
    }
    return {
      message: 'Kutilmagan xatolik yuz berdi. Iltimos, qayta urinib ko‘ring.',
      isNetworkError: false,
    };
  }

  const axiosErr = error as AxiosError<ApiErrorResponse>;

  // Network / Connection Error
  if (!axiosErr.response) {
    if (axiosErr.code === 'ECONNABORTED' || axiosErr.message.includes('timeout')) {
      return {
        message: 'Serverga ulanish vaqti tugadi (Timeout). Internet aloqasini tekshiring.',
        isNetworkError: true,
      };
    }
    return {
      message: 'Server bilan aloqa o‘rnatib bo‘lmadi. Internet yoki backend faolligini tekshiring.',
      isNetworkError: true,
    };
  }

  const status = axiosErr.response.status;
  const data = axiosErr.response.data;

  const serverMessage = data?.message;
  const fieldErrors = data?.errors;

  switch (status) {
    case 400:
      return {
        message: serverMessage || 'Kiritilgan ma’lumotlarda xatolik mavjud.',
        fieldErrors,
        statusCode: 400,
        isNetworkError: false,
      };
    case 401:
      return {
        message: serverMessage || 'Avtorizatsiyadan o‘tilmagan yoki sessiya muddati tugagan.',
        fieldErrors,
        statusCode: 401,
        isNetworkError: false,
      };
    case 403:
      return {
        message: serverMessage || 'Ushbu amalni bajarish uchun ruxsat yo‘q.',
        fieldErrors,
        statusCode: 403,
        isNetworkError: false,
      };
    case 404:
      return {
        message: serverMessage || 'So‘ralgan ma’lumot yoki manzil topilmadi.',
        fieldErrors,
        statusCode: 404,
        isNetworkError: false,
      };
    case 409:
      return {
        message: serverMessage || 'Ushbu ma’lumot tizimda allaqachon mavjud.',
        fieldErrors,
        statusCode: 409,
        isNetworkError: false,
      };
    case 422:
      return {
        message: serverMessage || 'Kiritilgan ma’lumotlar formati noto‘g‘ri.',
        fieldErrors,
        statusCode: 422,
        isNetworkError: false,
      };
    case 429:
      return {
        message: serverMessage || 'Juda ko‘p so‘rov yuborildi. Iltimos, birozdan so‘ng urinib ko‘ring.',
        fieldErrors,
        statusCode: 429,
        isNetworkError: false,
      };
    case 500:
    case 502:
    case 503:
      return {
        message: 'Serverda ichki xatolik yuz berdi. Mutaxassislarimiz buni bartaraf etmoqda.',
        fieldErrors,
        statusCode: status,
        isNetworkError: false,
      };
    default:
      return {
        message: serverMessage || `Server xatoligi (${status}). Iltimos, qayta urinib ko‘ring.`,
        fieldErrors,
        statusCode: status,
        isNetworkError: false,
      };
  }
}

