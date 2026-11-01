import React, { useEffect } from 'react';
import { X } from 'lucide-react';
import { cn } from '@/lib/utils/cn';

export interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
}

export function Modal({
  isOpen,
  onClose,
  title,
  description,
  children,
  className,
}: ModalProps) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) {
        onClose();
      }
    };

    if (isOpen) {
      document.body.style.overflow = 'hidden';
      window.addEventListener('keydown', handleKeyDown);
    }

    return () => {
      document.body.style.overflow = 'unset';
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby={title ? 'modal-title' : undefined}
      aria-describedby={description ? 'modal-description' : undefined}
      className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 overflow-y-auto"
    >
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-brand-navy/40 backdrop-blur-sm transition-opacity animate-fade-in"
        onClick={onClose}
      />

      {/* Modal Content */}
      <div
        className={cn(
          'relative w-full max-w-lg rounded-2xl bg-white p-6 shadow-modal border border-slate-100 z-10 transition-all animate-slide-up',
          className
        )}
      >
        <button
          type="button"
          onClick={onClose}
          aria-label="Yopish"
          className="absolute right-4 top-4 p-1 rounded-xl text-slate-400 hover:text-brand-navy hover:bg-slate-100 transition-colors"
        >
          <X className="h-5 w-5" />
        </button>

        {title && (
          <h3 id="modal-title" className="text-xl font-bold text-brand-navy pr-8">
            {title}
          </h3>
        )}

        {description && (
          <p id="modal-description" className="text-sm text-brand-muted mt-1.5">
            {description}
          </p>
        )}

        <div className="mt-4">{children}</div>
      </div>
    </div>
  );
}

