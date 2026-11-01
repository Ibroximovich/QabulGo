import React from 'react';
import { cn } from '@/lib/utils/cn';

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'blue' | 'teal' | 'navy' | 'success' | 'warning' | 'error' | 'neutral';
  size?: 'sm' | 'md';
}

export function Badge({
  className,
  variant = 'blue',
  size = 'md',
  ...props
}: BadgeProps) {
  const variants = {
    blue: 'bg-brand-blue-50 text-brand-blue-700 border-brand-blue-200',
    teal: 'bg-brand-teal-50 text-brand-teal-800 border-brand-teal-200',
    navy: 'bg-brand-navy-50 text-brand-navy-800 border-brand-navy-200',
    success: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    warning: 'bg-amber-50 text-amber-700 border-amber-200',
    error: 'bg-red-50 text-red-700 border-red-200',
    neutral: 'bg-slate-100 text-slate-700 border-slate-200',
  };

  const sizes = {
    sm: 'text-xs px-2 py-0.5 rounded-md font-medium',
    md: 'text-xs px-2.5 py-1 rounded-lg font-semibold',
  };

  return (
    <span
      className={cn(
        'inline-flex items-center border transition-colors select-none',
        variants[variant],
        sizes[size],
        className
      )}
      {...props}
    />
  );
}

