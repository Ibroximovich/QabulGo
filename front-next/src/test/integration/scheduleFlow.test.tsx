import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { DoctorBookingCalendar } from '@/features/search/components/DoctorBookingCalendar';
import { searchApi } from '@/features/search/api/searchApi';
import { DoctorClinicWorkplace } from '@/features/search/types';

vi.mock('@/features/search/api/searchApi', () => ({
  searchApi: {
    getDoctorAvailability: vi.fn(),
  },
}));

const mockWorkplaces: DoctorClinicWorkplace[] = [
  {
    clinic_doctor_id: 'cd-1',
    clinic_id: 'clinic-1',
    clinic_name: 'Shifo Nur Klinikasi',
    region: 'Toshkent shahri',
    district: 'Yunusobod',
    address: 'Amir Temur ko‘chasi 12',
    phone: '+998712001122',
    logo_url: '',
    consultation_price: 150000,
    appointment_duration_minutes: 30,
  },
  {
    clinic_doctor_id: 'cd-2',
    clinic_id: 'clinic-2',
    clinic_name: 'Medion Markazi',
    region: 'Toshkent shahri',
    district: 'Shayxontohur',
    address: 'Navoiy ko‘chasi 5',
    phone: '+998712334455',
    logo_url: '',
    consultation_price: 200000,
    appointment_duration_minutes: 45,
  },
];

function renderWithClient(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });
  return render(<QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>);
}

describe('Doctor Booking Calendar Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders booking calendar, clinic selector, and date strips', async () => {
    (searchApi.getDoctorAvailability as any).mockResolvedValue({
      data: {
        date: '2026-11-01',
        timezone: 'Asia/Tashkent',
        clinic_id: 'clinic-1',
        doctor_id: 'doc-1',
        clinic_doctor_id: 'cd-1',
        slots: [
          { start_time: '09:00', end_time: '10:00' },
          { start_time: '10:00', end_time: '11:00' },
          { start_time: '11:00', end_time: '12:00' },
        ],
      },
    });

    renderWithClient(<DoctorBookingCalendar doctorId="doc-1" workplaces={mockWorkplaces} />);

    // 1. Header
    expect(screen.getByText(/shifokor qabul vaqtlari/i)).toBeInTheDocument();

    // 2. Multi-clinic pills
    expect(screen.getByRole('button', { name: /shifo nur klinikasi/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /medion markazi/i })).toBeInTheDocument();

    // 3. Waiting for slots
    await waitFor(() => {
      expect(screen.getByText(/09:00 – 10:00/i)).toBeInTheDocument();
      expect(screen.getByText(/10:00 – 11:00/i)).toBeInTheDocument();
      expect(screen.getByText(/11:00 – 12:00/i)).toBeInTheDocument();
    });

    // 4. Booking button is initially disabled until slot is chosen
    expect(screen.getByRole('button', { name: /qabulga yozilish/i })).toBeDisabled();
  });

  it('renders empty state when no slots are available for a date', async () => {
    (searchApi.getDoctorAvailability as any).mockResolvedValue({
      data: {
        date: '2026-11-01',
        timezone: 'Asia/Tashkent',
        clinic_id: 'clinic-1',
        doctor_id: 'doc-1',
        clinic_doctor_id: 'cd-1',
        slots: [],
      },
    });

    renderWithClient(<DoctorBookingCalendar doctorId="doc-1" workplaces={mockWorkplaces} />);

    await waitFor(() => {
      expect(
        screen.getByText(/bu kunga qabul vaqtlari belgilanmagan/i)
      ).toBeInTheDocument();
    });
  });

  it('allows selecting a slot and highlights selected time', async () => {
    (searchApi.getDoctorAvailability as any).mockResolvedValue({
      data: {
        date: '2026-11-01',
        timezone: 'Asia/Tashkent',
        clinic_id: 'clinic-1',
        doctor_id: 'doc-1',
        clinic_doctor_id: 'cd-1',
        slots: [
          { start_time: '09:00', end_time: '10:00' },
          { start_time: '14:00', end_time: '15:00' },
        ],
      },
    });

    renderWithClient(<DoctorBookingCalendar doctorId="doc-1" workplaces={mockWorkplaces} />);

    await waitFor(() => {
      expect(screen.getByText(/09:00 – 10:00/i)).toBeInTheDocument();
    });

    // Click slot
    const slotBtn = screen.getByText(/09:00 – 10:00/i);
    fireEvent.click(slotBtn);

    // Verify selection notice
    expect(screen.getByText(/tanlangan vaqt:/i)).toBeInTheDocument();
  });
});
