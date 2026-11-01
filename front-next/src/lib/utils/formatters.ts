const UZ_MONTHS = [
  'yanvar', 'fevral', 'mart', 'aprel', 'may', 'iyun',
  'iyul', 'avgust', 'sentabr', 'oktabr', 'noyabr', 'dekabr'
];

export function formatDateUz(dateInput: string | Date | undefined | null): string {
  if (!dateInput) return '—';
  try {
    const d = new Date(dateInput);
    if (isNaN(d.getTime())) return '—';
    const day = d.getDate();
    const month = UZ_MONTHS[d.getMonth()];
    const year = d.getFullYear();
    return `${day}-${month}, ${year}`;
  } catch {
    return '—';
  }
}

export const formatDate = formatDateUz;

export function formatDateTimeUz(dateInput: string | Date | undefined | null): string {
  if (!dateInput) return '—';
  try {
    const d = new Date(dateInput);
    if (isNaN(d.getTime())) return '—';
    const day = d.getDate();
    const month = UZ_MONTHS[d.getMonth()];
    const year = d.getFullYear();
    const hours = String(d.getHours()).padStart(2, '0');
    const minutes = String(d.getMinutes()).padStart(2, '0');
    return `${day}-${month}, ${year} ${hours}:${minutes}`;
  } catch {
    return '—';
  }
}

export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('uz-UZ').format(amount) + " so'm";
}

export function getInitials(firstName?: string, lastName?: string): string {
  const f = firstName?.trim().charAt(0).toUpperCase() || '';
  const l = lastName?.trim().charAt(0).toUpperCase() || '';
  return `${f}${l}` || 'QG';
}

export function formatRoleUz(role?: string): string {
  switch (role) {
    case 'patient':
      return 'Bemor';
    case 'doctor':
      return 'Shifokor';
    case 'clinic_admin':
      return 'Klinika administratori';
    case 'clinic_owner':
      return 'Klinika egasi';
    case 'super_admin':
      return 'Super administrator';
    default:
      return role || 'Foydalanuvchi';
  }
}

