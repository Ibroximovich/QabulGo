import React from 'react';
import {
  AppointmentStatus,
  APPOINTMENT_STATUS_LABELS,
  APPOINTMENT_STATUS_COLORS,
} from '../types';
import { CheckCircle2, XCircle, Check, AlertCircle } from 'lucide-react';

interface AppointmentStatusBadgeProps {
  status: AppointmentStatus | string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function AppointmentStatusBadge({
  status,
  size = 'md',
  className = '',
}: AppointmentStatusBadgeProps) {
  const validStatus = (status in APPOINTMENT_STATUS_LABELS ? status : 'confirmed') as AppointmentStatus;
  const label = APPOINTMENT_STATUS_LABELS[validStatus] || status;
  const colors = APPOINTMENT_STATUS_COLORS[validStatus] || APPOINTMENT_STATUS_COLORS.confirmed;

  const sizeClasses = {
    sm: 'px-2 py-0.5 text-[10px]',
    md: 'px-2.5 py-1 text-xs',
    lg: 'px-3 py-1.5 text-xs font-bold',
  }[size];

  const renderIcon = () => {
    switch (validStatus) {
      case 'confirmed':
        return <CheckCircle2 className={size === 'sm' ? 'w-3 h-3' : 'w-3.5 h-3.5'} />;
      case 'cancelled':
        return <XCircle className={size === 'sm' ? 'w-3 h-3' : 'w-3.5 h-3.5'} />;
      case 'completed':
        return <Check className={size === 'sm' ? 'w-3 h-3' : 'w-3.5 h-3.5'} />;
      case 'no_show':
        return <AlertCircle className={size === 'sm' ? 'w-3 h-3' : 'w-3.5 h-3.5'} />;
    }
  };

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full font-semibold border ${colors.bg} ${colors.text} ${colors.border} ${sizeClasses} ${className}`}
    >
      {renderIcon()}
      <span>{label}</span>
    </span>
  );
}
