import { z } from 'zod';
import { isValidUzbekPhone } from '@/lib/utils/phone';

export const registerSchema = z
  .object({
    first_name: z
      .string()
      .min(1, 'Ism kiritilishi shart')
      .min(2, 'Ism kamida 2 ta belgidan iborat bo‘lishi kerak')
      .max(50, 'Ism ko‘pi bilan 50 ta belgidan iborat bo‘lishi kerak'),
    last_name: z
      .string()
      .min(1, 'Familiya kiritilishi shart')
      .min(2, 'Familiya kamida 2 ta belgidan iborat bo‘lishi kerak')
      .max(50, 'Familiya ko‘pi bilan 50 ta belgidan iborat bo‘lishi kerak'),
    email: z
      .string()
      .min(1, 'Email kiritilishi shart')
      .email('Email formati noto‘g‘ri (masalan: ali@example.com)')
      .max(255, 'Email juda uzun'),
    phone: z
      .string()
      .min(1, 'Telefon raqam kiritilishi shart')
      .refine(
        (val) => isValidUzbekPhone(val),
        'Telefon formati noto‘g‘ri (+998 90 123-45-67 ko‘rinishida bo‘lishi kerak)'
      ),
    password: z
      .string()
      .min(1, 'Parol kiritilishi shart')
      .min(8, 'Parol kamida 8 ta belgidan iborat bo‘lishi kerak')
      .max(128, 'Parol ko‘pi bilan 128 ta belgidan iborat bo‘lishi kerak')
      .regex(/[A-Za-z]/, 'Parolda kamida bitta harf bo‘lishi kerak')
      .regex(/[0-9]/, 'Parolda kamida bitta raqam bo‘lishi kerak'),
    confirm_password: z.string().min(1, 'Parolni tasdiqlash shart'),
    accept_terms: z.boolean().refine((val) => val === true, {
      message: 'Foydalanish shartlariga rozilik bildirilishi shart',
    }),
  })
  .refine((data) => data.password === data.confirm_password, {
    message: 'Parollar mos kelmadi',
    path: ['confirm_password'],
  });

export type RegisterFormValues = z.infer<typeof registerSchema>;

export const loginSchema = z.object({
  email: z
    .string()
    .min(1, 'Email kiritilishi shart')
    .email('Email formati noto‘g‘ri'),
  password: z.string().min(1, 'Parol kiritilishi shart'),
});

export type LoginFormValues = z.infer<typeof loginSchema>;

export const verifyEmailSchema = z.object({
  email: z.string().email('Email noto‘g‘ri'),
  code: z
    .string()
    .length(6, 'Tasdiqlash kodi 6 ta raqamdan iborat bo‘lishi kerak')
    .regex(/^\d{6}$/, 'Kod faqat raqamlardan iborat bo‘lishi kerak'),
});

export type VerifyEmailFormValues = z.infer<typeof verifyEmailSchema>;

export const forgotPasswordSchema = z.object({
  email: z
    .string()
    .min(1, 'Email kiritilishi shart')
    .email('Email formati noto‘g‘ri'),
});

export type ForgotPasswordFormValues = z.infer<typeof forgotPasswordSchema>;

export const resetPasswordSchema = z
  .object({
    email: z.string().email('Email noto‘g‘ri'),
    code: z
      .string()
      .length(6, 'Tasdiqlash kodi 6 ta raqamdan iborat bo‘lishi kerak')
      .regex(/^\d{6}$/, 'Kod faqat raqamlardan iborat bo‘lishi kerak'),
    new_password: z
      .string()
      .min(8, 'Parol kamida 8 ta belgidan iborat bo‘lishi kerak')
      .max(128, 'Parol ko‘pi bilan 128 ta belgidan iborat bo‘lishi kerak')
      .regex(/[A-Za-z]/, 'Parolda kamida bitta harf bo‘lishi kerak')
      .regex(/[0-9]/, 'Parolda kamida bitta raqam bo‘lishi kerak'),
    confirm_password: z.string().min(1, 'Parolni tasdiqlash shart'),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: 'Parollar mos kelmadi',
    path: ['confirm_password'],
  });

export type ResetPasswordFormValues = z.infer<typeof resetPasswordSchema>;

