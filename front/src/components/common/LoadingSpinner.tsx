import { cn } from '@/lib/utils/cn';

export interface LoadingSpinnerProps {
  className?: string;
  size?: 'sm' | 'md' | 'lg' | 'full';
  label?: string;
}

export function LoadingSpinner({
  className,
  size = 'md',
  label = 'Yuklanmoqda...',
}: LoadingSpinnerProps) {
  if (size === 'full') {
    return (
      <div className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-brand-bg/90 backdrop-blur-md">
        <div className="relative flex items-center justify-center">
          {/* Tashqi aylanuvchi zamonaviy tibbiy gradient halqa */}
          <div className="h-24 w-24 rounded-full border-[3px] border-slate-200/80 border-t-brand-blue border-r-brand-teal animate-spin" />

          {/* Markazdagi ultra-tiniq QabulGo belgisi */}
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="h-14 w-14 flex items-center justify-center animate-pulse">
              <img
                src="/logo-mark.png"
                alt="QabulGo"
                className="w-full h-full object-contain filter drop-shadow-[0_4px_16px_rgba(8,102,245,0.3)]"
              />
            </div>
          </div>
        </div>
        {label && (
          <div className="mt-6 text-center">
            <p className="text-sm font-bold text-brand-navy tracking-tight">{label}</p>
            <p className="text-xs text-brand-muted mt-0.5">Iltimos, kuting...</p>
          </div>
        )}
      </div>
    );
  }

  const sizes = {
    sm: 'h-4 w-4 border-2',
    md: 'h-8 w-8 border-3',
    lg: 'h-12 w-12 border-4',
  };

  return (
    <div className={cn('flex flex-col items-center justify-center gap-2', className)}>
      <div
        className={cn(
          'rounded-full border-brand-blue border-t-transparent animate-spin',
          sizes[size]
        )}
      />
      {label && <span className="text-xs text-brand-muted">{label}</span>}
    </div>
  );
}
