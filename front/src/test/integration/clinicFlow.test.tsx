import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ClinicLoginPage } from '@/features/clinic/pages/ClinicLoginPage';
import { ClinicRegisterPage } from '@/features/clinic/pages/ClinicRegisterPage';
import { AppProviders } from '@/app/providers';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

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
      <MemoryRouter initialEntries={['/clinic/login']}>
        <AppProviders>
          <Routes>
            <Route path="/clinic/login" element={<ClinicLoginPage />} />
          </Routes>
        </AppProviders>
      </MemoryRouter>
    );

    expect(
      screen.getByRole('heading', { name: /klinika portaliga kirish/i })
    ).toBeInTheDocument();
    expect(screen.getByLabelText(/email manzili/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^kirish$/i })).toBeInTheDocument();
  });

  it('renders clinic registration form with terms checkbox', async () => {
    render(
      <MemoryRouter initialEntries={['/clinic/register']}>
        <AppProviders>
          <Routes>
            <Route path="/clinic/register" element={<ClinicRegisterPage />} />
          </Routes>
        </AppProviders>
      </MemoryRouter>
    );

    expect(
      screen.getByRole('heading', { name: /klinikani ro‘yxatdan o‘tkazish/i })
    ).toBeInTheDocument();
    expect(screen.getByLabelText(/^ism$/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^familiya$/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/telefon raqami/i)).toBeInTheDocument();
  });
});

