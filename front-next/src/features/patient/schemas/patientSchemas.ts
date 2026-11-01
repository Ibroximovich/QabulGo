import { z } from 'zod';
import { isValidUzbekPhone } from '@/lib/utils/phone';

export const updateProfileSchema = z.object({
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
  phone: z
    .string()
    .min(1, 'Telefon raqam kiritilishi shart')
    .refine(
      (val) => isValidUzbekPhone(val),
      'Telefon formati noto‘g‘ri (+998 90 123-45-67 ko‘rinishida bo‘lishi kerak)'
    ),
});

export type UpdateProfileFormValues = z.infer<typeof updateProfileSchema>;

