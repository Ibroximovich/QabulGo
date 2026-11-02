import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ClinicTelegramPage } from '../pages/ClinicTelegramPage';
import { clinicApi } from '../api/clinicApi';
import * as ClinicLayoutModule from '@/components/layout/clinic/ClinicLayout';

// Mock clinicApi
vi.mock('../api/clinicApi', () => ({
  clinicApi: {
    getTelegramStatus: vi.fn(),
    generateTelegramLinkToken: vi.fn(),
    disconnectTelegramRecipient: vi.fn(),
    toggleTelegramNotifications: vi.fn(),
  },
}));

// Mock Sonner toast
vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

describe('ClinicTelegramPage Component', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });
    vi.clearAllMocks();

    // Mock Clinic Context
    vi.spyOn(ClinicLayoutModule, 'useClinicContext').mockReturnValue({
      clinic: {
        id: 'c111-2222-3333-4444',
        name: 'Shifo Med Clinic',
        status: 'approved',
        phone: '+998901234567',
        region: 'Toshkent',
        district: 'Yunusobod',
        address: 'Amir Temur ko‘chasi',
        owner_id: 'u111',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        doctors_count: 5,
        working_hours: [],
        description: 'Test klinikasi',
        email: 'shifo@med.uz',
        logo_url: '',
        cover_image_url: '',
      },
      isLoadingClinic: false,
    });
  });

  it('renders disconnected state when clinic has no linked Telegram accounts', async () => {
    vi.mocked(clinicApi.getTelegramStatus).mockResolvedValueOnce({
      data: {
        success: true,
        message: 'OK',
        data: {
          is_connected: false,
          bot_username: 'QabulGoClinicBot',
          recipients: [],
        },
      },
    } as any);

    render(
      <QueryClientProvider client={queryClient}>
        <ClinicTelegramPage />
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Telegram Bot Integratsiyasi')).toBeInTheDocument();
      expect(screen.getByText('Ulanmagan')).toBeInTheDocument();
      expect(screen.getByText('Telegram botni ulash')).toBeInTheDocument();
      expect(screen.getByText('@QabulGoClinicBot')).toBeInTheDocument();
    });
  });

  it('renders connected recipients and handles deep link generation', async () => {
    const windowOpenSpy = vi.spyOn(window, 'open').mockImplementation(() => null);

    vi.mocked(clinicApi.getTelegramStatus).mockResolvedValueOnce({
      data: {
        success: true,
        message: 'OK',
        data: {
          is_connected: true,
          bot_username: 'QabulGoClinicBot',
          recipients: [
            {
              id: 'rec-1',
              telegram_user_id: 12345678,
              telegram_username: 'clinic_admin_user',
              telegram_first_name: 'Sarvar',
              is_active: true,
              notifications_enabled: true,
              created_at: new Date().toISOString(),
            },
          ],
        },
      },
    } as any);

    vi.mocked(clinicApi.generateTelegramLinkToken).mockResolvedValueOnce({
      data: {
        success: true,
        message: 'Token created',
        data: {
          token: 'raw_token_123',
          deep_link: 'https://t.me/QabulGoClinicBot?start=raw_token_123',
          bot_username: 'QabulGoClinicBot',
          expires_at: new Date().toISOString(),
        },
      },
    } as any);

    render(
      <QueryClientProvider client={queryClient}>
        <ClinicTelegramPage />
      </QueryClientProvider>
    );

    await waitFor(() => {
      expect(screen.getByText('Ulangan')).toBeInTheDocument();
      expect(screen.getByText('Sarvar')).toBeInTheDocument();
      expect(screen.getByText('@clinic_admin_user')).toBeInTheDocument();
    });

    const connectBtn = screen.getByText('Yana bitta akkaunt ulash');
    fireEvent.click(connectBtn);

    await waitFor(() => {
      expect(clinicApi.generateTelegramLinkToken).toHaveBeenCalledWith('c111-2222-3333-4444');
      expect(windowOpenSpy).toHaveBeenCalledWith('https://t.me/QabulGoClinicBot?start=raw_token_123', '_blank');
    });

    windowOpenSpy.mockRestore();
  });
});

