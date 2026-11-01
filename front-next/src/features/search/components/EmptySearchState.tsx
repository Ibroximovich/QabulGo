'use client';

import React from 'react';
import { SearchX, RotateCcw } from 'lucide-react';
import { Button } from '@/components/ui/Button';

interface EmptySearchStateProps {
  title?: string;
  description?: string;
  onReset?: () => void;
}

export function EmptySearchState({
  title = 'Natija topilmadi',
  description = 'Kiritilgan qidiruv yoki filtr mezonlariga mos ma’lumotlar mavjud emas. Filtrlarni o‘zgartirib qayta urinib ko‘ring.',
  onReset,
}: EmptySearchStateProps) {
  return (
    <div className="flex flex-col items-center justify-center p-12 text-center bg-white rounded-3xl border border-slate-200/80 shadow-subtle my-6 max-w-lg mx-auto">
      <div className="w-16 h-16 rounded-2xl bg-slate-100 text-slate-400 flex items-center justify-center mb-4">
        <SearchX className="w-8 h-8 stroke-[1.5]" />
      </div>
      <h3 className="text-lg font-bold text-brand-navy">{title}</h3>
      <p className="text-sm text-slate-500 mt-2 max-w-sm">{description}</p>
      {onReset && (
        <div className="mt-6">
          <Button variant="outline" size="sm" onClick={onReset} className="gap-2">
            <RotateCcw className="w-4 h-4" />
            <span>Filtrlarni tozalash</span>
          </Button>
        </div>
      )}
    </div>
  );
}

