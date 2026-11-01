import React from 'react';
import { AlertCircle, CheckCircle2, Info, AlertTriangle, X } from 'lucide-react';
import { cn } from '@/lib/utils/cn';

export interface AlertProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'info' | 'success' | 'warning' | 'error' | 'danger';
  title?: string;
  message?: string;
  onClose?: () => void;
}

export function Alert({
  className,
  variant = 'info',
  title,
  message,
  onClose,
  children,
  ...props
}: AlertProps) {
  const normalizedVariant = variant === 'danger' ? 'error' : variant;

  const icons = {
    info: <Info className="h-5 w-5 text-brand-blue shrink-0" />,
    success: <CheckCircle2 className="h-5 w-5 text-emerald-600 shrink-0" />,
    warning: <AlertTriangle className="h-5 w-5 text-amber-600 shrink-0" />,
    error: <AlertCircle className="h-5 w-5 text-brand-error shrink-0" />,
  };

  const variants = {
    info: 'bg-brand-blue-50/70 border-brand-blue-200 text-brand-navy',
    success: 'bg-emerald-50/70 border-emerald-200 text-emerald-950',
    warning: 'bg-amber-50/70 border-amber-200 text-amber-950',
    error: 'bg-red-50/70 border-red-200 text-red-950',
  };

  return (
    <div
      role="alert"
      className={cn(
        'relative w-full rounded-xl border p-4 flex gap-3 text-sm transition-all',
        variants[normalizedVariant],
        className
      )}
      {...props}
    >
      <div className="mt-0.5">{icons[normalizedVariant]}</div>
      <div className="flex-1 space-y-1">
        {title && <h5 className="font-semibold leading-tight">{title}</h5>}
        <div className="text-xs sm:text-sm leading-relaxed opacity-90">
          {message || children}
        </div>
      </div>
      {onClose && (
        <button
          type="button"
          onClick={onClose}
          aria-label="Yopish"
          className="shrink-0 p-1 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-200/50 transition-colors"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  );
}
