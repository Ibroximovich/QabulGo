import { describe, it, expect } from 'vitest';
import { formatDateUz, getInitials, formatRoleUz } from '../formatters';

describe('Formatter utilities', () => {
  it('formatDateUz should format dates with Uzbek Latin month names', () => {
    const formatted = formatDateUz('2026-10-31T10:00:00Z');
    expect(formatted).toContain('oktabr');
    expect(formatted).toContain('2026');
  });

  it('getInitials should extract uppercase first letters', () => {
    expect(getInitials('Ali', 'Valiyev')).toBe('AV');
    expect(getInitials('Sardor', 'Rahimov')).toBe('SR');
    expect(getInitials('', '')).toBe('QG');
  });

  it('formatRoleUz should translate system roles', () => {
    expect(formatRoleUz('patient')).toBe('Bemor');
    expect(formatRoleUz('doctor')).toBe('Shifokor');
    expect(formatRoleUz('clinic_admin')).toBe('Klinika administratori');
  });
});

