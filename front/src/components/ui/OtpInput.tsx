import React, { useRef, useEffect } from 'react';
import { cn } from '@/lib/utils/cn';

export interface OtpInputProps {
  value: string;
  onChange: (value: string) => void;
  length?: number;
  disabled?: boolean;
  hasError?: boolean;
  autoFocus?: boolean;
}

export function OtpInput({
  value,
  onChange,
  length = 6,
  disabled = false,
  hasError = false,
  autoFocus = true,
}: OtpInputProps) {
  const inputRefs = useRef<(HTMLInputElement | null)[]>([]);

  // Split value into array of characters
  const digits = Array.from({ length }, (_, i) => value[i] || '');

  useEffect(() => {
    if (autoFocus && inputRefs.current[0]) {
      inputRefs.current[0].focus();
    }
  }, [autoFocus]);

  const handleDigitChange = (index: number, char: string) => {
    // Only accept numeric digit
    const cleaned = char.replace(/\D/g, '');
    if (!cleaned) {
      // Clear this index
      const newDigits = [...digits];
      newDigits[index] = '';
      onChange(newDigits.join(''));
      return;
    }

    const lastChar = cleaned[cleaned.length - 1];
    const newDigits = [...digits];
    newDigits[index] = lastChar;
    const combined = newDigits.join('');
    onChange(combined);

    // Auto-advance focus
    if (index < length - 1) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace') {
      if (!digits[index] && index > 0) {
        // Move back and clear previous
        const newDigits = [...digits];
        newDigits[index - 1] = '';
        onChange(newDigits.join(''));
        inputRefs.current[index - 1]?.focus();
      } else {
        const newDigits = [...digits];
        newDigits[index] = '';
        onChange(newDigits.join(''));
      }
    } else if (e.key === 'ArrowLeft' && index > 0) {
      e.preventDefault();
      inputRefs.current[index - 1]?.focus();
    } else if (e.key === 'ArrowRight' && index < length - 1) {
      e.preventDefault();
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handlePaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault();
    const pastedData = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, length);
    if (!pastedData) return;

    onChange(pastedData);
    const nextIndex = Math.min(pastedData.length, length - 1);
    inputRefs.current[nextIndex]?.focus();
  };

  return (
    <div className="flex items-center justify-center gap-2 sm:gap-3 w-full" role="group" aria-label="Tasdiqlash kodi">
      {digits.map((digit, index) => (
        <input
          key={index}
          ref={(el) => {
            inputRefs.current[index] = el;
          }}
          type="text"
          inputMode="numeric"
          pattern="[0-9]*"
          maxLength={1}
          autoComplete={index === 0 ? 'one-time-code' : 'off'}
          disabled={disabled}
          value={digit}
          onChange={(e) => handleDigitChange(index, e.target.value)}
          onKeyDown={(e) => handleKeyDown(index, e)}
          onPaste={handlePaste}
          onFocus={(e) => e.target.select()}
          aria-label={`${index + 1}-raqam`}
          className={cn(
            'w-11 h-13 sm:w-14 sm:h-16 text-center text-xl sm:text-2xl font-bold rounded-xl border bg-white text-brand-navy shadow-sm transition-all',
            'focus:outline-none focus:ring-2 focus:ring-brand-blue/30 focus:border-brand-blue focus:shadow-md',
            'disabled:bg-slate-50 disabled:text-slate-400 disabled:cursor-not-allowed',
            hasError
              ? 'border-brand-error focus:ring-brand-error/20 focus:border-brand-error text-brand-error bg-red-50/20'
              : digit
              ? 'border-brand-blue/60 bg-brand-blue-50/20'
              : 'border-slate-200 hover:border-slate-300'
          )}
        />
      ))}
    </div>
  );
}

