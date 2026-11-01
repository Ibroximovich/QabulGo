import { describe, it, expect } from 'vitest';
import { cleanPhoneNumber, formatUzbekPhoneDisplay, isValidUzbekPhone } from '../phone';

describe('Phone utilities', () => {
  it('cleanPhoneNumber should format raw input to +998XXXXXXXXX', () => {
    expect(cleanPhoneNumber('+998 (90) 123-45-67')).toBe('+998901234567');
    expect(cleanPhoneNumber('901234567')).toBe('+998901234567');
    expect(cleanPhoneNumber('+998901234567')).toBe('+998901234567');
  });

  it('formatUzbekPhoneDisplay should display pretty formatted string', () => {
    expect(formatUzbekPhoneDisplay('90')).toBe('+998 (90)');
    expect(formatUzbekPhoneDisplay('90123')).toBe('+998 (90) 123');
    expect(formatUzbekPhoneDisplay('9012345')).toBe('+998 (90) 123-45');
    expect(formatUzbekPhoneDisplay('901234567')).toBe('+998 (90) 123-45-67');
  });

  it('isValidUzbekPhone should correctly validate +998 numbers', () => {
    expect(isValidUzbekPhone('+998901234567')).toBe(true);
    expect(isValidUzbekPhone('+998 (90) 123-45-67')).toBe(true);
    expect(isValidUzbekPhone('901234567')).toBe(true);
    // Invalid
    expect(isValidUzbekPhone('12345')).toBe(false);
    expect(isValidUzbekPhone('+79012345678')).toBe(false);
    expect(isValidUzbekPhone('')).toBe(false);
  });
});

