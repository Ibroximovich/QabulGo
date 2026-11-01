import React, { forwardRef, useState, useId } from 'react';
import { Eye, EyeOff } from 'lucide-react';
import { cn } from '@/lib/utils/cn';

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  helperText?: string;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  isPassword?: boolean;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  (
    {
      className,
      type = 'text',
      label,
      error,
      helperText,
      leftIcon,
      rightIcon,
      isPassword = false,
      disabled,
      id,
      ...props
    },
    ref
  ) => {
    const generatedId = useId();
    const inputId = id || generatedId;
    const errorId = `${inputId}-error`;
    const helperId = `${inputId}-helper`;

    const [showPassword, setShowPassword] = useState(false);

    const computedType = isPassword ? (showPassword ? 'text' : 'password') : type;

    return (
      <div className="w-full flex flex-col gap-1.5">
        {label && (
          <label
            htmlFor={inputId}
            className="text-sm font-medium text-brand-navy flex items-center justify-between"
          >
            <span>{label}</span>
          </label>
        )}

        <div className="relative flex items-center">
          {leftIcon && (
            <div className="absolute left-3.5 flex items-center pointer-events-none text-brand-muted">
              {leftIcon}
            </div>
          )}

          <input
            ref={ref}
            id={inputId}
            type={computedType}
            disabled={disabled}
            aria-invalid={Boolean(error)}
            aria-describedby={error ? errorId : helperText ? helperId : undefined}
            className={cn(
              'w-full h-11 px-3.5 text-sm rounded-xl border bg-white text-brand-navy placeholder:text-brand-muted/70 transition-all duration-150',
              'focus:outline-none focus:ring-2 focus:ring-brand-blue/20 focus:border-brand-blue',
              'disabled:bg-slate-50 disabled:text-slate-400 disabled:cursor-not-allowed',
              error
                ? 'border-brand-error focus:ring-brand-error/20 focus:border-brand-error text-brand-error'
                : 'border-slate-200 hover:border-slate-300',
              leftIcon && 'pl-10',
              (rightIcon || isPassword) && 'pr-11',
              className
            )}
            {...props}
          />

          {isPassword && (
            <button
              type="button"
              tabIndex={-1}
              onClick={() => setShowPassword((prev) => !prev)}
              aria-label={showPassword ? 'Parolni yashirish' : 'Parolni ko‘rsatish'}
              className="absolute right-3.5 p-1 text-brand-muted hover:text-brand-navy focus:outline-none transition-colors"
            >
              {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
            </button>
          )}

          {!isPassword && rightIcon && (
            <div className="absolute right-3.5 flex items-center text-brand-muted pointer-events-none">
              {rightIcon}
            </div>
          )}
        </div>

        {error && (
          <p id={errorId} role="alert" className="text-xs font-medium text-brand-error flex items-center gap-1 mt-0.5">
            <span>{error}</span>
          </p>
        )}

        {!error && helperText && (
          <p id={helperId} className="text-xs text-brand-muted mt-0.5">
            {helperText}
          </p>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';

