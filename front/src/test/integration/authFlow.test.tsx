import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { LoginPage } from '@/features/auth/pages/LoginPage';
import { RegisterPage } from '@/features/auth/pages/RegisterPage';
import { VerifyEmailPage } from '@/features/auth/pages/VerifyEmailPage';
import { AppProviders } from '@/app/providers';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

// Mock auth API
vi.mock('@/features/auth/api/authApi', () => ({
  authApi: {
    login: vi.fn(),
    register: vi.fn(),
    verifyEmail: vi.fn(),
    resendVerification: vi.fn(),
    refreshToken: vi.fn().mockRejectedValue(new Error('No session')),
    logout: vi.fn().mockResolvedValue({ success: true, message: 'Logged out' }),
  },
}));

describe('Authentication User Flows Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders login form and validates inputs before submitting', async () => {
    render(
      <MemoryRouter initialEntries={['/login']}>
        <AppProviders>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
          </Routes>
        </AppProviders>
      </MemoryRouter>
    );

    expect(screen.getByRole('heading', { name: /tizimga kirish/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/email manzili/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^parol$/i)).toBeInTheDocument();

    const submitBtn = screen.getByRole('button', { name: /^kirish$/i });
    await userEvent.click(submitBtn);

    // Form validation triggers
    await waitFor(() => {
      expect(screen.getByText(/email kiritilishi shart/i)).toBeInTheDocument();
    });
  });

  it('renders register form and checks required terms acceptance', async () => {
    render(
      <MemoryRouter initialEntries={['/register']}>
        <AppProviders>
          <Routes>
            <Route path="/register" element={<RegisterPage />} />
          </Routes>
        </AppProviders>
      </MemoryRouter>
    );

    expect(screen.getByRole('heading', { name: /ro‘yxatdan o‘tish/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/ismingiz/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/familiyangiz/i)).toBeInTheDocument();

    const submitBtn = screen.getByRole('button', { name: /ro‘yxatdan o‘tish/i });
    await userEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByText(/ism kiritilishi shart/i)).toBeInTheDocument();
    });
  });

  it('renders verify email page and displays email info', () => {
    render(
      <MemoryRouter initialEntries={['/verify-email?email=test%40example.com']}>
        <AppProviders>
          <Routes>
            <Route path="/verify-email" element={<VerifyEmailPage />} />
          </Routes>
        </AppProviders>
      </MemoryRouter>
    );

    expect(screen.getByRole('heading', { name: /emailni tasdiqlang/i })).toBeInTheDocument();
    expect(screen.getByText(/test@example.com/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /emailni tasdiqlash/i })).toBeInTheDocument();
  });
});
