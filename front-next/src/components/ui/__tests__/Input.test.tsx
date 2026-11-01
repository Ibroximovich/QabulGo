import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect } from 'vitest';
import { Input } from '../Input';

describe('Input component', () => {
  it('renders input with label and helper text', () => {
    render(
      <Input
        label="Emailingiz"
        helperText="Pochtangizni kiriting"
        placeholder="ali@example.com"
      />
    );
    expect(screen.getByLabelText(/emailingiz/i)).toBeInTheDocument();
    expect(screen.getByText(/pochtangizni kiriting/i)).toBeInTheDocument();
  });

  it('renders error message and sets aria-invalid', () => {
    render(<Input label="Parol" error="Parol noto‘g‘ri" />);
    const input = screen.getByLabelText(/parol/i);
    expect(input).toHaveAttribute('aria-invalid', 'true');
    expect(screen.getByRole('alert')).toHaveTextContent(/parol noto‘g‘ri/i);
  });

  it('toggles password visibility with eye button', async () => {
    render(<Input label="Parol" isPassword defaultValue="Secret123" />);
    const input = screen.getByLabelText(/^parol$/i);
    expect(input).toHaveAttribute('type', 'password');

    const toggleBtn = screen.getByRole('button', { name: /parolni ko‘rsatish/i });
    await userEvent.click(toggleBtn);
    expect(input).toHaveAttribute('type', 'text');

    await userEvent.click(toggleBtn);
    expect(input).toHaveAttribute('type', 'password');
  });
});
