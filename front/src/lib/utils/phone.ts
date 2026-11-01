/**
 * Phone formatting utility for Uzbekistan numbers (+998)
 * Backend expects: +998XXXXXXXXX (exactly 13 characters: +998 followed by 9 digits)
 */

export function cleanPhoneNumber(raw: string): string {
  const trimmed = raw.trim();
  if (trimmed.startsWith('+') && !trimmed.startsWith('+998')) {
    return trimmed;
  }
  const digits = raw.replace(/\D/g, '');
  if (!digits) return '';

  // If starts with 998, keep it; otherwise prepend 998 if it looks like local 9 digits
  if (digits.startsWith('998')) {
    return `+${digits.slice(0, 12)}`;
  }
  return `+998${digits.slice(0, 9)}`;
}

export const normalizePhone = cleanPhoneNumber;

export function formatUzbekPhoneDisplay(raw: string): string {
  if (!raw) return '+998 ';
  
  // Extract digits
  let digits = raw.replace(/\D/g, '');
  
  // Normalize prefix
  if (digits.startsWith('998')) {
    digits = digits.slice(3);
  }

  // Restrict to max 9 operator/subscriber digits
  digits = digits.slice(0, 9);

  let formatted = '+998';
  if (digits.length > 0) {
    formatted += ` (${digits.slice(0, 2)}`;
  }
  if (digits.length >= 2) {
    formatted += ')';
  }
  if (digits.length > 2) {
    formatted += ` ${digits.slice(2, 5)}`;
  }
  if (digits.length > 5) {
    formatted += `-${digits.slice(5, 7)}`;
  }
  if (digits.length > 7) {
    formatted += `-${digits.slice(7, 9)}`;
  }

  return formatted;
}

export const formatPhone = formatUzbekPhoneDisplay;

export function isValidUzbekPhone(phone: string): boolean {
  const cleaned = cleanPhoneNumber(phone);
  return /^\+998[0-9]{9}$/.test(cleaned);
}
