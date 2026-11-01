import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi } from 'vitest';
import { OtpInput } from '../OtpInput';

describe('OtpInput component', () => {
  it('renders 6 inputs by default', () => {
    render(<OtpInput value="" onChange={() => {}} />);
    const inputs = screen.getAllByRole('textbox');
    expect(inputs).toHaveLength(6);
  });

  it('renders values in corresponding boxes', () => {
    render(<OtpInput value="123" onChange={() => {}} />);
    const inputs = screen.getAllByRole('textbox') as HTMLInputElement[];
    expect(inputs[0].value).toBe('1');
    expect(inputs[1].value).toBe('2');
    expect(inputs[2].value).toBe('3');
    expect(inputs[3].value).toBe('');
  });

  it('calls onChange with updated digit on input', async () => {
    const handleChange = vi.fn();
    render(<OtpInput value="" onChange={handleChange} />);
    const inputs = screen.getAllByRole('textbox');
    await userEvent.type(inputs[0], '5');
    expect(handleChange).toHaveBeenCalledWith('5');
  });
});
