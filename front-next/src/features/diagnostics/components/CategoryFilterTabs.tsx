'use client';

import React from 'react';
import { DiagnosticCategory } from '../types';
import {
  Activity,
  Radio,
  Eye,
  Layers,
  HeartPulse,
  TestTube,
  Microscope,
  Scan,
  Bone,
  Brain,
  Zap,
  Waves,
  Timer,
  Wind,
  Ear,
  Radiation,
  Moon,
  Atom,
} from 'lucide-react';
import { cn } from '@/lib/utils/cn';

interface CategoryFilterTabsProps {
  categories: DiagnosticCategory[];
  selectedSlug: string;
  onSelect: (slug: string) => void;
  isLoading?: boolean;
}

export function CategoryFilterTabs({
  categories,
  selectedSlug,
  onSelect,
  isLoading,
}: CategoryFilterTabsProps) {
  const getCategoryIcon = (slug: string) => {
    switch (slug.toLowerCase()) {
      case 'mrt':
        return Radio;
      case 'mskt':
        return Layers;
      case 'uzi':
        return Eye;
      case 'rentgen':
        return Activity;
      case 'ekg-exo-kg':
        return HeartPulse;
      case 'laboratoriya':
        return TestTube;
      case 'endoskopiya':
        return Microscope;
      case 'mammografiya':
        return Scan;
      case 'densitometriya':
        return Bone;
      case 'fluorografiya':
        return Scan;
      case 'eeg':
        return Brain;
      case 'enmg':
        return Zap;
      case 'dopplerografiya':
        return Waves;
      case 'holter-monitoring':
        return Timer;
      case 'spirometriya':
        return Wind;
      case 'kolposkopiya':
        return Microscope;
      case 'audiometriya':
        return Ear;
      case 'oftalmologik-diagnostika':
        return Eye;
      case 'pet-kt':
        return Radiation;
      case 'polisomnografiya':
        return Moon;
      case 'scintigrafiya':
        return Atom;
      default:
        return Activity;
    }
  };

  if (isLoading) {
    return (
      <div className="flex gap-2 overflow-x-auto pb-2 scrollbar-none">
        {[...Array(6)].map((_, i) => (
          <div key={i} className="h-10 w-28 bg-slate-100 rounded-full animate-pulse flex-shrink-0" />
        ))}
      </div>
    );
  }

  return (
    <div className="flex gap-2 overflow-x-auto pb-2 scrollbar-none items-center">
      <button
        onClick={() => onSelect('')}
        className={cn(
          'flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-all whitespace-nowrap flex-shrink-0',
          selectedSlug === ''
            ? 'bg-brand-blue text-white shadow-sm shadow-blue-500/20'
            : 'bg-white text-slate-700 hover:bg-slate-50 border border-slate-200'
        )}
      >
        <Activity className="w-4 h-4" />
        Barchasi
      </button>

      {categories.map((cat) => {
        const Icon = getCategoryIcon(cat.slug);
        const isSelected = selectedSlug === cat.slug;
        return (
          <button
            key={cat.id}
            onClick={() => onSelect(cat.slug)}
            className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-all whitespace-nowrap flex-shrink-0',
              isSelected
                ? 'bg-brand-blue text-white shadow-sm shadow-blue-500/20'
                : 'bg-white text-slate-700 hover:bg-slate-50 border border-slate-200'
            )}
          >
            <Icon className="w-4 h-4" />
            {cat.name}
          </button>
        );
      })}
    </div>
  );
}

