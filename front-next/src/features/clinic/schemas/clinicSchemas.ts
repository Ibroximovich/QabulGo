import { z } from 'zod';
import { isValidUzbekPhone } from '@/lib/utils/phone';

export const clinicRegisterSchema = z
  .object({
    first_name: z
      .string()
      .min(2, "Ism kamida 2 ta belgidan iborat bo'lishi kerak")
      .max(50, "Ism 50 ta belgidan oshmasligi kerak"),
    last_name: z
      .string()
      .min(2, "Familiya kamida 2 ta belgidan iborat bo'lishi kerak")
      .max(50, "Familiya 50 ta belgidan oshmasligi kerak"),
    email: z
      .string()
      .email("Email formati noto'g'ri")
      .toLowerCase()
      .trim(),
    phone: z
      .string()
      .refine(
        (val) => isValidUzbekPhone(val),
        "Telefon raqami noto'g'ri. Masalan: +998 90 123 45 67"
      ),
    password: z
      .string()
      .min(8, "Parol kamida 8 ta belgidan iborat bo'lishi kerak")
      .regex(/[A-Z]/, "Parolda kamida bitta bosh harf (A-Z) bo'lishi kerak")
      .regex(/[a-z]/, "Parolda kamida bitta kichik harf (a-z) bo'lishi kerak")
      .regex(/[0-9]/, "Parolda kamida bitta raqam (0-9) bo'lishi kerak"),
    confirm_password: z.string().min(1, 'Parolni tasdiqlash kiritilishi shart'),
    accept_terms: z.literal(true, {
      errorMap: () => ({ message: 'Foydalanish shartlariga rozilik bildirish shart' }),
    }),
  })
  .refine((data) => data.password === data.confirm_password, {
    message: 'Parollar mos kelmadi',
    path: ['confirm_password'],
  });

export type ClinicRegisterFormData = z.infer<typeof clinicRegisterSchema>;

export const clinicLoginSchema = z.object({
  email: z.string().email("Email formati noto'g'ri").toLowerCase().trim(),
  password: z.string().min(1, 'Parol kiritilishi shart'),
});

export type ClinicLoginFormData = z.infer<typeof clinicLoginSchema>;

export const adminLoginSchema = z.object({
  email: z.string().email("Email formati noto'g'ri").toLowerCase().trim(),
  password: z.string().min(1, 'Parol kiritilishi shart'),
});

export type AdminLoginFormData = z.infer<typeof adminLoginSchema>;

export const clinicStep1Schema = z.object({
  name: z
    .string()
    .min(2, "Klinika nomi kamida 2 ta belgidan iborat bo'lishi kerak")
    .max(120, "Klinika nomi juda uzun"),
  phone: z
    .string()
    .refine(
      (val) => isValidUzbekPhone(val),
      "Telefon raqami noto'g'ri. Masalan: +998 71 200 11 22"
    ),
  email: z
    .string()
    .email("Email formati noto'g'ri")
    .optional()
    .or(z.literal('')),
  description: z.string().max(1500, "Tavsif 1500 ta belgidan oshmasligi kerak").optional(),
});

export const clinicStep2Schema = z.object({
  region: z.string().min(2, "Viloyat tanlanishi shart"),
  district: z.string().min(2, "Tuman tanlanishi shart"),
  address: z.string().min(5, "Aniq manzil kiritilishi shart (kamida 5 ta belgi)"),
});

export const workingHourSchema = z.object({
  weekday: z.number().min(1).max(7),
  opening_time: z.string().default('08:00'),
  closing_time: z.string().default('18:00'),
  is_closed: z.boolean().default(false),
});

export const doctorFormSchema = z.object({
  first_name: z.string().min(2, "Ism kamida 2 ta belgidan iborat bo'lishi kerak").max(50),
  last_name: z.string().min(2, "Familiya kamida 2 ta belgidan iborat bo'lishi kerak").max(50),
  specialty_id: z.string().min(1, 'Mutaxassislik tanlanishi shart'),
  bio: z.string().max(1000, "Bio 1000 ta belgidan oshmasligi kerak").optional().default(''),
  experience_years: z.coerce
    .number()
    .min(0, "Tajriba manfiy bo'lishi mumkin emas")
    .max(70, "Tajriba 70 yildan oshmasligi kerak"),
  photo_url: z.string().optional().default(''),
  consultation_price: z.coerce
    .number()
    .min(0, "Qabul narxi 0 dan kam bo'lishi mumkin emas"),
  appointment_duration_minutes: z.coerce
    .number()
    .min(5, "Qabul davomiyligi kamida 5 daqiqa bo'lishi kerak")
    .max(240, "Qabul davomiyligi 240 daqiqadan oshmasligi kerak"),
  is_active: z.boolean().default(true),
});

export type DoctorFormData = z.infer<typeof doctorFormSchema>;

export const adminReviewSchema = z.object({
  reason: z
    .string()
    .min(5, "Sabab kamida 5 ta belgidan iborat bo'lishi kerak")
    .max(500, "Sabab juda uzun"),
});

export type AdminReviewFormData = z.infer<typeof adminReviewSchema>;
