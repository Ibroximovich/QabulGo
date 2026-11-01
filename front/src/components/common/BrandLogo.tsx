import { Link } from 'react-router-dom';
import { cn } from '@/lib/utils/cn';

export interface BrandLogoProps {
  className?: string;
  size?: 'sm' | 'md' | 'lg' | 'xl';
  withLink?: boolean;
  to?: string;
  showText?: boolean;
  subtitle?: string;
  isDark?: boolean;
}

export function BrandLogo({
  className,
  size = 'md',
  withLink = true,
  to = '/',
  showText = true,
  subtitle,
  isDark = false,
}: BrandLogoProps) {
  const sizeClasses = {
    sm: 'h-8 w-8',
    md: 'h-10 w-10',
    lg: 'h-12 w-12',
    xl: 'h-16 w-16',
  };

  const textSizes = {
    sm: 'text-lg',
    md: 'text-xl',
    lg: 'text-2xl',
    xl: 'text-3xl',
  };

  const content = (
    <div className={cn('inline-flex items-center gap-2.5 select-none group', className)}>
      {/* Tiniq, yuqori sifatli QabulGo Q belgisi */}
      <div
        className={cn(
          'relative flex items-center justify-center shrink-0 transition-transform duration-200 group-hover:scale-105',
          sizeClasses[size]
        )}
      >
        <img
          src="/logo-mark.png"
          alt="QabulGo"
          className="w-full h-full object-contain filter drop-shadow-[0_2px_8px_rgba(8,102,245,0.22)]"
          loading="eager"
        />
      </div>

      {showText && (
        <div className="flex flex-col">
          <span className={cn('font-black tracking-tight leading-none', textSizes[size])}>
            <span className={isDark ? 'text-white' : 'text-brand-navy'}>Qabul</span>
            <span className="text-brand-teal">Go</span>
          </span>
          <span
            className={cn(
              'text-[10px] font-bold tracking-wider uppercase mt-1',
              isDark ? 'text-blue-200/80' : 'text-brand-muted'
            )}
          >
            {subtitle || 'Onlayn qabul'}
          </span>
        </div>
      )}
    </div>
  );

  if (withLink) {
    return (
      <Link
        to={to}
        className="inline-block outline-none focus-visible:ring-2 focus-visible:ring-brand-blue rounded-lg"
      >
        {content}
      </Link>
    );
  }

  return content;
}
