import { describe, it, expect } from 'vitest';
import {
  clinicRegisterSchema,
  doctorFormSchema,
  adminReviewSchema,
} from '../clinicSchemas';

describe('Clinic Schema Validations', () => {
  describe('clinicRegisterSchema', () => {
    it('validates a valid clinic registration form', () => {
      const valid = {
        first_name: 'Jamshid',
        last_name: 'Karimov',
        email: 'jamshid@clinic.uz',
        phone: '+998 90 123 45 67',
        password: 'Password123',
        confirm_password: 'Password123',
        accept_terms: true,
      };
      const result = clinicRegisterSchema.safeParse(valid);
      expect(result.success).toBe(true);
    });

    it('fails when password does not have numbers or uppercase letters', () => {
      const invalid = {
        first_name: 'Jamshid',
        last_name: 'Karimov',
        email: 'jamshid@clinic.uz',
        phone: '+998 90 123 45 67',
        password: 'password',
        confirm_password: 'password',
        accept_terms: true,
      };
      const result = clinicRegisterSchema.safeParse(invalid);
      expect(result.success).toBe(false);
    });

    it('fails when passwords do not match', () => {
      const invalid = {
        first_name: 'Jamshid',
        last_name: 'Karimov',
        email: 'jamshid@clinic.uz',
        phone: '+998 90 123 45 67',
        password: 'Password123',
        confirm_password: 'Password999',
        accept_terms: true,
      };
      const result = clinicRegisterSchema.safeParse(invalid);
      expect(result.success).toBe(false);
    });

    it('fails when terms are not accepted', () => {
      const invalid = {
        first_name: 'Jamshid',
        last_name: 'Karimov',
        email: 'jamshid@clinic.uz',
        phone: '+998 90 123 45 67',
        password: 'Password123',
        confirm_password: 'Password123',
        accept_terms: false,
      };
      const result = clinicRegisterSchema.safeParse(invalid);
      expect(result.success).toBe(false);
    });
  });

  describe('doctorFormSchema', () => {
    it('validates a correct doctor input', () => {
      const valid = {
        first_name: 'Anvar',
        last_name: 'Saidov',
        specialty_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
        bio: 'Oliy toifali kardiolog',
        experience_years: 15,
        consultation_price: 150000,
        appointment_duration_minutes: 30,
        is_active: true,
      };
      const result = doctorFormSchema.safeParse(valid);
      expect(result.success).toBe(true);
    });

    it('fails when consultation price is negative', () => {
      const invalid = {
        first_name: 'Anvar',
        last_name: 'Saidov',
        specialty_id: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
        experience_years: 15,
        consultation_price: -500,
        appointment_duration_minutes: 30,
      };
      const result = doctorFormSchema.safeParse(invalid);
      expect(result.success).toBe(false);
    });
  });

  describe('adminReviewSchema', () => {
    it('accepts reasons with >= 5 characters', () => {
      const result = adminReviewSchema.safeParse({
        reason: 'Litsenziya talablariga javob bermaydi',
      });
      expect(result.success).toBe(true);
    });

    it('rejects short reasons (< 5 chars)', () => {
      const result = adminReviewSchema.safeParse({
        reason: 'Yoq',
      });
      expect(result.success).toBe(false);
    });
  });
});

