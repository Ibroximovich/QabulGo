import React, { forwardRef } from 'react';
import { Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils/cn';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive' | 'danger';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
  loadingText?: string;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  fullWidth?: boolean;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      className,
      variant = 'primary',
      size = 'md',
      isLoading = false,
      loadingText,
      leftIcon,
      rightIcon,
      fullWidth = false,
      disabled,
      children,
      type = 'button',
      ...props
    },
    ref
  ) => {
    const baseStyles =
      'inline-flex items-center justify-center font-medium rounded-xl transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-blue focus-visible:ring-offset-2 disabled:opacity-60 disabled:cursor-not-allowed select-none active:scale-[0.98]';

    const normalizedVariant = variant === 'danger' ? 'destructive' : variant;

    const variants = {
      primary:
        'bg-brand-blue text-white hover:bg-brand-blue-600 active:bg-brand-blue-700 shadow-sm hover:shadow-md hover:shadow-brand-blue/20',
      secondary:
        'bg-brand-teal text-white hover:bg-brand-teal-600 active:bg-brand-teal-700 shadow-sm hover:shadow-md hover:shadow-brand-teal/20',
      outline:
        'border-2 border-brand-navy/15 bg-white text-brand-navy hover:bg-brand-navy/5 hover:border-brand-navy/30 active:bg-brand-navy/10',
      ghost:
        'text-brand-navy hover:bg-brand-navy/5 active:bg-brand-navy/10',
      destructive:
        'bg-brand-error text-white hover:bg-red-700 active:bg-red-800 shadow-sm',
    };

    const sizes = {
      sm: 'text-xs h-9 px-3 gap-1.5',
      md: 'text-sm h-11 px-4 gap-2',
      lg: 'text-base h-12 px-6 gap-2.5 font-semibold',
    };

    const widthStyle = fullWidth ? 'w-full' : '';

    return (
      <button
        ref={ref}
        type={type}
        disabled={disabled || isLoading}
        className={cn(baseStyles, variants[normalizedVariant], sizes[size], widthStyle, className)}
        {...props}
      >
        {isLoading ? (
          <>
            <Loader2 className="h-4 w-4 animate-spin shrink-0" />
            <span>{loadingText || children}</span>
          </>
        ) : (
          <>
            {leftIcon && <span className="shrink-0">{leftIcon}</span>}
            <span>{children}</span>
            {rightIcon && <span className="shrink-0">{rightIcon}</span>}
          </>
        )}
      </button>
    );
  }
);

Button.displayName = 'Button';
