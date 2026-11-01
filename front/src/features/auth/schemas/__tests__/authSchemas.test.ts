import { describe, it, expect } from 'vitest';
import { registerSchema, loginSchema, verifyEmailSchema } from '../authSchemas';

describe('Auth validation schemas', () => {
  it('registerSchema validates correct payload', () => {
    const valid = {
      first_name: 'Ali',
      last_name: 'Valiyev',
      email: 'ali@example.com',
      phone: '+998 (90) 123-45-67',
      password: 'StrongPass123!',
      confirm_password: 'StrongPass123!',
      accept_terms: true,
    };
    const result = registerSchema.safeParse(valid);
    expect(result.success).toBe(true);
  });

  it('registerSchema rejects mismatched passwords', () => {
    const invalid = {
      first_name: 'Ali',
      last_name: 'Valiyev',
      email: 'ali@example.com',
      phone: '+998 (90) 123-45-67',
      password: 'StrongPass123!',
      confirm_password: 'DifferentPass123!',
      accept_terms: true,
    };
    const result = registerSchema.safeParse(invalid);
    expect(result.success).toBe(false);
  });

  it('registerSchema rejects unaccepted terms', () => {
    const invalid = {
      first_name: 'Ali',
      last_name: 'Valiyev',
      email: 'ali@example.com',
      phone: '+998 (90) 123-45-67',
      password: 'StrongPass123!',
      confirm_password: 'StrongPass123!',
      accept_terms: false,
    };
    const result = registerSchema.safeParse(invalid);
    expect(result.success).toBe(false);
  });

  it('loginSchema validates email and password', () => {
    expect(loginSchema.safeParse({ email: 'ali@example.com', password: '123' }).success).toBe(true);
    expect(loginSchema.safeParse({ email: 'not-an-email', password: '123' }).success).toBe(false);
    expect(loginSchema.safeParse({ email: '', password: '' }).success).toBe(false);
  });

  it('verifyEmailSchema requires 6-digit numeric code', () => {
    expect(verifyEmailSchema.safeParse({ email: 'ali@example.com', code: '123456' }).success).toBe(true);
    expect(verifyEmailSchema.safeParse({ email: 'ali@example.com', code: '123' }).success).toBe(false);
    expect(verifyEmailSchema.safeParse({ email: 'ali@example.com', code: 'abcdef' }).success).toBe(false);
  });
});

