import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { Button } from '../Button';

describe('Button component', () => {
  it('renders button with label', () => {
    render(<Button>Davom etish</Button>);
    expect(screen.getByRole('button', { name: /davom etish/i })).toBeInTheDocument();
  });

  it('triggers onClick handler when clicked', async () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Bosing</Button>);
    await userEvent.click(screen.getByRole('button', { name: /bosing/i }));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('renders loading spinner and disables button when isLoading is true', () => {
    render(<Button isLoading loadingText="Yuklanmoqda...">Saqlash</Button>);
    const button = screen.getByRole('button');
    expect(button).toBeDisabled();
    expect(screen.getByText(/yuklanmoqda/i)).toBeInTheDocument();
  });
});
