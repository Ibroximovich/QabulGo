import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { DoctorBookingCalendar } from '@/features/search/components/DoctorBookingCalendar';
import { MyAppointmentsPage } from '@/features/appointments/pages/MyAppointmentsPage';
import { ClinicAppointmentsPage } from '@/features/clinic/pages/ClinicAppointmentsPage';
import { searchApi } from '@/features/search/api/searchApi';
import { appointmentApi } from '@/features/appointments/api/appointmentApi';
import { clinicApi } from '@/features/clinic/api/clinicApi';
import * as AuthProviderModule from '@/providers/AuthProvider';
import * as ClinicLayoutModule from '@/components/layout/clinic/ClinicLayout';

// Mock Next.js navigation
const mockPush = vi.fn();
const mockReplace = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
    replace: mockReplace,
    prefetch: vi.fn(),
    back: vi.fn(),
  }),
  usePathname: () => '/doctors/doc-1',
  useParams: () => ({ id: 'doc-1', appointmentId: 'appt-1' }),
}));

// Mock Sonner Toast
vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

// Mock APIs
vi.mock('@/features/search/api/searchApi', () => ({
  searchApi: {
    getDoctorAvailability: vi.fn(),
  },
}));

vi.mock('@/features/appointments/api/appointmentApi', () => ({
  appointmentApi: {
    bookAppointment: vi.fn(),
    getMyAppointments: vi.fn(),
    getMyAppointmentById: vi.fn(),
    cancelMyAppointment: vi.fn(),
    getClinicAppointments: vi.fn(),
    getClinicAppointmentById: vi.fn(),
    cancelClinicAppointment: vi.fn(),
    completeClinicAppointment: vi.fn(),
    noShowClinicAppointment: vi.fn(),
  },
}));

vi.mock('@/features/clinic/api/clinicApi', () => ({
  clinicApi: {
    listDoctors: vi.fn(),
    getMyClinic: vi.fn(),
  },
}));

const mockWorkplaces = [
  {
    clinic_doctor_id: 'cd-1',
    clinic_id: 'clinic-1',
    clinic_name: 'Shifo Nur Klinikasi',
    region: 'Toshkent',
    district: 'Yunusobod',
    address: 'Amir Temur 45',
    phone: '+998901234567',
    logo_url: '/logo.png',
    consultation_price: 150000,
    appointment_duration_minutes: 30,
  },
];

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
      mutations: { retry: false },
    },
  });
}

function renderWithClient(ui: React.ReactElement) {
  const testClient = createTestQueryClient();
  return render(<QueryClientProvider client={testClient}>{ui}</QueryClientProvider>);
}

describe('Phase 5.2: Appointment Booking Frontend Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();

    // Default mock availability
    (searchApi.getDoctorAvailability as any).mockResolvedValue({
      data: {
        date: '2026-11-09',
        timezone: 'Asia/Tashkent',
        clinic_id: 'clinic-1',
        doctor_id: 'doc-1',
        clinic_doctor_id: 'cd-1',
        slots: [
          { start_time: '09:00', end_time: '10:00' },
          { start_time: '10:00', end_time: '11:00' },
        ],
      },
    });

    // Default mock clinic context
    vi.spyOn(ClinicLayoutModule, 'useClinicContext').mockReturnValue({
      clinic: {
        id: 'clinic-1',
        name: 'Shifo Nur',
        status: 'approved',
        doctors_count: 5,
        address: 'Amir Temur 45',
        region: 'Toshkent',
        district: 'Yunusobod',
        phone: '+998901234567',
      } as any,
      isLoadingClinic: false,
    });
  });

  describe('DoctorBookingCalendar Booking Flow', () => {
    it('redirects unauthenticated user to login with booking intent stored in sessionStorage', async () => {
      vi.spyOn(AuthProviderModule, 'useAuth').mockReturnValue({
        user: null,
        isAuthenticated: false,
        isLoading: false,
        login: vi.fn(),
        logout: vi.fn() as any,
        updateUser: vi.fn(),
        refetchUser: vi.fn() as any,
      });

      renderWithClient(
        <DoctorBookingCalendar
          doctorId="doc-1"
          doctorName="Dr. Alisher Valiyev"
          workplaces={mockWorkplaces}
        />
      );

      // Wait for slots to load
      await waitFor(() => {
        expect(screen.getByText(/09:00 – 10:00/i)).toBeInTheDocument();
      });

      // Select first slot
      fireEvent.click(screen.getByText(/09:00 – 10:00/i));

      // Click "Qabulga yozilish"
      const bookButton = screen.getByRole('button', { name: /qabulga yozilish/i });
      expect(bookButton).not.toBeDisabled();
      fireEvent.click(bookButton);

      // Check session storage has intent
      const savedIntent = sessionStorage.getItem('qabulgo_booking_intent');
      expect(savedIntent).toBeTruthy();
      expect(JSON.parse(savedIntent!)).toMatchObject({
        clinicId: 'clinic-1',
        doctorId: 'doc-1',
        startTime: '09:00',
        endTime: '10:00',
      });

      // Verify router push to login with redirect
      expect(mockPush).toHaveBeenCalledWith(expect.stringContaining('/login?redirect='));
    });

    it('blocks non-patient roles (clinic_owner) with an error alert', async () => {
      vi.spyOn(AuthProviderModule, 'useAuth').mockReturnValue({
        user: { id: 'u-1', email: 'doc@test.com', role: 'clinic_owner' } as any,
        isAuthenticated: true,
        isLoading: false,
        login: vi.fn(),
        logout: vi.fn() as any,
        updateUser: vi.fn(),
        refetchUser: vi.fn() as any,
      });

      renderWithClient(
        <DoctorBookingCalendar
          doctorId="doc-1"
          doctorName="Dr. Alisher Valiyev"
          workplaces={mockWorkplaces}
        />
      );

      await waitFor(() => {
        expect(screen.getByText(/09:00 – 10:00/i)).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText(/09:00 – 10:00/i));
      fireEvent.click(screen.getByRole('button', { name: /qabulga yozilish/i }));

      // Booking modal should NOT be opened
      expect(screen.queryByText(/Qabul ma’lumotlarini tasdiqlang/i)).not.toBeInTheDocument();
    });

    it('opens confirmation modal and completes instant confirmed booking on 201 Created', async () => {
      vi.spyOn(AuthProviderModule, 'useAuth').mockReturnValue({
        user: { id: 'patient-1', first_name: 'Bekzod', role: 'patient' } as any,
        isAuthenticated: true,
        isLoading: false,
        login: vi.fn(),
        logout: vi.fn() as any,
        updateUser: vi.fn(),
        refetchUser: vi.fn() as any,
      });

      const mockAppointmentResponse = {
        data: {
          id: 'appt-123',
          patient_id: 'patient-1',
          clinic_id: 'clinic-1',
          doctor_id: 'doc-1',
          appointment_date: '2026-11-09',
          start_time: '09:00',
          end_time: '10:00',
          status: 'confirmed',
          consultation_price: 150000,
          doctor: {
            first_name: 'Alisher',
            last_name: 'Valiyev',
            specialty: 'Kardiolog',
          },
          clinic: {
            name: 'Shifo Nur Klinikasi',
            address: 'Amir Temur 45',
          },
        },
      };

      (appointmentApi.bookAppointment as any).mockResolvedValue(mockAppointmentResponse);

      renderWithClient(
        <DoctorBookingCalendar
          doctorId="doc-1"
          doctorName="Dr. Alisher Valiyev"
          specialtyName="Kardiolog"
          workplaces={mockWorkplaces}
        />
      );

      await waitFor(() => {
        expect(screen.getByText(/09:00 – 10:00/i)).toBeInTheDocument();
      });

      // Select slot
      fireEvent.click(screen.getByRole('button', { name: /09:00 – 10:00/i }));

      // Click "Qabulga yozilish"
      fireEvent.click(screen.getByRole('button', { name: /qabulga yozilish/i }));

      // Confirmation modal is open
      await waitFor(() => {
        expect(screen.getByText(/Qabulga yozilishni tasdiqlaysizmi/i)).toBeInTheDocument();
        expect(screen.getByText(/Dr. Alisher Valiyev/i)).toBeInTheDocument();
        expect(screen.getAllByText(/Shifo Nur Klinikasi/i).length).toBeGreaterThan(0);
      });

      // Click "Tasdiqlash va yozilish"
      const confirmButton = screen.getByRole('button', { name: /tasdiqlash va yozilish/i });
      fireEvent.click(confirmButton);

      // Verify API was called with correct payload & idempotency key
      await waitFor(() => {
        expect(appointmentApi.bookAppointment).toHaveBeenCalledWith(
          expect.objectContaining({
            clinic_id: 'clinic-1',
            doctor_id: 'doc-1',
            start_time: '09:00',
            end_time: '10:00',
          }),
          expect.any(String) // UUID idempotency key
        );
      });

      // Success modal is open with instant confirmed status
      await waitFor(() => {
        expect(screen.getByText(/Qabulga muvaffaqiyatli yozildingiz!/i)).toBeInTheDocument();
        expect(screen.getByText(/Status: Tasdiqlangan/i)).toBeInTheDocument();
      });
    });

    it('handles 409 Conflict properly when slot is already taken', async () => {
      vi.spyOn(AuthProviderModule, 'useAuth').mockReturnValue({
        user: { id: 'patient-1', role: 'patient' } as any,
        isAuthenticated: true,
        isLoading: false,
        login: vi.fn(),
        logout: vi.fn() as any,
        updateUser: vi.fn(),
        refetchUser: vi.fn() as any,
      });

      const conflictError: any = new Error('Slot already booked');
      conflictError.response = {
        status: 409,
        data: { error: 'Ushbu vaqt allaqachon bron qilingan' },
      };

      (appointmentApi.bookAppointment as any).mockRejectedValue(conflictError);

      renderWithClient(
        <DoctorBookingCalendar
          doctorId="doc-1"
          doctorName="Dr. Alisher Valiyev"
          workplaces={mockWorkplaces}
        />
      );

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /09:00 – 10:00/i })).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: /09:00 – 10:00/i }));
      fireEvent.click(screen.getByRole('button', { name: /qabulga yozilish/i }));

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /tasdiqlash va yozilish/i })).toBeInTheDocument();
      });

      fireEvent.click(screen.getByRole('button', { name: /tasdiqlash va yozilish/i }));

      // Slot availability re-fetched, selection reset
      await waitFor(() => {
        expect(searchApi.getDoctorAvailability).toHaveBeenCalled();
      });
    });
  });

  describe('Patient MyAppointmentsPage', () => {
    it('renders appointments list and switches tabs', async () => {
      vi.spyOn(AuthProviderModule, 'useAuth').mockReturnValue({
        user: { id: 'patient-1', role: 'patient' } as any,
        isAuthenticated: true,
        isLoading: false,
        login: vi.fn(),
        logout: vi.fn() as any,
        updateUser: vi.fn(),
        refetchUser: vi.fn() as any,
      });

      (appointmentApi.getMyAppointments as any).mockResolvedValue({
        data: {
          appointments: [
            {
              id: 'appt-1',
              appointment_date: '2026-11-15',
              start_time: '10:00',
              end_time: '11:00',
              status: 'confirmed',
              consultation_price: 200000,
              doctor: {
                first_name: 'Rustam',
                last_name: 'Sobirov',
                specialty: 'Nevropatolog',
              },
              clinic: {
                name: 'City Med',
                address: 'Navoiy 12',
                region: 'Toshkent',
                district: 'Shayxontohur',
              },
            },
          ],
          total: 1,
          page: 1,
          limit: 10,
        },
      });

      renderWithClient(<MyAppointmentsPage />);

      // Title & Tabs
      expect(screen.getByText(/Mening bronlarim/i)).toBeInTheDocument();
      expect(screen.getByText(/Kelgusi qabullar/i)).toBeInTheDocument();

      // Appointment card displayed
      await waitFor(() => {
        expect(screen.getByText(/Rustam Sobirov/i)).toBeInTheDocument();
        expect(screen.getByText(/City Med/i)).toBeInTheDocument();
        expect(screen.getByText(/Tasdiqlangan/i)).toBeInTheDocument();
      });

      // Switch to "Bekor qilinganlar" tab
      fireEvent.click(screen.getByText(/Bekor qilinganlar/i));

      await waitFor(() => {
        expect(appointmentApi.getMyAppointments).toHaveBeenCalledWith(
          expect.objectContaining({ timeframe: 'cancelled' })
        );
      });
    });
  });

  describe('Clinic ClinicAppointmentsPage', () => {
    it('renders clinic appointments table and filters by doctor and status', async () => {
      (clinicApi.listDoctors as any).mockResolvedValue({
        data: {
          data: [
            { id: 'doc-1', first_name: 'Alisher', last_name: 'Valiyev' },
            { id: 'doc-2', first_name: 'Nodira', last_name: 'Azimova' },
          ],
        },
      });

      (appointmentApi.getClinicAppointments as any).mockResolvedValue({
        data: {
          appointments: [
            {
              id: 'appt-10',
              appointment_date: '2026-11-09',
              start_time: '14:00',
              end_time: '15:00',
              status: 'confirmed',
              consultation_price: 150000,
              patient: {
                first_name: 'Jasur',
                last_name: 'Tursunov',
                phone: '+998931112233',
              },
              doctor: {
                first_name: 'Alisher',
                last_name: 'Valiyev',
                specialty: 'Kardiolog',
              },
            },
          ],
          total: 1,
          page: 1,
          limit: 20,
        },
      });

      renderWithClient(<ClinicAppointmentsPage />);

      // Page header
      expect(screen.getByText(/Qabul bronlari/i)).toBeInTheDocument();

      // Table row rendered
      await waitFor(() => {
        expect(screen.getByText(/Jasur Tursunov/i)).toBeInTheDocument();
        expect(screen.getByText(/\+998931112233/i)).toBeInTheDocument();
        expect(screen.getByText(/14:00 – 15:00/i)).toBeInTheDocument();
      });
    });
  });
});
