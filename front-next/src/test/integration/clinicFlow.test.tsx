import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ClinicLoginPage } from '@/features/clinic/pages/ClinicLoginPage';
import { ClinicRegisterPage } from '@/features/clinic/pages/ClinicRegisterPage';
import { AppProviders } from '@/providers';

vi.mock('@/features/clinic/api/clinicApi', () => ({
  clinicApi: {
    login: vi.fn(),
    register: vi.fn(),
    getMyClinic: vi.fn(),
  },
}));

vi.mock('@/features/auth/api/authApi', () => ({
  authApi: {
    refreshToken: vi.fn().mockRejectedValue(new Error('No session')),
    logout: vi.fn(),
  },
}));

describe('Clinic Authentication Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders clinic login form with clinic badge and inputs', async () => {
    render(
      <AppProviders>
        <ClinicLoginPage />
      </AppProviders>
    );

    expect(
      screen.getByRole('heading', { name: /klinika portaliga kirish/i })
    ).toBeInTheDocument();
    expect(screen.getByLabelText(/email manzili/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^kirish$/i })).toBeInTheDocument();
  });

  it('renders clinic registration form with terms checkbox', async () => {
    render(
      <AppProviders>
        <ClinicRegisterPage />
      </AppProviders>
    );

    expect(
      screen.getByRole('heading', { name: /klinikani ro‘yxatdan o‘tkazish/i })
    ).toBeInTheDocument();
    expect(screen.getByLabelText(/ism/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/familiya/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/telefon raqami/i)).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: /klinika hisobini yaratish/i })
    ).toBeInTheDocument();
  });
});
